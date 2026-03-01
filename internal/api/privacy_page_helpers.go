package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/ovumcy/internal/models"
	"github.com/terraincognita07/ovumcy/internal/services"
)

func buildPrivacyPageData(messages map[string]string, backQuery string, user *models.User) fiber.Map {
	navigation := services.BuildPrivacyBackNavigation(backQuery, user != nil)
	data := fiber.Map{
		"Title":                  localizedPageTitle(messages, "meta.title.privacy", "Ovumcy | Privacy Policy"),
		"MetaDescription":        services.ResolvePrivacyMetaDescription(translateMessage(messages, "meta.description.privacy")),
		"BackPath":               navigation.BackPath,
		"BreadcrumbBackLabelKey": navigation.BreadcrumbBackLabelKey,
	}

	if user != nil {
		data["CurrentUser"] = user
	}
	return data
}
