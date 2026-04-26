package db

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/ovumcy/ovumcy-web/internal/models"
)

func TestPartnerInvitationRepositoryCreateAndLookup(t *testing.T) {
	database := openSQLiteForMigrationBootstrapTest(t, filepath.Join(t.TempDir(), "partner-invitations.db"))

	if err := database.Exec(
		`INSERT INTO users (email, password_hash, role, created_at, local_auth_enabled) VALUES (?, ?, ?, CURRENT_TIMESTAMP, 1)`,
		"partner-invite-owner@example.com",
		"hash",
		"owner",
	).Error; err != nil {
		t.Fatalf("insert owner user: %v", err)
	}

	repository := NewPartnerInvitationRepository(database)
	now := time.Now().UTC()
	invitation := models.PartnerInvitation{
		OwnerUserID: 1,
		CodeHash:    "invite-hash",
		CodeHint:    "ABCD",
		ExpiresAt:   now.Add(72 * time.Hour),
		CreatedAt:   now,
	}

	if err := repository.Create(&invitation); err != nil {
		t.Fatalf("create invitation: %v", err)
	}
	if invitation.ID == 0 {
		t.Fatal("expected invitation ID to be assigned")
	}
	if invitation.Status != models.PartnerInvitationStatusPending {
		t.Fatalf("expected default invitation status %q, got %q", models.PartnerInvitationStatusPending, invitation.Status)
	}

	stored, found, err := repository.FindByCodeHash(invitation.CodeHash)
	if err != nil {
		t.Fatalf("find invitation by hash: %v", err)
	}
	if !found {
		t.Fatal("expected invitation lookup to find stored record")
	}
	if stored.OwnerUserID != invitation.OwnerUserID {
		t.Fatalf("expected owner_user_id %d, got %d", invitation.OwnerUserID, stored.OwnerUserID)
	}
	if stored.CodeHint != invitation.CodeHint {
		t.Fatalf("expected code_hint %q, got %q", invitation.CodeHint, stored.CodeHint)
	}
}
