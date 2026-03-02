package services

import (
	"testing"

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
	service := NewSymptomService(repositories.Symptoms, repositories.DailyLogs)

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
