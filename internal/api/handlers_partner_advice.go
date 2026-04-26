package api

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/ovumcy/ovumcy-web/internal/models"
)

func (handler *Handler) HandlePartnerAdvice(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok || user.Role != models.RolePartner {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	language := currentLanguage(c)
	if language == "" {
		language = handler.i18n.DefaultLanguage()
	}
	phase := c.Query("phase", "unknown")
	skipCache := c.Query("fresh") == "true"

	advice, err := handler.partnerAdviceSvc.GetAdvice(c.Context(), phase, language, skipCache)
	if err != nil {
		log.Printf("partner advice unavailable: phase=%q language=%q err=%v", phase, language, err)
		return handler.renderPartnerAdviceUnavailable(c, phase)
	}

	messages := currentMessages(c)

	return handler.renderPartial(c, "partner_advice_partial", fiber.Map{
		"Advice":   advice,
		"Phase":    phase,
		"Messages": messages,
	})
}

func (handler *Handler) renderPartnerAdviceUnavailable(c *fiber.Ctx, phase string) error {
	return handler.renderPartial(c, "partner_advice_partial", fiber.Map{
		"AdviceUnavailable": true,
		"Phase":             phase,
		"Messages":          currentMessages(c),
	})
}
