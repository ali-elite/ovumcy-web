package api

import (
	"testing"

	"github.com/terraincognita07/ovumcy/internal/models"
)

func TestBuildPrivacyPageDataGuestOmitsCurrentUser(t *testing.T) {
	t.Parallel()

	data := buildPrivacyPageData(map[string]string{}, "/dashboard", nil)
	if _, exists := data["CurrentUser"]; exists {
		t.Fatalf("did not expect CurrentUser for guest payload")
	}
	if title, ok := data["Title"].(string); !ok || title == "" {
		t.Fatalf("expected non-empty title, got %#v", data["Title"])
	}
}

func TestBuildPrivacyPageDataAuthenticatedIncludesCurrentUser(t *testing.T) {
	t.Parallel()

	user := &models.User{Email: "privacy@example.com"}
	data := buildPrivacyPageData(map[string]string{}, "/dashboard", user)

	currentUser, ok := data["CurrentUser"].(*models.User)
	if !ok || currentUser != user {
		t.Fatalf("expected CurrentUser pointer to be preserved, got %#v", data["CurrentUser"])
	}
}
