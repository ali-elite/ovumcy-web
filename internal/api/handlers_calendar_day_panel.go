package api

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/ovumcy/ovumcy-web/internal/models"
	"github.com/ovumcy/ovumcy-web/internal/services"
)

func (handler *Handler) CalendarDayPanel(c *fiber.Ctx) error {
	user, handled, err := currentUserOrUnauthorized(c)
	if err != nil {
		return err
	}
	if handled {
		return nil
	}

	location := handler.requestLocation(c)
	day, err := services.ParseDayDate(c.Params("date"), location)
	if err != nil {
		return handler.respondMappedError(c, invalidDateErrorSpec())
	}

	viewer := user
	subject, hasPartnerSubject, err := handler.dashboardSubjectUser(viewer)
	if err != nil {
		return handler.respondMappedError(c, mapDayEditorViewError(err))
	}

	return handler.renderDayEditorPartialForViewer(c, viewer, &subject, hasPartnerSubject, day)
}

func (handler *Handler) renderDayEditorPartial(c *fiber.Ctx, user *models.User, day time.Time) error {
	return handler.renderDayEditorPartialForViewer(c, user, user, false, day)
}

func (handler *Handler) renderDayEditorPartialForViewer(c *fiber.Ctx, viewer *models.User, subject *models.User, hasPartnerSubject bool, day time.Time) error {
	language, messages, now := handler.currentPageViewContext(c)
	location := handler.requestLocation(c)
	payload, err := handler.buildDayEditorPartialData(subject, language, messages, day, now, location, c.Query("mode") == "edit")
	if err != nil {
		return handler.respondMappedError(c, mapDayEditorViewError(err))
	}
	handler.applyPartnerDayEditorViewState(viewer, subject, hasPartnerSubject, payload)
	return handler.renderPartial(c, "day_editor_partial", payload)
}

func (handler *Handler) applyPartnerDayEditorViewState(viewer *models.User, subject *models.User, hasPartnerSubject bool, payload fiber.Map) {
	if !services.IsPartnerUser(viewer) {
		payload["SubjectUser"] = subject
		payload["IsPartnerView"] = false
		payload["HasPartnerSubject"] = false
		return
	}

	payload["SubjectUser"] = subject
	payload["IsPartnerView"] = true
	payload["HasPartnerSubject"] = hasPartnerSubject
	payload["IsOwner"] = false
	payload["EditMode"] = false
	payload["ShowSexChip"] = false
	payload["ShowBBTField"] = false
	payload["ShowCycleFactors"] = false
	payload["ShowNotesField"] = false
	payload["ShowCervicalMucus"] = false
	payload["AllowManualCycleStart"] = false

	if logEntry, ok := payload["Log"].(models.DailyLog); ok {
		sanitized := services.SanitizeLogForViewer(viewer, logEntry)
		payload["Log"] = sanitized
		payload["SelectedSymptomID"] = services.SymptomIDSet(sanitized.SymptomIDs)
		payload["SelectedCycleFactorKey"] = services.DayCycleFactorKeySet(sanitized.CycleFactorKeys)
	}
}
