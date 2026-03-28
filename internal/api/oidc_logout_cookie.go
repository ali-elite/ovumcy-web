package api

import (
	"encoding/json"
	"net/url"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/ovumcy/ovumcy-web/internal/services"
)

type oidcLogoutCookiePayload struct {
	EndSessionEndpoint    string `json:"end_session_endpoint"`
	IDTokenHint           string `json:"id_token_hint"`
	PostLogoutRedirectURL string `json:"post_logout_redirect_url"`
}

type oidcLogoutBridgeCookiePayload struct {
	EndSessionEndpoint    string `json:"end_session_endpoint"`
	IDTokenHint           string `json:"id_token_hint"`
	PostLogoutRedirectURL string `json:"post_logout_redirect_url"`
	ExpiresAtUnix         int64  `json:"expires_at_unix"`
}

func (handler *Handler) setOIDCLogoutCookie(c *fiber.Ctx, state services.OIDCLogoutState) error {
	payload := oidcLogoutCookiePayload{
		EndSessionEndpoint:    strings.TrimSpace(state.EndSessionEndpoint),
		IDTokenHint:           strings.TrimSpace(state.IDTokenHint),
		PostLogoutRedirectURL: strings.TrimSpace(state.PostLogoutRedirectURL),
	}
	if !payload.valid() {
		handler.clearOIDCLogoutTransportCookies(c)
		return fiber.ErrBadRequest
	}

	serialized, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	encoded, err := handler.sealCookieValue(oidcLogoutCookieName, serialized)
	if err != nil {
		return err
	}

	c.Cookie(&fiber.Cookie{
		Name:     oidcLogoutCookieName,
		Value:    encoded,
		Path:     "/",
		HTTPOnly: true,
		Secure:   handler.cookieSecure,
		SameSite: "Lax",
	})
	handler.clearOIDCLogoutBridgeCookie(c)
	return nil
}

func (handler *Handler) readOIDCLogoutCookie(c *fiber.Ctx) oidcLogoutCookiePayload {
	raw := strings.TrimSpace(c.Cookies(oidcLogoutCookieName))
	if raw == "" {
		return oidcLogoutCookiePayload{}
	}

	decoded, err := handler.openCookieValue(oidcLogoutCookieName, raw)
	if err != nil {
		handler.clearOIDCLogoutCookie(c)
		return oidcLogoutCookiePayload{}
	}

	payload := oidcLogoutCookiePayload{}
	if err := json.Unmarshal(decoded, &payload); err != nil || !payload.valid() {
		handler.clearOIDCLogoutCookie(c)
		return oidcLogoutCookiePayload{}
	}
	return payload
}

func (handler *Handler) setOIDCLogoutBridgeCookie(c *fiber.Ctx, payload oidcLogoutCookiePayload, now time.Time) error {
	if !payload.valid() {
		handler.clearOIDCLogoutBridgeCookie(c)
		return fiber.ErrBadRequest
	}
	if now.IsZero() {
		now = time.Now()
	}
	expiresAt := now.UTC().Add(time.Minute)
	bridgePayload := oidcLogoutBridgeCookiePayload{
		EndSessionEndpoint:    payload.EndSessionEndpoint,
		IDTokenHint:           payload.IDTokenHint,
		PostLogoutRedirectURL: payload.PostLogoutRedirectURL,
		ExpiresAtUnix:         expiresAt.Unix(),
	}

	serialized, err := json.Marshal(bridgePayload)
	if err != nil {
		return err
	}
	encoded, err := handler.sealCookieValue(oidcLogoutBridgeCookieName, serialized)
	if err != nil {
		return err
	}

	c.Cookie(&fiber.Cookie{
		Name:     oidcLogoutBridgeCookieName,
		Value:    encoded,
		Path:     oidcLogoutBridgePath,
		HTTPOnly: true,
		Secure:   handler.cookieSecure,
		SameSite: "Lax",
		Expires:  expiresAt,
	})
	return nil
}

