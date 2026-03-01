package api

import "github.com/terraincognita07/ovumcy/internal/services"

func templateTranslate(messages map[string]string, key string) string {
	return translateMessage(messages, key)
}

func templatePhaseLabel(messages map[string]string, phase string) string {
	return translateMessage(messages, services.PhaseTranslationKey(phase))
}

func templatePhaseIcon(phase string) string {
	return services.PhaseIcon(phase)
}

func templateFlowLabel(messages map[string]string, flow string) string {
	return translateMessage(messages, services.FlowTranslationKey(flow))
}

func templateSymptomLabel(messages map[string]string, name string) string {
	return localizedSymptomName(messages, name)
}

func templateSymptomGroup(name string) string {
	return services.SymptomGroup(name)
}

func templateRoleLabel(messages map[string]string, role string) string {
	return translateMessage(messages, services.RoleTranslationKey(role))
}
