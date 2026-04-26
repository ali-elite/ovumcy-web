package api

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

const defaultPartnerInvitationTTL = 72 * time.Hour

func (handler *Handler) CreatePartnerInvitation(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		spec := unauthorizedErrorSpec()
		handler.logSecurityError(c, "partner.invitation_create", spec)
		return handler.respondMappedError(c, spec)
	}

	invitation, code, err := handler.partnerInviteSvc.CreateInvitation(user.ID, defaultPartnerInvitationTTL, time.Now().UTC())
	if err != nil {
		handler.logSecurityError(c, "partner.invitation_create", globalErrorSpec(fiber.StatusInternalServerError, APIErrorCategoryInternal, "failed to create partner invitation"))

		if isHTMX(c) {
			return handler.renderPartial(c, "settings_partner_invitation_partial", fiber.Map{
				"HasInvitation": false,
				"Lang":          currentLanguage(c),
				"Messages":      currentMessages(c),
				"CSRFToken":     csrfToken(c),
				"ErrorMessage":  translateMessage(currentMessages(c), "settings.partner.generate_error"),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"ok": false})
	}

	handler.logSecurityEvent(c, "partner.invitation_create", "success")

	if isHTMX(c) {
		return handler.renderPartial(c, "settings_partner_invitation_partial", fiber.Map{
			"HasInvitation":       true,
			"InvitationCode":      code,
			"InvitationExpiresAt": invitation.ExpiresAt,
			"Lang":                currentLanguage(c),
			"Messages":            currentMessages(c),
			"CSRFToken":           csrfToken(c),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"ok":         true,
		"code":       code,
		"status":     invitation.Status,
		"expires_at": invitation.ExpiresAt.UTC().Format(time.RFC3339),
	})
}
