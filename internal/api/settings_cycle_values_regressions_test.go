package api

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/ovumcy/internal/models"
)

func TestSettingsPageRendersPersistedCycleValues(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "settings-values@example.com", "StrongPass1", true)
	if err := database.Model(&models.User{}).Where("id = ?", user.ID).Updates(map[string]any{
		"cycle_length":     29,
		"period_length":    6,
		"auto_period_fill": true,
	}).Error; err != nil {
		t.Fatalf("update cycle values: %v", err)
	}
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	rendered := renderSettingsPageForTest(t, app, authCookie)
	assertBodyContainsAll(t, rendered,
		bodyStringMatch{fragment: `id="settings-period-length"`, message: "expected settings period slider max=14"},
		bodyStringMatch{fragment: `max="14"`, message: "expected settings period slider max=14"},
		bodyStringMatch{fragment: `id="settings-last-period-start"`, message: "expected settings cycle form to include editable last-period-start field"},
		bodyStringMatch{fragment: `name="auto_period_fill" value="true"`, message: "expected settings cycle form to include auto-period-fill toggle"},
		bodyStringMatch{fragment: `id="export-from"`, message: "expected export date range inputs to be rendered"},
		bodyStringMatch{fragment: `id="export-to"`, message: "expected export date range inputs to be rendered"},
	)

	cycleInputPattern := regexp.MustCompile(`(?s)name="cycle_length".*?value="29"`)
	if !cycleInputPattern.MatchString(rendered) {
		t.Fatalf("expected cycle slider value attribute to be rendered from DB")
	}
	periodInputPattern := regexp.MustCompile(`(?s)name="period_length".*?value="6"`)
	if !periodInputPattern.MatchString(rendered) {
		t.Fatalf("expected period slider value attribute to be rendered from DB")
	}
	autoPeriodFillPattern := regexp.MustCompile(`(?s)name="auto_period_fill".*?checked`)
	if !autoPeriodFillPattern.MatchString(rendered) {
		t.Fatalf("expected auto_period_fill checkbox to reflect persisted enabled state")
	}
	exportInputPattern := regexp.MustCompile(`(?s)id="export-from".*?type="date".*?id="export-to".*?type="date"`)
	if !exportInputPattern.MatchString(rendered) {
		t.Fatalf("expected export inputs to use native date type for client-side validation")
	}
	lastPeriodInputAccessibilityPattern := regexp.MustCompile(`(?s)id="settings-last-period-start".*?lang="en".*?aria-label="Last period start date".*?min="\d{4}-01-01"`)
	if !lastPeriodInputAccessibilityPattern.MatchString(rendered) {
		t.Fatalf("expected settings last-period-start field to include localized language, aria-label, and min date attributes")
	}
}

func renderSettingsPageForTest(t *testing.T, app *fiber.App, authCookie string) string {
	t.Helper()

	request := httptest.NewRequest(http.MethodGet, "/settings", nil)
	request.Header.Set("Accept-Language", "en")
	request.Header.Set("Cookie", authCookie)

	response := mustAppResponse(t, app, request)
	assertStatusCode(t, response, http.StatusOK)
	return mustReadBodyString(t, response.Body)
}
