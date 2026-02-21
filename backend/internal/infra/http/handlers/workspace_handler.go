package handlers

import (
	"backend/internal/application"
	"backend/internal/infra/http/middleware"
	"backend/pkg/contracts"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type WorkspaceHandler struct {
	workspaceService application.WorkspaceService
}

func NewWorkspaceHandler(workspaceService application.WorkspaceService) *WorkspaceHandler {
	return &WorkspaceHandler{
		workspaceService: workspaceService,
	}
}

func (h *WorkspaceHandler) RegisterRoutes(router fiber.Router) {
	router.Post("/workspaces", h.CreateWorkspace)
	router.Get("/workspaces/admin/:admin_id", h.GetWorkspacesByAdmin)
	router.Get("/workspaces/:id", h.GetWorkspace)
	router.Put("/workspaces/:id", h.UpdateWorkspace)
	router.Delete("/workspaces/:id", h.DeleteWorkspace)
	router.Get("/workspaces", h.ListWorkspaces)
}

// CreateWorkspace handles POST /api/v1/workspaces
func (h *WorkspaceHandler) CreateWorkspace(c *fiber.Ctx) error {
	var request contracts.CreateWorkspace

	if err := c.BodyParser(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	workspace, serviceErr := h.workspaceService.CreateWorkspace(middleware.ContextWithClaims(c), request)
	if serviceErr != nil {
		return serviceErr
	}

	return c.Status(fiber.StatusCreated).JSON(workspace)
}

// GetWorkspace handles GET /api/v1/workspaces/:id
func (h *WorkspaceHandler) GetWorkspace(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid workspace ID")
	}

	workspace, serviceErr := h.workspaceService.GetWorkspace(middleware.ContextWithClaims(c), contracts.GetWorkspace{ID: id})
	if serviceErr != nil {
		return serviceErr
	}

	return c.JSON(workspace)
}

// GetWorkspacesByAdmin handles GET /api/v1/workspaces/admin/:admin_id
func (h *WorkspaceHandler) GetWorkspacesByAdmin(c *fiber.Ctx) error {
	adminID, err := uuid.Parse(c.Params("admin_id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid admin ID")
	}

	workspaces, serviceErr := h.workspaceService.GetWorkspacesByAdmin(middleware.ContextWithClaims(c), contracts.GetWorkspacesByAdmin{AdminID: adminID})
	if serviceErr != nil {
		return serviceErr
	}

	return c.JSON(workspaces)
}

// UpdateWorkspace handles PUT /api/v1/workspaces/:id
func (h *WorkspaceHandler) UpdateWorkspace(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid workspace ID")
	}

	var request contracts.UpdateWorkspace
	if err := c.BodyParser(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	request.ID = id

	workspace, serviceErr := h.workspaceService.UpdateWorkspace(middleware.ContextWithClaims(c), request)
	if serviceErr != nil {
		return serviceErr
	}

	return c.JSON(workspace)
}

// DeleteWorkspace handles DELETE /api/v1/workspaces/:id
func (h *WorkspaceHandler) DeleteWorkspace(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid workspace ID")
	}

	if serviceErr := h.workspaceService.DeleteWorkspace(middleware.ContextWithClaims(c), contracts.DeleteWorkspace{ID: id}); serviceErr != nil {
		return serviceErr
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ListWorkspaces handles GET /api/v1/workspaces
func (h *WorkspaceHandler) ListWorkspaces(c *fiber.Ctx) error {
	var request contracts.ListWorkspaces

	if err := c.QueryParser(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid query parameters")
	}

	workspaces, serviceErr := h.workspaceService.ListWorkspaces(middleware.ContextWithClaims(c), request)
	if serviceErr != nil {
		return serviceErr
	}

	return c.JSON(workspaces)
}
