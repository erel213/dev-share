package handlers

import (
	"backend/internal/application"
	apphandlers "backend/internal/application/handlers"
	"backend/internal/infra/http/middleware"
	"backend/pkg/contracts"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type WorkspaceHandler struct {
	serviceFactory func() (application.WorkspaceService, apphandlers.UnitOfWork)
}

func NewWorkspaceHandler(serviceFactory func() (application.WorkspaceService, apphandlers.UnitOfWork)) *WorkspaceHandler {
	return &WorkspaceHandler{
		serviceFactory: serviceFactory,
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

	service, uow := h.serviceFactory()
	workspace, serviceErr := service.CreateWorkspace(middleware.ContextWithClaims(c), uow, request)
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

	service, _ := h.serviceFactory()
	workspace, serviceErr := service.GetWorkspace(middleware.ContextWithClaims(c), contracts.GetWorkspace{ID: id})
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

	service, _ := h.serviceFactory()
	workspaces, serviceErr := service.GetWorkspacesByAdmin(middleware.ContextWithClaims(c), contracts.GetWorkspacesByAdmin{AdminID: adminID})
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

	service, uow := h.serviceFactory()
	workspace, serviceErr := service.UpdateWorkspace(middleware.ContextWithClaims(c), uow, request)
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

	service, uow := h.serviceFactory()
	if serviceErr := service.DeleteWorkspace(middleware.ContextWithClaims(c), uow, contracts.DeleteWorkspace{ID: id}); serviceErr != nil {
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

	service, _ := h.serviceFactory()
	workspaces, serviceErr := service.ListWorkspaces(middleware.ContextWithClaims(c), request)
	if serviceErr != nil {
		return serviceErr
	}

	return c.JSON(workspaces)
}
