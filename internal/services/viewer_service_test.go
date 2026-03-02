package services

import (
	"errors"
	"testing"
	"time"

	"github.com/terraincognita07/ovumcy/internal/models"
)

type stubViewerDayReader struct {
	entry models.DailyLog
	err   error
}

func (stub *stubViewerDayReader) FetchLogByDate(uint, time.Time, *time.Location) (models.DailyLog, error) {
	if stub.err != nil {
		return models.DailyLog{}, stub.err
	}
	return stub.entry, nil
}

type stubViewerSymptomReader struct {
	symptoms []models.SymptomType
	err      error
}

func (stub *stubViewerSymptomReader) FetchSymptoms(uint) ([]models.SymptomType, error) {
	if stub.err != nil {
		return nil, stub.err
	}
	result := make([]models.SymptomType, len(stub.symptoms))
	copy(result, stub.symptoms)
	return result, nil
}

func TestViewerServiceFetchSymptomsForViewer_OnlyOwner(t *testing.T) {
	service := NewViewerService(&stubViewerDayReader{}, &stubViewerSymptomReader{
		symptoms: []models.SymptomType{{Name: "Headache"}},
	})

	ownerSymptoms, err := service.FetchSymptomsForViewer(&models.User{ID: 10, Role: models.RoleOwner})
	if err != nil {
		t.Fatalf("FetchSymptomsForViewer(owner) unexpected error: %v", err)
	}
	if len(ownerSymptoms) != 1 {
		t.Fatalf("expected owner symptoms to load, got %#v", ownerSymptoms)
	}

	partnerSymptoms, err := service.FetchSymptomsForViewer(&models.User{ID: 11, Role: models.RolePartner})
	if err != nil {
		t.Fatalf("FetchSymptomsForViewer(partner) unexpected error: %v", err)
	}
	if len(partnerSymptoms) != 0 {
		t.Fatalf("expected partner symptoms to be hidden, got %#v", partnerSymptoms)
	}
}

func TestViewerServiceFetchDayLogForViewer_SanitizesPartnerAndSkipsSymptoms(t *testing.T) {
	service := NewViewerService(
		&stubViewerDayReader{
			entry: models.DailyLog{
				IsPeriod:   true,
				Flow:       models.FlowMedium,
				Notes:      "private",
				SymptomIDs: []uint{1, 2},
			},
		},
		&stubViewerSymptomReader{
			symptoms: []models.SymptomType{{Name: "Headache"}},
		},
	)

	logEntry, symptoms, err := service.FetchDayLogForViewer(&models.User{ID: 10, Role: models.RolePartner}, time.Now().UTC(), time.UTC)
	if err != nil {
		t.Fatalf("FetchDayLogForViewer(partner) unexpected error: %v", err)
	}
	if logEntry.Notes != "" || len(logEntry.SymptomIDs) != 0 {
		t.Fatalf("expected partner log to be sanitized, got %#v", logEntry)
	}
	if len(symptoms) != 0 {
		t.Fatalf("expected partner symptoms hidden, got %#v", symptoms)
	}
}

func TestViewerServiceFetchDayLogForViewer_PropagatesErrors(t *testing.T) {
	dayErr := errors.New("day fetch failed")
	service := NewViewerService(&stubViewerDayReader{err: dayErr}, &stubViewerSymptomReader{})

	_, _, err := service.FetchDayLogForViewer(&models.User{ID: 10, Role: models.RoleOwner}, time.Now().UTC(), time.UTC)
	if !errors.Is(err, dayErr) {
		t.Fatalf("expected day fetch error, got %v", err)
	}

	symptomErr := errors.New("symptom fetch failed")
	service = NewViewerService(
		&stubViewerDayReader{
			entry: models.DailyLog{Notes: "owner-note"},
		},
		&stubViewerSymptomReader{err: symptomErr},
	)

	_, _, err = service.FetchDayLogForViewer(&models.User{ID: 10, Role: models.RoleOwner}, time.Now().UTC(), time.UTC)
	if !errors.Is(err, symptomErr) {
		t.Fatalf("expected symptom fetch error, got %v", err)
	}
}
