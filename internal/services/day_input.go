package services

import (
	"errors"

	"github.com/terraincognita07/ovumcy/internal/models"
)

const MaxDayNotesLength = 2000

const (
	MinDayMood = 1
	MaxDayMood = 5
)

var (
	ErrInvalidDayFlow = errors.New("invalid day flow")
	ErrInvalidDayMood = errors.New("invalid day mood")
)

func NormalizeDayEntryInput(input DayEntryInput) (DayEntryInput, error) {
	if !IsValidDayFlow(input.Flow) {
		return input, ErrInvalidDayFlow
	}
	if !IsValidDayMood(input.Mood) {
		return input, ErrInvalidDayMood
	}
	if !input.IsPeriod {
		input.Flow = models.FlowNone
	}
	input.Notes = TrimDayNotes(input.Notes)
	return input, nil
}

func IsValidDayFlow(flow string) bool {
	switch flow {
	case models.FlowNone, models.FlowLight, models.FlowMedium, models.FlowHeavy:
		return true
	default:
		return false
	}
}

func IsValidDayMood(value int) bool {
	return value == 0 || (value >= MinDayMood && value <= MaxDayMood)
}

func TrimDayNotes(value string) string {
	if len(value) <= MaxDayNotesLength {
		return value
	}
	return value[:MaxDayNotesLength]
}
