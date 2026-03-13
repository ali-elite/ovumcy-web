package api

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"
)

func TestSettingsPageRendersSingleIrregularCycleExplanation(t *testing.T) {
	ctx := newSettingsSecurityTestContext(t, "settings-irregular-copy@example.com")

	request := httptest.NewRequest(http.MethodGet, "/settings", nil)
	request.Header.Set("Accept-Language", "en")
	request.Header.Set("Cookie", ctx.authCookie)

	response, err := ctx.app.Test(request, -1)
	if err != nil {
		t.Fatalf("settings request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected settings status 200, got %d", response.StatusCode)
	}

	rendered := mustReadBodyString(t, response.Body)
	const hint = "Turn this on if your cycles vary by more than 7 days. A range will be shown instead of a single date."
	if strings.Count(rendered, hint) != 1 {
		t.Fatalf("expected a single irregular-cycle explanation, got %q", rendered)
	}
}

func TestSettingsClearDataFormUsesPasswordValidationBeforeConfirm(t *testing.T) {
	ctx := newSettingsSecurityTestContext(t, "settings-clear-data-ui@example.com")

	request := httptest.NewRequest(http.MethodGet, "/settings", nil)
	request.Header.Set("Accept-Language", "en")
	request.Header.Set("Cookie", ctx.authCookie)

	response, err := ctx.app.Test(request, -1)
	if err != nil {
		t.Fatalf("settings request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected settings status 200, got %d", response.StatusCode)
	}

	rendered := mustReadBodyString(t, response.Body)
	formPattern := regexp.MustCompile(`(?s)<form\s+action="/api/settings/clear-data".*?</form>`)
	formMarkup := formPattern.FindString(rendered)
	if formMarkup == "" {
		t.Fatalf("expected clear-data form in settings page")
	}
	if strings.Contains(formMarkup, `data-confirm=`) {
		t.Fatalf("expected clear-data confirmation to be deferred until password validation")
	}
	if !strings.Contains(formMarkup, `data-clear-data-verify-form`) {
		t.Fatalf("expected clear-data form to opt into password validation flow")
	}
	if !strings.Contains(formMarkup, `data-clear-data-validate-action="/api/settings/clear-data/validate"`) {
		t.Fatalf("expected clear-data form to point at password validation endpoint")
	}
	if !strings.Contains(formMarkup, `id="settings-clear-data-status"`) {
		t.Fatalf("expected clear-data form to provide a local status target")
	}
}

func TestForgotPasswordEmailStepUsesGenericEnumerationSafeSubtitle(t *testing.T) {
	app, _ := newOnboardingTestApp(t)

	form := url.Values{"email": {"unknown-owner@example.com"}}
	request := httptest.NewRequest(http.MethodPost, "/api/auth/forgot-password", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("forgot-password email step request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusSeeOther {
		t.Fatalf("expected status 303, got %d", response.StatusCode)
	}

	flashValue := responseCookieValue(response.Cookies(), flashCookieName)
	if flashValue == "" {
		t.Fatalf("expected sealed flash cookie after forgot-password email step")
	}

	followRequest := httptest.NewRequest(http.MethodGet, "/forgot-password", nil)
	followRequest.Header.Set("Accept-Language", "en")
	followRequest.Header.Set("Cookie", flashCookieName+"="+flashValue)

	followResponse, err := app.Test(followRequest, -1)
	if err != nil {
		t.Fatalf("forgot-password follow-up request failed: %v", err)
	}
	defer followResponse.Body.Close()

	if followResponse.StatusCode != http.StatusOK {
		t.Fatalf("expected forgot-password follow-up status 200, got %d", followResponse.StatusCode)
	}

	rendered := mustReadBodyString(t, followResponse.Body)
	if !strings.Contains(rendered, "If this address is registered, enter your recovery code to continue.") {
		t.Fatalf("expected generic recovery subtitle after email step, got %q", rendered)
	}
}
