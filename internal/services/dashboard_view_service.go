package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/terraincognita07/ovumcy/internal/models"
)

var (
	ErrDashboardViewLoadStats    = errors.New("dashboard view load stats")
	ErrDashboardViewLoadTodayLog = errors.New("dashboard view load today log")
	ErrDashboardViewLoadDayState = errors.New("dashboard view load day state")
	ErrDashboardViewLoadDayLog   = errors.New("dashboard view load day log")
)

type DashboardStatsProvider interface {
	BuildCycleStatsForRange(user *models.User, from time.Time, to time.Time, now time.Time, location *time.Location) (CycleStats, []models.DailyLog, error)
}

type DashboardViewerProvider interface {
	FetchDayLogForViewer(user *models.User, day time.Time, location *time.Location) (models.DailyLog, []models.SymptomType, error)
}

type DashboardDayStateProvider interface {
	DayHasDataForDate(userID uint, day time.Time, location *time.Location) (bool, error)
}

type DashboardViewService struct {
	stats  DashboardStatsProvider
	viewer DashboardViewerProvider
	days   DashboardDayStateProvider
}

type DashboardViewData struct {
	Stats             CycleStats
	CycleContext      DashboardCycleContext
	Today             time.Time
	FormattedDate     string
	TodayLog          models.DailyLog
	TodayHasData      bool
	Symptoms          []models.SymptomType
	SelectedSymptomID map[uint]bool
	IsOwner           bool
}

type DayEditorViewData struct {
	Date              time.Time
	DateString        string
	DateLabel         string
	IsFutureDate      bool
	Log               models.DailyLog
	Symptoms          []models.SymptomType
	SelectedSymptomID map[uint]bool
	HasDayData        bool
	IsOwner           bool
}

func NewDashboardViewService(stats DashboardStatsProvider, viewer DashboardViewerProvider, days DashboardDayStateProvider) *DashboardViewService {
	return &DashboardViewService{
		stats:  stats,
		viewer: viewer,
		days:   days,
	}
}

func (service *DashboardViewService) BuildDashboardViewData(user *models.User, language string, now time.Time, location *time.Location) (DashboardViewData, error) {
	today := DateAtLocation(now, location)

	stats, _, err := service.stats.BuildCycleStatsForRange(user, today.AddDate(-2, 0, 0), today, now, location)
	if err != nil {
		return DashboardViewData{}, fmt.Errorf("%w: %v", ErrDashboardViewLoadStats, err)
	}

	todayLog, symptoms, err := service.viewer.FetchDayLogForViewer(user, today, location)
	if err != nil {
		return DashboardViewData{}, fmt.Errorf("%w: %v", ErrDashboardViewLoadTodayLog, err)
	}

	cycleContext := BuildDashboardCycleContext(user, stats, today, location)

	return DashboardViewData{
		Stats:             stats,
		CycleContext:      cycleContext,
		Today:             today,
		FormattedDate:     LocalizedDashboardDate(language, today),
		TodayLog:          todayLog,
		TodayHasData:      DayHasData(todayLog),
		Symptoms:          symptoms,
		SelectedSymptomID: SymptomIDSet(todayLog.SymptomIDs),
		IsOwner:           IsOwnerUser(user),
	}, nil
}

func (service *DashboardViewService) BuildDayEditorViewData(user *models.User, language string, day time.Time, now time.Time, location *time.Location) (DayEditorViewData, error) {
	hasDayData, err := service.days.DayHasDataForDate(user.ID, day, location)
	if err != nil {
		return DayEditorViewData{}, fmt.Errorf("%w: %v", ErrDashboardViewLoadDayState, err)
	}

	logEntry, symptoms, err := service.viewer.FetchDayLogForViewer(user, day, location)
	if err != nil {
		return DayEditorViewData{}, fmt.Errorf("%w: %v", ErrDashboardViewLoadDayLog, err)
	}

	return DayEditorViewData{
		Date:              day,
		DateString:        day.Format("2006-01-02"),
		DateLabel:         LocalizedDateLabel(language, day),
		IsFutureDate:      day.After(DateAtLocation(now.In(location), location)),
		Log:               logEntry,
		Symptoms:          symptoms,
		SelectedSymptomID: SymptomIDSet(logEntry.SymptomIDs),
		HasDayData:        hasDayData,
		IsOwner:           IsOwnerUser(user),
	}, nil
}
