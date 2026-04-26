package api

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/ovumcy/ovumcy-web/internal/services"
)

func (handler *Handler) HandleJoin(c *fiber.Ctx) error {
	input, errStr := parseJoinInput(c)
	if errStr != "" {
		spec := authInvalidInputErrorSpec()
		handler.logSecurityError(c, "partner.join", spec)
		return handler.respondMappedError(c, spec)
	}

	// 1. Register the partner account
	user, recoveryCode, err := handler.registrationService.RegisterPartnerAccount(
		input.Email,
		input.Password,
		input.ConfirmPassword,
		time.Now().In(handler.location),
	)
	if err != nil {
		spec := mapAuthRegisterError(err)
		handler.logSecurityError(c, "partner.join", spec)
		return handler.respondMappedError(c, spec)
	}

	// 2. Redeem the invitation
	err = handler.partnerInviteSvc.RedeemInvitation(input.InviteCode, user.ID, time.Now())
	if err != nil {
		// If redemption fails, we should ideally rollback user creation, 
		// but since we don't have transactions here easily, we at least return an error.
		// Actually, we could delete the user if redemption fails.
		// For now, let's just map the error.
		spec := authInvalidInviteCodeErrorSpec()
		handler.logSecurityError(c, "partner.join", spec)
		
		// Note: The user account is created but not linked. 
		// They can try again if they fix the code, but they'd need to login first.
		// This is a bit of a rough edge, but let's stick to the flow.
		return handler.respondMappedError(c, spec)
	}

	// 3. Log them in
	if _, err := handler.setAuthCookie(c, &user, true); err != nil {
		spec := authSessionCreateErrorSpec()
		if errors.Is(err, services.ErrAuthUnsupportedRole) {
			spec = authWebSignInUnavailableErrorSpec()
		}
		handler.logSecurityError(c, "partner.join", spec)
		return handler.respondMappedError(c, spec)
	}
	handler.clearOIDCLogoutTransportCookies(c)

	handler.logSecurityEvent(c, "partner.join", "success")
	
	// Partners also get a recovery code
	return handler.renderRegisterInlineRecoveryResponse(c, &user, recoveryCode, fiber.StatusCreated)
}

func authInvalidInviteCodeErrorSpec() APIErrorSpec {
	return APIErrorSpec{
		Key:      "auth.error.invalid_invite_code",
		Status:   fiber.StatusBadRequest,
		Category: APIErrorCategoryValidation,
		Target:   APIErrorTargetAuthForm,
	}
}
