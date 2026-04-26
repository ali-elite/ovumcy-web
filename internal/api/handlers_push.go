package api

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/ovumcy/ovumcy-web/internal/models"
)

type PushSubscriptionRequest struct {
	Endpoint string `json:"endpoint"`
	Keys     struct {
		P256dh string `json:"p256dh"`
		Auth   string `json:"auth"`
	} `json:"keys"`
}

func (h *Handler) HandlePushSubscribe(c *fiber.Ctx) error {
	user, handled, err := currentUserOrUnauthorized(c)
	if err != nil {
		return err
	}
	if handled {
		return nil
	}

	var req PushSubscriptionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
	}

	if req.Endpoint == "" || req.Keys.P256dh == "" || req.Keys.Auth == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "missing subscription details"})
	}

	sub := &models.PushSubscription{
		UserID:   user.ID,
		Endpoint: req.Endpoint,
		P256dh:   req.Keys.P256dh,
		Auth:     req.Keys.Auth,
	}

	if err := h.pushSubscriptions.Save(sub); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to save subscription"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"ok": true})
}

func (h *Handler) HandleInternalCronDaily(c *fiber.Ctx) error {
	secret := os.Getenv("CRON_SECRET")
	provided := c.Get("Authorization")

	if secret == "" || provided != "Bearer "+secret {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	if h.reminderService != nil {
		if err := h.reminderService.CheckAndSendDailyReminders(); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"ok": true})
}
