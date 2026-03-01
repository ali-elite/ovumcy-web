package services

import "strings"

const defaultPrivacyMetaDescription = "Ovumcy Privacy Policy - Zero data collection, self-hosted period tracker."

type PrivacyBackNavigation struct {
	BackPath               string
	BreadcrumbBackLabelKey string
}

func ResolvePrivacyMetaDescription(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" || value == "meta.description.privacy" {
		return defaultPrivacyMetaDescription
	}
	return value
}

func BuildPrivacyBackNavigation(backQuery string, isAuthenticated bool) PrivacyBackNavigation {
	backFallback := "/login"
	breadcrumbBackLabelKey := "common.home"
	if isAuthenticated {
		backFallback = "/dashboard"
		breadcrumbBackLabelKey = "nav.dashboard"
	}

	return PrivacyBackNavigation{
		BackPath:               SanitizeRedirectPath(backQuery, backFallback),
		BreadcrumbBackLabelKey: breadcrumbBackLabelKey,
	}
}
