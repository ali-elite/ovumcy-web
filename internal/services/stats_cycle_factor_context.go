package services

import (
	"sort"
	"time"

	"github.com/terraincognita07/ovumcy/internal/models"
)

const (
	statsCycleFactorContextWindowDays = 90
	statsCycleFactorContextLimit      = 3
)

type StatsCycleFactorContextItem struct {
	Key   string
	Count int
}

func StatsCycleFactorContextWindowDays() int {
	return statsCycleFactorContextWindowDays
}

func buildStatsCycleFactorContext(user *models.User, logs []models.DailyLog, stats CycleStats, flags StatsFlags, now time.Time, location *time.Location) ([]StatsCycleFactorContextItem, bool) {
	if !IsOwnerUser(user) || len(logs) == 0 || isStatsPredictionDisabled(user) || flags.CompletedCycleCount < statsMinimumInsightsCycles {
		return nil, false
	}

	variablePattern := user.IrregularCycle || (flags.CompletedCycleCount >= minimumPhaseInsightCycles && IsIrregularCycleSpread(stats))
	if !variablePattern {
		return nil, false
	}

	today := DateAtLocation(now, location)
	windowStart := today.AddDate(0, 0, -(statsCycleFactorContextWindowDays - 1))
	counts := make(map[string]int, len(supportedDayCycleFactorKeys))
	orderIndex := make(map[string]int, len(supportedDayCycleFactorKeys))
	for index, key := range supportedDayCycleFactorKeys {
		orderIndex[key] = index
	}

	for _, logEntry := range logs {
		logDay := DateAtLocation(logEntry.Date, location)
		if logDay.Before(windowStart) || logDay.After(today) {
			continue
		}

		keys, _ := NormalizeDayCycleFactorKeys(logEntry.CycleFactorKeys)
		for _, key := range keys {
			counts[key]++
		}
	}

	items := make([]StatsCycleFactorContextItem, 0, len(counts))
	for key, count := range counts {
		if count <= 0 {
			continue
		}
		items = append(items, StatsCycleFactorContextItem{Key: key, Count: count})
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].Count == items[j].Count {
			return orderIndex[items[i].Key] < orderIndex[items[j].Key]
		}
		return items[i].Count > items[j].Count
	})

	if len(items) > statsCycleFactorContextLimit {
		items = items[:statsCycleFactorContextLimit]
	}

	return items, len(items) > 0
}
