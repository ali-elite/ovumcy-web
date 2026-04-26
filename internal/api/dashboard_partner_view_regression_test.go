package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ovumcy/ovumcy-web/internal/models"
)

func TestPartnerDashboardUsesLinkedOwnerCycleDataWithoutPrivateFields(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	owner := createOnboardingTestUser(t, database, "partner-dashboard-owner@example.com", "StrongPass1", true)
	partner := createOnboardingTestUser(t, database, "partner-dashboard-viewer@example.com", "StrongPass1", true)
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	oldLastPeriodStart := today.AddDate(0, 0, -90)
	if err := database.Model(&models.User{}).Where("id = ?", owner.ID).Updates(map[string]any{
		"display_name":      "Owner Name",
		"last_period_start": oldLastPeriodStart,
	}).Error; err != nil {
		t.Fatalf("set owner display name: %v", err)
	}
	if err := database.Model(&models.User{}).Where("id = ?", partner.ID).Updates(map[string]any{
		"display_name": "Partner Name",
		"role":         models.RolePartner,
	}).Error; err != nil {
		t.Fatalf("set partner role: %v", err)
	}
	partner.DisplayName = "Partner Name"
	partner.Role = models.RolePartner

	if err := database.Create(&models.PartnerLink{
		OwnerUserID:     owner.ID,
		PartnerUserID:   partner.ID,
		InvitedByUserID: owner.ID,
		Status:          models.PartnerLinkStatusActive,
		CreatedAt:       now.UTC(),
	}).Error; err != nil {
		t.Fatalf("create partner link: %v", err)
	}
	symptom := models.SymptomType{UserID: owner.ID, Name: "Partner-visible tenderness", Icon: "T", Color: "#FF7755"}
	if err := database.Create(&symptom).Error; err != nil {
		t.Fatalf("create owner symptom: %v", err)
	}
	if err := database.Create(&models.DailyLog{
		UserID:     owner.ID,
		Date:       today,
		IsPeriod:   true,
		Flow:       models.FlowMedium,
		Mood:       5,
		SymptomIDs: []uint{symptom.ID},
		Notes:      "private owner note",
	}).Error; err != nil {
		t.Fatalf("create owner daily log: %v", err)
	}

	rendered := mustRenderDashboard(t, app, issueAuthCookieForUser(t, partner), "en")

	if !strings.Contains(rendered, `data-partner-overview`) {
		t.Fatalf("expected partner overview section, got %q", rendered)
	}
	if !strings.Contains(rendered, "Owner Name&#39;s cycle overview") {
		t.Fatalf("expected personalized partner overview heading with owner name, got %q", rendered)
	}
	if !strings.Contains(rendered, "Period day") || !strings.Contains(rendered, "Medium") {
		t.Fatalf("expected linked owner period summary, got %q", rendered)
	}
	if !strings.Contains(rendered, "5/5") || !strings.Contains(rendered, "Partner-visible tenderness") {
		t.Fatalf("expected partner-visible mood and symptoms in dashboard, got %q", rendered)
	}
	if strings.Contains(rendered, "Cycle data may be outdated") {
		t.Fatalf("did not expect stale cycle warning when owner logged a fresh period day, got %q", rendered)
	}
	if strings.Contains(rendered, "private owner note") {
		t.Fatalf("did not expect private owner notes in partner dashboard, got %q", rendered)
	}
	if !strings.Contains(rendered, "Partner Name") {
		t.Fatalf("expected partner identity in navigation, got %q", rendered)
	}
}

func TestPartnerCalendarUsesLinkedOwnerCycleDataReadOnly(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	owner := createOnboardingTestUser(t, database, "partner-calendar-owner@example.com", "StrongPass1", true)
	partner := createOnboardingTestUser(t, database, "partner-calendar-viewer@example.com", "StrongPass1", true)
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	if err := database.Model(&models.User{}).Where("id = ?", partner.ID).Update("role", models.RolePartner).Error; err != nil {
		t.Fatalf("set partner role: %v", err)
	}
	partner.Role = models.RolePartner

	if err := database.Create(&models.PartnerLink{
		OwnerUserID:     owner.ID,
		PartnerUserID:   partner.ID,
		InvitedByUserID: owner.ID,
		Status:          models.PartnerLinkStatusActive,
		CreatedAt:       now.UTC(),
	}).Error; err != nil {
		t.Fatalf("create partner link: %v", err)
	}
	if err := database.Create(&models.DailyLog{
		UserID:   owner.ID,
		Date:     today,
		IsPeriod: true,
		Flow:     models.FlowHeavy,
		Notes:    "private calendar note",
	}).Error; err != nil {
		t.Fatalf("create owner daily log: %v", err)
	}

	authCookie := issueAuthCookieForUser(t, partner)
	request := httptest.NewRequest(http.MethodGet, "/calendar", nil)
	request.Header.Set("Cookie", authCookie)
	response := mustAppResponse(t, app, request)
	assertStatusCode(t, response, http.StatusOK)
	rendered := mustReadBodyString(t, response.Body)

	if !strings.Contains(rendered, `data-calendar-state="period"`) {
		t.Fatalf("expected partner calendar to show linked owner period cells, got %q", rendered)
	}

	dayRequest := httptest.NewRequest(http.MethodGet, "/calendar/day/"+today.Format("2006-01-02")+"?mode=edit", nil)
	dayRequest.Header.Set("Cookie", authCookie)
	dayResponse := mustAppResponse(t, app, dayRequest)
	assertStatusCode(t, dayResponse, http.StatusOK)
	dayRendered := mustReadBodyString(t, dayResponse.Body)

	if strings.Contains(dayRendered, "<form") || strings.Contains(dayRendered, "private calendar note") {
		t.Fatalf("expected read-only sanitized partner day panel, got %q", dayRendered)
	}
	if !strings.Contains(dayRendered, "Period day") || !strings.Contains(dayRendered, "Heavy") {
		t.Fatalf("expected linked owner period summary in day panel, got %q", dayRendered)
	}
}
