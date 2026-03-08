package services

import (
	"errors"
	"testing"
	"time"

	"github.com/terraincognita07/ovumcy/internal/models"
)

type stubSettingsViewLoader struct {
	user models.User
	err  error
}

func (stub *stubSettingsViewLoader) LoadSettings(_ uint) (models.User, error) {
	if stub.err != nil {
		return models.User{}, stub.err
	}
	return stub.user, nil
}

type stubSettingsViewExportBuilder struct {
	summary ExportSummary
	err     error
	called  bool
}

func (stub *stubSettingsViewExportBuilder) BuildSummary(_ uint, _ *time.Time, _ *time.Time, _ *time.Location) (ExportSummary, error) {
	stub.called = true
	if stub.err != nil {
		return ExportSummary{}, stub.err
	}
	return stub.summary, nil
}

type stubSettingsViewSymptomProvider struct {
	symptoms []models.SymptomType
	err      error
	called   bool
}

func (stub *stubSettingsViewSymptomProvider) FetchSymptoms(_ uint) ([]models.SymptomType, error) {
	stub.called = true
	if stub.err != nil {
		return nil, stub.err
	}
	result := make([]models.SymptomType, len(stub.symptoms))
	copy(result, stub.symptoms)
	return result, nil
}

func TestBuildSettingsPageViewDataClassifiesChangePasswordError(t *testing.T) {
	settingsLoader := &stubSettingsViewLoader{
		user: models.User{
			CycleLength:     28,
			PeriodLength:    5,
			AutoPeriodFill:  true,
			LastPeriodStart: nil,
		},
	}
	service := NewSettingsViewService(settingsLoader, NewNotificationService(), nil, nil)

	user := &models.User{ID: 1, Role: models.RoleOwner}
	viewData, err := service.BuildSettingsPageViewData(user, "en", SettingsViewInput{
		FlashError: "invalid current password",
	}, mustParseSettingsViewDay(t, "2026-02-21"), time.UTC)
	if err != nil {
		t.Fatalf("BuildSettingsPageViewData() unexpected error: %v", err)
	}

	if viewData.ChangePasswordErrorKey != "settings.error.invalid_current_password" {
		t.Fatalf("expected change-password error key, got %q", viewData.ChangePasswordErrorKey)
	}
	if viewData.ErrorKey != "" {
		t.Fatalf("expected empty general ErrorKey, got %q", viewData.ErrorKey)
	}
}

func TestBuildSettingsPageViewDataOwnerLoadsExportSummary(t *testing.T) {
	settingsLoader := &stubSettingsViewLoader{
		user: models.User{
			CycleLength:    28,
			PeriodLength:   5,
			AutoPeriodFill: true,
		},
	}
	exportBuilder := &stubSettingsViewExportBuilder{
		summary: ExportSummary{
			TotalEntries: 2,
			HasData:      true,
			DateFrom:     "2026-02-01",
			DateTo:       "2026-02-21",
		},
	}
	symptomProvider := &stubSettingsViewSymptomProvider{
		symptoms: []models.SymptomType{
			{ID: 1, Name: "Headache", IsBuiltin: true},
			{ID: 2, Name: "Joint stiffness"},
			{ID: 3, Name: "Caffeine crash", ArchivedAt: ptrSettingsViewTime(mustParseSettingsViewDay(t, "2026-02-01"))},
		},
	}
	service := NewSettingsViewService(settingsLoader, NewNotificationService(), exportBuilder, symptomProvider)

	user := &models.User{ID: 2, Role: models.RoleOwner}
	viewData, err := service.BuildSettingsPageViewData(user, "ru", SettingsViewInput{}, mustParseSettingsViewDay(t, "2026-02-21"), time.UTC)
	if err != nil {
		t.Fatalf("BuildSettingsPageViewData() unexpected error: %v", err)
	}

	if !exportBuilder.called {
		t.Fatalf("expected BuildSummary to be called for owner")
	}
	if !symptomProvider.called {
		t.Fatalf("expected FetchSymptoms to be called for owner")
	}
	if !viewData.HasOwnerExportViewState || !viewData.Export.HasSummaryForOwner {
		t.Fatalf("expected owner export state in view data")
	}
	if !viewData.HasOwnerSymptomsView {
		t.Fatalf("expected owner symptoms view state")
	}
	if len(viewData.Symptoms.ActiveCustomSymptoms) != 1 || viewData.Symptoms.ActiveCustomSymptoms[0].Name != "Joint stiffness" {
		t.Fatalf("expected one active custom symptom, got %#v", viewData.Symptoms.ActiveCustomSymptoms)
	}
	if len(viewData.Symptoms.ArchivedCustomSymptoms) != 1 || viewData.Symptoms.ArchivedCustomSymptoms[0].Name != "Caffeine crash" {
		t.Fatalf("expected one archived custom symptom, got %#v", viewData.Symptoms.ArchivedCustomSymptoms)
	}
	if viewData.Export.DateFromDisplay != "01.02.2026" {
		t.Fatalf("expected localized from display, got %q", viewData.Export.DateFromDisplay)
	}
	if viewData.Export.DateToDisplay != "21.02.2026" {
		t.Fatalf("expected localized to display, got %q", viewData.Export.DateToDisplay)
	}
}

