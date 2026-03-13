package api

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/go-pdf/fpdf"
	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/ovumcy/internal/services"
)

func (handler *Handler) ExportPDF(c *fiber.Ctx) error {
	user, from, to, spec := handler.exportUserAndRange(c)
	if spec != nil {
		return handler.respondMappedError(c, *spec)
	}

	location := handler.requestLocation(c)
	now := time.Now().In(location)
	report, err := handler.exportService.BuildPDFReport(user.ID, from, to, now, location)
	if err != nil {
		spec := exportFetchLogsErrorSpec()
		handler.logSecurityError(c, "data.export", spec, securityEventField("export_format", "pdf"))
		return handler.respondMappedError(c, spec)
	}

	document, err := buildExportPDFDocument(report, currentMessages(c))
	if err != nil {
		spec := exportBuildErrorSpec()
		handler.logSecurityError(c, "data.export", spec, securityEventField("export_format", "pdf"))
		return handler.respondMappedError(c, spec)
	}

	setExportAttachmentHeaders(c, "application/pdf", buildExportFilename(now, "pdf"))
	handler.logSecurityEvent(c, "data.export", "success", securityEventField("export_format", "pdf"))
	return c.Send(document)
}

func buildExportPDFDocument(report services.ExportPDFReport, messages map[string]string) ([]byte, error) {
	pdf := fpdf.New("L", "mm", "A4", "")
	if err := configureExportPDFFonts(pdf); err != nil {
		return nil, err
	}
	pdf.SetTitle(exportPDFText(messages, "export.pdf.report_title", "Ovumcy report for doctor"), false)
	pdf.SetAuthor("Ovumcy", false)
	pdf.SetMargins(10, 10, 10)
	pdf.SetAutoPageBreak(true, 10)
	pdf.AddPage()

	pdf.SetFont(exportPDFFontFamily, "B", 16)
	pdf.CellFormat(0, 10, exportPDFText(messages, "export.pdf.report_title", "Ovumcy report for doctor"), "", 1, "", false, 0, "")
	pdf.SetFont(exportPDFFontFamily, "", 9)
	pdf.CellFormat(0, 6, fmt.Sprintf("%s: %s", exportPDFText(messages, "export.pdf.generated_at", "Generated"), report.GeneratedAt), "", 1, "", false, 0, "")
	pdf.Ln(2)

	pdf.SetFont(exportPDFFontFamily, "B", 11)
	pdf.CellFormat(0, 7, exportPDFText(messages, "export.pdf.summary", "Summary"), "", 1, "", false, 0, "")
	pdf.SetFont(exportPDFFontFamily, "", 9)
	summaryLines := []string{
		fmt.Sprintf("%s: %d", exportPDFText(messages, "export.pdf.logged_days", "Logged days"), report.Summary.LoggedDays),
		fmt.Sprintf("%s: %d", exportPDFText(messages, "export.pdf.completed_cycles", "Completed cycles"), report.Summary.CompletedCycles),
	}
	if report.Summary.AverageCycleLength > 0 {
		summaryLines = append(summaryLines, fmt.Sprintf("%s: %.1f", exportPDFText(messages, "export.pdf.average_cycle", "Average cycle length"), report.Summary.AverageCycleLength))
	}
	if report.Summary.AveragePeriodLength > 0 {
		summaryLines = append(summaryLines, fmt.Sprintf("%s: %.1f", exportPDFText(messages, "export.pdf.average_period", "Average period length"), report.Summary.AveragePeriodLength))
	}
	if report.Summary.HasAverageMood {
		summaryLines = append(summaryLines, fmt.Sprintf("%s: %.1f / 5", exportPDFText(messages, "export.pdf.average_mood", "Average mood"), report.Summary.AverageMood))
	}
	if strings.TrimSpace(report.Summary.RangeStart) != "" && strings.TrimSpace(report.Summary.RangeEnd) != "" {
		summaryLines = append(summaryLines, fmt.Sprintf("%s: %s - %s", exportPDFText(messages, "export.pdf.range", "Range"), report.Summary.RangeStart, report.Summary.RangeEnd))
	}
	for _, line := range summaryLines {
		pdf.CellFormat(0, 5.5, line, "", 1, "", false, 0, "")
	}

	renderExportPDFCalendar(pdf, report, messages)

	if len(report.Cycles) == 0 {
		pdf.Ln(3)
		pdf.MultiCell(0, 5.5, exportPDFText(messages, "export.pdf.no_cycles", "Not enough completed cycles to build a doctor-focused report yet."), "", "L", false)
		var output bytes.Buffer
		if err := pdf.Output(&output); err != nil {
			return nil, err
		}
		return output.Bytes(), nil
	}

	headers := []string{
		exportPDFText(messages, "export.pdf.column.date", "Date"),
		exportPDFText(messages, "export.pdf.column.cycle_day", "Cycle day"),
		exportPDFText(messages, "dashboard.period_day", "Period day"),
		exportPDFText(messages, "dashboard.flow", "Flow"),
		exportPDFText(messages, "dashboard.mood", "Mood"),
		exportPDFText(messages, "dashboard.sex", "Sex"),
		exportPDFText(messages, "dashboard.bbt", "BBT"),
		exportPDFText(messages, "dashboard.cervical_mucus", "Cervical mucus"),
		exportPDFText(messages, "dashboard.symptoms", "Symptoms"),
		exportPDFText(messages, "dashboard.notes", "Notes"),
	}
	widths := []float64{22, 18, 16, 18, 16, 24, 18, 28, 52, 65}

	for index, cycle := range report.Cycles {
		pdf.Ln(4)
		pdf.SetFont(exportPDFFontFamily, "B", 11)
		title := fmt.Sprintf(
			"%s %d: %s - %s (%s %d, %s %d)",
			exportPDFText(messages, "export.pdf.cycle_heading", "Cycle"),
			index+1,
			cycle.StartDate,
			cycle.EndDate,
			exportPDFText(messages, "export.pdf.cycle_length_short", "len"),
			cycle.CycleLength,
			exportPDFText(messages, "export.pdf.period_length_short", "period"),
			cycle.PeriodLength,
		)
		pdf.CellFormat(0, 7, title, "", 1, "", false, 0, "")

		pdf.SetFont(exportPDFFontFamily, "B", 8)
		for headerIndex, header := range headers {
			pdf.CellFormat(widths[headerIndex], 7, header, "1", 0, "L", false, 0, "")
		}
		pdf.Ln(-1)

		pdf.SetFont(exportPDFFontFamily, "", 8)
		for _, entry := range cycle.Entries {
			values := []string{
				entry.Date,
				fmt.Sprintf("%d", entry.CycleDay),
				exportPDFBooleanLabel(messages, entry.IsPeriod),
				exportPDFFlowLabel(messages, entry.Flow),
				exportPDFMoodLabel(entry.MoodRating),
				exportPDFSexActivityLabel(messages, entry.SexActivity),
				exportPDFBBTLabel(entry.BBT),
				exportPDFCervicalMucusLabel(messages, entry.CervicalMucus),
				exportPDFSymptomList(messages, entry.Symptoms),
				entry.Notes,
			}
			for valueIndex, value := range values {
				pdf.CellFormat(widths[valueIndex], 6.5, truncatePDFText(value, widths[valueIndex]), "1", 0, "L", false, 0, "")
			}
			pdf.Ln(-1)
		}
	}

	var output bytes.Buffer
	if err := pdf.Output(&output); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

func renderExportPDFCalendar(pdf *fpdf.Fpdf, report services.ExportPDFReport, messages map[string]string) {
	months := exportPDFCalendarMonths(report.CalendarDays)
	if len(months) == 0 {
		pdf.Ln(3)
		pdf.SetFont(exportPDFFontFamily, "", 9)
		pdf.MultiCell(0, 5.5, exportPDFText(messages, "export.pdf.calendar_none", "No recorded days"), "", "L", false)
		return
	}

	pdf.Ln(4)
	pdf.SetFont(exportPDFFontFamily, "B", 11)
	pdf.CellFormat(0, 7, exportPDFText(messages, "export.pdf.calendar_title", "Color calendar"), "", 1, "", false, 0, "")
	pdf.SetFont(exportPDFFontFamily, "", 8)
	pdf.SetFillColor(199, 117, 109)
	pdf.Rect(pdf.GetX(), pdf.GetY()+1.6, 3.8, 3.8, "F")
	pdf.SetX(pdf.GetX() + 5.4)
	pdf.CellFormat(24, 7, exportPDFText(messages, "export.pdf.calendar_period", "Period"), "", 0, "L", false, 0, "")
	pdf.SetFillColor(232, 196, 168)
	pdf.Rect(pdf.GetX(), pdf.GetY()+1.6, 3.8, 3.8, "F")
	pdf.SetX(pdf.GetX() + 5.4)
	pdf.CellFormat(32, 7, exportPDFText(messages, "export.pdf.calendar_logged", "Logged day"), "", 1, "L", false, 0, "")

	dayMap := make(map[string]services.ExportPDFCalendarDay, len(report.CalendarDays))
	for _, day := range report.CalendarDays {
		dayMap[day.Date] = day
	}

	pageWidth, pageHeight := pdf.GetPageSize()
	leftMargin, _, rightMargin, _ := pdf.GetMargins()
	topY := pdf.GetY()
	usableWidth := pageWidth - leftMargin - rightMargin
	gap := 4.0
	columns := 3
	monthWidth := (usableWidth - gap*float64(columns-1)) / float64(columns)
	monthHeight := 51.0
	requiredRows := float64((len(months) + columns - 1) / columns)
	requiredHeight := requiredRows*monthHeight + gap*(requiredRows-1)
	if topY+requiredHeight > pageHeight-12 {
		pdf.AddPage()
		topY = pdf.GetY()
	}

	for index, monthStart := range months {
		column := index % columns
		row := index / columns
		x := leftMargin + float64(column)*(monthWidth+gap)
		y := topY + float64(row)*(monthHeight+gap)
		drawExportPDFMonth(pdf, monthStart, x, y, monthWidth, monthHeight, dayMap, messages)
	}
}

func exportPDFCalendarMonths(days []services.ExportPDFCalendarDay) []time.Time {
	if len(days) == 0 {
		return nil
	}

	monthSet := make(map[string]time.Time)
	for _, day := range days {
		parsed, ok := parseExportPDFDate(day.Date)
		if !ok {
			continue
		}
		monthStart := time.Date(parsed.Year(), parsed.Month(), 1, 0, 0, 0, 0, time.UTC)
		monthSet[monthStart.Format("2006-01")] = monthStart
	}

	months := make([]time.Time, 0, len(monthSet))
	for _, month := range monthSet {
		months = append(months, month)
	}
	sort.Slice(months, func(i, j int) bool {
		return months[i].Before(months[j])
	})
	if len(months) > 6 {
		months = months[len(months)-6:]
	}
	return months
}

func drawExportPDFMonth(pdf *fpdf.Fpdf, monthStart time.Time, x float64, y float64, width float64, height float64, dayMap map[string]services.ExportPDFCalendarDay, messages map[string]string) {
	pdf.SetDrawColor(216, 204, 190)
	pdf.SetFillColor(255, 255, 255)
	pdf.RoundedRect(x, y, width, height, 2.2, "1234", "DF")

	pdf.SetXY(x+2.2, y+2.4)
	pdf.SetFont(exportPDFFontFamily, "B", 9)
	pdf.CellFormat(width-4.4, 5, monthStart.Format("01/2006"), "", 1, "L", false, 0, "")

	weekdayLabels := []string{
		exportPDFWeekdayLabel(messages, "calendar.weekday.sun", "Sun"),
		exportPDFWeekdayLabel(messages, "calendar.weekday.mon", "Mon"),
		exportPDFWeekdayLabel(messages, "calendar.weekday.tue", "Tue"),
		exportPDFWeekdayLabel(messages, "calendar.weekday.wed", "Wed"),
		exportPDFWeekdayLabel(messages, "calendar.weekday.thu", "Thu"),
		exportPDFWeekdayLabel(messages, "calendar.weekday.fri", "Fri"),
		exportPDFWeekdayLabel(messages, "calendar.weekday.sat", "Sat"),
	}
	cellWidth := (width - 4.4) / 7
	cellHeight := 6.0
	headerY := y + 9

	pdf.SetFont(exportPDFFontFamily, "", 6.5)
	pdf.SetTextColor(120, 102, 84)
	for index, label := range weekdayLabels {
		pdf.SetXY(x+2.2+float64(index)*cellWidth, headerY)
		pdf.CellFormat(cellWidth, 3.6, label, "", 0, "C", false, 0, "")
	}

	gridStart := monthStart.AddDate(0, 0, -int(monthStart.Weekday()))
	for row := 0; row < 6; row++ {
		for column := 0; column < 7; column++ {
			currentDay := gridStart.AddDate(0, 0, row*7+column)
			key := currentDay.Format("2006-01-02")
			cellX := x + 2.2 + float64(column)*cellWidth
			cellY := headerY + 4.5 + float64(row)*cellHeight
			entry, hasEntry := dayMap[key]

			pdf.SetDrawColor(228, 220, 210)
			pdf.SetFillColor(255, 255, 255)
			if hasEntry && entry.IsPeriod {
				pdf.SetFillColor(199, 117, 109)
			} else if hasEntry && entry.HasData {
				pdf.SetFillColor(232, 196, 168)
			}
			fillMode := "D"
			if hasEntry && (entry.IsPeriod || entry.HasData) {
				fillMode = "DF"
			}
			pdf.Rect(cellX, cellY, cellWidth, cellHeight-0.6, fillMode)

			if currentDay.Month() != monthStart.Month() {
				pdf.SetTextColor(190, 182, 174)
			} else if hasEntry && entry.IsPeriod {
				pdf.SetTextColor(255, 255, 255)
			} else {
				pdf.SetTextColor(88, 74, 58)
			}

			pdf.SetXY(cellX, cellY+1)
			pdf.CellFormat(cellWidth, 3.2, fmt.Sprintf("%d", currentDay.Day()), "", 0, "C", false, 0, "")
		}
	}

	pdf.SetTextColor(88, 74, 58)
}

func exportPDFWeekdayLabel(messages map[string]string, key string, fallback string) string {
	label := exportPDFText(messages, key, fallback)
	runes := []rune(strings.TrimSpace(label))
	if len(runes) <= 2 {
		return string(runes)
	}
	return string(runes[:2])
}

func parseExportPDFDate(value string) (time.Time, bool) {
	parsed, err := time.Parse("2006-01-02", strings.TrimSpace(value))
	if err != nil {
		return time.Time{}, false
	}
	return parsed, true
}

func exportPDFText(messages map[string]string, key string, fallback string) string {
	translated := translateMessage(messages, key)
	if translated == "" || translated == key {
		return fallback
	}
	return translated
}

func exportPDFBooleanLabel(messages map[string]string, value bool) string {
	if value {
		return exportPDFText(messages, "common.yes", "Yes")
	}
	return exportPDFText(messages, "common.no", "No")
}

func exportPDFFlowLabel(messages map[string]string, value string) string {
	return exportPDFText(messages, services.FlowTranslationKey(value), strings.TrimSpace(value))
}

func exportPDFMoodLabel(value int) string {
	if value <= 0 {
		return ""
	}
	return fmt.Sprintf("%d/5", value)
}

func exportPDFSexActivityLabel(messages map[string]string, value string) string {
	return exportPDFText(messages, services.SexActivityTranslationKey(value), value)
}

func exportPDFBBTLabel(value float64) string {
	if value <= 0 {
		return ""
	}
	return fmt.Sprintf("%.2f C", value)
}

func exportPDFCervicalMucusLabel(messages map[string]string, value string) string {
	return exportPDFText(messages, services.CervicalMucusTranslationKey(value), value)
}

func exportPDFSymptomList(messages map[string]string, names []string) string {
	if len(names) == 0 {
		return ""
	}

	translated := make([]string, 0, len(names))
	for _, name := range names {
		key := services.BuiltinSymptomTranslationKey(name)
		if key == "" {
			translated = append(translated, name)
			continue
		}
		translated = append(translated, exportPDFText(messages, key, name))
	}
	return strings.Join(translated, ", ")
}

func truncatePDFText(value string, width float64) string {
	limit := int(width * 1.6)
	trimmed := strings.TrimSpace(value)
	runes := []rune(trimmed)
	if len(runes) <= limit {
		return trimmed
	}
	if limit <= 1 {
		return string(runes[:1])
	}
	return string(runes[:limit-1]) + "…"
}
