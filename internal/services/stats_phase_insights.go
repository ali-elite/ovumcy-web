package services

import (
	"sort"
	"time"

	"github.com/terraincognita07/ovumcy/internal/models"
)

const minimumPhaseInsightCycles = 3

type StatsPhaseMoodInsight struct {
	Phase       string
	AverageMood float64
	Percentage  float64
	EntryCount  int
	HasData     bool
}

type StatsPhaseSymptomInsightItem struct {
	Name       string
	Icon       string
	Count      int
	TotalDays  int
	Percentage float64
}

type StatsPhaseSymptomInsight struct {
	Phase     string
	TotalDays int
	Items     []StatsPhaseSymptomInsightItem
	HasData   bool
}

type completedCyclePhaseContext struct {
	Start        time.Time
	NextStart    time.Time
	CycleLength  int
	PeriodLength int
	OvulationDay int
}

func buildCompletedCyclePhaseContexts(logs []models.DailyLog, location *time.Location) []completedCyclePhaseContext {
	starts := DetectCycleStarts(logs)
	if len(starts) < 2 {
		return nil
	}

	sorted := sortDailyLogs(logs)
	cycles := buildCycles(starts, sorted)
	contexts := make([]completedCyclePhaseContext, 0, len(starts)-1)
	for index := 0; index+1 < len(starts) && index < len(cycles); index++ {
		start := DateAtLocation(starts[index], location)
		nextStart := DateAtLocation(starts[index+1], location)
		cycleLength := int(nextStart.Sub(start).Hours() / 24)
		if cycleLength <= 0 {
			continue
		}

		periodLength := cycles[index].PeriodLength
		if periodLength <= 0 {
			periodLength = models.DefaultPeriodLength
		}

		ovulationDay, _ := CalcOvulationDay(cycleLength, periodLength)
		if ovulationDay <= 0 {
			continue
		}

		contexts = append(contexts, completedCyclePhaseContext{
			Start:        start,
			NextStart:    nextStart,
			CycleLength:  cycleLength,
			PeriodLength: periodLength,
			OvulationDay: ovulationDay,
		})
	}

	return contexts
}

func phaseForCompletedCycleDay(day time.Time, cycle completedCyclePhaseContext, location *time.Location) string {
	localDay := DateAtLocation(day, location)
	if localDay.Before(cycle.Start) || !localDay.Before(cycle.NextStart) {
		return ""
	}

	dayNumber := int(localDay.Sub(cycle.Start).Hours()/24) + 1
	switch {
	case dayNumber <= cycle.PeriodLength:
		return "menstrual"
	case dayNumber == cycle.OvulationDay:
		return "ovulation"
	case dayNumber < cycle.OvulationDay:
		return "follicular"
	default:
		return "luteal"
	}
}

func findCompletedCycleForDay(day time.Time, cycles []completedCyclePhaseContext, location *time.Location) (completedCyclePhaseContext, bool) {
	localDay := DateAtLocation(day, location)
	for _, cycle := range cycles {
		if !localDay.Before(cycle.Start) && localDay.Before(cycle.NextStart) {
			return cycle, true
		}
	}
	return completedCyclePhaseContext{}, false
}

