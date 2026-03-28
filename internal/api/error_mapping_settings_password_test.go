package api

import (
	"errors"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/ovumcy/ovumcy-web/internal/services"
)

func TestMapSettingsPasswordChangeError(t *testing.T) {
	testCases := []struct {
		name string
		err  error
		want APIErrorSpec
	}{
		{
			name: "invalid input",
			err:  services.ErrSettingsPasswordChangeInvalidInput,
			want: settingsFormErrorSpec(fiber.StatusBadRequest, APIErrorCategoryValidation, "invalid settings input"),
		},
		{
			name: "password mismatch",
			err:  services.ErrSettingsPasswordMismatch,
			want: settingsFormErrorSpec(fiber.StatusBadRequest, APIErrorCategoryValidation, "password mismatch"),
		},
		{
			name: "invalid current password",
			err:  services.ErrSettingsInvalidCurrentPassword,
			want: settingsFormErrorSpec(fiber.StatusUnauthorized, APIErrorCategoryUnauthorized, "invalid current password"),
		},
		{
			name: "local password required",
			err:  services.ErrSettingsLocalPasswordNotSet,
			want: settingsLocalPasswordRequiredErrorSpec(),
		},
		{
			name: "new password must differ",
			err:  services.ErrSettingsNewPasswordMustDiffer,
			want: settingsFormErrorSpec(fiber.StatusBadRequest, APIErrorCategoryValidation, "new password must differ"),
		},
		{
			name: "weak password",
			err:  services.ErrSettingsWeakPassword,
			want: settingsFormErrorSpec(fiber.StatusBadRequest, APIErrorCategoryValidation, "weak password"),
		},
		{
			name: "hash failed",
			err:  services.ErrSettingsPasswordHashFailed,
			want: globalErrorSpec(fiber.StatusInternalServerError, APIErrorCategoryInternal, "failed to secure password"),
		},
		{
			name: "recovery code failed",
			err:  services.ErrSettingsRecoveryCodeGenerateFailed,
			want: globalErrorSpec(fiber.StatusInternalServerError, APIErrorCategoryInternal, "failed to secure password"),
		},
		{
			name: "update failed",
			err:  services.ErrSettingsPasswordUpdateFailed,
			want: globalErrorSpec(fiber.StatusInternalServerError, APIErrorCategoryInternal, "failed to update password"),
		},
		{
			name: "unknown",
			err:  errors.New("unknown"),
			want: globalErrorSpec(fiber.StatusInternalServerError, APIErrorCategoryInternal, "failed to update password"),
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			if got := mapSettingsPasswordChangeError(testCase.err); got != testCase.want {
				t.Fatalf("unexpected mapped error: got %#v want %#v", got, testCase.want)
			}
		})
	}
}
