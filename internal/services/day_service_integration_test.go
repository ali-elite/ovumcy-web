package services

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/terraincognita07/ovumcy/internal/db"
	"github.com/terraincognita07/ovumcy/internal/models"
	"gorm.io/gorm"
)

func newDayServiceIntegration(t *testing.T) (*DayService, *gorm.DB) {
	t.Helper()

	databasePath := filepath.Join(t.TempDir(), "ovumcy-day-service-int.db")
	database, err := db.OpenSQLite(databasePath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sqlDB, err := database.DB()
	if err != nil {
		t.Fatalf("open sql db: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	repositories := db.NewRepositories(database)
	service := NewDayService(repositories.DailyLogs, repositories.Users)
	return service, database
}

func createDayServiceTestUser(t *testing.T, database *gorm.DB, email string) models.User {
	t.Helper()

	user := models.User{
		Email:               email,
		PasswordHash:        "test-hash",
		Role:                models.RoleOwner,
		OnboardingCompleted: true,
		CycleLength:         28,
		PeriodLength:        5,
		CreatedAt:           time.Now().UTC(),
	}
	if err := database.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	return user
}

func TestDayServiceFetchLogByDateFindsZuluStoredRowForLocalCalendarDay(t *testing.T) {
	service, database := newDayServiceIntegration(t)
	user := createDayServiceTestUser(t, database, "zulu-fetch-service@example.com")

	now := time.Now().UTC()
	if err := database.Exec(
		`INSERT INTO daily_logs (user_id, date, is_period, flow, symptom_ids, notes, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		user.ID,
		"2026-02-17T00:00:00Z",
		true,
		models.FlowLight,
		"[]",
		"",
		now,
		now,
	).Error; err != nil {
		t.Fatalf("insert zulu row: %v", err)
	}

	moscow := time.FixedZone("UTC+3", 3*60*60)
	day, err := ParseDayDate("2026-02-17", moscow)
	if err != nil {
		t.Fatalf("parse day: %v", err)
	}

	entry, err := service.FetchLogByDate(user.ID, day, moscow)
	if err != nil {
		t.Fatalf("FetchLogByDate: %v", err)
	}
	if !entry.IsPeriod {
		t.Fatalf("expected is_period=true for local day 2026-02-17")
	}
	if entry.Flow != models.FlowLight {
		t.Fatalf("expected flow %q, got %q", models.FlowLight, entry.Flow)
	}
}

func TestDayServiceFetchLogByDateIgnoresUTCShiftedRowForLocalCalendarDay(t *testing.T) {
	service, database := newDayServiceIntegration(t)
	user := createDayServiceTestUser(t, database, "zulu-shifted-service@example.com")

	now := time.Now().UTC()
	if err := database.Exec(
		`INSERT INTO daily_logs (user_id, date, is_period, flow, symptom_ids, notes, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		user.ID,
		"2026-02-21T21:00:00Z",
		true,
		models.FlowMedium,
		"[]",
		"",
		now,
		now,
	).Error; err != nil {
		t.Fatalf("insert utc shifted row: %v", err)
	}

	moscow := time.FixedZone("UTC+3", 3*60*60)
	day, err := ParseDayDate("2026-02-22", moscow)
	if err != nil {
		t.Fatalf("parse day: %v", err)
	}

	entry, err := service.FetchLogByDate(user.ID, day, moscow)
	if err != nil {
		t.Fatalf("FetchLogByDate: %v", err)
	}
	if entry.IsPeriod {
		t.Fatalf("expected no period row for local day 2026-02-22")
	}
	if entry.Flow != models.FlowNone {
		t.Fatalf("expected default flow %q, got %q", models.FlowNone, entry.Flow)
	}
}

func TestDayServiceFetchLogsForUserExcludesUTCShiftedRowForLocalDayRange(t *testing.T) {
	service, database := newDayServiceIntegration(t)
	user := createDayServiceTestUser(t, database, "zulu-shifted-range-service@example.com")

	now := time.Now().UTC()
	if err := database.Exec(
		`INSERT INTO daily_logs (user_id, date, is_period, flow, symptom_ids, notes, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		user.ID,
		"2026-02-21T21:00:00Z",
		true,
		models.FlowHeavy,
		"[]",
		"",
		now,
		now,
	).Error; err != nil {
		t.Fatalf("insert utc shifted row: %v", err)
	}

	moscow := time.FixedZone("UTC+3", 3*60*60)
	from, err := ParseDayDate("2026-02-22", moscow)
	if err != nil {
		t.Fatalf("parse from day: %v", err)
	}
	to := from

	logs, err := service.FetchLogsForUser(user.ID, from, to, moscow)
	if err != nil {
		t.Fatalf("FetchLogsForUser: %v", err)
	}
	if len(logs) != 0 {
		t.Fatalf("expected no rows in strict local-day range, got %d", len(logs))
	}
}

func TestDayServiceDayHasDataForDate(t *testing.T) {
	service, database := newDayServiceIntegration(t)
	user := createDayServiceTestUser(t, database, "day-has-data-service@example.com")

	day := time.Date(2026, time.February, 18, 0, 0, 0, 0, time.UTC)
	hasData, err := service.DayHasDataForDate(user.ID, day, time.UTC)
	if err != nil {
		t.Fatalf("DayHasDataForDate returned error: %v", err)
	}
	if hasData {
		t.Fatal("expected false when no entries exist")
	}

	entry := models.DailyLog{
		UserID:   user.ID,
		Date:     day,
		IsPeriod: false,
		Flow:     models.FlowNone,
		Notes:    "note",
	}
	if err := database.Create(&entry).Error; err != nil {
		t.Fatalf("create log: %v", err)
	}

	hasData, err = service.DayHasDataForDate(user.ID, day, time.UTC)
	if err != nil {
		t.Fatalf("DayHasDataForDate returned error: %v", err)
	}
	if !hasData {
		t.Fatal("expected true when notes exist for the day")
	}
}

func TestDayServiceRefreshUserLastPeriodStart(t *testing.T) {
	service, database := newDayServiceIntegration(t)
	user := createDayServiceTestUser(t, database, "refresh-last-period-service@example.com")

	first := time.Date(2026, time.January, 10, 0, 0, 0, 0, time.UTC)
	second := time.Date(2026, time.February, 10, 0, 0, 0, 0, time.UTC)
	logs := []models.DailyLog{
		{UserID: user.ID, Date: first, IsPeriod: true, Flow: models.FlowMedium, SymptomIDs: []uint{}},
		{UserID: user.ID, Date: second, IsPeriod: true, Flow: models.FlowMedium, SymptomIDs: []uint{}},
	}
	if err := database.Create(&logs).Error; err != nil {
		t.Fatalf("create period logs: %v", err)
	}

	if err := service.RefreshUserLastPeriodStart(user.ID, time.UTC); err != nil {
		t.Fatalf("RefreshUserLastPeriodStart returned error: %v", err)
	}

	updated := models.User{}
	if err := database.First(&updated, user.ID).Error; err != nil {
		t.Fatalf("load updated user: %v", err)
	}
	if updated.LastPeriodStart == nil {
		t.Fatal("expected last_period_start to be populated")
	}
	if updated.LastPeriodStart.Format("2006-01-02") != second.Format("2006-01-02") {
		t.Fatalf("expected latest period start %s, got %s", second.Format("2006-01-02"), updated.LastPeriodStart.Format("2006-01-02"))
	}

	if err := database.Where("user_id = ?", user.ID).Delete(&models.DailyLog{}).Error; err != nil {
		t.Fatalf("delete logs: %v", err)
	}
	if err := service.RefreshUserLastPeriodStart(user.ID, time.UTC); err != nil {
		t.Fatalf("RefreshUserLastPeriodStart second call returned error: %v", err)
	}

	updated = models.User{}
	if err := database.First(&updated, user.ID).Error; err != nil {
		t.Fatalf("reload user: %v", err)
	}
	if updated.LastPeriodStart != nil {
		t.Fatalf("expected last_period_start to be cleared, got %v", updated.LastPeriodStart)
	}
}
