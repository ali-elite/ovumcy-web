package models

import "time"

type OIDCIdentity struct {
	ID         uint       `gorm:"primaryKey"`
	UserID     uint       `gorm:"column:user_id;not null"`
	Issuer     string     `gorm:"column:issuer;not null"`
	Subject    string     `gorm:"column:subject;not null"`
	CreatedAt  time.Time  `gorm:"column:created_at;not null"`
	LastUsedAt *time.Time `gorm:"column:last_used_at"`
}
