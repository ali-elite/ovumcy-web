package services

import (
	"testing"

	"github.com/ovumcy/ovumcy-web/internal/models"
)

func TestNormalizeDayCycleFactorKeysStabilizesOrderAndRejectsUnknownValues(t *testing.T) {
	keys, allValid := NormalizeDayCycleFactorKeys([]string{
		models.CycleFactorTravel,
		"  stress ",
		models.CycleFactorTravel,
		"unknown",
	})

	if allValid {
		t.Fatal("expected invalid values to be reported")
	}
	if len(keys) != 2 {
		t.Fatalf("expected two normalized keys, got %#v", keys)
	}
	if keys[0] != models.CycleFactorStress || keys[1] != models.CycleFactorTravel {
		t.Fatalf("expected stable supported order, got %#v", keys)
	}
}

func TestDayCycleFactorKeySetUsesNormalizedKeys(t *testing.T) {
	set := DayCycleFactorKeySet([]string{" travel ", models.CycleFactorStress, models.CycleFactorStress})
	if len(set) != 2 {
		t.Fatalf("expected two unique factors, got %#v", set)
	}
	if !set[models.CycleFactorStress] || !set[models.CycleFactorTravel] {
		t.Fatalf("expected normalized factor keys, got %#v", set)
	}
}
