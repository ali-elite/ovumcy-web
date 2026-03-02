package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/ovumcy/internal/services"
)

func (handler *Handler) OwnerOnly(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return apiError(c, fiber.StatusUnauthorized, "unauthorized")
	}
	if !services.IsOwnerUser(user) {
		return apiError(c, fiber.StatusForbidden, "owner access required")
	}
	return c.Next()
}
