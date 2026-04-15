package api

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/ovumcy/ovumcy-web/internal/models"
	"github.com/ovumcy/ovumcy-web/internal/services"
)

func (handler *Handler) ChangePassword(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		spec := unauthorizedErrorSpec()
		handler.logSecurityError(c, "auth.password_change", spec)
		return handler.respondMappedError(c, spec)
	}

	input, err := parseChangePasswordInput(c)
	if err != nil {
		spec := settingsInvalidInputErrorSpec()
		handler.logSecurityError(c, "auth.password_change", spec)
		return handler.respondMappedError(c, spec)
	}

	if !user.LocalAuthEnabled {
		return handler.enableLocalPasswordAndRenderRecovery(c, user, input)
	}

	if err := handler.settingsService.ChangePassword(user, input.CurrentPassword, input.NewPassword, input.ConfirmPassword); err != nil {
		return handler.respondPasswordChangeError(c, err)
	}

	if err := handler.refreshPasswordChangeSession(c, user); err != nil {
		return err
	}

	handler.logSecurityEvent(c, "auth.password_change", "success")
	return handler.respondPasswordChanged(c)
}

func parseChangePasswordInput(c *fiber.Ctx) (changePasswordInput, error) {
	input := changePasswordInput{}
	if err := c.BodyParser(&input); err != nil {
		return changePasswordInput{}, err
	}
	return input, nil
}

func (handler *Handler) enableLocalPasswordAndRenderRecovery(c *fiber.Ctx, user *models.User, input changePasswordInput) error {
	recoveryCode, err := handler.settingsService.EnableLocalPassword(user, input.NewPassword, input.ConfirmPassword)
	if err != nil {
		return handler.respondPasswordChangeError(c, err)
	}
	if err := handler.refreshPasswordChangeSession(c, user); err != nil {
		return err
	}
	handler.logSecurityEvent(c, "auth.password_change", "local_password_enabled")
	return handler.renderRecoveryCodeResponseWithContinuePath(c, user, recoveryCode, fiber.StatusOK, "/settings")
}

func (handler *Handler) refreshPasswordChangeSession(c *fiber.Ctx, user *models.User) error {
	sessionID, err := handler.setAuthCookie(c, user, false)
	if err != nil {
		handler.clearAuthCookie(c)
		spec := authSessionCreateErrorSpec()
		if errors.Is(err, services.ErrAuthUnsupportedRole) {
			spec = authWebSignInUnavailableErrorSpec()
		}
		handler.logSecurityError(c, "auth.password_change", spec)
		return handler.respondMappedError(c, spec)
	}
	if err := handler.rotateOIDCLogoutState(c, sessionID); err != nil {
		handler.logSecurityEvent(c, "auth.password_change", "provider_logout_state_rotation_failed")
	}
	return nil
}

func (handler *Handler) respondPasswordChangeError(c *fiber.Ctx, err error) error {
	spec := mapSettingsPasswordChangeError(err)
	handler.logSecurityError(c, "auth.password_change", spec)
	return handler.respondMappedError(c, spec)
}

func (handler *Handler) respondPasswordChanged(c *fiber.Ctx) error {
	if acceptsJSON(c) {
		return c.JSON(fiber.Map{"ok": true})
	}
	if isHTMX(c) {
		messageKey := services.SettingsStatusTranslationKey("password_changed")
		message := translateMessage(currentMessages(c), messageKey)
		if message == "" || message == messageKey {
			message = "Password changed successfully."
		}
		return c.SendString(htmxDismissibleSuccessStatusMarkup(currentMessages(c), message))
	}
	handler.setFlashCookie(c, FlashPayload{SettingsSuccess: "password_changed"})
	return redirectOrJSON(c, "/settings")
}
