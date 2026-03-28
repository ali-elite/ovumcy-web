package api

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/ovumcy/ovumcy-web/internal/security"
)

func TestPopOIDCStateCookieRejectsExpiredPayload(t *testing.T) {
	t.Parallel()

	handler := &Handler{
		secretKey:    []byte("0123456789abcdef0123456789abcdef"),
		cookieSecure: true,
	}
	codec, err := newSecureCookieCodec(handler.secretKey)
	if err != nil {
		t.Fatalf("newSecureCookieCodec() error: %v", err)
	}

	payload, err := json.Marshal(oidcAuthState{
		State:        "state-value",
		Nonce:        "nonce-value",
		CodeVerifier: "verifier-value",
		ExpiresAt:    time.Now().UTC().Add(-time.Minute).Format(time.RFC3339Nano),
	})
	if err != nil {
		t.Fatalf("marshal state payload: %v", err)
	}
	sealed, err := codec.seal(oidcStateCookieName, payload)
	if err != nil {
		t.Fatalf("seal state payload: %v", err)
	}

	app := fiber.New()
	app.Get(security.OIDCCallbackPath, func(c *fiber.Ctx) error {
		state := handler.popOIDCStateCookie(c)
		if state.State != "" || state.Nonce != "" || state.CodeVerifier != "" {
			t.Fatalf("expected expired OIDC state cookie to be rejected, got %+v", state)
		}
		return c.SendStatus(fiber.StatusNoContent)
	})

	request := httptest.NewRequest("GET", security.OIDCCallbackPath, nil)
	request.Header.Set("Cookie", oidcStateCookieName+"="+sealed)
	response, testErr := app.Test(request, -1)
	if testErr != nil {
		t.Fatalf("request failed: %v", testErr)
	}
	defer response.Body.Close()

	if response.StatusCode != fiber.StatusNoContent {
		t.Fatalf("expected status 204, got %d", response.StatusCode)
	}
}
