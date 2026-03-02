package httpx

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

type JSONMode uint8

const (
	JSONModeAcceptOnly JSONMode = iota
	JSONModeAcceptOrContentType
)

func IsHTMX(c *fiber.Ctx) bool {
	return strings.EqualFold(c.Get("HX-Request"), "true")
}

func AcceptsJSON(c *fiber.Ctx, mode JSONMode) bool {
	accept := strings.ToLower(c.Get("Accept"))
	if strings.Contains(accept, fiber.MIMEApplicationJSON) {
		return true
	}

	if mode == JSONModeAcceptOrContentType {
		contentType := strings.ToLower(c.Get(fiber.HeaderContentType))
		if strings.Contains(contentType, fiber.MIMEApplicationJSON) {
			return true
		}
	}

	return false
}
