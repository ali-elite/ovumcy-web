package models

import "time"

import "gorm.io/gorm"

const (
	PartnerInvitationStatusPending  = "pending"
	PartnerInvitationStatusRedeemed = "redeemed"
	PartnerInvitationStatusRevoked  = "revoked"
	PartnerInvitationStatusExpired  = "expired"
)

type PartnerInvitation struct {
	ID               uint       `gorm:"primaryKey"`
	OwnerUserID      uint       `gorm:"column:owner_user_id;not null"`
	CodeHash         string     `gorm:"column:code_hash;not null;uniqueIndex"`
	CodeHint         string     `gorm:"column:code_hint;not null;size:4"`
	Status           string     `gorm:"column:status;not null;default:pending"`
	ExpiresAt        time.Time  `gorm:"column:expires_at;not null"`
	RedeemedAt       *time.Time `gorm:"column:redeemed_at"`
	RedeemedByUserID *uint      `gorm:"column:redeemed_by_user_id"`
	RevokedAt        *time.Time `gorm:"column:revoked_at"`
	CreatedAt        time.Time  `gorm:"column:created_at;not null"`
}

func (invitation *PartnerInvitation) BeforeCreate(_ *gorm.DB) error {
	if invitation == nil {
		return nil
	}
	if invitation.Status == "" {
		invitation.Status = PartnerInvitationStatusPending
	}
	return nil
}

func (PartnerInvitation) TableName() string {
	return "partner_invitations"
}
