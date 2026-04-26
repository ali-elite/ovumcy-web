package services

import (
	"errors"
	"time"

	"github.com/ovumcy/ovumcy-web/internal/models"
)

var (
	ErrInviteCodeInvalid  = errors.New("invite code is invalid or not found")
	ErrInviteCodeExpired  = errors.New("invite code has expired")
	ErrInviteCodeRedeemed = errors.New("invite code has already been used")
	ErrInviteCodeRevoked  = errors.New("invite code has been revoked")
)

type partnerInvitationStore interface {
	Create(invitation *models.PartnerInvitation) error
	Save(invitation *models.PartnerInvitation) error
	FindByCodeHash(codeHash string) (models.PartnerInvitation, bool, error)
	RevokeAllPendingByOwner(ownerUserID uint) error
}

type partnerLinkStore interface {
	Create(link *models.PartnerLink) error
}

type PartnerInvitationRepository interface {
	Create(invitation *models.PartnerInvitation) error
}

type PartnerInvitationService struct {
	invitations partnerInvitationStore
	links       partnerLinkStore
	secretKey   []byte
}

func NewPartnerInvitationService(invitations partnerInvitationStore, secretKey []byte) *PartnerInvitationService {
	return &PartnerInvitationService{invitations: invitations, secretKey: secretKey}
}

// WithPartnerLinks wires in the link store so RedeemInvitation can create partner links.
func (service *PartnerInvitationService) WithPartnerLinks(links partnerLinkStore) *PartnerInvitationService {
	service.links = links
	return service
}

func (service *PartnerInvitationService) CreateInvitation(ownerUserID uint, ttl time.Duration, now time.Time) (models.PartnerInvitation, string, error) {
	if service == nil || service.invitations == nil {
		return models.PartnerInvitation{}, "", errors.New("partner invitation repository is required")
	}

	if err := service.invitations.RevokeAllPendingByOwner(ownerUserID); err != nil {
		return models.PartnerInvitation{}, "", err
	}

	invitation, code, err := BuildPartnerInvitationRecord(ownerUserID, service.secretKey, ttl, now)
	if err != nil {
		return models.PartnerInvitation{}, "", err
	}
	if err := service.invitations.Create(&invitation); err != nil {
		return models.PartnerInvitation{}, "", err
	}
	return invitation, code, nil
}

// RedeemInvitation validates a raw invite code, creates the partner link, and marks the invitation redeemed.
// partnerUserID is the newly registered partner's user ID.
func (service *PartnerInvitationService) RedeemInvitation(rawCode string, partnerUserID uint, now time.Time) error {
	if service == nil || service.invitations == nil {
		return errors.New("invitation store required")
	}

	hash, err := HashPartnerInviteCode(service.secretKey, rawCode)
	if err != nil {
		return ErrInviteCodeInvalid
	}

	invitation, found, err := service.invitations.FindByCodeHash(hash)
	if err != nil {
		return err
	}
	if !found {
		return ErrInviteCodeInvalid
	}

	switch invitation.Status {
	case models.PartnerInvitationStatusRedeemed:
		return ErrInviteCodeRedeemed
	case models.PartnerInvitationStatusRevoked:
		return ErrInviteCodeRevoked
	}
	if now.After(invitation.ExpiresAt) {
		return ErrInviteCodeExpired
	}

	// Create the PartnerLink
	if service.links != nil {
		link := &models.PartnerLink{
			OwnerUserID:     invitation.OwnerUserID,
			PartnerUserID:   partnerUserID,
			InvitedByUserID: invitation.OwnerUserID,
			Status:          models.PartnerLinkStatusActive,
			CreatedAt:       now,
		}
		if err := service.links.Create(link); err != nil {
			return err
		}
	}

	// Mark invitation redeemed
	redeemedAt := now
	invitation.Status = models.PartnerInvitationStatusRedeemed
	invitation.RedeemedAt = &redeemedAt
	invitation.RedeemedByUserID = &partnerUserID
	return service.invitations.Save(&invitation)
}
