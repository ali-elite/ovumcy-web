package services

import (
	"testing"
	"time"

	"github.com/terraincognita07/ovumcy/internal/db"
	"github.com/terraincognita07/ovumcy/internal/models"
)

func TestSymptomServiceFetchSymptomsBackfillsMissingBuiltinSymptoms(t *testing.T) {
	_, database := newDayServiceIntegration(t)
	user := createDayServiceTestUser(t, database, "symptoms-service@example.com")

	oldBuiltin := models.DefaultBuiltinSymptoms()[:7]
	records := make([]models.SymptomType, 0, len(oldBuiltin))
	for _, symptom := range oldBuiltin {
		records = append(records, models.SymptomType{
			UserID:    user.ID,
			Name:      symptom.Name,
			Icon:      symptom.Icon,
			Color:     symptom.Color,
			IsBuiltin: true,
		})
	}
	if err := database.Create(&records).Error; err != nil {
		t.Fatalf("seed old builtin symptoms: %v", err)
	}

	repositories := db.NewRepositories(database)
	service := NewSymptomService(repositories.Symptoms)

	symptoms, err := service.FetchSymptoms(user.ID)
	if err != nil {
		t.Fatalf("FetchSymptoms returned error: %v", err)
	}

	expected := models.DefaultBuiltinSymptoms()
	if len(symptoms) != len(expected) {
		t.Fatalf("expected %d symptoms after backfill, got %d", len(expected), len(symptoms))
	}
	for index, symptom := range expected {
		if symptoms[index].Name != symptom.Name {
			t.Fatalf("expected symptom %q at index %d, got %q", symptom.Name, index, symptoms[index].Name)
		}
		if !symptoms[index].IsBuiltin {
			t.Fatalf("expected symptom %q to be builtin", symptom.Name)
		}
	}
}

func TestSymptomServiceArchiveKeepsHistoryAndHidesFromPicker(t *testing.T) {
	_, database := newDayServiceIntegration(t)
	user := createDayServiceTestUser(t, database, "symptoms-archive@example.com")

	customSymptom := models.SymptomType{
		UserID: user.ID,
		Name:   "Joint stiffness",
		Icon:   "J",
		Color:  "#334455",
	}
	if err := database.Create(&customSymptom).Error; err != nil {
		t.Fatalf("create custom symptom: %v", err)
	}

	logEntry := models.DailyLog{
		UserID:     user.ID,
		Date:       time.Date(2026, time.March, 7, 0, 0, 0, 0, time.UTC),
		SymptomIDs: []uint{customSymptom.ID},
	}
	if err := database.Create(&logEntry).Error; err != nil {
		t.Fatalf("create daily log: %v", err)
	}

	repositories := db.NewRepositories(database)
	service := NewSymptomService(repositories.Symptoms)

	if err := service.ArchiveSymptomForUser(user.ID, customSymptom.ID, time.Date(2026, time.March, 8, 10, 0, 0, 0, time.UTC)); err != nil {
		t.Fatalf("archive symptom: %v", err)
	}

	allSymptoms, err := service.FetchSymptoms(user.ID)
	if err != nil {
		t.Fatalf("FetchSymptoms returned error: %v", err)
	}
	foundArchived := false
	for _, symptom := range allSymptoms {
		if symptom.ID == customSymptom.ID {
			foundArchived = true
			if symptom.IsActive() {
				t.Fatalf("expected archived symptom to become inactive")
			}
		}
	}
	if !foundArchived {
		t.Fatalf("expected archived custom symptom to remain queryable")
	}

	pickerSymptoms, err := service.FetchPickerSymptoms(user.ID, nil)
	if err != nil {
		t.Fatalf("FetchPickerSymptoms returned error: %v", err)
	}
	for _, symptom := range pickerSymptoms {
		if symptom.ID == customSymptom.ID {
			t.Fatalf("did not expect archived custom symptom in empty picker")
		}
	}

	selectedPickerSymptoms, err := service.FetchPickerSymptoms(user.ID, []uint{customSymptom.ID})
	if err != nil {
		t.Fatalf("FetchPickerSymptoms(selected) returned error: %v", err)
	}
	foundSelectedArchived := false
	for _, symptom := range selectedPickerSymptoms {
		if symptom.ID == customSymptom.ID {
			foundSelectedArchived = true
		}
	}
	if !foundSelectedArchived {
		t.Fatalf("expected selected archived symptom to remain available for existing logs")
	}

	updatedLog := models.DailyLog{}
	if err := database.First(&updatedLog, logEntry.ID).Error; err != nil {
		t.Fatalf("load daily log after archive: %v", err)
	}
	if len(updatedLog.SymptomIDs) != 1 || updatedLog.SymptomIDs[0] != customSymptom.ID {
		t.Fatalf("expected archived symptom to remain in history, got %#v", updatedLog.SymptomIDs)
	}
}
