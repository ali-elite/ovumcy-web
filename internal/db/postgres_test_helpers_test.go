package db

import (
	"testing"

	"github.com/terraincognita07/ovumcy/internal/testdb"
)

func startPostgresTestConfig(t *testing.T) Config {
	t.Helper()

	return Config{
		Driver:      DriverPostgres,
		PostgresURL: testdb.StartPostgresDSN(t, "ovumcy_test"),
	}
}
