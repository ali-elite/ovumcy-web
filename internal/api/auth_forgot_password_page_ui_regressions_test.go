package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestForgotPasswordPageStartsWithEmailStep(t *testing.T) {
	app, _ := newOnboardingTestApp(t)

	request := httptest.NewRequest(http.MethodGet, "/forgot-password", nil)
	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("forgot-password page request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read forgot-password body: %v", err)
	}
	rendered := string(body)
	if !strings.Contains(rendered, `name="email"`) {
		t.Fatalf("expected email input on initial forgot-password page")
	}
	if strings.Contains(rendered, `name="recovery_code"`) {
		t.Fatalf("did not expect recovery code input before email step")
	}
}

func TestForgotPasswordEmailStepTransitionsToRecoveryCodeStep(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "forgot-page-step@example.com", "StrongPass1", true)

	form := url.Values{"email": {user.Email}}
	request := httptest.NewRequest(http.MethodPost, "/api/auth/forgot-password", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("forgot-password email-step request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusSeeOther {
		t.Fatalf("expected status 303, got %d", response.StatusCode)
	}
	if location := response.Header.Get("Location"); location != "/forgot-password" {
		t.Fatalf("expected redirect to /forgot-password, got %q", location)
	}

	flashCookie := responseCookieValue(response.Cookies(), flashCookieName)
	if strings.TrimSpace(flashCookie) == "" {
		t.Fatalf("expected flash cookie for forgot-password email step")
	}

	pageRequest := httptest.NewRequest(http.MethodGet, "/forgot-password", nil)
	pageRequest.Header.Set("Cookie", flashCookieName+"="+flashCookie)
	pageResponse, err := app.Test(pageRequest, -1)
	if err != nil {
		t.Fatalf("forgot-password step2 page request failed: %v", err)
	}
	defer pageResponse.Body.Close()

	if pageResponse.StatusCode != http.StatusOK {
		t.Fatalf("expected step2 page status 200, got %d", pageResponse.StatusCode)
	}

	body, err := io.ReadAll(pageResponse.Body)
	if err != nil {
		t.Fatalf("read forgot-password step2 body: %v", err)
	}
	rendered := string(body)
	if !strings.Contains(rendered, `name="recovery_code"`) {
		t.Fatalf("expected recovery code input on step2 page")
	}
	if !strings.Contains(rendered, `type="hidden" name="email"`) {
		t.Fatalf("expected hidden email field on step2 page")
	}
}
