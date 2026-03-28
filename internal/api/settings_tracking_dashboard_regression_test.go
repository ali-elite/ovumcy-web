package api

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"testing"

	"github.com/ovumcy/ovumcy-web/internal/models"
)

func TestTrackingSettingsExposeBBTAndCervicalMucusOnDashboard(t *testing.T) {
	ctx := newSettingsSecurityTestContext(t, "settings-tracking-dashboard@example.com")

	response := settingsFormRequestWithCSRF(t, ctx, http.MethodPost, "/api/settings/tracking", url.Values{
		"track_bbt":            {"true"},
		"track_cervical_mucus": {"true"},
		"temperature_unit":     {"c"},
	}, map[string]string{
		"HX-Request": "true",
	})
	assertStatusCode(t, response, http.StatusOK)

	dashboardRequest := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	dashboardRequest.Header.Set("Accept-Language", "en")
	dashboardRequest.Header.Set("Cookie", ctx.authCookie)

	dashboardResponse := mustAppResponse(t, ctx.app, dashboardRequest)
	assertStatusCode(t, dashboardResponse, http.StatusOK)
	rendered := mustReadBodyString(t, dashboardResponse.Body)

	assertBodyContainsAll(t, rendered,
		bodyStringMatch{fragment: `id="dashboard-bbt"`, message: "expected dashboard BBT field after enabling tracking"},
		bodyStringMatch{fragment: `name="cervical_mucus"`, message: "expected dashboard cervical mucus controls after enabling tracking"},
	)
}

func TestSettingsPageKeepsPersistedCycleValuesAfterRecoveryCodeRegeneration(t *testing.T) {
	ctx := newSettingsSecurityTestContext(t, "settings-recovery-return@example.com")

	response := settingsFormRequestWithCSRF(t, ctx, http.MethodPost, "/api/settings/regenerate-recovery-code", url.Values{}, nil)
	assertStatusCode(t, response, http.StatusSeeOther)

	recoveryCookie := responseCookieValue(response.Cookies(), recoveryCodeCookieName)
	if recoveryCookie == "" {
		t.Fatal("expected recovery-code page cookie after regeneration")
	}

	recoveryPageRequest := httptest.NewRequest(http.MethodGet, "/recovery-code", nil)
	recoveryPageRequest.Header.Set("Accept-Language", "en")
	recoveryPageRequest.Header.Set("Cookie", ctx.authCookie+"; "+recoveryCodeCookieName+"="+recoveryCookie)

	recoveryPageResponse := mustAppResponse(t, ctx.app, recoveryPageRequest)
	assertStatusCode(t, recoveryPageResponse, http.StatusOK)
	recoveryPage := mustReadBodyString(t, recoveryPageResponse.Body)
	assertBodyContainsAll(t, recoveryPage,
		bodyStringMatch{fragment: `form action="/settings"`, message: "expected recovery confirmation to return to settings"},
	)
	assertBodyNotContainsAll(t, recoveryPage,
		bodyStringMatch{fragment: `name="saved"`, message: "did not expect recovery confirmation checkbox to submit a saved query parameter"},
	)

	var persisted struct {
		PeriodLength       int
		UnpredictableCycle bool
	}
	if err := ctx.database.Model(&models.User{}).
		Select("period_length", "unpredictable_cycle").
		Where("id = ?", ctx.user.ID).
		First(&persisted).Error; err != nil {
		t.Fatalf("load persisted settings after recovery-code regeneration: %v", err)
	}
	if persisted.PeriodLength != 5 {
		t.Fatalf("expected persisted settings period length to stay at 5 after recovery-code regeneration, got %d", persisted.PeriodLength)
	}
	if persisted.UnpredictableCycle {
		t.Fatalf("did not expect persisted unpredictable_cycle to change after recovery-code regeneration")
	}

	rendered := renderSettingsPageForTest(t, ctx.app, ctx.authCookie)
	if !regexp.MustCompile(`name="period_length"[^>]*value="5"`).MatchString(rendered) {
		t.Fatalf("expected persisted settings period length to stay at 5 days after recovery-code regeneration")
	}
	if regexp.MustCompile(`name="unpredictable_cycle"[^>]*checked`).MatchString(rendered) {
		t.Fatalf("did not expect unpredictable_cycle to become checked after recovery-code regeneration")
	}
}
