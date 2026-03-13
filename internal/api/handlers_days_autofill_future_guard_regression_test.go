package api

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/terraincognita07/ovumcy/internal/models"
	"github.com/terraincognita07/ovumcy/internal/services"
)

func TestUpsertDayAutoFillDoesNotCreateFutureDays(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "upsert-day-autofill-future-guard@example.com", "StrongPass1", true)
	if err := database.Model(&models.User{}).Where("id = ?", user.ID).Updates(map[string]any{
		"period_length":    4,
		"auto_period_fill": true,
	}).Error; err != nil {
		t.Fatalf("seed autofill settings: %v", err)
	}

	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")
	today := services.DateAtLocation(time.Now().In(time.UTC), time.UTC)
	todayRaw := today.Format("2006-01-02")

	form := url.Values{
		"is_period": {"true"},
		"flow":      {models.FlowLight},
	}
	request := httptest.NewRequest(http.MethodPost, "/api/days/"+todayRaw, strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("HX-Request", "true")
	request.Header.Set("Accept-Language", "en")
	request.Header.Set("Cookie", authCookie)

	response := mustAppResponse(t, app, request)
	assertStatusCode(t, response, http.StatusOK)

	tomorrow := today.AddDate(0, 0, 1)
	entry, err := fetchLogByDateForTest(database, user.ID, tomorrow, time.UTC)
	if err != nil {
		t.Fatalf("load tomorrow entry after autofill attempt: %v", err)
	}
	if entry.ID != 0 {
		t.Fatalf("did not expect autofill to create future entry for %s", tomorrow.Format("2006-01-02"))
	}
}
