package services

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/terraincognita07/ovumcy/internal/db"
	"github.com/terraincognita07/ovumcy/internal/models"
	"github.com/terraincognita07/ovumcy/internal/testdb"
	"gorm.io/gorm"
)

func newDayServiceIntegration(t *testing.T) (*DayService, *gorm.DB) {
	t.Helper()

	return newDayServiceIntegrationWithConfig(t, db.Config{
		Driver:     db.DriverSQLite,
		SQLitePath: filepath.Join(t.TempDir(), "ovumcy-day-service-int.db"),
	})
}

func newDayServicePostgresIntegration(t *testing.T) (*DayService, *gorm.DB) {
	t.Helper()

	return newDayServiceIntegrationWithConfig(t, db.Config{
		Driver:      db.DriverPostgres,
		PostgresURL: testdb.StartPostgresDSN(t, "ovumcy_day_service_test"),
	})
}

func newDayServiceIntegrationWithConfig(t *testing.T, databaseConfig db.Config) (*DayService, *gorm.DB) {
	t.Helper()

	database, err := db.OpenDatabase(databaseConfig)
	if err != nil {
		t.Fatalf("open integration database: %v", err)
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

func assertDayServiceFetchLogByDateFindsZuluStoredRowForLocalCalendarDay(t *testing.T, setup func(*testing.T) (*DayService, *gorm.DB), email string) {
	t.Helper()

	service, database := setup(t)
	user := createDayServiceTestUser(t, database, email)

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
	assertDayServiceFetchLogByDateFindsZuluStoredRowForLocalCalendarDay(t, newDayServiceIntegration, "zulu-fetch-service@example.com")
}

func TestDayServiceFetchLogByDateFindsZuluStoredRowForLocalCalendarDayPostgres(t *testing.T) {
	assertDayServiceFetchLogByDateFindsZuluStoredRowForLocalCalendarDay(t, newDayServicePostgresIntegration, "zulu-fetch-service-postgres@example.com")
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

func TestDayServiceMarkCycleStartManuallyPreservesEntryAndUpdatesAnchor(t *testing.T) {
	service, database := newDayServiceIntegration(t)
	user := createDayServiceTestUser(t, database, "manual-cycle-start-service@example.com")

	targetDay := time.Date(2026, time.February, 18, 0, 0, 0, 0, time.UTC)
	entry := models.DailyLog{
		UserID:        user.ID,
		Date:          targetDay,
		IsPeriod:      false,
		Flow:          models.FlowHeavy,
		Mood:          4,
		SexActivity:   models.SexActivityProtected,
		CervicalMucus: models.CervicalMucusCreamy,
		Notes:         "keep this note",
		SymptomIDs:    []uint{11, 22},
	}
	if err := database.Create(&entry).Error; err != nil {
		t.Fatalf("create log: %v", err)
	}

	if err := service.MarkCycleStartManually(user.ID, targetDay, time.UTC); err != nil {
		t.Fatalf("MarkCycleStartManually returned error: %v", err)
	}

	updatedEntry := models.DailyLog{}
	if err := database.Where("user_id = ? AND date = ?", user.ID, targetDay).First(&updatedEntry).Error; err != nil {
		t.Fatalf("load updated log: %v", err)
	}
	if !updatedEntry.IsPeriod {
		t.Fatalf("expected selected day to become a period day")
	}
	if updatedEntry.Flow != models.FlowHeavy {
		t.Fatalf("expected flow to be preserved, got %q", updatedEntry.Flow)
	}
	if updatedEntry.Notes != "keep this note" {
		t.Fatalf("expected notes to be preserved, got %q", updatedEntry.Notes)
	}
	if len(updatedEntry.SymptomIDs) != 2 || updatedEntry.SymptomIDs[0] != 11 || updatedEntry.SymptomIDs[1] != 22 {
		t.Fatalf("expected symptom ids to be preserved, got %#v", updatedEntry.SymptomIDs)
	}

	updatedUser := models.User{}
	if err := database.First(&updatedUser, user.ID).Error; err != nil {
		t.Fatalf("load updated user: %v", err)
	}
	if updatedUser.LastPeriodStart == nil {
		t.Fatalf("expected last_period_start to be updated")
	}
	if got := updatedUser.LastPeriodStart.Format("2006-01-02"); got != "2026-02-18" {
		t.Fatalf("expected last_period_start 2026-02-18, got %s", got)
	}
}
