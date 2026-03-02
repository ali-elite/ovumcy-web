package api

import (
	"errors"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/ovumcy/internal/services"
)

func TestMapSymptomCreateError(t *testing.T) {
	testCases := []struct {
		name string
		err  error
		want APIErrorSpec
	}{
		{
			name: "invalid name",
			err:  services.ErrInvalidSymptomName,
			want: globalErrorSpec(fiber.StatusBadRequest, APIErrorCategoryValidation, "invalid symptom name"),
		},
		{
			name: "invalid color",
			err:  services.ErrInvalidSymptomColor,
			want: globalErrorSpec(fiber.StatusBadRequest, APIErrorCategoryValidation, "invalid symptom color"),
		},
		{
			name: "create failed",
			err:  services.ErrCreateSymptomFailed,
			want: globalErrorSpec(fiber.StatusInternalServerError, APIErrorCategoryInternal, "failed to create symptom"),
		},
		{
			name: "unknown",
			err:  errors.New("unknown"),
			want: globalErrorSpec(fiber.StatusInternalServerError, APIErrorCategoryInternal, "failed to create symptom"),
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			if got := mapSymptomCreateError(testCase.err); got != testCase.want {
				t.Fatalf("unexpected mapped error: got %#v want %#v", got, testCase.want)
			}
		})
	}
}

func TestMapSymptomDeleteError(t *testing.T) {
	testCases := []struct {
		name string
		err  error
		want APIErrorSpec
	}{
		{
			name: "not found",
			err:  services.ErrSymptomNotFound,
			want: globalErrorSpec(fiber.StatusNotFound, APIErrorCategoryNotFound, "symptom not found"),
		},
		{
			name: "builtin forbidden",
			err:  services.ErrBuiltinSymptomDeleteForbidden,
			want: globalErrorSpec(fiber.StatusBadRequest, APIErrorCategoryValidation, "built-in symptom cannot be deleted"),
		},
		{
			name: "delete failed",
			err:  services.ErrDeleteSymptomFailed,
			want: globalErrorSpec(fiber.StatusInternalServerError, APIErrorCategoryInternal, "failed to delete symptom"),
		},
		{
			name: "clean logs failed",
			err:  services.ErrCleanSymptomLogsFailed,
			want: globalErrorSpec(fiber.StatusInternalServerError, APIErrorCategoryInternal, "failed to clean symptom logs"),
		},
		{
			name: "unknown",
			err:  errors.New("unknown"),
			want: globalErrorSpec(fiber.StatusInternalServerError, APIErrorCategoryInternal, "failed to delete symptom"),
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			if got := mapSymptomDeleteError(testCase.err); got != testCase.want {
				t.Fatalf("unexpected mapped error: got %#v want %#v", got, testCase.want)
			}
		})
	}
}
