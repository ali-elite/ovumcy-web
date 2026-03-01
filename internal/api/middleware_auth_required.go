package api

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/ovumcy/internal/services"
)

func (handler *Handler) AuthRequired(c *fiber.Ctx) error {
	user, err := handler.authenticateRequest(c)
	if err != nil {
		if strings.HasPrefix(c.Path(), "/api/") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
		}
		return c.Redirect("/login", fiber.StatusSeeOther)
	}

	c.Locals(contextUserKey, user)
	if services.RequiresOnboarding(user) && services.ShouldEnforceOnboardingAccess(c.Path()) {
		if strings.HasPrefix(c.Path(), "/api/") {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "onboarding required"})
		}
		return c.Redirect("/onboarding", fiber.StatusSeeOther)
	}

	return c.Next()
}