func TestBuildSettingsPageViewDataPartnerSkipsExportSummary(t *testing.T) {
	settingsLoader := &stubSettingsViewLoader{
		user: models.User{
			CycleLength:    28,
			PeriodLength:   5,
			AutoPeriodFill: true,
		},
	}
	exportBuilder := &stubSettingsViewExportBuilder{}
	symptomProvider := &stubSettingsViewSymptomProvider{}
	service := NewSettingsViewService(settingsLoader, NewNotificationService(), exportBuilder, symptomProvider)

	user := &models.User{ID: 3, Role: models.RolePartner}
	viewData, err := service.BuildSettingsPageViewData(user, "en", SettingsViewInput{}, mustParseSettingsViewDay(t, "2026-02-21"), time.UTC)
	if err != nil {
		t.Fatalf("BuildSettingsPageViewData() unexpected error: %v", err)
	}
	if exportBuilder.called {
		t.Fatalf("did not expect BuildSummary call for partner")
	}
	if symptomProvider.called {
		t.Fatalf("did not expect FetchSymptoms call for partner")
	}
	if viewData.HasOwnerExportViewState {
		t.Fatalf("expected no owner export state for partner")
	}
	if viewData.HasOwnerSymptomsView {
		t.Fatalf("expected no owner symptoms view state for partner")
	}
}

func TestBuildSettingsPageViewDataReturnsTypedErrors(t *testing.T) {
	user := &models.User{ID: 4, Role: models.RoleOwner}

	settingsErrService := NewSettingsViewService(
		&stubSettingsViewLoader{err: errors.New("settings fail")},
		NewNotificationService(),
		nil,
		nil,
	)
	if _, err := settingsErrService.BuildSettingsPageViewData(user, "en", SettingsViewInput{}, mustParseSettingsViewDay(t, "2026-02-21"), time.UTC); !errors.Is(err, ErrSettingsViewLoadSettings) {
		t.Fatalf("expected ErrSettingsViewLoadSettings, got %v", err)
	}

	exportErrService := NewSettingsViewService(
		&stubSettingsViewLoader{user: models.User{CycleLength: 28, PeriodLength: 5, AutoPeriodFill: true}},
		NewNotificationService(),
		&stubSettingsViewExportBuilder{err: errors.New("export fail")},
		nil,
	)
	if _, err := exportErrService.BuildSettingsPageViewData(user, "en", SettingsViewInput{}, mustParseSettingsViewDay(t, "2026-02-21"), time.UTC); !errors.Is(err, ErrSettingsViewLoadExport) {
		t.Fatalf("expected ErrSettingsViewLoadExport, got %v", err)
	}

	symptomErrService := NewSettingsViewService(
		&stubSettingsViewLoader{user: models.User{CycleLength: 28, PeriodLength: 5, AutoPeriodFill: true}},
		NewNotificationService(),
		nil,
		&stubSettingsViewSymptomProvider{err: errors.New("symptom fail")},
	)
	if _, err := symptomErrService.BuildSettingsPageViewData(user, "en", SettingsViewInput{}, mustParseSettingsViewDay(t, "2026-02-21"), time.UTC); !errors.Is(err, ErrSettingsViewLoadSymptoms) {
		t.Fatalf("expected ErrSettingsViewLoadSymptoms, got %v", err)
	}
}

func mustParseSettingsViewDay(t *testing.T, raw string) time.Time {
	t.Helper()
	parsed, err := time.ParseInLocation("2006-01-02", raw, time.UTC)
	if err != nil {
		t.Fatalf("parse day %q: %v", raw, err)
	}
	return parsed
}

func ptrSettingsViewTime(value time.Time) *time.Time {
	return &value
}
