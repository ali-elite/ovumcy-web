package api

import (
	"strings"

	"github.com/terraincognita07/ovumcy/internal/models"
)

type settingsSymptomIconOption struct {
	Value    string
	Selected bool
	IsCustom bool
}

type settingsSymptomColorOption struct {
	Value     string
	ClassName string
	Selected  bool
}

type settingsSymptomRowView struct {
	Symptom        models.SymptomType
	FormName       string
	FormIcon       string
	FormColor      string
	IconOptions    []settingsSymptomIconOption
	ColorOptions   []settingsSymptomColorOption
	ErrorMessage   string
	SuccessMessage string
}

type settingsSymptomRowState struct {
	SymptomID      uint
	SuccessStatus  string
	ErrorMessage   string
	Draft          symptomPayload
	UseDraftValues bool
}

type settingsSymptomSectionState struct {
	SuccessStatus string
	ErrorMessage  string
	Draft         symptomPayload
	Row           settingsSymptomRowState
}

type symptomColorPresetConfig struct {
	Value     string
	ClassName string
}

var settingsSymptomColorPresets = []symptomColorPresetConfig{
	{Value: "#E8799F", ClassName: "symptom-color-preset-rose"},
	{Value: "#D4A574", ClassName: "symptom-color-preset-gold"},
	{Value: "#F59E0B", ClassName: "symptom-color-preset-amber"},
	{Value: "#FB7185", ClassName: "symptom-color-preset-coral"},
	{Value: "#8B5CF6", ClassName: "symptom-color-preset-violet"},
	{Value: "#38BDF8", ClassName: "symptom-color-preset-sky"},
	{Value: "#14B8A6", ClassName: "symptom-color-preset-teal"},
	{Value: "#64748B", ClassName: "symptom-color-preset-slate"},
}

var settingsSymptomIconCatalog = []string{
	"✨",
	"🔥",
	"💧",
	"⚡",
	"🌙",
	"🤕",
	"🌀",
	"🍫",
}

func buildSettingsSymptomRows(symptoms []models.SymptomType, rowState settingsSymptomRowState, statusLocalizer func(string) string, errorLocalizer func(string) string) []settingsSymptomRowView {
	rows := make([]settingsSymptomRowView, 0, len(symptoms))
	for _, symptom := range symptoms {
		useDraft := rowState.SymptomID != 0 && rowState.SymptomID == symptom.ID && rowState.UseDraftValues
		formName := symptom.Name
		formIcon := symptom.Icon
		formColor := symptom.Color
		if useDraft {
			formName = sanitizeDraftName(rowState.Draft.Name)
			formIcon = defaultSymptomDraftIcon(rowState.Draft.Icon)
			formColor = defaultSymptomDraftColor(rowState.Draft.Color)
		}

		row := settingsSymptomRowView{
			Symptom:      symptom,
			FormName:     formName,
			FormIcon:     formIcon,
			FormColor:    formColor,
			IconOptions:  buildSettingsSymptomIconOptions(formIcon),
			ColorOptions: buildSettingsSymptomColorOptions(formColor),
		}

		if rowState.SymptomID == symptom.ID {
			row.ErrorMessage = errorLocalizer(rowState.ErrorMessage)
			row.SuccessMessage = statusLocalizer(rowState.SuccessStatus)
		}

		rows = append(rows, row)
	}

	return rows
}

func buildSettingsSymptomIconOptions(current string) []settingsSymptomIconOption {
	selected := defaultSymptomDraftIcon(current)
	options := make([]settingsSymptomIconOption, 0, len(settingsSymptomIconCatalog)+1)
	if !settingsSymptomIconInCatalog(selected) {
		options = append(options, settingsSymptomIconOption{
			Value:    selected,
			Selected: true,
			IsCustom: true,
		})
	}
	for _, value := range settingsSymptomIconCatalog {
		options = append(options, settingsSymptomIconOption{
			Value:    value,
			Selected: value == selected,
		})
	}
	return options
}

func buildSettingsSymptomColorOptions(current string) []settingsSymptomColorOption {
	selected := defaultSymptomDraftColor(current)
	options := make([]settingsSymptomColorOption, 0, len(settingsSymptomColorPresets))
	for _, preset := range settingsSymptomColorPresets {
		options = append(options, settingsSymptomColorOption{
			Value:     preset.Value,
			ClassName: preset.ClassName,
			Selected:  strings.EqualFold(preset.Value, selected),
		})
	}
	return options
}

func settingsSymptomIconInCatalog(value string) bool {
	for _, option := range settingsSymptomIconCatalog {
		if option == value {
			return true
		}
	}
	return false
}

func sanitizeDraftName(raw string) string {
	return strings.ToValidUTF8(raw, "")
}
