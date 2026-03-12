package handlers

import (
	"backend/internal/application"
	"backend/internal/infra/http/middleware"
	"backend/pkg/contracts"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type EnvironmentVariableValueHandler struct {
	serviceFactory func() application.EnvironmentVariableValueService
}

func NewEnvironmentVariableValueHandler(serviceFactory func() application.EnvironmentVariableValueService) *EnvironmentVariableValueHandler {
	return &EnvironmentVariableValueHandler{serviceFactory: serviceFactory}
}

func (h *EnvironmentVariableValueHandler) RegisterRoutes(router fiber.Router) {
	router.Put("/environments/:id/variables", h.SetVariableValues)
	router.Get("/environments/:id/variables", h.GetVariableValues)
}

func (h *EnvironmentVariableValueHandler) SetVariableValues(c *fiber.Ctx) error {
	environmentID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid environment ID")
	}

	var request contracts.SetEnvironmentVariableValues
	if err := c.BodyParser(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}
	request.EnvironmentID = environmentID

	service := h.serviceFactory()
	if serviceErr := service.SetVariableValues(middleware.ContextWithClaims(c), request); serviceErr != nil {
		return serviceErr
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *EnvironmentVariableValueHandler) GetVariableValues(c *fiber.Ctx) error {
	environmentID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid environment ID")
	}

	service := h.serviceFactory()
	values, serviceErr := service.GetVariableValues(middleware.ContextWithClaims(c), contracts.GetEnvironmentVariableValues{EnvironmentID: environmentID})
	if serviceErr != nil {
		return serviceErr
	}

	return c.JSON(values)
}
