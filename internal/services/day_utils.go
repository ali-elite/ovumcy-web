package services

import (
	"strings"
	"time"

	"github.com/terraincognita07/ovumcy/internal/models"
)

func DateAtLocation(value time.Time, location *time.Location) time.Time {
	if location == nil {
		location = time.UTC
	}
	localized := value.In(location)
	year, month, day := localized.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, location)
}

func DayRange(value time.Time, location *time.Location) (time.Time, time.Time) {
	start := DateAtLocation(value, location)
	return start, start.AddDate(0, 0, 1)
}

func SameCalendarDay(a time.Time, b time.Time) bool {
	return a.Format("2006-01-02") == b.Format("2006-01-02")
}

func BetweenCalendarDaysInclusive(day time.Time, start time.Time, end time.Time) bool {
	if start.IsZero() || end.IsZero() {
		return false
	}
	return (day.Equal(start) || day.After(start)) && (day.Equal(end) || day.Before(end))
}

func SymptomIDSet(ids []uint) map[uint]bool {
	set := make(map[uint]bool, len(ids))
	for _, id := range ids {
		set[id] = true
	}
	return set
}

func DayHasData(entry models.DailyLog) bool {
	if entry.IsPeriod {
		return true
	}
	if entry.Mood >= MinDayMood && entry.Mood <= MaxDayMood {
		return true
	}
	if NormalizeDaySexActivity(entry.SexActivity) != models.SexActivityNone {
		return true
	}
	if IsValidDayBBT(entry.BBT) && entry.BBT > 0 {
		return true
	}
	if NormalizeDayCervicalMucus(entry.CervicalMucus) != models.CervicalMucusNone {
		return true
	}
	if len(entry.SymptomIDs) > 0 {
		return true
	}
	if strings.TrimSpace(entry.Notes) != "" {
		return true
	}
	return strings.TrimSpace(entry.Flow) != "" && entry.Flow != models.FlowNone
}

func RemoveUint(values []uint, needle uint) []uint {
	filtered := make([]uint, 0, len(values))
	for _, value := range values {
		if value != needle {
			filtered = append(filtered, value)
		}
	}
	return filtered
}
