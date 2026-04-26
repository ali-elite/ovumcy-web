package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"github.com/ovumcy/ovumcy-web/internal/models"
	"github.com/ovumcy/ovumcy-web/internal/security"
)

const (
	partnerInviteCodeLength    = 16
	partnerInviteCodeGroupSize = 4
	partnerInviteAlphabet      = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	partnerInviteHashContext   = "ovumcy.partner-invite.v1:"
)

var (
	ErrPartnerInviteCodeInvalid   = errors.New("partner invite code invalid")
	ErrPartnerInviteSecretMissing = errors.New("partner invite secret missing")
)

func GeneratePartnerInviteCode() (string, error) {
	raw, err := security.RandomString(partnerInviteCodeLength, partnerInviteAlphabet)
	if err != nil {
		return "", err
	}
	return formatPartnerInviteCode(raw), nil
}

func BuildPartnerInvitationRecord(ownerUserID uint, secretKey []byte, ttl time.Duration, now time.Time) (models.PartnerInvitation, string, error) {
	if ownerUserID == 0 {
		return models.PartnerInvitation{}, "", errors.New("partner invite owner required")
	}
	if ttl <= 0 {
		ttl = 72 * time.Hour
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}

	code, err := GeneratePartnerInviteCode()
	if err != nil {
		return models.PartnerInvitation{}, "", err
	}
	hash, err := HashPartnerInviteCode(secretKey, code)
	if err != nil {
		return models.PartnerInvitation{}, "", err
	}

	return models.PartnerInvitation{
		OwnerUserID: ownerUserID,
		CodeHash:    hash,
		CodeHint:    PartnerInviteCodeHint(code),
		Status:      models.PartnerInvitationStatusPending,
		ExpiresAt:   now.Add(ttl),
		CreatedAt:   now,
	}, code, nil
}

func NormalizePartnerInviteCode(raw string) string {
	condensed := condensePartnerInviteCode(raw)
	if len(condensed) != partnerInviteCodeLength {
		return strings.ToUpper(strings.TrimSpace(raw))
	}
	return formatPartnerInviteCode(condensed)
}

func ValidatePartnerInviteCodeFormat(raw string) error {
	condensed := condensePartnerInviteCode(raw)
	if len(condensed) != partnerInviteCodeLength {
		return ErrPartnerInviteCodeInvalid
	}
	for _, char := range condensed {
		if !strings.ContainsRune(partnerInviteAlphabet, char) {
			return ErrPartnerInviteCodeInvalid
		}
	}
	return nil
}

func PartnerInviteCodeHint(raw string) string {
	condensed := condensePartnerInviteCode(raw)
	if len(condensed) < partnerInviteCodeGroupSize {
		return ""
	}
	return condensed[len(condensed)-partnerInviteCodeGroupSize:]
}

func HashPartnerInviteCode(secretKey []byte, raw string) (string, error) {
	if len(secretKey) == 0 {
		return "", ErrPartnerInviteSecretMissing
	}
	if err := ValidatePartnerInviteCodeFormat(raw); err != nil {
		return "", err
	}

	condensed := condensePartnerInviteCode(raw)
	mac := hmac.New(sha256.New, secretKey)
	_, _ = mac.Write([]byte(partnerInviteHashContext))
	_, _ = mac.Write([]byte(condensed))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}

func IsPartnerInviteCodeMatch(secretKey []byte, raw string, expectedHash string) bool {
	if strings.TrimSpace(expectedHash) == "" {
		return false
	}
	actual, err := HashPartnerInviteCode(secretKey, raw)
	if err != nil {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(actual), []byte(strings.TrimSpace(expectedHash))) == 1
}

func condensePartnerInviteCode(raw string) string {
	condensed := strings.ToUpper(strings.TrimSpace(raw))
	condensed = strings.ReplaceAll(condensed, " ", "")
	condensed = strings.ReplaceAll(condensed, "-", "")
	return condensed
}

func formatPartnerInviteCode(raw string) string {
	condensed := condensePartnerInviteCode(raw)
	if len(condensed) != partnerInviteCodeLength {
		return strings.ToUpper(strings.TrimSpace(raw))
	}

	parts := make([]string, 0, partnerInviteCodeLength/partnerInviteCodeGroupSize)
	for index := 0; index < len(condensed); index += partnerInviteCodeGroupSize {
		parts = append(parts, condensed[index:index+partnerInviteCodeGroupSize])
	}
	return strings.Join(parts, "-")
}
