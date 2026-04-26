package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/ovumcy/ovumcy-web/internal/models"
	"github.com/ovumcy/ovumcy-web/internal/services"
)

func (handler *Handler) ShowDashboard(c *fiber.Ctx) error {
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
		return handler.respondMappedError(c, mapDashboardViewError(err))
	}

	language, messages, now := handler.currentPageViewContext(c)
	location := handler.requestLocation(c)
	data, err := handler.buildDashboardViewData(&subject, language, messages, now, location)
	if err != nil {
		return handler.respondMappedError(c, mapDashboardViewError(err))
	}
	handler.applyPartnerDashboardViewState(viewer, &subject, hasPartnerSubject, data)

	return handler.render(c, "dashboard", data)
}

func (handler *Handler) dashboardSubjectUser(viewer *models.User) (models.User, bool, error) {
	if !services.IsPartnerUser(viewer) {
		return *viewer, false, nil
	}

	link, found, err := handler.partnerLinks.FindActiveOwnerForPartner(viewer.ID)
	if err != nil {
		return models.User{}, false, err
	}
	if !found {
		return *viewer, false, nil
	}

	owner, err := handler.authService.FindByID(link.OwnerUserID)
	if err != nil {
		return models.User{}, false, err
	}
	return owner, true, nil
}

func (handler *Handler) applyPartnerDashboardViewState(viewer *models.User, subject *models.User, hasPartnerSubject bool, data fiber.Map) {
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
	data["ShowMissedDaysLink"] = false
	data["ShowHighFertilityBadge"] = false

	ownerName := subject.DisplayName
	if ownerName == "" {
		ownerName = services.EmailLocalPart(subject.Email)
	}
	data["OwnerDisplayName"] = ownerName

	if todayLog, ok := data["TodayEntry"].(models.DailyLog); ok {
		sanitized := services.SanitizeLogForViewer(viewer, todayLog)
		data["TodayEntry"] = sanitized
		data["TodayLog"] = sanitized
		data["TodayHasData"] = services.DayHasData(sanitized)
	}
}
