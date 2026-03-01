package services

import (
	"testing"

	"github.com/terraincognita07/ovumcy/internal/models"
)

func TestSanitizeLogForViewerPartnerHidesPrivateFields(t *testing.T) {
	partner := &models.User{Role: models.RolePartner}
	entry := models.DailyLog{
		Notes:      "private",
		SymptomIDs: []uint{1, 2},
	}

	sanitized := SanitizeLogForViewer(partner, entry)
	if sanitized.Notes != "" {
		t.Fatalf("expected notes to be hidden, got %q", sanitized.Notes)
	}
	if len(sanitized.SymptomIDs) != 0 {
		t.Fatalf("expected symptom IDs to be hidden, got %#v", sanitized.SymptomIDs)
	}
}

func TestSanitizeLogForViewerOwnerKeepsFields(t *testing.T) {
	owner := &models.User{Role: models.RoleOwner}
	entry := models.DailyLog{
		Notes:      "private",
		SymptomIDs: []uint{1, 2},
	}

	sanitized := SanitizeLogForViewer(owner, entry)
	if sanitized.Notes != entry.Notes {
		t.Fatalf("expected owner notes preserved, got %q", sanitized.Notes)
	}
	if len(sanitized.SymptomIDs) != 2 {
		t.Fatalf("expected owner symptom IDs preserved, got %#v", sanitized.SymptomIDs)
	}
}

func TestSanitizeLogsForViewerPartnerHidesPrivateFieldsInAllEntries(t *testing.T) {
	partner := &models.User{Role: models.RolePartner}
	logs := []models.DailyLog{
		{Notes: "a", SymptomIDs: []uint{1}},
		{Notes: "b", SymptomIDs: []uint{2, 3}},
	}

	SanitizeLogsForViewer(partner, logs)

	for index := range logs {
		if logs[index].Notes != "" {
			t.Fatalf("expected notes to be hidden for entry %d, got %q", index, logs[index].Notes)
		}
		if len(logs[index].SymptomIDs) != 0 {
			t.Fatalf("expected symptom IDs to be hidden for entry %d, got %#v", index, logs[index].SymptomIDs)
		}
	}
}

func TestShouldExposeSymptomsForViewer(t *testing.T) {
	if !ShouldExposeSymptomsForViewer(&models.User{Role: models.RoleOwner}) {
		t.Fatal("expected owner to see symptoms")
	}
	if ShouldExposeSymptomsForViewer(&models.User{Role: models.RolePartner}) {
		t.Fatal("expected partner not to see symptoms")
	}
}
