package api

import (
	"net/url"
	"testing"

	"github.com/ovumcy/ovumcy-web/internal/models"
)

func TestSettingsCycleUpdatePersistsWithHTMXAndCSRF(t *testing.T) {
	app, database := newOnboardingTestAppWithCSRF(t)
	user := createOnboardingTestUser(t, database, "settings-cycle-persist@example.com", "StrongPass1", true)
	if err := database.Model(&models.User{}).Where("id = ?", user.ID).Updates(map[string]any{
		"cycle_length":     15,
		"period_length":    5,
		"auto_period_fill": false,
		"irregular_cycle":  false,
	}).Error; err != nil {
		t.Fatalf("set initial cycle values: %v", err)
	}

	authCookie := loginAndExtractAuthCookieWithCSRF(t, app, user.Email, "StrongPass1")
	csrfCookie, csrfToken := loadSettingsCSRFContext(t, app, authCookie)

	form := url.Values{
		"cycle_length":      {"28"},
		"period_length":     {"6"},
		"auto_period_fill":  {"true"},
		"irregular_cycle":   {"true"},
		"last_period_start": {"2026-02-10"},
	}
	updateBody := submitSettingsCycleUpdate(t, app, authCookie, csrfCookie, csrfToken, form)
	assertSettingsCycleHTMXSuccess(t, updateBody)

	persisted := models.User{}
	if err := database.Select("cycle_length", "period_length", "auto_period_fill", "irregular_cycle", "last_period_start").First(&persisted, user.ID).Error; err != nil {
		t.Fatalf("load persisted user cycle values: %v", err)
	}
	if persisted.CycleLength != 28 {
		t.Fatalf("expected persisted cycle_length=28, got %d", persisted.CycleLength)
	}
	if persisted.PeriodLength != 6 {
		t.Fatalf("expected persisted period_length=6, got %d", persisted.PeriodLength)
	}
	if !persisted.AutoPeriodFill {
		t.Fatalf("expected persisted auto_period_fill=true")
	}
	if !persisted.IrregularCycle {
		t.Fatalf("expected persisted irregular_cycle=true")
	}
	if persisted.LastPeriodStart == nil || persisted.LastPeriodStart.Format("2006-01-02") != "2026-02-10" {
		t.Fatalf("expected persisted last_period_start=2026-02-10, got %v", persisted.LastPeriodStart)
	}
}
