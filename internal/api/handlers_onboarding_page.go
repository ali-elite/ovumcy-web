package api

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/ovumcy/ovumcy-web/internal/services"
)

func (handler *Handler) ShowOnboarding(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return c.Redirect("/login", fiber.StatusSeeOther)
	}
	if !services.RequiresOnboarding(user) {
		return c.Redirect("/dashboard", fiber.StatusSeeOther)
	}

	location := handler.requestLocation(c)
	now := services.DateAtLocation(time.Now().In(location), location)
	data := handler.buildOnboardingViewData(c, user, now, location)
	return handler.render(c, "onboarding", data)
}
