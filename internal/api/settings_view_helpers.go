package api

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/ovumcy/internal/models"
	"github.com/terraincognita07/ovumcy/internal/services"
)

func (handler *Handler) buildSettingsViewData(c *fiber.Ctx, user *models.User, flash FlashPayload) (fiber.Map, error) {
	messages := currentMessages(c)
	language := currentLanguage(c)
	location := handler.requestLocation(c)

	viewData, err := handler.settingsViewService.BuildSettingsPageViewData(
		user,
		language,
		services.SettingsViewInput{
			FlashSuccess: flash.SettingsSuccess,
			FlashError:   flash.SettingsError,
		},
		time.Now().In(location),
		location,
	)
	if err != nil {
		return nil, err
	}

	*user = viewData.CurrentUser

	data := fiber.Map{
		"Title":                  localizedPageTitle(messages, "meta.title.settings", "Ovumcy | Settings"),
		"CurrentUser":            user,
		"ErrorKey":               viewData.ErrorKey,
		"ChangePasswordErrorKey": viewData.ChangePasswordErrorKey,
		"SuccessKey":             viewData.SuccessKey,
		"CycleLength":            viewData.CycleLength,
		"PeriodLength":           viewData.PeriodLength,
		"AutoPeriodFill":         viewData.AutoPeriodFill,
		"LastPeriodStart":        viewData.LastPeriodStart,
		"TodayISO":               viewData.TodayISO,
		"CycleStartMinISO":       viewData.CycleStartMinISO,
	}

	if viewData.HasOwnerExportViewState {
		data["ExportTotalEntries"] = viewData.Export.TotalEntries
		data["HasExportData"] = viewData.Export.HasData
		data["ExportDateFrom"] = viewData.Export.DateFrom
		data["ExportDateTo"] = viewData.Export.DateTo
		data["ExportDateFromDisplay"] = viewData.Export.DateFromDisplay
		data["ExportDateToDisplay"] = viewData.Export.DateToDisplay
	}

	if viewData.HasOwnerSymptomsView {
		data["ActiveCustomSymptoms"] = buildSettingsSymptomRows(viewData.Symptoms.ActiveCustomSymptoms, settingsSymptomRowState{}, func(source string) string {
			return localizedSettingsSymptomStatus(c, source)
		}, func(source string) string {
			return localizedSettingsSymptomError(c, source)
		})
		data["ArchivedCustomSymptoms"] = buildSettingsSymptomRows(viewData.Symptoms.ArchivedCustomSymptoms, settingsSymptomRowState{}, func(source string) string {
			return localizedSettingsSymptomStatus(c, source)
		}, func(source string) string {
			return localizedSettingsSymptomError(c, source)
		})
		data["HasCustomSymptoms"] = viewData.Symptoms.HasCustomSymptoms
		data["HasArchivedSymptoms"] = viewData.Symptoms.HasArchivedSymptoms
		data["SymptomStatusMessage"] = ""
		data["SymptomErrorMessage"] = ""
		data["SymptomDraftName"] = ""
		data["SymptomDraftIcon"] = defaultSymptomDraftIcon("")
		data["SymptomDraftColor"] = defaultSymptomDraftColor("")
		data["SymptomIconOptions"] = buildSettingsSymptomIconOptions("")
		data["SymptomColorOptions"] = buildSettingsSymptomColorOptions("")
	}

	return data, nil
}
