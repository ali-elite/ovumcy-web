package cli

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/ovumcy/ovumcy-web/internal/db"
)

func TestRunResetPasswordCommandRejectsBlankEmail(t *testing.T) {
	t.Parallel()

	err := runResetPasswordCommand(db.Config{}, "   ", nil, io.Discard)
	if err == nil || err.Error() != "email is required" {
		t.Fatalf("expected blank email error, got %v", err)
	}
}

func TestRunResetPasswordCommandRejectsInvalidEmail(t *testing.T) {
	t.Parallel()

	err := runResetPasswordCommand(db.Config{}, "not-an-email", nil, io.Discard)
	if err == nil || !strings.Contains(err.Error(), "invalid email address") {
		t.Fatalf("expected invalid email error, got %v", err)
	}
}

func TestRunResetPasswordCommandRequiresPasswordPrompt(t *testing.T) {
	t.Parallel()

	databasePath := createCLIResetDatabase(t)
	createCLIResetUser(t, databasePath, "cli-reset-nil-prompt@example.com", "StrongPass1")

	err := runResetPasswordCommand(
		db.Config{Driver: db.DriverSQLite, SQLitePath: databasePath},
		"cli-reset-nil-prompt@example.com",
		nil,
		io.Discard,
	)
	if err == nil || err.Error() != "password prompt is required" {
		t.Fatalf("expected nil prompt error, got %v", err)
	}
}

func TestRunResetPasswordCommandRejectsEmptyPromptedPassword(t *testing.T) {
	t.Parallel()

	databasePath := createCLIResetDatabase(t)
	createCLIResetUser(t, databasePath, "cli-reset-empty-password@example.com", "StrongPass1")

	err := runResetPasswordCommand(
		db.Config{Driver: db.DriverSQLite, SQLitePath: databasePath},
		"cli-reset-empty-password@example.com",
		func() ([]byte, error) { return []byte{}, nil },
		io.Discard,
	)
	if err == nil || err.Error() != "password is required" {
		t.Fatalf("expected empty password error, got %v", err)
	}
}

func TestRunResetPasswordCommandRejectsWeakPassword(t *testing.T) {
	t.Parallel()

	databasePath := createCLIResetDatabase(t)
	createCLIResetUser(t, databasePath, "cli-reset-weak-password@example.com", "StrongPass1")

	err := runResetPasswordCommand(
		db.Config{Driver: db.DriverSQLite, SQLitePath: databasePath},
		"cli-reset-weak-password@example.com",
		func() ([]byte, error) { return []byte("weakpass"), nil },
		io.Discard,
	)
	if err == nil || err.Error() != "password does not meet strength requirements" {
		t.Fatalf("expected weak password error, got %v", err)
	}
}

func TestRunResetPasswordCommandReportsMissingUser(t *testing.T) {
	t.Parallel()

	databasePath := createCLIResetDatabase(t)

	err := runResetPasswordCommand(
		db.Config{Driver: db.DriverSQLite, SQLitePath: databasePath},
		"missing-reset-user@example.com",
		func() ([]byte, error) { return []byte("StrongPass2"), nil },
		io.Discard,
	)
	if err == nil || !strings.Contains(err.Error(), "user missing-reset-user@example.com not found") {
		t.Fatalf("expected missing user error, got %v", err)
	}
}

func TestRunResetPasswordCommandWrapsPromptReadFailure(t *testing.T) {
	t.Parallel()

	databasePath := createCLIResetDatabase(t)
	createCLIResetUser(t, databasePath, "cli-reset-read-failure@example.com", "StrongPass1")

	err := runResetPasswordCommand(
		db.Config{Driver: db.DriverSQLite, SQLitePath: databasePath},
		"cli-reset-read-failure@example.com",
		func() ([]byte, error) { return nil, errors.New("terminal unavailable") },
		io.Discard,
	)
	if err == nil || !strings.Contains(err.Error(), "read new password") {
		t.Fatalf("expected wrapped prompt error, got %v", err)
	}
}
