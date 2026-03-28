package api

import (
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/ovumcy/ovumcy-web/internal/db"
	"github.com/ovumcy/ovumcy-web/internal/i18n"
)

func newTestAppWithPrivacyRoute(t *testing.T) *fiber.App {
	t.Helper()

	_, testFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve current test file path")
	}

	apiDir := filepath.Dir(testFile)
	internalDir := filepath.Dir(apiDir)
	templatesDir := filepath.Join(internalDir, "templates")
	localesDir := filepath.Join(internalDir, "i18n", "locales")
	databasePath := filepath.Join(t.TempDir(), "ovumcy-test.db")

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

	i18nManager, err := i18n.NewManager("en", localesDir)
	if err != nil {
		t.Fatalf("init i18n: %v", err)
	}

	handler, err := NewHandler("test-secret-key", templatesDir, time.UTC, i18nManager, false, newTestHandlerDependencies(database, i18nManager))
	if err != nil {
		t.Fatalf("init handler: %v", err)
	}

	app := fiber.New()
	app.Use(handler.LanguageMiddleware)
	RegisterRoutes(app, handler)
	return app
}
