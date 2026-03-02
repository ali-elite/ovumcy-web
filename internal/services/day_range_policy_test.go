package services

import (
	"errors"
	"testing"
	"time"
)

func TestParseDayRange(t *testing.T) {
	location := time.UTC

	t.Run("valid range", func(t *testing.T) {
		from, to, err := ParseDayRange("2026-02-10", "2026-02-20", location)
		if err != nil {
			t.Fatalf("ParseDayRange() unexpected error: %v", err)
		}
		if from.Format("2006-01-02") != "2026-02-10" {
			t.Fatalf("expected from=2026-02-10, got %s", from.Format("2006-01-02"))
		}
		if to.Format("2006-01-02") != "2026-02-20" {
			t.Fatalf("expected to=2026-02-20, got %s", to.Format("2006-01-02"))
		}
	})

	t.Run("invalid from", func(t *testing.T) {
		_, _, err := ParseDayRange("invalid-date", "2026-02-20", location)
		if !errors.Is(err, ErrDayRangeFromInvalid) {
			t.Fatalf("expected ErrDayRangeFromInvalid, got %v", err)
		}
	})

	t.Run("invalid to", func(t *testing.T) {
		_, _, err := ParseDayRange("2026-02-10", "invalid-date", location)
		if !errors.Is(err, ErrDayRangeToInvalid) {
			t.Fatalf("expected ErrDayRangeToInvalid, got %v", err)
		}
	})

	t.Run("reversed range", func(t *testing.T) {
		_, _, err := ParseDayRange("2026-02-20", "2026-02-10", location)
		if !errors.Is(err, ErrDayRangeInvalid) {
			t.Fatalf("expected ErrDayRangeInvalid, got %v", err)
		}
	})
}
