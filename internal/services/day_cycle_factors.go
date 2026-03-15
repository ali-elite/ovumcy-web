package services

import (
	"strings"

	"github.com/terraincognita07/ovumcy/internal/models"
)

var supportedDayCycleFactorKeys = []string{
	models.CycleFactorStress,
	models.CycleFactorIllness,
	models.CycleFactorTravel,
	models.CycleFactorSleepDisruption,
	models.CycleFactorMedicationChange,
}

func SupportedDayCycleFactorKeys() []string {
	keys := make([]string, len(supportedDayCycleFactorKeys))
	copy(keys, supportedDayCycleFactorKeys)
	return keys
}

func DayCycleFactorTranslationKey(key string) string {
	switch normalizeDayCycleFactorKey(key) {
	case models.CycleFactorStress:
		return "dashboard.cycle_factors.stress"
	case models.CycleFactorIllness:
		return "dashboard.cycle_factors.illness"
	case models.CycleFactorTravel:
		return "dashboard.cycle_factors.travel"
	case models.CycleFactorSleepDisruption:
		return "dashboard.cycle_factors.sleep_disruption"
	case models.CycleFactorMedicationChange:
		return "dashboard.cycle_factors.medication_change"
	default:
		return ""
	}
}

func DayCycleFactorIcon(key string) string {
	switch normalizeDayCycleFactorKey(key) {
	case models.CycleFactorStress:
		return "⚡"
	case models.CycleFactorIllness:
		return "🤒"
	case models.CycleFactorTravel:
		return "✈️"
	case models.CycleFactorSleepDisruption:
		return "🌙"
	case models.CycleFactorMedicationChange:
		return "💊"
	default:
		return "•"
	}
}

func NormalizeDayCycleFactorKeys(keys []string) ([]string, bool) {
	selected := make(map[string]struct{}, len(keys))
	allValid := true

	for _, raw := range keys {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			continue
		}

		normalized := normalizeDayCycleFactorKey(trimmed)
		if normalized == "" {
			allValid = false
			continue
		}
		selected[normalized] = struct{}{}
	}

	result := make([]string, 0, len(selected))
	for _, key := range supportedDayCycleFactorKeys {
		if _, ok := selected[key]; ok {
			result = append(result, key)
		}
	}

	return result, allValid
}

func DayCycleFactorKeySet(keys []string) map[string]bool {
	normalized, _ := NormalizeDayCycleFactorKeys(keys)
	set := make(map[string]bool, len(normalized))
	for _, key := range normalized {
		set[key] = true
	}
	return set
}

func normalizeDayCycleFactorKey(key string) string {
	switch strings.ToLower(strings.TrimSpace(key)) {
	case models.CycleFactorStress:
		return models.CycleFactorStress
	case models.CycleFactorIllness:
		return models.CycleFactorIllness
	case models.CycleFactorTravel:
		return models.CycleFactorTravel
	case models.CycleFactorSleepDisruption:
		return models.CycleFactorSleepDisruption
	case models.CycleFactorMedicationChange:
		return models.CycleFactorMedicationChange
	default:
		return ""
	}
}
