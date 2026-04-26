package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/ovumcy/ovumcy-web/internal/models"
	"github.com/ovumcy/ovumcy-web/internal/services"
)

func TestCreatePartnerInvitationRequiresAuthJSON(t *testing.T) {
	app, _ := newOnboardingTestApp(t)

	request := httptest.NewRequest(http.MethodPost, "/api/partner/invitations", strings.NewReader(url.Values{}.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Accept", "application/json")

	response := mustAppResponse(t, app, request)
	assertStatusCode(t, response, http.StatusUnauthorized)
	if got := readAPIError(t, response.Body); got != "unauthorized" {
		t.Fatalf("expected unauthorized error, got %q", got)
	}
}

func TestCreatePartnerInvitationRejectsPartnerRoleJSON(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "partner-invite-partner@example.com", "StrongPass1", true)
	if err := database.Model(&models.User{}).Where("id = ?", user.ID).Update("role", models.RolePartner).Error; err != nil {
		t.Fatalf("set partner role: %v", err)
	}
	user.Role = models.RolePartner
	authCookie := issueAuthCookieForUser(t, user)

	request := httptest.NewRequest(http.MethodPost, "/api/partner/invitations", strings.NewReader(url.Values{}.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Cookie", authCookie)

	response := mustAppResponse(t, app, request)
	assertStatusCode(t, response, http.StatusForbidden)
	if got := readAPIError(t, response.Body); got != "owner access required" {
		t.Fatalf("expected owner access required error, got %q", got)
	}
}

func TestCreatePartnerInvitationMissingCSRFRejectedByMiddleware(t *testing.T) {
	app, database := newOnboardingTestAppWithCSRF(t)
	user := createOnboardingTestUser(t, database, "partner-invite-csrf@example.com", "StrongPass1", true)
	authCookie := loginAndExtractAuthCookieWithCSRF(t, app, user.Email, "StrongPass1")

	request := httptest.NewRequest(http.MethodPost, "/api/partner/invitations", strings.NewReader(url.Values{}.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Cookie", authCookie)

	response := mustAppResponse(t, app, request)
	assertStatusCode(t, response, http.StatusForbidden)
}

func TestCreatePartnerInvitationPersistsHashedCodeAndReturnsPlainCode(t *testing.T) {
	ctx := newSettingsSecurityTestContext(t, "partner-invite-owner@example.com")

	response := settingsFormRequestWithCSRF(t, ctx, http.MethodPost, "/api/partner/invitations", url.Values{}, map[string]string{
		"Accept": "application/json",
	})
	defer func() { _ = response.Body.Close() }()

	if response.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", response.StatusCode)
	}

	payload := struct {
		OK        bool   `json:"ok"`
		Code      string `json:"code"`
		Status    string `json:"status"`
		ExpiresAt string `json:"expires_at"`
	}{}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode invitation response: %v", err)
	}
	if !payload.OK {
		t.Fatal("expected ok=true in invitation response")
	}
	if err := services.ValidatePartnerInviteCodeFormat(payload.Code); err != nil {
		t.Fatalf("expected returned code in canonical format, got %q (%v)", payload.Code, err)
	}
	if payload.Status != models.PartnerInvitationStatusPending {
		t.Fatalf("expected pending status, got %q", payload.Status)
	}
	expiresAt, err := time.Parse(time.RFC3339, payload.ExpiresAt)
	if err != nil {
		t.Fatalf("parse expires_at: %v", err)
	}
	if expiresAt.Before(time.Now().UTC().Add(71 * time.Hour)) {
		t.Fatalf("expected expires_at at least ~72h out, got %v", expiresAt)
	}

	stored := models.PartnerInvitation{}
	if err := ctx.database.Where("owner_user_id = ?", ctx.user.ID).Order("id DESC").First(&stored).Error; err != nil {
		t.Fatalf("load persisted invitation: %v", err)
	}
	if stored.CodeHash == "" || stored.CodeHash == payload.Code {
		t.Fatalf("expected persisted code hash to be non-empty and distinct from plain code, got %q", stored.CodeHash)
	}
	if !services.IsPartnerInviteCodeMatch([]byte("test-secret-key"), payload.Code, stored.CodeHash) {
		t.Fatal("expected stored hash to match returned invitation code")
	}
	if stored.CodeHint != services.PartnerInviteCodeHint(payload.Code) {
		t.Fatalf("expected code hint %q, got %q", services.PartnerInviteCodeHint(payload.Code), stored.CodeHint)
	}
}

func TestCreatePartnerInvitationHTMXRendersRegeneratedCode(t *testing.T) {
	ctx := newSettingsSecurityTestContext(t, "partner-invite-htmx@example.com")

	response := settingsFormRequestWithCSRF(t, ctx, http.MethodPost, "/api/partner/invitations", url.Values{}, map[string]string{
		"HX-Request": "true",
	})
	defer func() { _ = response.Body.Close() }()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read htmx invitation response: %v", err)
	}
	rendered := string(body)
	if !strings.Contains(rendered, "Active invitation code") {
		t.Fatalf("expected active invitation markup, got %q", rendered)
	}
	if !strings.Contains(rendered, `hx-post="/api/partner/invitations"`) {
		t.Fatalf("expected regenerated invite form in htmx response, got %q", rendered)
	}
	if !strings.Contains(rendered, "Regenerate code") {
		t.Fatalf("expected regenerate button in htmx response, got %q", rendered)
	}
}
