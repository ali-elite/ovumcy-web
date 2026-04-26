package api

import (
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/ovumcy/ovumcy-web/internal/models"
	"github.com/ovumcy/ovumcy-web/internal/services"
)

func (handler *Handler) HandlePartnerAdvice(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok || user.Role != models.RolePartner {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	language := currentLanguage(c)
	if language == "" {
		language = handler.i18n.DefaultLanguage()
	}
	skipCache := c.Query("fresh") == "true"

	adviceContext, err := handler.buildPartnerAdviceContext(user, currentMessages(c))
	if err != nil {
		log.Printf("partner advice context unavailable: partner_id=%d err=%v", user.ID, err)
		adviceContext = services.PartnerAdviceContext{Phase: c.Query("phase", "unknown")}
	}

	advice, err := handler.partnerAdviceSvc.GetAdvice(c.Context(), adviceContext, language, skipCache)
	if err != nil {
		log.Printf("partner advice unavailable: phase=%q language=%q err=%v", adviceContext.Phase, language, err)
		return handler.renderPartnerAdviceUnavailable(c, adviceContext.Phase)
	}

	messages := currentMessages(c)

	return handler.renderPartial(c, "partner_advice_partial", fiber.Map{
		"Advice":   advice,
		"Phase":    adviceContext.Phase,
		"Messages": messages,
	})
}

func (handler *Handler) buildPartnerAdviceContext(partner *models.User, messages map[string]string) (services.PartnerAdviceContext, error) {
	link, found, err := handler.partnerLinks.FindActiveOwnerForPartner(partner.ID)
	if err != nil {
		return services.PartnerAdviceContext{}, err
	}
	if !found {
		return services.PartnerAdviceContext{Phase: "unknown"}, nil
	}

	owner, err := handler.authService.FindByID(link.OwnerUserID)
	if err != nil {
		return services.PartnerAdviceContext{}, err
	}

	now := time.Now().In(handler.location)
	today := services.DateAtLocation(now, handler.location)
	stats, _, err := handler.statsService.BuildCycleStatsForRange(&owner, today.AddDate(-2, 0, 0), today, now, handler.location)
	if err != nil {
		return services.PartnerAdviceContext{}, err
	}
	todayLog, symptoms, err := handler.viewerService.FetchDayLogForViewer(&owner, today, handler.location)
	if err != nil {
		return services.PartnerAdviceContext{}, err
	}
	todayLog = services.SanitizePartnerViewerLog(todayLog)

	return services.PartnerAdviceContext{
		Phase:                 stats.CurrentPhase,
		CycleDay:              stats.CurrentCycleDay,
		MedianCycleLength:     stats.MedianCycleLength,
		CompletedCycleCount:   stats.CompletedCycleCount,
		AveragePeriodLength:   stats.AveragePeriodLength,
		LastPeriodLength:      stats.LastPeriodLength,
		NextPeriodEstimate:    adviceDateRange(stats.NextPeriodStart, stats.NextPeriodStart, handler.location),
		FertilityWindow:       adviceDateRange(stats.FertilityWindowStart, stats.FertilityWindowEnd, handler.location),
		OvulationEstimate:     adviceDate(stats.OvulationDate, handler.location),
		AgeGroup:              translatedLabel(messages, services.AgeGroupTranslationKey(owner.AgeGroup), owner.AgeGroup),
		UsageGoal:             translatedLabel(messages, services.UsageGoalTranslationKey(owner.UsageGoal), owner.UsageGoal),
		IrregularCycle:        owner.IrregularCycle,
		UnpredictableCycle:    owner.UnpredictableCycle,
		TodayPeriodLogged:     todayLog.IsPeriod,
		TodayFlow:             adviceFlowLabel(messages, todayLog),
		TodayMood:             adviceMoodLabel(todayLog.Mood),
		TodaySymptoms:         adviceSymptomLabels(messages, todayLog, symptoms),
		TodayCervicalMucus:    adviceCervicalMucusLabel(messages, todayLog.CervicalMucus),
		TodayBBT:              adviceBBTLabel(todayLog.BBT, owner.TemperatureUnit),
		TodayCycleFactors:     adviceCycleFactorLabels(messages, todayLog.CycleFactorKeys),
		PartnerSupportContext: "Provide supportive partner guidance without identity details or private notes.",
	}, nil
}

func translatedLabel(messages map[string]string, key string, fallback string) string {
	label := translateMessage(messages, key)
	if label == "" || label == key {
		return fallback
	}
	return label
}

func adviceDate(value time.Time, location *time.Location) string {
	if value.IsZero() {
		return ""
	}
	return services.DateAtLocation(value, location).Format("2006-01-02")
}

func adviceDateRange(start time.Time, end time.Time, location *time.Location) string {
	startLabel := adviceDate(start, location)
	endLabel := adviceDate(end, location)
	switch {
	case startLabel == "":
		return ""
	case endLabel == "" || startLabel == endLabel:
		return startLabel
	default:
		return startLabel + " to " + endLabel
	}
}

func adviceFlowLabel(messages map[string]string, logEntry models.DailyLog) string {
	if !logEntry.IsPeriod && strings.TrimSpace(logEntry.Flow) == models.FlowNone {
		return ""
	}
	return translatedLabel(messages, services.FlowTranslationKey(logEntry.Flow), logEntry.Flow)
}

func adviceMoodLabel(value int) string {
	if value < services.MinDayMood || value > services.MaxDayMood {
		return ""
	}
	return strconv.Itoa(value) + "/5"
}

func adviceSymptomLabels(messages map[string]string, logEntry models.DailyLog, symptoms []models.SymptomType) []string {
	if len(logEntry.SymptomIDs) == 0 || len(symptoms) == 0 {
		return nil
	}
	selected := make(map[uint]struct{}, len(logEntry.SymptomIDs))
	for _, id := range logEntry.SymptomIDs {
		selected[id] = struct{}{}
	}
	labels := make([]string, 0, len(logEntry.SymptomIDs))
	for _, symptom := range symptoms {
		if _, ok := selected[symptom.ID]; !ok {
			continue
		}
		label := translatedLabel(messages, services.BuiltinSymptomTranslationKey(symptom.Name), symptom.Name)
		if strings.TrimSpace(label) != "" {
			labels = append(labels, label)
		}
	}
	return labels
}

func adviceCervicalMucusLabel(messages map[string]string, value string) string {
	normalized := services.NormalizeDayCervicalMucus(value)
	if normalized == models.CervicalMucusNone {
		return ""
	}
	return translatedLabel(messages, services.CervicalMucusTranslationKey(normalized), normalized)
}

func adviceBBTLabel(value float64, unit string) string {
	if !services.IsValidDayBBT(value) || value <= 0 {
		return ""
	}
	formatted := services.FormatDayBBTForInput(value, unit)
	if formatted == "" {
		return ""
	}
	return formatted + " " + strings.ToUpper(services.NormalizeTemperatureUnit(unit))
}

func adviceCycleFactorLabels(messages map[string]string, keys []string) []string {
	labels := make([]string, 0, len(keys))
	for _, key := range keys {
		translationKey := services.DayCycleFactorTranslationKey(key)
		label := translatedLabel(messages, translationKey, key)
		if strings.TrimSpace(label) != "" {
			labels = append(labels, label)
		}
	}
	return labels
}

func (handler *Handler) renderPartnerAdviceUnavailable(c *fiber.Ctx, phase string) error {
	return handler.renderPartial(c, "partner_advice_partial", fiber.Map{
		"AdviceUnavailable": true,
		"Phase":             phase,
		"Messages":          currentMessages(c),
	})
}
