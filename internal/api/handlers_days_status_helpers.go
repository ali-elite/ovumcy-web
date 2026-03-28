package api

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/ovumcy/ovumcy-web/internal/httpx"
)

func htmxDismissibleSuccessStatusMarkup(messages map[string]string, message string) string {
	return httpx.DismissibleStatusOKMarkup(message, localizedStatusDismissLabel(messages))
}

func localizedStatusDismissLabel(messages map[string]string) string {
	closeLabel := translateMessage(messages, "common.close")
	if closeLabel == "" || closeLabel == "common.close" {
		return "Close"
	}
	return closeLabel
}

func setEncodedResponseNotice(c *fiber.Ctx, message string) {
	trimmed := strings.TrimSpace(message)
	if trimmed == "" {
		return
	}
	c.Set("X-Ovumcy-Notice", url.QueryEscape(trimmed))
}

func (handler *Handler) sendDaySaveStatus(c *fiber.Ctx, messageKey string) error {
	timestamp := time.Now().In(handler.requestLocation(c)).Format("15:04")
	patternKey := messageKey
	if patternKey == "" {
		patternKey = "common.saved_at"
	}
	pattern := translateMessage(currentMessages(c), patternKey)
	if pattern == "" || pattern == patternKey {
		if patternKey == "common.saved_at" {
			pattern = "Saved at %s"
		} else {
			pattern = "Saved."
		}
	}
	message := pattern
	if patternKey == "common.saved_at" {
		message = fmt.Sprintf(pattern, timestamp)
	}
	return c.SendString(htmxDismissibleSuccessStatusMarkup(currentMessages(c), message))
}
