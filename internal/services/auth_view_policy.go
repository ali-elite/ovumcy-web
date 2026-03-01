package services

import (
	"strings"
	"time"
)

func ResolveAuthErrorSource(flashAuthError string, queryError string) string {
	return firstNonEmptyTrimmed(flashAuthError, queryError)
}

func ResolveAuthPageEmail(flashEmail string, queryEmail string) string {
	email := NormalizeAuthEmail(flashEmail)
	if email == "" {
		email = NormalizeAuthEmail(queryEmail)
	}
	return email
}

func IsResetPasswordTokenValid(secretKey []byte, rawToken string, now time.Time) bool {
	if strings.TrimSpace(rawToken) == "" {
		return false
	}
	_, err := ParsePasswordResetToken(secretKey, rawToken, now)
	return err == nil
}
