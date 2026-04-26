package services

import (
	"strings"
	"testing"
	"time"

	"github.com/ovumcy/ovumcy-web/internal/models"
)

func TestBuildPartnerInvitationRecord(t *testing.T) {
	t.Parallel()

	secret := []byte("partner-invite-secret")
	now := time.Date(2026, time.April, 23, 10, 0, 0, 0, time.UTC)
	invitation, code, err := BuildPartnerInvitationRecord(42, secret, 48*time.Hour, now)
	if err != nil {
		t.Fatalf("BuildPartnerInvitationRecord returned error: %v", err)
	}
	if invitation.OwnerUserID != 42 {
		t.Fatalf("expected owner user id 42, got %d", invitation.OwnerUserID)
	}
	if invitation.Status != models.PartnerInvitationStatusPending {
		t.Fatalf("expected pending invitation status, got %q", invitation.Status)
	}
	if invitation.CodeHint != PartnerInviteCodeHint(code) {
		t.Fatalf("expected code hint %q, got %q", PartnerInviteCodeHint(code), invitation.CodeHint)
	}
	if invitation.ExpiresAt != now.Add(48*time.Hour) {
		t.Fatalf("expected expiry %v, got %v", now.Add(48*time.Hour), invitation.ExpiresAt)
	}
	if invitation.CreatedAt != now {
		t.Fatalf("expected created_at %v, got %v", now, invitation.CreatedAt)
	}
	if !IsPartnerInviteCodeMatch(secret, code, invitation.CodeHash) {
		t.Fatal("expected generated code to match invitation hash")
	}
}

func TestGeneratePartnerInviteCodeProducesCanonicalFormat(t *testing.T) {
	t.Parallel()

	code, err := GeneratePartnerInviteCode()
	if err != nil {
		t.Fatalf("GeneratePartnerInviteCode returned error: %v", err)
	}
	if len(code) != 19 {
		t.Fatalf("expected formatted invite code length 19, got %d for %q", len(code), code)
	}
	if normalized := NormalizePartnerInviteCode(strings.ToLower("  " + code + "  ")); normalized != code {
		t.Fatalf("expected normalized code %q, got %q", code, normalized)
	}
	if err := ValidatePartnerInviteCodeFormat(code); err != nil {
		t.Fatalf("expected formatted code to validate, got %v", err)
	}
	if hint := PartnerInviteCodeHint(code); hint != code[len(code)-4:] {
		t.Fatalf("expected hint %q, got %q", code[len(code)-4:], hint)
	}
}

func TestPartnerInviteCodeHashMatchesNormalizedInput(t *testing.T) {
	t.Parallel()

	secret := []byte("partner-invite-secret")
	code := "ABCD-EFGH-JKMP-QRST"
	hash, err := HashPartnerInviteCode(secret, code)
	if err != nil {
		t.Fatalf("HashPartnerInviteCode returned error: %v", err)
	}

	if !IsPartnerInviteCodeMatch(secret, "abcd efgh jkmp qrst", hash) {
		t.Fatal("expected normalized invite code to match hash")
	}
	if IsPartnerInviteCodeMatch(secret, "ZZZZ-ZZZZ-ZZZZ-ZZZZ", hash) {
		t.Fatal("expected different invite code not to match hash")
	}
}

func TestPartnerInviteCodeValidationRejectsInvalidFormats(t *testing.T) {
	t.Parallel()

	invalidCodes := []string{"", "abc", "ABCD-EFGH-IJKL-MNO", "ABCD-EFGH-IJKL-MN0O", "ABCD-EFGH-IJKL-MNO!"}
	for _, code := range invalidCodes {
		code := code
		t.Run(code, func(t *testing.T) {
			t.Parallel()

			if err := ValidatePartnerInviteCodeFormat(code); err == nil {
				t.Fatalf("expected %q to be rejected", code)
			}
		})
	}
}
