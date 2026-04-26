package db

import (
	"errors"

	"github.com/ovumcy/ovumcy-web/internal/models"
	"gorm.io/gorm"
)

type PushSubscriptionRepository struct {
	database *gorm.DB
}

func NewPushSubscriptionRepository(database *gorm.DB) *PushSubscriptionRepository {
	return &PushSubscriptionRepository{database: database}
}

func (r *PushSubscriptionRepository) Save(subscription *models.PushSubscription) error {
	// Upsert based on endpoint
	var existing models.PushSubscription
	err := r.database.Where("endpoint = ?", subscription.Endpoint).First(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return r.database.Create(subscription).Error
		}
		return err
	}

	// Update existing
	existing.UserID = subscription.UserID
	existing.P256dh = subscription.P256dh
	existing.Auth = subscription.Auth
	return r.database.Save(&existing).Error
}

func (r *PushSubscriptionRepository) FindByUserID(userID uint) ([]models.PushSubscription, error) {
	var subscriptions []models.PushSubscription
	err := r.database.Where("user_id = ?", userID).Find(&subscriptions).Error
	return subscriptions, err
}

func (r *PushSubscriptionRepository) DeleteByEndpoint(endpoint string) error {
	return r.database.Where("endpoint = ?", endpoint).Delete(&models.PushSubscription{}).Error
}
