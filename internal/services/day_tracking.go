package services

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/terraincognita07/ovumcy/internal/models"
)

const (
	MinDayBBT = 34.0
	MaxDayBBT = 43.0
)

func NormalizeDaySexActivity(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case models.SexActivityProtected:
		return models.SexActivityProtected
	case models.SexActivityUnprotected:
		return models.SexActivityUnprotected
	default:
		return models.SexActivityNone
	}
}

func IsValidDaySexActivity(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", models.SexActivityNone, models.SexActivityProtected, models.SexActivityUnprotected:
		return true
	default:
		return false
	}
}

func NormalizeDayCervicalMucus(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case models.CervicalMucusDry:
		return models.CervicalMucusDry
	case models.CervicalMucusMoist:
		return models.CervicalMucusMoist
	case models.CervicalMucusCreamy:
		return models.CervicalMucusCreamy
	case models.CervicalMucusEggWhite:
		return models.CervicalMucusEggWhite
	default:
		return models.CervicalMucusNone
	}
}

func IsValidDayCervicalMucus(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", models.CervicalMucusNone, models.CervicalMucusDry, models.CervicalMucusMoist, models.CervicalMucusCreamy, models.CervicalMucusEggWhite:
		return true
	default:
		return false
	}
}

func IsValidDayBBT(value float64) bool {
	return value == 0 || (value >= MinDayBBT && value <= MaxDayBBT)
}

func ParseDayBBTRaw(raw string) (float64, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return 0, nil
	}

	normalized := strings.ReplaceAll(trimmed, ",", ".")
	value, err := strconv.ParseFloat(normalized, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid day bbt: %w", err)
	}
	return value, nil
}
