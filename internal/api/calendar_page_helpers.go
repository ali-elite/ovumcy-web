package api

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/ovumcy/ovumcy-web/internal/models"
	"github.com/ovumcy/ovumcy-web/internal/services"
)

func (handler *Handler) buildCalendarViewData(user *models.User, language string, messages map[string]string, now time.Time, monthStart time.Time, selectedDate string, location *time.Location) (fiber.Map, error) {
	viewData, err := handler.calendarViewService.BuildCalendarPageViewData(user, language, now, monthStart, selectedDate, location)
	if err != nil {
		return nil, err
	}

	days := handler.buildCalendarDays(viewData.DayStates)

	data := fiber.Map{
		"Title":                             localizedPageTitle(messages, "meta.title.calendar", "Ovumcy | Calendar"),
		"CurrentUser":                       user,
		"SubjectUser":                       user,
		"MonthLabel":                        viewData.MonthLabel,
		"MonthValue":                        viewData.MonthValue,
		"PrevMonth":                         viewData.PrevMonth,
		"NextMonth":                         viewData.NextMonth,
		"SelectedDate":                      viewData.SelectedDate,
		"CalendarDays":                      days,
		"Today":                             viewData.TodayISO,
		"Stats":                             viewData.Stats,
		"PredictionExplanationPrimaryKey":   viewData.PredictionExplanationPrimaryKey,
		"PredictionExplanationSecondaryKey": viewData.PredictionExplanationSecondaryKey,
		"HasPredictionExplanationPrimary":   viewData.HasPredictionExplanationPrimary,
		"HasPredictionExplanationSecondary": viewData.HasPredictionExplanationSecondary,
		"IsOwner":                           viewData.IsOwner,
	}
	return data, nil
}

func (handler *Handler) applyPartnerCalendarViewState(viewer *models.User, subject *models.User, hasPartnerSubject bool, data fiber.Map) {
	if !services.IsPartnerUser(viewer) {
		data["SubjectUser"] = subject
		data["IsPartnerView"] = false
		data["HasPartnerSubject"] = false
		return
	}

	data["CurrentUser"] = viewer
	data["SubjectUser"] = subject
	data["IsOwner"] = false
	data["IsPartnerView"] = true
	data["HasPartnerSubject"] = hasPartnerSubject
}
