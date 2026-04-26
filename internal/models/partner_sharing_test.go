package models

import "testing"

func TestPartnerInvitationBeforeCreateDefaultsPendingStatus(t *testing.T) {
	t.Parallel()

	invitation := &PartnerInvitation{}
	if err := invitation.BeforeCreate(nil); err != nil {
		t.Fatalf("BeforeCreate returned error: %v", err)
	}
	if invitation.Status != PartnerInvitationStatusPending {
		t.Fatalf("expected pending status, got %q", invitation.Status)
	}
}

func TestPartnerInvitationBeforeCreateKeepsExplicitStatus(t *testing.T) {
	t.Parallel()

	invitation := &PartnerInvitation{Status: PartnerInvitationStatusRedeemed}
	if err := invitation.BeforeCreate(nil); err != nil {
		t.Fatalf("BeforeCreate returned error: %v", err)
	}
	if invitation.Status != PartnerInvitationStatusRedeemed {
		t.Fatalf("expected explicit status to remain redeemed, got %q", invitation.Status)
	}
}

func TestPartnerLinkBeforeCreateDefaultsActiveStatus(t *testing.T) {
	t.Parallel()

	link := &PartnerLink{}
	if err := link.BeforeCreate(nil); err != nil {
		t.Fatalf("BeforeCreate returned error: %v", err)
	}
	if link.Status != PartnerLinkStatusActive {
		t.Fatalf("expected active status, got %q", link.Status)
	}
}

func TestPartnerLinkBeforeCreateKeepsExplicitStatus(t *testing.T) {
	t.Parallel()

	link := &PartnerLink{Status: PartnerLinkStatusRevoked}
	if err := link.BeforeCreate(nil); err != nil {
		t.Fatalf("BeforeCreate returned error: %v", err)
	}
	if link.Status != PartnerLinkStatusRevoked {
		t.Fatalf("expected explicit status to remain revoked, got %q", link.Status)
	}
}
