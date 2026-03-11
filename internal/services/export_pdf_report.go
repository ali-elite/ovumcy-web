package services

import (
	"sort"
	"strings"
	"time"

	"github.com/terraincognita07/ovumcy/internal/models"
)

const maxExportPDFCycles = 6

type ExportPDFCycleDay struct {
	Date          string
	CycleDay      int
	IsPeriod      bool
	Flow          string
	MoodRating    int
	SexActivity   string
	BBT           float64
	CervicalMucus string
	Symptoms      []string
	Notes         string
}

type ExportPDFCycle struct {
	StartDate    string
	EndDate      string
	CycleLength  int
	PeriodLength int
	Entries      []ExportPDFCycleDay
}

type ExportPDFSummary struct {
	LoggedDays          int
	CompletedCycles     int
	AverageCycleLength  float64
	AveragePeriodLength float64
	AverageMood         float64
	HasAverageMood      bool
	RangeStart          string
	RangeEnd            string
}

type ExportPDFReport struct {
	GeneratedAt string
	Summary     ExportPDFSummary
	Cycles      []ExportPDFCycle
}

func (service *ExportService) BuildPDFReport(userID uint, from *time.Time, to *time.Time, now time.Time, location *time.Location) (ExportPDFReport, error) {
	logs, symptomNames, err := service.LoadDataForRange(userID, from, to, location)
	if err != nil {
		return ExportPDFReport{}, err
	}

	report := ExportPDFReport{
		GeneratedAt: now.In(location).Format(time.RFC3339),
		Summary:     buildExportPDFSummary(logs, location),
		Cycles:      []ExportPDFCycle{},
	}
	if len(logs) == 0 {
		return report, nil
	}

	cycles := buildCompletedCyclePhaseContexts(logs, location)
	if len(cycles) > maxExportPDFCycles {
		cycles = cycles[len(cycles)-maxExportPDFCycles:]
	}
	if len(cycles) == 0 {
		return report, nil
	}

	report.Summary.CompletedCycles = len(cycles)
	report.Summary.AverageCycleLength = averageCompletedCycleLength(cycles)
	report.Summary.AveragePeriodLength = averageCompletedPeriodLength(cycles)

	filteredLogs := make([]models.DailyLog, 0, len(logs))
	for _, logEntry := range logs {
		cycle, ok := findCompletedCycleForDay(logEntry.Date, cycles, location)
		if !ok {
			continue
		}
		filteredLogs = append(filteredLogs, logEntry)
		day := ExportPDFCycleDay{
			Date:          DateAtLocation(logEntry.Date, location).Format(exportDateLayout),
			CycleDay:      int(DateAtLocation(logEntry.Date, location).Sub(cycle.Start).Hours()/24) + 1,
			IsPeriod:      logEntry.IsPeriod,
			Flow:          normalizeExportFlow(logEntry.Flow),
			MoodRating:    normalizeExportMood(logEntry.Mood),
			SexActivity:   normalizeExportSexActivity(logEntry.SexActivity),
			BBT:           normalizeExportBBT(logEntry.BBT),
			CervicalMucus: normalizeExportCervicalMucus(logEntry.CervicalMucus),
			Symptoms:      exportPDFSymptoms(logEntry.SymptomIDs, symptomNames),
			Notes:         strings.TrimSpace(logEntry.Notes),
		}

		if len(report.Cycles) == 0 || report.Cycles[len(report.Cycles)-1].StartDate != cycle.Start.Format(exportDateLayout) {
			report.Cycles = append(report.Cycles, ExportPDFCycle{
				StartDate:    cycle.Start.Format(exportDateLayout),
				EndDate:      cycle.NextStart.AddDate(0, 0, -1).Format(exportDateLayout),
				CycleLength:  cycle.CycleLength,
				PeriodLength: cycle.PeriodLength,
				Entries:      []ExportPDFCycleDay{},
			})
		}
		report.Cycles[len(report.Cycles)-1].Entries = append(report.Cycles[len(report.Cycles)-1].Entries, day)
	}

	report.Summary.LoggedDays = len(filteredLogs)
	report.Summary.AverageMood, report.Summary.HasAverageMood = exportPDFAverageMood(filteredLogs)
	if len(filteredLogs) > 0 {
		report.Summary.RangeStart = DateAtLocation(filteredLogs[0].Date, location).Format(exportDateLayout)
		report.Summary.RangeEnd = DateAtLocation(filteredLogs[len(filteredLogs)-1].Date, location).Format(exportDateLayout)
	}

	return report, nil
}

func buildExportPDFSummary(logs []models.DailyLog, location *time.Location) ExportPDFSummary {
	if len(logs) == 0 {
		return ExportPDFSummary{}
	}
	first := DateAtLocation(logs[0].Date, location)
	last := first
	for _, logEntry := range logs[1:] {
		day := DateAtLocation(logEntry.Date, location)
		if day.Before(first) {
			first = day
		}
		if day.After(last) {
			last = day
		}
	}

	return ExportPDFSummary{
		LoggedDays: len(logs),
		RangeStart: first.Format(exportDateLayout),
		RangeEnd:   last.Format(exportDateLayout),
	}
}

func averageCompletedCycleLength(cycles []completedCyclePhaseContext) float64 {
	if len(cycles) == 0 {
		return 0
	}
	total := 0
	for _, cycle := range cycles {
		total += cycle.CycleLength
	}
	return float64(total) / float64(len(cycles))
}

func averageCompletedPeriodLength(cycles []completedCyclePhaseContext) float64 {
	if len(cycles) == 0 {
		return 0
	}
	total := 0
	for _, cycle := range cycles {
		total += cycle.PeriodLength
	}
	return float64(total) / float64(len(cycles))
}

func exportPDFAverageMood(logs []models.DailyLog) (float64, bool) {
	total := 0
	count := 0
	for _, logEntry := range logs {
		if logEntry.Mood < MinDayMood || logEntry.Mood > MaxDayMood {
			continue
		}
		total += logEntry.Mood
		count++
	}
	if count == 0 {
		return 0, false
	}
	return float64(total) / float64(count), true
}

func exportPDFSymptoms(symptomIDs []uint, symptomNames map[uint]string) []string {
	if len(symptomIDs) == 0 {
		return nil
	}

	names := make([]string, 0, len(symptomIDs))
	for _, symptomID := range symptomIDs {
		name := strings.TrimSpace(symptomNames[symptomID])
		if name == "" {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
