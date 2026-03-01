package api

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

func requestLimiterKey(c *fiber.Ctx) string {
	key := strings.TrimSpace(c.IP())
	if key == "" {
		return "unknown"
	}
	return key
}
