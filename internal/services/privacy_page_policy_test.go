package services

import "testing"

func TestResolvePrivacyMetaDescriptionFallback(t *testing.T) {
	t.Parallel()

	if got := ResolvePrivacyMetaDescription(""); got != defaultPrivacyMetaDescription {
		t.Fatalf("unexpected fallback description: %q", got)
	}
	if got := ResolvePrivacyMetaDescription("meta.description.privacy"); got != defaultPrivacyMetaDescription {
		t.Fatalf("expected key fallback description, got %q", got)
	}
}

func TestBuildPrivacyBackNavigationGuestUsesLoginFallback(t *testing.T) {
	t.Parallel()

	navigation := BuildPrivacyBackNavigation("https://evil.example/path", false)
	if navigation.BackPath != "/login" {
		t.Fatalf("expected guest back path /login, got %q", navigation.BackPath)
	}
	if navigation.BreadcrumbBackLabelKey != "common.home" {
		t.Fatalf("expected guest breadcrumb key common.home, got %q", navigation.BreadcrumbBackLabelKey)
	}
}

func TestBuildPrivacyBackNavigationAuthenticatedUsesDashboardFallback(t *testing.T) {
	t.Parallel()

	navigation := BuildPrivacyBackNavigation("https://evil.example/path", true)
	if navigation.BackPath != "/dashboard" {
		t.Fatalf("expected auth back path /dashboard, got %q", navigation.BackPath)
	}
	if navigation.BreadcrumbBackLabelKey != "nav.dashboard" {
		t.Fatalf("expected auth breadcrumb key nav.dashboard, got %q", navigation.BreadcrumbBackLabelKey)
	}
}

func TestBuildPrivacyBackNavigationUsesCalendarBackLabelWhenRequested(t *testing.T) {
	t.Parallel()

	navigation := BuildPrivacyBackNavigation("/calendar?month=2026-03", true)
	if navigation.BackPath != "/calendar?month=2026-03" {
		t.Fatalf("expected sanitized calendar back path, got %q", navigation.BackPath)
	}
	if navigation.BreadcrumbBackLabelKey != "nav.calendar" {
		t.Fatalf("expected nav.calendar breadcrumb key, got %q", navigation.BreadcrumbBackLabelKey)
	}
}
