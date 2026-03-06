package db

import (
	"io"
	"log"
	"os"
	"time"

	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func newGORMConfig(output io.Writer) *gorm.Config {
	return &gorm.Config{
		Logger:         newDatabaseLogger(output),
		TranslateError: true,
	}
}

func newDatabaseLogger(output io.Writer) gormlogger.Interface {
	if output == nil {
		output = os.Stdout
	}

	return gormlogger.New(
		log.New(output, "\r\n", log.LstdFlags),
		gormlogger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  gormlogger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
			ParameterizedQueries:      true,
		},
	)
}
