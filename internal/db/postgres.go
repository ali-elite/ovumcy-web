package db

import (
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func openPostgresConnection(databaseURL string) (*gorm.DB, error) {
	database, err := gorm.Open(postgres.Open(databaseURL), newGORMConfig(os.Stdout))
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	return database, nil
}
