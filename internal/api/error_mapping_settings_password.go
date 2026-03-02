package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/ovumcy/internal/services"
)

func mapSettingsPasswordChangeError(err error) APIErrorSpec {
	switch services.ClassifySettingsPasswordChangeError(err) {
	case services.SettingsPasswordChangeErrorInvalidInput:
		return settingsFormErrorSpec(fiber.StatusBadRequest, APIErrorCategoryValidation, "invalid settings input")
	case services.SettingsPasswordChangeErrorPasswordMismatch:
		return settingsFormErrorSpec(fiber.StatusBadRequest, APIErrorCategoryValidation, "password mismatch")
	case services.SettingsPasswordChangeErrorInvalidCurrentPassword:
		return settingsFormErrorSpec(fiber.StatusUnauthorized, APIErrorCategoryUnauthorized, "invalid current password")
	case services.SettingsPasswordChangeErrorNewPasswordMustDiffer:
		return settingsFormErrorSpec(fiber.StatusBadRequest, APIErrorCategoryValidation, "new password must differ")
	case services.SettingsPasswordChangeErrorWeakPassword:
		return settingsFormErrorSpec(fiber.StatusBadRequest, APIErrorCategoryValidation, "weak password")
	case services.SettingsPasswordChangeErrorHashFailed:
		return globalErrorSpec(fiber.StatusInternalServerError, APIErrorCategoryInternal, "failed to secure password")
	case services.SettingsPasswordChangeErrorUpdateFailed:
		return globalErrorSpec(fiber.StatusInternalServerError, APIErrorCategoryInternal, "failed to update password")
	default:
		return globalErrorSpec(fiber.StatusInternalServerError, APIErrorCategoryInternal, "failed to update password")
	}
}
