package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/ovumcy/ovumcy-web/internal/services"
)

func mapSymptomCreateError(err error) APIErrorSpec {
	switch services.ClassifySymptomCreateError(err) {
	case services.SymptomCreateErrorNameRequired:
		return settingsFormErrorSpec(fiber.StatusBadRequest, APIErrorCategoryValidation, "symptom name is required")
	case services.SymptomCreateErrorNameTooLong:
		return settingsFormErrorSpec(fiber.StatusBadRequest, APIErrorCategoryValidation, "symptom name is too long")
	case services.SymptomCreateErrorNameInvalidCharacters:
		return settingsFormErrorSpec(fiber.StatusBadRequest, APIErrorCategoryValidation, "symptom name contains invalid characters")
	case services.SymptomCreateErrorInvalidColor:
		return settingsFormErrorSpec(fiber.StatusBadRequest, APIErrorCategoryValidation, "invalid symptom color")
	case services.SymptomCreateErrorDuplicateName:
		return settingsFormErrorSpec(fiber.StatusConflict, APIErrorCategoryConflict, "symptom name already exists")
	case services.SymptomCreateErrorFailed:
		return settingsFormErrorSpec(fiber.StatusInternalServerError, APIErrorCategoryInternal, "failed to create symptom")
	default:
		return settingsFormErrorSpec(fiber.StatusInternalServerError, APIErrorCategoryInternal, "failed to create symptom")
	}
}

func mapSymptomUpdateError(err error) APIErrorSpec {
	switch services.ClassifySymptomUpdateError(err) {
	case services.SymptomUpdateErrorNotFound:
		return settingsFormErrorSpec(fiber.StatusNotFound, APIErrorCategoryNotFound, "symptom not found")
	case services.SymptomUpdateErrorNameRequired:
		return settingsFormErrorSpec(fiber.StatusBadRequest, APIErrorCategoryValidation, "symptom name is required")
	case services.SymptomUpdateErrorNameTooLong:
		return settingsFormErrorSpec(fiber.StatusBadRequest, APIErrorCategoryValidation, "symptom name is too long")
	case services.SymptomUpdateErrorNameInvalidCharacters:
		return settingsFormErrorSpec(fiber.StatusBadRequest, APIErrorCategoryValidation, "symptom name contains invalid characters")
	case services.SymptomUpdateErrorInvalidColor:
		return settingsFormErrorSpec(fiber.StatusBadRequest, APIErrorCategoryValidation, "invalid symptom color")
	case services.SymptomUpdateErrorDuplicateName:
		return settingsFormErrorSpec(fiber.StatusConflict, APIErrorCategoryConflict, "symptom name already exists")
	case services.SymptomUpdateErrorBuiltinForbidden:
		return settingsFormErrorSpec(fiber.StatusBadRequest, APIErrorCategoryValidation, "built-in symptom cannot be edited")
	case services.SymptomUpdateErrorFailed:
		return settingsFormErrorSpec(fiber.StatusInternalServerError, APIErrorCategoryInternal, "failed to update symptom")
	default:
		return settingsFormErrorSpec(fiber.StatusInternalServerError, APIErrorCategoryInternal, "failed to update symptom")
	}
}

func mapSymptomArchiveError(err error) APIErrorSpec {
	switch services.ClassifySymptomArchiveError(err) {
	case services.SymptomArchiveErrorNotFound:
		return settingsFormErrorSpec(fiber.StatusNotFound, APIErrorCategoryNotFound, "symptom not found")
	case services.SymptomArchiveErrorBuiltinForbidden:
		return settingsFormErrorSpec(fiber.StatusBadRequest, APIErrorCategoryValidation, "built-in symptom cannot be hidden")
	case services.SymptomArchiveErrorFailed:
		return settingsFormErrorSpec(fiber.StatusInternalServerError, APIErrorCategoryInternal, "failed to hide symptom")
	default:
		return settingsFormErrorSpec(fiber.StatusInternalServerError, APIErrorCategoryInternal, "failed to hide symptom")
	}
}

func mapSymptomRestoreError(err error) APIErrorSpec {
	switch services.ClassifySymptomRestoreError(err) {
	case services.SymptomRestoreErrorNotFound:
		return settingsFormErrorSpec(fiber.StatusNotFound, APIErrorCategoryNotFound, "symptom not found")
	case services.SymptomRestoreErrorBuiltinForbidden:
		return settingsFormErrorSpec(fiber.StatusBadRequest, APIErrorCategoryValidation, "built-in symptom cannot be restored")
	case services.SymptomRestoreErrorDuplicateName:
		return settingsFormErrorSpec(fiber.StatusConflict, APIErrorCategoryConflict, "symptom name already exists")
	case services.SymptomRestoreErrorFailed:
		return settingsFormErrorSpec(fiber.StatusInternalServerError, APIErrorCategoryInternal, "failed to restore symptom")
	default:
		return settingsFormErrorSpec(fiber.StatusInternalServerError, APIErrorCategoryInternal, "failed to restore symptom")
	}
}

func symptomsFetchErrorSpec() APIErrorSpec {
	return globalErrorSpec(fiber.StatusInternalServerError, APIErrorCategoryInternal, "failed to fetch symptoms")
}