func (handler *Handler) readOIDCLogoutBridgeCookie(c *fiber.Ctx, now time.Time) oidcLogoutCookiePayload {
	raw := strings.TrimSpace(c.Cookies(oidcLogoutBridgeCookieName))
	if raw == "" {
		return oidcLogoutCookiePayload{}
	}

	decoded, err := handler.openCookieValue(oidcLogoutBridgeCookieName, raw)
	if err != nil {
		handler.clearOIDCLogoutBridgeCookie(c)
		return oidcLogoutCookiePayload{}
	}

	payload := oidcLogoutBridgeCookiePayload{}
	if err := json.Unmarshal(decoded, &payload); err != nil || !payload.validAt(now) {
		handler.clearOIDCLogoutBridgeCookie(c)
		return oidcLogoutCookiePayload{}
	}

	return oidcLogoutCookiePayload{
		EndSessionEndpoint:    payload.EndSessionEndpoint,
		IDTokenHint:           payload.IDTokenHint,
		PostLogoutRedirectURL: payload.PostLogoutRedirectURL,
	}
}

func (handler *Handler) clearOIDCLogoutCookie(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     oidcLogoutCookieName,
		Value:    "",
		Path:     "/",
		HTTPOnly: true,
		Secure:   handler.cookieSecure,
		SameSite: "Lax",
		Expires:  time.Now().Add(-1 * time.Hour),
	})
}

func (handler *Handler) clearOIDCLogoutBridgeCookie(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     oidcLogoutBridgeCookieName,
		Value:    "",
		Path:     oidcLogoutBridgePath,
		HTTPOnly: true,
		Secure:   handler.cookieSecure,
		SameSite: "Lax",
		Expires:  time.Now().Add(-1 * time.Hour),
	})
}

func (handler *Handler) clearOIDCLogoutTransportCookies(c *fiber.Ctx) {
	handler.clearOIDCLogoutCookie(c)
	handler.clearOIDCLogoutBridgeCookie(c)
}

func (handler *Handler) providerLogoutRedirectURLFromPayload(payload oidcLogoutCookiePayload) string {
	if !payload.valid() {
		return ""
	}
	logoutURL, err := url.Parse(payload.EndSessionEndpoint)
	if err != nil || !logoutURL.IsAbs() {
		return ""
	}

	query := logoutURL.Query()
	query.Set("id_token_hint", payload.IDTokenHint)
	query.Set("post_logout_redirect_uri", payload.PostLogoutRedirectURL)
	logoutURL.RawQuery = query.Encode()
	return logoutURL.String()
}

func (handler *Handler) sealCookieValue(cookieName string, plaintext []byte) (string, error) {
	codec, err := newSecureCookieCodec(handler.secretKey)
	if err != nil {
		return "", err
	}
	return codec.seal(cookieName, plaintext)
}

func (handler *Handler) openCookieValue(cookieName string, raw string) ([]byte, error) {
	codec, err := newSecureCookieCodec(handler.secretKey)
	if err != nil {
		return nil, err
	}
	return codec.open(cookieName, raw)
}

func (payload oidcLogoutCookiePayload) valid() bool {
	endSessionEndpoint := strings.TrimSpace(payload.EndSessionEndpoint)
	idTokenHint := strings.TrimSpace(payload.IDTokenHint)
	postLogoutRedirectURL := strings.TrimSpace(payload.PostLogoutRedirectURL)
	if endSessionEndpoint == "" || idTokenHint == "" || postLogoutRedirectURL == "" {
		return false
	}

	endpointURL, err := url.Parse(endSessionEndpoint)
	if err != nil || !endpointURL.IsAbs() || !strings.EqualFold(endpointURL.Scheme, "https") || endpointURL.Fragment != "" {
		return false
	}

	redirectURL, err := url.Parse(postLogoutRedirectURL)
	if err != nil || !redirectURL.IsAbs() || !strings.EqualFold(redirectURL.Scheme, "https") {
		return false
	}
	if redirectURL.RawQuery != "" || redirectURL.Fragment != "" {
		return false
	}

	return true
}

func (payload oidcLogoutBridgeCookiePayload) validAt(now time.Time) bool {
	if !(oidcLogoutCookiePayload{
		EndSessionEndpoint:    payload.EndSessionEndpoint,
		IDTokenHint:           payload.IDTokenHint,
		PostLogoutRedirectURL: payload.PostLogoutRedirectURL,
	}.valid()) {
		return false
	}
	if payload.ExpiresAtUnix <= 0 {
		return false
	}
	if now.IsZero() {
		now = time.Now()
	}
	return now.UTC().Unix() <= payload.ExpiresAtUnix
}
