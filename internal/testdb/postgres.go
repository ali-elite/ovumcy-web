package testdb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	postgresTestUser     = "ovumcy"
	postgresTestPassword = "ovumcy"
)

// StartPostgresDSN launches an isolated Postgres container for tests and
// returns a DSN suitable for gorm.io/driver/postgres.
func StartPostgresDSN(t *testing.T, databaseName string) string {
	t.Helper()

	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker is required for postgres tests")
	}

	databaseName = strings.TrimSpace(databaseName)
	if databaseName == "" {
		t.Fatal("postgres test database name is required")
	}

	containerID := runDockerCommand(t, "run", "-d", "--rm", "-P",
		"-e", "POSTGRES_USER="+postgresTestUser,
		"-e", "POSTGRES_PASSWORD="+postgresTestPassword,
		"-e", "POSTGRES_DB="+databaseName,
		"postgres:17-alpine",
	)

	t.Cleanup(func() {
		_ = runDockerCommandAllowFailure("rm", "-f", containerID)
	})

	waitForPostgresReadiness(t, containerID, databaseName)
	port := loadPostgresMappedPort(t, containerID)
	dsn := fmt.Sprintf(
		"host=127.0.0.1 port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=UTC",
		port,
		postgresTestUser,
		postgresTestPassword,
		databaseName,
	)
	waitForHostSQLReadiness(t, dsn)

	return dsn
}

func waitForPostgresReadiness(t *testing.T, containerID string, databaseName string) {
	t.Helper()

	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := runDockerCommandWithError("exec", containerID, "pg_isready", "-U", postgresTestUser, "-d", databaseName); err == nil {
			return
		}
		time.Sleep(500 * time.Millisecond)
	}

	logs, _ := runDockerCommandWithError("logs", containerID)
	t.Fatalf("postgres test container %s did not become ready in time; logs: %s", containerID, logs)
}

func loadPostgresMappedPort(t *testing.T, containerID string) string {
	t.Helper()

	output := runDockerCommand(t, "port", containerID, "5432/tcp")
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) == "" {
		t.Fatalf("docker port returned no mapping for postgres container %s", containerID)
	}

	mapping := strings.TrimSpace(lines[0])
	lastColon := strings.LastIndex(mapping, ":")
	if lastColon < 0 || lastColon == len(mapping)-1 {
		t.Fatalf("unexpected docker port mapping %q", mapping)
	}
	return mapping[lastColon+1:]
}

func waitForHostSQLReadiness(t *testing.T, dsn string) {
	t.Helper()

	deadline := time.Now().Add(60 * time.Second)
	var lastErr error
	for time.Now().Before(deadline) {
		pingErr := pingPostgresDSN(dsn)
		if pingErr == nil {
			return
		}
		lastErr = pingErr
		time.Sleep(500 * time.Millisecond)
	}

	t.Fatalf("postgres test database did not become reachable from host in time: %v", lastErr)
}

func pingPostgresDSN(dsn string) error {
	database, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	defer database.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return database.PingContext(ctx)
}

func runDockerCommand(t *testing.T, args ...string) string {
	t.Helper()

	output, err := runDockerCommandWithError(args...)
	if err != nil {
		t.Skipf("docker is unavailable for postgres tests: %v", err)
	}
	return output
}

func runDockerCommandWithError(args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	command := exec.CommandContext(ctx, "docker", args...)
	output, err := command.CombinedOutput()
	if ctx.Err() != nil && errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return "", fmt.Errorf("docker %s timed out", strings.Join(args, " "))
	}
	if err != nil {
		return "", fmt.Errorf("docker %s failed: %v: %s", strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}
	return strings.TrimSpace(string(output)), nil
}

func runDockerCommandAllowFailure(args ...string) error {
	_, err := runDockerCommandWithError(args...)
	return err
}
