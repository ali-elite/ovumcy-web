package api

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

func (handler *Handler) GetSymptoms(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return handler.respondMappedError(c, unauthorizedErrorSpec())
	}
	symptoms, err := handler.symptomService.FetchSymptoms(user.ID)
	if err != nil {
		return handler.respondMappedError(c, symptomsFetchErrorSpec())
	}
	return c.JSON(symptoms)
}

func (handler *Handler) CreateSymptom(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return handler.respondMappedError(c, unauthorizedErrorSpec())
	}

	payload := symptomPayload{}
	if err := c.BodyParser(&payload); err != nil {
		return handler.respondSymptomMutationError(c, user, settingsInvalidInputErrorSpec(), settingsSymptomSectionState{
			Draft: payload,
		})
	}

	symptom, err := handler.symptomService.CreateSymptomForUser(user.ID, payload.Name, payload.Icon, payload.Color)
	if err != nil {
		return handler.respondSymptomMutationError(c, user, mapSymptomCreateError(err), settingsSymptomSectionState{
			Draft: payload,
		})
	}

	if acceptsJSON(c) {
		return c.Status(fiber.StatusCreated).JSON(symptom)
	}
	return handler.respondSymptomMutationSuccess(c, user, fiber.StatusCreated, "symptom_created")
}

func (handler *Handler) UpdateSymptom(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return handler.respondMappedError(c, unauthorizedErrorSpec())
	}

	id, err := parseRequestUint(c.Params("id"))
	if err != nil {
		return handler.respondSymptomMutationError(c, user, invalidSymptomIDErrorSpec(), settingsSymptomSectionState{})
	}

	payload := symptomPayload{}
	if err := c.BodyParser(&payload); err != nil {
		return handler.respondSymptomMutationError(c, user, settingsInvalidInputErrorSpec(), settingsSymptomSectionState{})
	}

	symptom, err := handler.symptomService.UpdateSymptomForUser(user.ID, id, payload.Name, payload.Icon, payload.Color)
	if err != nil {
		return handler.respondSymptomMutationError(c, user, mapSymptomUpdateError(err), settingsSymptomSectionState{})
	}

	if acceptsJSON(c) {
		return c.JSON(symptom)
	}
	return handler.respondSymptomMutationSuccess(c, user, fiber.StatusOK, "symptom_updated")
}

func (handler *Handler) ArchiveSymptom(c *fiber.Ctx) error {
	return handler.archiveSymptom(c)
}

func (handler *Handler) DeleteSymptom(c *fiber.Ctx) error {
	return handler.archiveSymptom(c)
}

func (handler *Handler) RestoreSymptom(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return handler.respondMappedError(c, unauthorizedErrorSpec())
	}

	id, err := parseRequestUint(c.Params("id"))
	if err != nil {
		return handler.respondSymptomMutationError(c, user, invalidSymptomIDErrorSpec(), settingsSymptomSectionState{})
	}
	if err := handler.symptomService.RestoreSymptomForUser(user.ID, id); err != nil {
		return handler.respondSymptomMutationError(c, user, mapSymptomRestoreError(err), settingsSymptomSectionState{})
	}

	if acceptsJSON(c) {
		return c.JSON(fiber.Map{"ok": true})
	}
	return handler.respondSymptomMutationSuccess(c, user, fiber.StatusOK, "symptom_restored")
}

func (handler *Handler) archiveSymptom(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return handler.respondMappedError(c, unauthorizedErrorSpec())
	}

	id, err := parseRequestUint(c.Params("id"))
	if err != nil {
		return handler.respondSymptomMutationError(c, user, invalidSymptomIDErrorSpec(), settingsSymptomSectionState{})
	}
	if err := handler.symptomService.ArchiveSymptomForUser(user.ID, id, time.Now()); err != nil {
		return handler.respondSymptomMutationError(c, user, mapSymptomArchiveError(err), settingsSymptomSectionState{})
	}

	if acceptsJSON(c) {
		return c.JSON(fiber.Map{"ok": true})
	}
	return handler.respondSymptomMutationSuccess(c, user, fiber.StatusOK, "symptom_hidden")
}
