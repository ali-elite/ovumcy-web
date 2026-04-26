package api

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/ovumcy/ovumcy-web/internal/services"
)

func (handler *Handler) ShowCalendar(c *fiber.Ctx) error {
	user, handled, err := handler.currentUserOrRedirectToLogin(c)
	if err != nil {
		return err
	}
	if handled {
		return nil
	}

	viewer := user
	subject, hasPartnerSubject, err := handler.dashboardSubjectUser(viewer)
	if err != nil {
		return handler.respondMappedError(c, mapCalendarViewError(err))
	}

	language, messages, now := handler.currentPageViewContext(c)
	location := handler.requestLocation(c)
	minMonth := services.CalendarMinimumNavigableMonth(&subject, location)
	selectedDateQuery := strings.TrimSpace(c.Query("day"))
	if selectedDateQuery == "" {
		selectedDateQuery = strings.TrimSpace(c.Query("selected"))
	}
	activeMonth, selectedDate, err := services.ResolveCalendarMonthAndSelectedDateWithinBounds(c.Query("month"), selectedDateQuery, now, location, minMonth)
	if err != nil {
		if acceptsJSON(c) {
			return handler.respondMappedError(c, invalidMonthErrorSpec())
		}
		return redirectOrJSON(c, "/calendar")
	}

	data, err := handler.buildCalendarViewData(&subject, language, messages, now, activeMonth, selectedDate, location)
	if err != nil {
		return handler.respondMappedError(c, mapCalendarViewError(err))
	}
	handler.applyPartnerCalendarViewState(viewer, &subject, hasPartnerSubject, data)
	data["SelectedDateEditMode"] = services.ParseBoolLike(c.Query("edit"))

	return handler.render(c, "calendar", data)
}
