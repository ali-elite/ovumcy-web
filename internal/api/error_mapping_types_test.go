package api

import (
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestAPIErrorSpecIsFormError(t *testing.T) {
	testCases := []struct {
		name string
		spec APIErrorSpec
		want bool
	}{
		{
			name: "global",
			spec: globalErrorSpec(fiber.StatusBadRequest, APIErrorCategoryValidation, "invalid input"),
			want: false,
		},
		{
			name: "auth form",
			spec: authFormErrorSpec(fiber.StatusUnauthorized, APIErrorCategoryUnauthorized, "invalid credentials"),
			want: true,
		},
		{
			name: "settings form",
			spec: settingsFormErrorSpec(fiber.StatusBadRequest, APIErrorCategoryValidation, "invalid settings input"),
			want: true,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			if got := testCase.spec.IsFormError(); got != testCase.want {
				t.Fatalf("IsFormError() = %t, want %t", got, testCase.want)
			}
		})
	}
}
