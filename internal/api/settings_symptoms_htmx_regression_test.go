package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/ovumcy/internal/models"
)

func TestSettingsSymptomsHTMXCreateArchiveRestoreRerendersSection(t *testing.T) {
	app, database := newOnboardingTestAppWithCSRF(t)
	user := createOnboardingTestUser(t, database, "settings-symptoms-htmx@example.com", "StrongPass1", true)
	authCookie := loginAndExtractAuthCookieWithCSRF(t, app, user.Email, "StrongPass1")
	csrfCookie, csrfToken := loadSettingsCSRFContext(t, app, authCookie)

	createForm := url.Values{
		"csrf_token": {csrfToken},
		"name":       {"Joint stiffness"},
		"icon":       {"J"},
		"color":      {"#334455"},
	}
	createRequest := httptest.NewRequest(http.MethodPost, "/api/symptoms", strings.NewReader(createForm.Encode()))
	createRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	createRequest.Header.Set("HX-Request", "true")
	createRequest.Header.Set("Cookie", joinCookieHeader(authCookie, cookiePair(csrfCookie)))

	createResponse, err := app.Test(createRequest, -1)
	if err != nil {
		t.Fatalf("create symptom htmx request failed: %v", err)
	}
	defer createResponse.Body.Close()

	if createResponse.StatusCode != http.StatusOK {
		t.Fatalf("expected htmx create status 200, got %d", createResponse.StatusCode)
	}
	createBody, err := io.ReadAll(createResponse.Body)
	if err != nil {
		t.Fatalf("read htmx create body: %v", err)
	}
	renderedCreate := string(createBody)
	if !strings.Contains(renderedCreate, `id="settings-symptoms-section"`) {
		t.Fatalf("expected symptom section rerender, got %q", renderedCreate)
	}
	if !strings.Contains(renderedCreate, `data-symptom-name="Joint stiffness"`) {
		t.Fatalf("expected new custom symptom row in htmx response, got %q", renderedCreate)
	}

	stored := models.SymptomType{}
	if err := database.Where("user_id = ? AND name = ?", user.ID, "Joint stiffness").First(&stored).Error; err != nil {
		t.Fatalf("load created custom symptom: %v", err)
	}

	archiveForm := url.Values{"csrf_token": {csrfToken}}
	archiveRequest := httptest.NewRequest(http.MethodPost, "/api/symptoms/"+strconv.FormatUint(uint64(stored.ID), 10)+"/archive", strings.NewReader(archiveForm.Encode()))
	archiveRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	archiveRequest.Header.Set("HX-Request", "true")
	archiveRequest.Header.Set("Cookie", joinCookieHeader(authCookie, cookiePair(csrfCookie)))

	archiveResponse, err := app.Test(archiveRequest, -1)
	if err != nil {
		t.Fatalf("archive symptom htmx request failed: %v", err)
	}
	defer archiveResponse.Body.Close()

	if archiveResponse.StatusCode != http.StatusOK {
		t.Fatalf("expected htmx archive status 200, got %d", archiveResponse.StatusCode)
	}
	archiveBody, err := io.ReadAll(archiveResponse.Body)
	if err != nil {
		t.Fatalf("read htmx archive body: %v", err)
	}
	renderedArchive := string(archiveBody)
	if !strings.Contains(renderedArchive, `data-symptom-state="archived"`) {
		t.Fatalf("expected archived custom symptom row, got %q", renderedArchive)
	}

	restoreForm := url.Values{"csrf_token": {csrfToken}}
	restoreRequest := httptest.NewRequest(http.MethodPost, "/api/symptoms/"+strconv.FormatUint(uint64(stored.ID), 10)+"/restore", strings.NewReader(restoreForm.Encode()))
	restoreRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	restoreRequest.Header.Set("HX-Request", "true")
	restoreRequest.Header.Set("Cookie", joinCookieHeader(authCookie, cookiePair(csrfCookie)))

	restoreResponse, err := app.Test(restoreRequest, -1)
	if err != nil {
		t.Fatalf("restore symptom htmx request failed: %v", err)
	}
	defer restoreResponse.Body.Close()

	if restoreResponse.StatusCode != http.StatusOK {
		t.Fatalf("expected htmx restore status 200, got %d", restoreResponse.StatusCode)
	}
	restoreBody, err := io.ReadAll(restoreResponse.Body)
	if err != nil {
		t.Fatalf("read htmx restore body: %v", err)
	}
	renderedRestore := string(restoreBody)
	if !strings.Contains(renderedRestore, `data-symptom-state="active"`) {
		t.Fatalf("expected active custom symptom row after restore, got %q", renderedRestore)
	}
}

func loadSettingsCSRFContext(t *testing.T, app *fiber.App, authCookie string) (*http.Cookie, string) {
	t.Helper()

	request := httptest.NewRequest(http.MethodGet, "/settings", nil)
	request.Header.Set("Accept-Language", "en")
	request.Header.Set("Cookie", authCookie)

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("settings request for csrf context failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected settings status 200, got %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read settings body for csrf context: %v", err)
	}
	csrfToken := extractCSRFTokenFromHTML(t, string(body))
	csrfCookie := responseCookie(response.Cookies(), "ovumcy_csrf")
	if csrfCookie == nil || strings.TrimSpace(csrfCookie.Value) == "" {
		t.Fatalf("expected csrf cookie in settings response")
	}

	return csrfCookie, csrfToken
}
