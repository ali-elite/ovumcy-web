package api

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/ovumcy/internal/models"
	"github.com/terraincognita07/ovumcy/internal/services"
)

type settingsSymptomSectionState struct {
	SuccessStatus string
	ErrorMessage  string
	Draft         symptomPayload
}

func (handler *Handler) buildSettingsSymptomsSectionData(c *fiber.Ctx, user *models.User, state settingsSymptomSectionState) (fiber.Map, error) {
	viewData, err := handler.settingsViewService.BuildSettingsSymptomsViewData(user)
	if err != nil {
		return nil, err
	}

	return fiber.Map{
		"ActiveCustomSymptoms":   viewData.ActiveCustomSymptoms,
		"ArchivedCustomSymptoms": viewData.ArchivedCustomSymptoms,
		"HasCustomSymptoms":      viewData.HasCustomSymptoms,
		"HasArchivedSymptoms":    viewData.HasArchivedSymptoms,
		"SymptomStatusMessage":   localizedSettingsSymptomStatus(c, state.SuccessStatus),
		"SymptomErrorMessage":    localizedSettingsSymptomError(c, state.ErrorMessage),
		"SymptomDraftName":       strings.TrimSpace(state.Draft.Name),
		"SymptomDraftIcon":       defaultSymptomDraftIcon(state.Draft.Icon),
		"SymptomDraftColor":      defaultSymptomDraftColor(state.Draft.Color),
	}, nil
}

func (handler *Handler) respondSymptomMutationError(c *fiber.Ctx, user *models.User, spec APIErrorSpec, state settingsSymptomSectionState) error {
	if isHTMX(c) {
		state.ErrorMessage = spec.Key
		data, err := handler.buildSettingsSymptomsSectionData(c, user, state)
		if err != nil {
			return handler.respondMappedError(c, settingsLoadErrorSpec())
		}
		c.Status(fiber.StatusOK)
		return handler.renderPartial(c, "settings_symptoms_section", data)
	}

	if !acceptsJSON(c) {
		handler.setFlashCookie(c, FlashPayload{SettingsError: spec.Key})
		return c.Redirect("/settings", fiber.StatusSeeOther)
	}
	return apiError(c, spec.Status, spec.Key)
}

func (handler *Handler) respondSymptomMutationSuccess(c *fiber.Ctx, user *models.User, statusCode int, successStatus string) error {
	if isHTMX(c) {
		data, err := handler.buildSettingsSymptomsSectionData(c, user, settingsSymptomSectionState{
			SuccessStatus: successStatus,
		})
		if err != nil {
			return handler.respondMappedError(c, settingsLoadErrorSpec())
		}
		c.Status(fiber.StatusOK)
		return handler.renderPartial(c, "settings_symptoms_section", data)
	}

	if !acceptsJSON(c) {
		handler.setFlashCookie(c, FlashPayload{SettingsSuccess: successStatus})
		return c.Redirect("/settings", fiber.StatusSeeOther)
	}

	return c.SendStatus(statusCode)
}

func localizedSettingsSymptomError(c *fiber.Ctx, source string) string {
	source = strings.TrimSpace(source)
	if source == "" {
		return ""
	}

	messages := currentMessages(c)
	if key := services.AuthErrorTranslationKey(source); key != "" {
		if localized := translateMessage(messages, key); localized != key {
			return localized
		}
	}
	return source
}

func localizedSettingsSymptomStatus(c *fiber.Ctx, status string) string {
	status = strings.TrimSpace(status)
	if status == "" {
		return ""
	}

	messages := currentMessages(c)
	if key := services.SettingsStatusTranslationKey(status); key != "" {
		if localized := translateMessage(messages, key); localized != key {
			return localized
		}
	}
	return status
}

func defaultSymptomDraftIcon(raw string) string {
	icon := strings.TrimSpace(raw)
	if icon == "" {
		return "✨"
	}
	return icon
}

func defaultSymptomDraftColor(raw string) string {
	color := strings.TrimSpace(raw)
	if color == "" {
		return "#E8799F"
	}
	return color
}
