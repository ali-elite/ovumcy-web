package db

import (
	"errors"
	"strings"
	"time"

	"github.com/ovumcy/ovumcy-web/internal/models"
	"gorm.io/gorm"
)

type PartnerInvitationRepository struct {
	database *gorm.DB
}

func NewPartnerInvitationRepository(database *gorm.DB) *PartnerInvitationRepository {
	return &PartnerInvitationRepository{database: database}
}

func (repo *PartnerInvitationRepository) Create(invitation *models.PartnerInvitation) error {
	return classifyPartnerInvitationCreateError(repo.database.Create(invitation).Error)
}

func (repo *PartnerInvitationRepository) Save(invitation *models.PartnerInvitation) error {
	return repo.database.Save(invitation).Error
}

func (repo *PartnerInvitationRepository) FindByCodeHash(codeHash string) (models.PartnerInvitation, bool, error) {
	var invitation models.PartnerInvitation
	if err := repo.database.Where("code_hash = ?", strings.TrimSpace(codeHash)).First(&invitation).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.PartnerInvitation{}, false, nil
		}
		return models.PartnerInvitation{}, false, err
	}
	return invitation, true, nil
}

func (repo *PartnerInvitationRepository) FindActiveByOwner(ownerUserID uint) ([]models.PartnerInvitation, error) {
	var invitations []models.PartnerInvitation
	err := repo.database.Where("owner_user_id = ? AND status = ? AND expires_at > ?", ownerUserID, models.PartnerInvitationStatusPending, time.Now()).
		Order("created_at DESC").Find(&invitations).Error
	return invitations, err
}

func (repo *PartnerInvitationRepository) RevokeAllPendingByOwner(ownerUserID uint) error {
	now := time.Now()
	return repo.database.Model(&models.PartnerInvitation{}).
		Where("owner_user_id = ? AND status = ?", ownerUserID, models.PartnerInvitationStatusPending).
		Updates(map[string]interface{}{
			"status":     models.PartnerInvitationStatusRevoked,
			"revoked_at": &now,
		}).Error
}
