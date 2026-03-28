package services

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ovumcy/ovumcy-web/internal/models"
)

var (
	ErrOperatorUserEmailRequired = errors.New("operator user email is required")
	ErrOperatorUserEmailInvalid  = errors.New("operator user email is invalid")
	ErrOperatorUserNotFound      = errors.New("operator user not found")
	ErrOperatorUserListFailed    = errors.New("operator user list failed")
	ErrOperatorUserLookupFailed  = errors.New("operator user lookup failed")
	ErrOperatorUserDeleteFailed  = errors.New("operator user delete failed")
)

type OperatorUserRepository interface {
	ListOperatorUserSummaries() ([]models.OperatorUserSummary, error)
	FindByNormalizedEmailOptional(email string) (models.User, bool, error)
	DeleteAccountAndRelatedData(userID uint) error
}

type OperatorUserService struct {
	users OperatorUserRepository
}

func NewOperatorUserService(users OperatorUserRepository) *OperatorUserService {
	return &OperatorUserService{users: users}
}

func (service *OperatorUserService) ListUsers() ([]models.OperatorUserSummary, error) {
	users, err := service.users.ListOperatorUserSummaries()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOperatorUserListFailed, err)
	}
	return users, nil
}

func (service *OperatorUserService) GetUserByEmail(email string) (models.OperatorUserSummary, error) {
	normalizedEmail, err := normalizeOperatorUserEmail(email)
	if err != nil {
		return models.OperatorUserSummary{}, err
	}

	user, found, lookupErr := service.users.FindByNormalizedEmailOptional(normalizedEmail)
	if lookupErr != nil {
		return models.OperatorUserSummary{}, fmt.Errorf("%w: %v", ErrOperatorUserLookupFailed, lookupErr)
	}
	if !found {
		return models.OperatorUserSummary{}, ErrOperatorUserNotFound
	}

	return operatorUserSummaryFromUser(user), nil
}

func (service *OperatorUserService) DeleteUserByEmail(email string) (models.OperatorUserSummary, error) {
	userSummary, err := service.GetUserByEmail(email)
	if err != nil {
		return models.OperatorUserSummary{}, err
	}

	if deleteErr := service.users.DeleteAccountAndRelatedData(userSummary.ID); deleteErr != nil {
		return models.OperatorUserSummary{}, fmt.Errorf("%w: %v", ErrOperatorUserDeleteFailed, deleteErr)
	}

	return userSummary, nil
}

func normalizeOperatorUserEmail(email string) (string, error) {
	trimmedRaw := strings.TrimSpace(email)
	if trimmedRaw == "" {
		return "", ErrOperatorUserEmailRequired
	}

	normalized := NormalizeAuthEmail(trimmedRaw)
	if normalized == "" {
		return "", ErrOperatorUserEmailInvalid
	}
	return normalized, nil
}

func operatorUserSummaryFromUser(user models.User) models.OperatorUserSummary {
	return models.OperatorUserSummary{
		ID:                  user.ID,
		DisplayName:         user.DisplayName,
		Email:               user.Email,
		Role:                user.Role,
		OnboardingCompleted: user.OnboardingCompleted,
		CreatedAt:           user.CreatedAt,
	}
}
