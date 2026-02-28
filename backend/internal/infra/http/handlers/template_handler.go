package handlers

import (
	"backend/internal/application"
	"backend/internal/infra/http/middleware"
	"backend/pkg/contracts"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type TemplateHandler struct {
	serviceFactory func() application.TemplateService
}

func NewTemplateHandler(serviceFactory func() application.TemplateService) *TemplateHandler {
	return &TemplateHandler{
		serviceFactory: serviceFactory,
	}
}

func (h *TemplateHandler) RegisterRoutes(router fiber.Router) {
	router.Post("/templates", h.CreateTemplate)
	router.Get("/templates/workspace/:workspace_id", h.GetTemplatesByWorkspace)
	router.Get("/templates/:id", h.GetTemplate)
	router.Put("/templates/:id", h.UpdateTemplate)
	router.Delete("/templates/:id", h.DeleteTemplate)
	router.Get("/templates", h.ListTemplates)
}

// CreateTemplate handles POST /api/v1/templates
func (h *TemplateHandler) CreateTemplate(c *fiber.Ctx) error {
	var request contracts.CreateTemplate

	if err := c.BodyParser(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	service := h.serviceFactory()
	template, serviceErr := service.CreateTemplate(middleware.ContextWithClaims(c), request)
	if serviceErr != nil {
		return serviceErr
	}

	return c.Status(fiber.StatusCreated).JSON(template)
}

// GetTemplate handles GET /api/v1/templates/:id
func (h *TemplateHandler) GetTemplate(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid template ID")
	}

	service := h.serviceFactory()
	template, serviceErr := service.GetTemplate(middleware.ContextWithClaims(c), contracts.GetTemplate{ID: id})
	if serviceErr != nil {
		return serviceErr
	}

	return c.JSON(template)
}

// GetTemplatesByWorkspace handles GET /api/v1/templates/workspace/:workspace_id
func (h *TemplateHandler) GetTemplatesByWorkspace(c *fiber.Ctx) error {
	workspaceID, err := uuid.Parse(c.Params("workspace_id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid workspace ID")
	}

	service := h.serviceFactory()
	templates, serviceErr := service.GetTemplatesByWorkspace(middleware.ContextWithClaims(c), contracts.GetTemplatesByWorkspace{WorkspaceID: workspaceID})
	if serviceErr != nil {
		return serviceErr
	}

	return c.JSON(templates)
}

// UpdateTemplate handles PUT /api/v1/templates/:id
func (h *TemplateHandler) UpdateTemplate(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid template ID")
	}

	var request contracts.UpdateTemplate
	if err := c.BodyParser(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	request.ID = id

	service := h.serviceFactory()
	template, serviceErr := service.UpdateTemplate(middleware.ContextWithClaims(c), request)
	if serviceErr != nil {
		return serviceErr
	}

	return c.JSON(template)
}

// DeleteTemplate handles DELETE /api/v1/templates/:id
func (h *TemplateHandler) DeleteTemplate(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid template ID")
	}

	service := h.serviceFactory()
	if serviceErr := service.DeleteTemplate(middleware.ContextWithClaims(c), contracts.DeleteTemplate{ID: id}); serviceErr != nil {
		return serviceErr
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ListTemplates handles GET /api/v1/templates
func (h *TemplateHandler) ListTemplates(c *fiber.Ctx) error {
	var request contracts.ListTemplates

	if err := c.QueryParser(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid query parameters")
	}

	service := h.serviceFactory()
	templates, serviceErr := service.ListTemplates(middleware.ContextWithClaims(c), request)
	if serviceErr != nil {
		return serviceErr
	}

	return c.JSON(templates)
}
