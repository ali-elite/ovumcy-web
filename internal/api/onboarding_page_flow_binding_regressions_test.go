package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOnboardingPageRendersStableOnboardingFlowControls(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "onboarding-flow-binding@example.com", "StrongPass1", false)
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	request := httptest.NewRequest(http.MethodGet, "/onboarding", nil)
	request.Header.Set("Cookie", authCookie)

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("onboarding request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected onboarding status 200, got %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read onboarding body: %v", err)
	}
	rendered := string(body)

	expectedFragments := []string{
		`hx-post="/onboarding/step1"`,
		`id="last-period-start"`,
		`hx-post="/onboarding/step2"`,
		`id="cycle-length"`,
		`id="period-length"`,
		`hx-post="/onboarding/complete"`,
	}
	for _, fragment := range expectedFragments {
		if !strings.Contains(rendered, fragment) {
			t.Fatalf("expected onboarding page to include %q", fragment)
		}
	}
}
