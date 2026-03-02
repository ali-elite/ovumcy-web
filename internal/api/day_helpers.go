package api

import (
	"time"

	"github.com/terraincognita07/ovumcy/internal/services"
)

func dateAtLocation(value time.Time, location *time.Location) time.Time {
	return services.DateAtLocation(value, location)
}
