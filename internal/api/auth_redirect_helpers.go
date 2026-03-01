package api

import (
	"github.com/terraincognita07/ovumcy/internal/models"
	"github.com/terraincognita07/ovumcy/internal/services"
)

func postLoginRedirectPath(user *models.User) string {
	return services.PostLoginRedirectPath(user)
}

func requiresOnboarding(user *models.User) bool {
	return services.RequiresOnboarding(user)
}
