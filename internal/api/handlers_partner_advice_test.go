package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ovumcy/ovumcy-web/internal/models"
)

func TestHandlePartnerAdvice_OwnerForbidden(t *testing.T) {
	app, database := newOnboardingTestApp(t)

	// User who is an owner
	user := createOnboardingTestUser(t, database, "partner-advice-owner@example.com", "StrongPass1", true)
	authCookie := issueAuthCookieForUser(t, user)

	req := httptest.NewRequest(http.MethodGet, "/api/partner/advice?phase=period", nil)
	req.Header.Set("Cookie", authCookie)

	resp := mustAppResponse(t, app, req)
	assertStatusCode(t, resp, http.StatusUnauthorized)
}

func TestHandlePartnerAdvice_PartnerAllowed(t *testing.T) {
	app, database := newOnboardingTestApp(t)

	// User who is a partner
	user := createOnboardingTestUser(t, database, "partner-advice-partner@example.com", "StrongPass1", true)
	if err := database.Model(&models.User{}).Where("id = ?", user.ID).Update("role", models.RolePartner).Error; err != nil {
		t.Fatalf("set partner role: %v", err)
	}
	user.Role = models.RolePartner
	authCookie := issueAuthCookieForUser(t, user)

	req := httptest.NewRequest(http.MethodGet, "/api/partner/advice?phase=period", nil)
	req.Header.Set("Cookie", authCookie)

	resp := mustAppResponse(t, app, req)
	assertStatusCode(t, resp, http.StatusOK)
}
