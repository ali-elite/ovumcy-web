package services

import (
	"errors"
	"testing"

	"github.com/ovumcy/ovumcy-web/internal/models"
)

func TestValidateSupportedWebUserAllowsOwnerAndPartner(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		user *models.User
	}{
		{name: "owner", user: &models.User{Role: models.RoleOwner}},
		{name: "partner", user: &models.User{Role: models.RolePartner}},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if err := ValidateSupportedWebUser(testCase.user); err != nil {
				t.Fatalf("ValidateSupportedWebUser(%s) returned error: %v", testCase.name, err)
			}
		})
	}
}

func TestValidateSupportedWebUserRejectsLegacyRoles(t *testing.T) {
	t.Parallel()

	if err := ValidateSupportedWebUser(&models.User{Role: "legacy_viewer"}); !errors.Is(err, ErrAuthUnsupportedRole) {
		t.Fatalf("expected ErrAuthUnsupportedRole, got %v", err)
	}
}
