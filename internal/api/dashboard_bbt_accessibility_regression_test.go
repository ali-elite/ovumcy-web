package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ovumcy/ovumcy-web/internal/models"
	"github.com/ovumcy/ovumcy-web/internal/services"
)

func TestDashboardAndCalendarExposeAccessibleBBTInputs(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "bbt-accessibility@example.com", "StrongPass1", true)
	if err := database.Model(&models.User{}).Where("id = ?", user.ID).Updates(map[string]any{
		"track_bbt":        true,
		"temperature_unit": services.TemperatureUnitCelsius,
	}).Error; err != nil {
		t.Fatalf("enable BBT tracking: %v", err)
	}

	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	dashboardRequest := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	dashboardRequest.Header.Set("Accept-Language", "en")
	dashboardRequest.Header.Set("Cookie", authCookie)
	dashboardResponse := mustAppResponse(t, app, dashboardRequest)
	assertStatusCode(t, dashboardResponse, http.StatusOK)

	dashboardBody := mustReadBodyString(t, dashboardResponse.Body)
	for _, fragment := range []string{
		`id="dashboard-bbt"`,
		`aria-labelledby="dashboard-bbt-legend"`,
		`aria-describedby="dashboard-bbt-hint"`,
	} {
		if !strings.Contains(dashboardBody, fragment) {
			t.Fatalf("expected dashboard BBT field markup %q", fragment)
		}
	}

	dayActionPrefix := `hx-post="/api/days/`
	startIndex := strings.Index(dashboardBody, dayActionPrefix)
	if startIndex < 0 {
		t.Fatal("expected dashboard day form action")
	}
	dayStart := startIndex + len(dayActionPrefix)
	dayEnd := dayStart + len("2006-01-02")
	if len(dashboardBody) < dayEnd {
		t.Fatal("expected dashboard day form date to be present")
	}
	dayRaw := dashboardBody[dayStart:dayEnd]

	panelRequest := httptest.NewRequest(http.MethodGet, "/calendar/day/"+dayRaw+"?mode=edit", nil)
	panelRequest.Header.Set("Accept-Language", "en")
	panelRequest.Header.Set("Cookie", authCookie)
	panelResponse := mustAppResponse(t, app, panelRequest)
	assertStatusCode(t, panelResponse, http.StatusOK)

	panelBody := mustReadBodyString(t, panelResponse.Body)
	for _, fragment := range []string{
		`id="calendar-bbt"`,
		`aria-labelledby="calendar-bbt-legend"`,
		`aria-describedby="calendar-bbt-hint"`,
	} {
		if !strings.Contains(panelBody, fragment) {
			t.Fatalf("expected calendar BBT field markup %q", fragment)
		}
	}
}
