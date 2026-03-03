package api

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/terraincognita07/ovumcy/internal/services"
)

func TestCalendarPageUsesRequestTimezoneForTodaySelectionAndBadge(t *testing.T) {
	app, database, _ := newOnboardingTestAppWithLocation(t, time.UTC)
	user := createOnboardingTestUser(t, database, "calendar-tz-request@example.com", "StrongPass1", true)
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	nowUTC := time.Now().UTC()
	timezoneName, location := timezoneWithDifferentCalendarDay(t, nowUTC)
	expectedToday := services.DateAtLocation(nowUTC.In(location), location).Format("2006-01-02")

	request := httptest.NewRequest(http.MethodGet, "/calendar", nil)
	request.Header.Set("Accept-Language", "en")
	request.Header.Set("Cookie", authCookie)
	request.Header.Set(timezoneHeaderName, timezoneName)

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("calendar request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}

	renderedBytes, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read calendar body: %v", err)
	}
	rendered := string(renderedBytes)

	expectedSelected := fmt.Sprintf(`selectedDate: "%s"`, expectedToday)
	if !strings.Contains(rendered, expectedSelected) {
		t.Fatalf("expected selectedDate %q in calendar page", expectedSelected)
	}

	todayBadgePattern := regexp.MustCompile(fmt.Sprintf(`(?s)data-day="%s".*?calendar-today-pill`, regexp.QuoteMeta(expectedToday)))
	if !todayBadgePattern.MatchString(rendered) {
		t.Fatalf("expected today badge on day %s for timezone %s", expectedToday, timezoneName)
	}
}
