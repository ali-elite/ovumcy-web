package services

import (
	"strings"
	"time"
)

func ResolveAuthErrorSource(flashAuthError string) string {
	return firstNonEmptyTrimmed(flashAuthError)
}

func ResolveAuthPageEmail(flashEmail string) string {
	return NormalizeAuthEmail(flashEmail)
}

func IsResetPasswordTokenValid(secretKey []byte, rawToken string, now time.Time) bool {
	if strings.TrimSpace(rawToken) == "" {
		return false
	}
	_, err := ParsePasswordResetToken(secretKey, rawToken, now)
	return err == nil
}
