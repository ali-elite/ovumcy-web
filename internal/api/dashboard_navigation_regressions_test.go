package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestDashboardLogoutFormsRequireConfirmation(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "logout-confirm@example.com", "StrongPass1", true)
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	request := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	request.Header.Set("Accept-Language", "en")
	request.Header.Set("Cookie", authCookie)

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("dashboard request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read dashboard body: %v", err)
	}
	rendered := string(body)
	if strings.Count(rendered, `action="/api/auth/logout"`) < 2 {
		t.Fatalf("expected desktop and mobile logout forms")
	}
	if strings.Count(rendered, `action="/api/auth/logout" method="post"`) < 2 {
		t.Fatalf("expected logout forms to use POST method")
	}
	if strings.Count(rendered, `name="csrf_token" value="`) < 2 {
		t.Fatalf("expected csrf token hidden fields on both logout forms")
	}
}

func TestDashboardNavigationOmitsCurrentUserIdentity(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "identity-owner@example.com", "StrongPass1", true)
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	request := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	request.Header.Set("Accept-Language", "en")
	request.Header.Set("Cookie", authCookie)

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("dashboard request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read dashboard body: %v", err)
	}
	rendered := string(body)
	if strings.Contains(rendered, "identity-owner") {
		t.Fatalf("did not expect local-part identity in navigation")
	}
	if strings.Contains(rendered, "identity-owner@example.com") {
		t.Fatalf("did not expect email identity in navigation")
	}
}

func TestDashboardHeaderOmitsLanguageSwitch(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "lang-switch-labels@example.com", "StrongPass1", true)
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	rendered := mustRenderDashboard(t, app, authCookie, "ru")
	for _, label := range []string{"RU", "EN", "ES"} {
		if strings.Contains(rendered, ">"+label+"</a>") {
			t.Fatalf("did not expect %s language shortcut in dashboard header", label)
		}
	}
}

func mustRenderDashboard(t *testing.T, app *fiber.App, authCookie string, languageCookie string) string {
	t.Helper()

	request := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	request.Header.Set("Accept-Language", "en")
	if strings.TrimSpace(languageCookie) == "" {
		request.Header.Set("Cookie", authCookie)
	} else {
		request.Header.Set("Cookie", authCookie+"; ovumcy_lang="+languageCookie)
	}

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("dashboard request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read dashboard body: %v", err)
	}
	return string(body)
}
