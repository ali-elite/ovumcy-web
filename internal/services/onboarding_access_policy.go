package services

import "strings"

func IsOnboardingPath(path string) bool {
	cleanPath := strings.TrimSpace(path)
	return cleanPath == "/onboarding" || strings.HasPrefix(cleanPath, "/onboarding/")
}

func ShouldEnforceOnboardingAccess(path string) bool {
	cleanPath := strings.TrimSpace(path)
	if cleanPath == "/api/auth/logout" {
		return false
	}
	return !IsOnboardingPath(cleanPath)
}
