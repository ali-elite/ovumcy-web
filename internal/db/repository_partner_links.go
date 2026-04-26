package db

import (
	"errors"

	"github.com/ovumcy/ovumcy-web/internal/models"
	"gorm.io/gorm"
)

type PartnerLinkRepository struct {
	database *gorm.DB
}

func NewPartnerLinkRepository(database *gorm.DB) *PartnerLinkRepository {
	return &PartnerLinkRepository{database: database}
}

func (r *PartnerLinkRepository) Create(link *models.PartnerLink) error {
	return r.database.Create(link).Error
}

func (r *PartnerLinkRepository) FindActivePartners(ownerUserID uint) ([]models.PartnerLink, error) {
	var links []models.PartnerLink
	err := r.database.Where("owner_user_id = ? AND status = ?", ownerUserID, models.PartnerLinkStatusActive).Find(&links).Error
	return links, err
}

func (r *PartnerLinkRepository) FindActiveOwnerForPartner(partnerUserID uint) (models.PartnerLink, bool, error) {
	var link models.PartnerLink
	err := r.database.Where("partner_user_id = ? AND status = ?", partnerUserID, models.PartnerLinkStatusActive).First(&link).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.PartnerLink{}, false, nil
		}
		return models.PartnerLink{}, false, err
	}
	return link, true, nil
}
