package api

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/ovumcy/internal/models"
	"github.com/terraincognita07/ovumcy/internal/services"
)

func (handler *Handler) buildDashboardViewData(user *models.User, language string, messages map[string]string, now time.Time) (fiber.Map, string, error) {
	today := services.DateAtLocation(now, handler.location)

	handler.ensureDependencies()
	stats, _, err := handler.statsService.BuildCycleStatsForRange(user, today.AddDate(-2, 0, 0), today, now, handler.location)
	if err != nil {
		return nil, "failed to load logs", err
	}

	todayLog, symptoms, err := handler.viewerService.FetchDayLogForViewer(user, today, handler.location)
	if err != nil {
		return nil, "failed to load today log", err
	}

	cycleContext := services.BuildDashboardCycleContext(user, stats, today, handler.location)

	data := fiber.Map{
		"Title":                      localizedPageTitle(messages, "meta.title.dashboard", "Ovumcy | Dashboard"),
		"CurrentUser":                user,
		"Stats":                      stats,
		"CycleDayReference":          cycleContext.CycleDayReference,
		"CycleDayWarning":            cycleContext.CycleDayWarning,
		"CycleDataStale":             cycleContext.CycleDataStale,
		"DisplayNextPeriodStart":     cycleContext.DisplayNextPeriodStart,
		"DisplayOvulationDate":       cycleContext.DisplayOvulationDate,
		"DisplayOvulationExact":      cycleContext.DisplayOvulationExact,
		"DisplayOvulationImpossible": cycleContext.DisplayOvulationImpossible,
		"NextPeriodInPast":           cycleContext.NextPeriodInPast,
		"OvulationInPast":            cycleContext.OvulationInPast,
		"Today":                      today.Format("2006-01-02"),
		"FormattedDate":              services.LocalizedDashboardDate(language, today),
		"TodayEntry":                 todayLog,
		"TodayLog":                   todayLog,
		"TodayHasData":               services.DayHasData(todayLog),
		"Symptoms":                   symptoms,
		"SelectedSymptomID":          services.SymptomIDSet(todayLog.SymptomIDs),
		"IsOwner":                    services.IsOwnerUser(user),
	}
	return data, "", nil
}

func (handler *Handler) buildDayEditorPartialData(user *models.User, language string, messages map[string]string, day time.Time, now time.Time) (fiber.Map, string, error) {
	handler.ensureDependencies()
	hasDayData, err := handler.dayService.DayHasDataForDate(user.ID, day, handler.location)
	if err != nil {
		return nil, "failed to load day state", err
	}

	logEntry, symptoms, err := handler.viewerService.FetchDayLogForViewer(user, day, handler.location)
	if err != nil {
		return nil, "failed to load day", err
	}

	payload := fiber.Map{
		"Date":              day,
		"DateString":        day.Format("2006-01-02"),
		"DateLabel":         services.LocalizedDateLabel(language, day),
		"IsFutureDate":      day.After(services.DateAtLocation(now.In(handler.location), handler.location)),
		"NoDataLabel":       translateMessage(messages, "common.not_available"),
		"Log":               logEntry,
		"Symptoms":          symptoms,
		"SelectedSymptomID": services.SymptomIDSet(logEntry.SymptomIDs),
		"HasDayData":        hasDayData,
		"IsOwner":           services.IsOwnerUser(user),
	}
	return payload, "", nil
}
