package services

import (
	"testing"
	"time"
)

func TestLocalizedDashboardDateRussian(t *testing.T) {
	value := time.Date(2026, time.February, 18, 0, 0, 0, 0, time.UTC)

	got := LocalizedDashboardDate("ru", value)
	want := "18 февраля 2026, среда"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestLocalizedDashboardDateEnglish(t *testing.T) {
	value := time.Date(2026, time.February, 18, 0, 0, 0, 0, time.UTC)

	got := LocalizedDashboardDate("en", value)
	want := "February 18, 2026, Wednesday"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestLocalizedMonthYear(t *testing.T) {
	value := time.Date(2026, time.February, 18, 0, 0, 0, 0, time.UTC)

	if got := LocalizedMonthYear("ru", value); got != "Февраль 2026" {
		t.Fatalf("expected russian month-year, got %q", got)
	}
	if got := LocalizedMonthYear("en", value); got != "February 2026" {
		t.Fatalf("expected english month-year, got %q", got)
	}
	if got := LocalizedMonthYear("de", value); got != "February 2026" {
		t.Fatalf("expected fallback month-year, got %q", got)
	}
}

func TestLocalizedDateLabel(t *testing.T) {
	value := time.Date(2026, time.February, 18, 0, 0, 0, 0, time.UTC)

	if got := LocalizedDateLabel("ru", value); got != "Ср, Фев 18" {
		t.Fatalf("expected russian date label, got %q", got)
	}
	if got := LocalizedDateLabel("en", value); got != "Wed, Feb 18" {
		t.Fatalf("expected english date label, got %q", got)
	}
	if got := LocalizedDateLabel("de", value); got != "Wed, Feb 18" {
		t.Fatalf("expected fallback date label, got %q", got)
	}
}

func TestLocalizedDateDisplay(t *testing.T) {
	value := time.Date(2026, time.January, 29, 0, 0, 0, 0, time.UTC)

	if got := LocalizedDateDisplay("ru", value); got != "29.01.2026" {
		t.Fatalf("expected russian display date, got %q", got)
	}
	if got := LocalizedDateDisplay("en", value); got != "Jan 29, 2026" {
		t.Fatalf("expected english display date, got %q", got)
	}
}

func TestLocalizedDateShort(t *testing.T) {
	value := time.Date(2026, time.January, 29, 0, 0, 0, 0, time.UTC)

	if got := LocalizedDateShort("ru", value); got != "29.01" {
		t.Fatalf("expected russian short date, got %q", got)
	}
	if got := LocalizedDateShort("en", value); got != "Jan 29" {
		t.Fatalf("expected english short date, got %q", got)
	}
}
