package handlers

import (
	"backend/internal/application"
	"backend/internal/infra/http/middleware"
	"backend/pkg/contracts"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type TemplateVariableHandler struct {
	serviceFactory func() application.TemplateVariableService
}

func NewTemplateVariableHandler(serviceFactory func() application.TemplateVariableService) *TemplateVariableHandler {
	return &TemplateVariableHandler{serviceFactory: serviceFactory}
}

func (h *TemplateVariableHandler) RegisterRoutes(router fiber.Router) {
	router.Post("/templates/:id/variables", h.CreateVariable)
	router.Get("/templates/:id/variables", h.ListVariables)
	router.Put("/templates/:id/variables/:varId", h.UpdateVariable)
	router.Delete("/templates/:id/variables/:varId", h.DeleteVariable)
	router.Post("/templates/:id/variables/parse", h.ParseAndReconcileVariables)
}

func (h *TemplateVariableHandler) CreateVariable(c *fiber.Ctx) error {
	templateID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid template ID")
	}

	var request contracts.CreateTemplateVariable
	if err := c.BodyParser(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}
	request.TemplateID = templateID

	service := h.serviceFactory()
	variable, serviceErr := service.CreateVariable(middleware.ContextWithClaims(c), request)
	if serviceErr != nil {
		return serviceErr
	}

	return c.Status(fiber.StatusCreated).JSON(variable)
}

func (h *TemplateVariableHandler) ListVariables(c *fiber.Ctx) error {
	templateID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid template ID")
	}

	service := h.serviceFactory()
	variables, serviceErr := service.ListVariables(middleware.ContextWithClaims(c), contracts.GetTemplateVariables{TemplateID: templateID})
	if serviceErr != nil {
		return serviceErr
	}

	return c.JSON(variables)
}

func (h *TemplateVariableHandler) UpdateVariable(c *fiber.Ctx) error {
	varID, err := uuid.Parse(c.Params("varId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid variable ID")
	}

	var request contracts.UpdateTemplateVariable
	if err := c.BodyParser(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}
	request.ID = varID

	service := h.serviceFactory()
	variable, serviceErr := service.UpdateVariable(middleware.ContextWithClaims(c), request)
	if serviceErr != nil {
		return serviceErr
	}

	return c.JSON(variable)
}

func (h *TemplateVariableHandler) DeleteVariable(c *fiber.Ctx) error {
	varID, err := uuid.Parse(c.Params("varId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid variable ID")
	}

	service := h.serviceFactory()
	if serviceErr := service.DeleteVariable(middleware.ContextWithClaims(c), contracts.DeleteTemplateVariable{ID: varID}); serviceErr != nil {
		return serviceErr
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *TemplateVariableHandler) ParseAndReconcileVariables(c *fiber.Ctx) error {
	templateID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid template ID")
	}

	service := h.serviceFactory()
	result, serviceErr := service.ParseAndReconcileVariables(middleware.ContextWithClaims(c), contracts.ParseTemplateVariables{TemplateID: templateID})
	if serviceErr != nil {
		return serviceErr
	}

	return c.JSON(result)
}
