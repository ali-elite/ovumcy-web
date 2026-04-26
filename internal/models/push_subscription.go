package models

import "time"

type PushSubscription struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	User      User      `json:"-"`
	Endpoint  string    `gorm:"type:text;not null;uniqueIndex:idx_push_subscription_endpoint" json:"endpoint"`
	P256dh    string    `gorm:"not null" json:"p256dh"`
	Auth      string    `gorm:"not null" json:"auth"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
