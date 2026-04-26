package models

import "time"

import "gorm.io/gorm"

const (
	PartnerLinkStatusActive  = "active"
	PartnerLinkStatusRevoked = "revoked"
)

type PartnerLink struct {
	ID              uint       `gorm:"primaryKey"`
	OwnerUserID     uint       `gorm:"column:owner_user_id;not null"`
	PartnerUserID   uint       `gorm:"column:partner_user_id;not null;uniqueIndex"`
	Status          string     `gorm:"column:status;not null;default:active"`
	InvitedByUserID uint       `gorm:"column:invited_by_user_id;not null"`
	CreatedAt       time.Time  `gorm:"column:created_at;not null"`
	RevokedAt       *time.Time `gorm:"column:revoked_at"`
}

func (link *PartnerLink) BeforeCreate(_ *gorm.DB) error {
	if link == nil {
		return nil
	}
	if link.Status == "" {
		link.Status = PartnerLinkStatusActive
	}
	return nil
}

func (PartnerLink) TableName() string {
	return "partner_links"
}