func (service *StatsService) BuildPhaseMoodInsights(user *models.User, logs []models.DailyLog, location *time.Location) ([]StatsPhaseMoodInsight, bool) {
	if !IsOwnerUser(user) {
		return nil, false
	}

	cycles := buildCompletedCyclePhaseContexts(logs, location)
	if len(cycles) < minimumPhaseInsightCycles {
		return nil, false
	}

	type phaseTotals struct {
		total int
		count int
	}

	totals := map[string]phaseTotals{
		"menstrual":  {},
		"follicular": {},
		"ovulation":  {},
		"luteal":     {},
	}

	for _, logEntry := range logs {
		if logEntry.Mood < MinDayMood || logEntry.Mood > MaxDayMood {
			continue
		}
		cycle, ok := findCompletedCycleForDay(logEntry.Date, cycles, location)
		if !ok {
			continue
		}
		phase := phaseForCompletedCycleDay(logEntry.Date, cycle, location)
		if phase == "" {
			continue
		}
		current := totals[phase]
		current.total += logEntry.Mood
		current.count++
		totals[phase] = current
	}

	insights := make([]StatsPhaseMoodInsight, 0, 4)
	hasData := false
	for _, phase := range []string{"menstrual", "follicular", "ovulation", "luteal"} {
		current := totals[phase]
		insight := StatsPhaseMoodInsight{Phase: phase, EntryCount: current.count}
		if current.count > 0 {
			insight.HasData = true
			insight.AverageMood = float64(current.total) / float64(current.count)
			insight.Percentage = insight.AverageMood * 20
			hasData = true
		}
		insights = append(insights, insight)
	}

	return insights, hasData
}

func (service *StatsService) BuildPhaseSymptomInsights(user *models.User, logs []models.DailyLog, location *time.Location) ([]StatsPhaseSymptomInsight, bool, error) {
	if !IsOwnerUser(user) {
		return nil, false, nil
	}
	if service == nil || service.symptoms == nil {
		return nil, false, nil
	}

	cycles := buildCompletedCyclePhaseContexts(logs, location)
	if len(cycles) < minimumPhaseInsightCycles {
		return nil, false, nil
	}

	symptoms, err := service.symptoms.FetchSymptoms(user.ID)
	if err != nil {
		return nil, false, err
	}

	symptomByID := make(map[uint]models.SymptomType, len(symptoms))
	for _, symptom := range symptoms {
		symptomByID[symptom.ID] = symptom
	}

	type phaseCounter struct {
		totalDays int
		counts    map[uint]int
	}

	counters := map[string]phaseCounter{
		"menstrual":  {counts: make(map[uint]int)},
		"follicular": {counts: make(map[uint]int)},
		"ovulation":  {counts: make(map[uint]int)},
		"luteal":     {counts: make(map[uint]int)},
	}

	for _, logEntry := range logs {
		cycle, ok := findCompletedCycleForDay(logEntry.Date, cycles, location)
		if !ok {
			continue
		}
		phase := phaseForCompletedCycleDay(logEntry.Date, cycle, location)
		if phase == "" {
			continue
		}

		current := counters[phase]
		current.totalDays++
		for _, symptomID := range logEntry.SymptomIDs {
			if _, exists := symptomByID[symptomID]; !exists {
				continue
			}
			current.counts[symptomID]++
		}
		counters[phase] = current
	}

	insights := make([]StatsPhaseSymptomInsight, 0, 4)
	hasData := false
	for _, phase := range []string{"menstrual", "follicular", "ovulation", "luteal"} {
		current := counters[phase]
		insight := StatsPhaseSymptomInsight{
			Phase:     phase,
			TotalDays: current.totalDays,
		}
		if current.totalDays > 0 && len(current.counts) > 0 {
			items := make([]StatsPhaseSymptomInsightItem, 0, len(current.counts))
			for symptomID, count := range current.counts {
				symptom := symptomByID[symptomID]
				items = append(items, StatsPhaseSymptomInsightItem{
					Name:       symptom.Name,
					Icon:       symptom.Icon,
					Count:      count,
					TotalDays:  current.totalDays,
					Percentage: float64(count) * 100 / float64(current.totalDays),
				})
			}
			sort.Slice(items, func(i, j int) bool {
				if items[i].Count == items[j].Count {
					return items[i].Name < items[j].Name
				}
				return items[i].Count > items[j].Count
			})
			if len(items) > 3 {
				items = items[:3]
			}
			insight.Items = items
			insight.HasData = len(items) > 0
			hasData = hasData || insight.HasData
		}
		insights = append(insights, insight)
	}

	return insights, hasData, nil
}
