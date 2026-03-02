package httpx

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

type negotiationSnapshot struct {
	HTMX                 bool `json:"htmx"`
	JSONAcceptOnly       bool `json:"json_accept_only"`
	JSONAcceptOrBodyType bool `json:"json_accept_or_body_type"`
}

func readNegotiationSnapshot(t *testing.T, headers map[string]string) negotiationSnapshot {
	t.Helper()

	app := fiber.New()
	app.Post("/", func(c *fiber.Ctx) error {
		return c.JSON(negotiationSnapshot{
			HTMX:                 IsHTMX(c),
			JSONAcceptOnly:       AcceptsJSON(c, JSONModeAcceptOnly),
			JSONAcceptOrBodyType: AcceptsJSON(c, JSONModeAcceptOrContentType),
		})
	})

	request := httptest.NewRequest(http.MethodPost, "/", nil)
	for key, value := range headers {
		request.Header.Set(key, value)
	}

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("app test request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}

	var payload negotiationSnapshot
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response payload: %v", err)
	}
	return payload
}

func TestIsHTMXAndAcceptsJSONViaAcceptHeader(t *testing.T) {
	snapshot := readNegotiationSnapshot(t, map[string]string{
		"HX-Request": "TrUe",
		"Accept":     "text/html, application/json",
	})

	if !snapshot.HTMX {
		t.Fatal("expected HTMX=true")
	}
	if !snapshot.JSONAcceptOnly {
		t.Fatal("expected JSONAcceptOnly=true")
	}
	if !snapshot.JSONAcceptOrBodyType {
		t.Fatal("expected JSONAcceptOrBodyType=true")
	}
}

func TestAcceptsJSONViaContentTypeOnlyWhenModeAllows(t *testing.T) {
	snapshot := readNegotiationSnapshot(t, map[string]string{
		"Content-Type": "application/json; charset=utf-8",
	})

	if snapshot.JSONAcceptOnly {
		t.Fatal("expected JSONAcceptOnly=false when Accept header has no json")
	}
	if !snapshot.JSONAcceptOrBodyType {
		t.Fatal("expected JSONAcceptOrBodyType=true for JSON Content-Type")
	}
}

func TestAcceptsJSONFalseWhenHeadersDoNotContainJSON(t *testing.T) {
	snapshot := readNegotiationSnapshot(t, map[string]string{
		"Accept":       "text/html",
		"Content-Type": "application/x-www-form-urlencoded",
	})

	if snapshot.HTMX {
		t.Fatal("expected HTMX=false")
	}
	if snapshot.JSONAcceptOnly {
		t.Fatal("expected JSONAcceptOnly=false")
	}
	if snapshot.JSONAcceptOrBodyType {
		t.Fatal("expected JSONAcceptOrBodyType=false")
	}
}
