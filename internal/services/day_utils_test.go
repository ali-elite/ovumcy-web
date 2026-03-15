package services

import (
	"testing"
	"time"

	"github.com/terraincognita07/ovumcy/internal/models"
)

func TestDayHasData(t *testing.T) {
	tests := []struct {
		name  string
		entry models.DailyLog
		want  bool
	}{
		{
			name:  "period day",
			entry: models.DailyLog{IsPeriod: true},
			want:  true,
		},
		{
			name:  "symptoms present",
			entry: models.DailyLog{SymptomIDs: []uint{1}},
			want:  true,
		},
		{
			name:  "notes present",
			entry: models.DailyLog{Notes: "note"},
			want:  true,
		},
		{
			name:  "cycle factors present",
			entry: models.DailyLog{CycleFactorKeys: []string{models.CycleFactorStress}},
			want:  true,
		},
		{
			name:  "flow present",
			entry: models.DailyLog{Flow: models.FlowLight},
			want:  true,
		},
		{
			name:  "empty entry",
			entry: models.DailyLog{Flow: models.FlowNone},
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DayHasData(tt.entry); got != tt.want {
				t.Fatalf("DayHasData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDayRangeNormalizesToLocationMidnight(t *testing.T) {
	location, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		t.Fatalf("load location: %v", err)
	}

	raw := time.Date(2026, 2, 1, 19, 35, 10, 0, time.UTC)
	start, end := DayRange(raw, location)

	if start.Hour() != 0 || start.Minute() != 0 || start.Second() != 0 {
		t.Fatalf("expected midnight start, got %s", start.Format(time.RFC3339))
	}
	if !end.Equal(start.AddDate(0, 0, 1)) {
		t.Fatalf("expected next day end, got %s", end.Format(time.RFC3339))
	}
}

func TestDateAtLocationShiftsToNextLocalDayAcrossUTCBoundary(t *testing.T) {
	location := time.FixedZone("UTC+3", 3*60*60)
	raw := time.Date(2026, time.March, 2, 21, 30, 0, 0, time.UTC)

	day := DateAtLocation(raw, location)
	if day.Format("2006-01-02") != "2026-03-03" {
		t.Fatalf("expected local date 2026-03-03, got %s", day.Format("2006-01-02"))
	}
	if day.Hour() != 0 || day.Minute() != 0 || day.Second() != 0 {
		t.Fatalf("expected normalized local midnight, got %s", day.Format(time.RFC3339))
	}
}

func TestCalendarDayHelpers(t *testing.T) {
	day := time.Date(2026, time.February, 17, 0, 0, 0, 0, time.UTC)
	sameDayLater := time.Date(2026, time.February, 17, 23, 59, 0, 0, time.UTC)
	if !SameCalendarDay(day, sameDayLater) {
		t.Fatal("expected same calendar day")
	}

	start := time.Date(2026, time.February, 10, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, time.February, 20, 0, 0, 0, 0, time.UTC)
	if !BetweenCalendarDaysInclusive(day, start, end) {
		t.Fatal("expected day to be between inclusive bounds")
	}
	if BetweenCalendarDaysInclusive(day, time.Time{}, end) {
		t.Fatal("expected false when start bound is zero")
	}
}

func TestSymptomIDSet(t *testing.T) {
	set := SymptomIDSet([]uint{3, 3, 5})
	if len(set) != 2 {
		t.Fatalf("expected unique set size 2, got %d", len(set))
	}
	if !set[3] || !set[5] {
		t.Fatal("expected set to contain ids 3 and 5")
	}
}
