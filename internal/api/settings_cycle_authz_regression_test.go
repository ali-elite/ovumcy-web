package api

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/terraincognita07/ovumcy/internal/models"
)

func TestSettingsCycleUpdateRedirectsUnauthenticatedUsersToLogin(t *testing.T) {
	app, _ := newOnboardingTestApp(t)

	request := httptest.NewRequest(http.MethodPost, "/settings/cycle", strings.NewReader(url.Values{
		"cycle_length":  {"28"},
		"period_length": {"5"},
	}.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response := mustAppResponse(t, app, request)
	assertStatusCode(t, response, http.StatusSeeOther)
	if location := response.Header.Get("Location"); location != "/login" {
		t.Fatalf("expected redirect to /login, got %q", location)
	}
}

func TestSettingsCycleUpdateRejectsPartnerRoleJSON(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "settings-cycle-partner@example.com", "StrongPass1", true)
	if err := database.Model(&models.User{}).Where("id = ?", user.ID).Update("role", models.RolePartner).Error; err != nil {
		t.Fatalf("set partner role: %v", err)
	}
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	request := httptest.NewRequest(http.MethodPost, "/settings/cycle", strings.NewReader(url.Values{
		"cycle_length":  {"28"},
		"period_length": {"5"},
	}.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Cookie", authCookie)

	response := mustAppResponse(t, app, request)
	assertStatusCode(t, response, http.StatusForbidden)
	if got := readAPIError(t, response.Body); got != "owner access required" {
		t.Fatalf("expected owner access required error, got %q", got)
	}
}

func TestSettingsCycleUpdateMissingCSRFRejectedByMiddleware(t *testing.T) {
	app, database := newOnboardingTestAppWithCSRF(t)
	user := createOnboardingTestUser(t, database, "settings-cycle-csrf@example.com", "StrongPass1", true)
	authCookie := loginAndExtractAuthCookieWithCSRF(t, app, user.Email, "StrongPass1")

	request := httptest.NewRequest(http.MethodPost, "/settings/cycle", strings.NewReader(url.Values{
		"cycle_length":  {"28"},
		"period_length": {"5"},
	}.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("HX-Request", "true")
	request.Header.Set("Cookie", authCookie)

	response := mustAppResponse(t, app, request)
	assertStatusCode(t, response, http.StatusForbidden)
}
