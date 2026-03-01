package api

import (
	"time"

	"github.com/terraincognita07/ovumcy/internal/models"
	"github.com/terraincognita07/ovumcy/internal/services"
)

func dayHasData(entry models.DailyLog) bool {
	return services.DayHasData(entry)
}

func sameCalendarDay(a time.Time, b time.Time) bool {
	return services.SameCalendarDay(a, b)
}

func betweenCalendarDaysInclusive(day time.Time, start time.Time, end time.Time) bool {
	return services.BetweenCalendarDaysInclusive(day, start, end)
}

func sanitizeLogForPartner(entry models.DailyLog) models.DailyLog {
	return services.SanitizeLogForPartner(entry)
}

func dateAtLocation(value time.Time, location *time.Location) time.Time {
	return services.DateAtLocation(value, location)
}

func dayRange(value time.Time, location *time.Location) (time.Time, time.Time) {
	return services.DayRange(value, location)
}

func symptomIDSet(ids []uint) map[uint]bool {
	return services.SymptomIDSet(ids)
}
