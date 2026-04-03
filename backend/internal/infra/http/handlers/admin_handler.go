package handlers

import (
	"backend/internal/application"
	handlererrors "backend/internal/application/errors"
	apphandlers "backend/internal/application/handlers"
	"backend/internal/infra/http/middleware"
	"backend/pkg/contracts"
	"backend/pkg/jwt"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type AdminHandler struct {
	serviceFactory func() (*application.AdminService, apphandlers.UnitOfWork)
	jwtService     *jwt.Service
	cookieCfg      jwt.CookieConfig
	adminInitToken string
}

func NewAdminHandler(serviceFactory func() (*application.AdminService, apphandlers.UnitOfWork), jwtService *jwt.Service, adminInitToken string) *AdminHandler {
	return &AdminHandler{
		serviceFactory: serviceFactory,
		jwtService:     jwtService,
		cookieCfg:      jwt.DefaultCookieConfig(),
		adminInitToken: adminInitToken,
	}
}

// InitializeSystem handles POST /admin/init
func (h *AdminHandler) InitializeSystem(c *fiber.Ctx) error {
	// Check optional ADMIN_INIT_TOKEN
	if h.adminInitToken != "" {
		providedToken := c.Get("X-Admin-Init-Token")
		if providedToken != h.adminInitToken {
			return fiber.NewError(fiber.StatusUnauthorized, "Invalid or missing initialization token")
		}
	}

	var request contracts.AdminInit

	// Parse and validate request body
	if err := c.BodyParser(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	// AdminService.InitializeSystem manages the transaction via defer uow.Rollback()
	service, uow := h.serviceFactory()
	response, serviceErr := service.InitializeSystem(c.Context(), uow, request)
	if serviceErr != nil {
		return serviceErr
	}
	token, err := h.jwtService.GenerateToken(response.AdminUserID.String(), response.UserName, "admin", response.WorkspaceID.String())
	if err != nil {
		return err
	}

	// Set JWT cookie
	middleware.SetTokenCookie(c, token, h.cookieCfg)

	// Return 201 Created with response
	return c.Status(fiber.StatusCreated).JSON(response)
}

// GetSystemStatus handles GET /admin/status
func (h *AdminHandler) GetSystemStatus(c *fiber.Ctx) error {
	service, _ := h.serviceFactory()
	initialized, err := service.IsInitialized(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to check system status")
	}
	return c.JSON(fiber.Map{
		"initialized": initialized,
	})
}

// RegisterAdminRoutes registers admin-only user management routes.
func (h *AdminHandler) RegisterAdminRoutes(router fiber.Router) {
	router.Get("/admin/users", h.ListUsers)
	router.Post("/admin/users/invite", h.InviteUser)
	router.Post("/admin/users/:id/reset-password", h.ResetPassword)
	router.Delete("/admin/users/:id", h.DeleteUser)
}

// ListUsers handles GET /admin/users
func (h *AdminHandler) ListUsers(c *fiber.Ctx) error {
	service, _ := h.serviceFactory()
	users, serviceErr := service.ListUsers(middleware.ContextWithClaims(c))
	if serviceErr != nil {
		return serviceErr
	}
	return c.JSON(users)
}

// InviteUser handles POST /admin/users/invite
func (h *AdminHandler) InviteUser(c *fiber.Ctx) error {
	var request contracts.InviteUser
	if err := c.BodyParser(&request); err != nil {
		return handlererrors.ReturnBadRequest("Invalid request body")
	}

	service, uow := h.serviceFactory()
	response, serviceErr := service.InviteUser(middleware.ContextWithClaims(c), uow, request)
	if serviceErr != nil {
		return serviceErr
	}

	return c.Status(fiber.StatusCreated).JSON(response)
}

// ResetPassword handles POST /admin/users/:id/reset-password
func (h *AdminHandler) ResetPassword(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return handlererrors.ReturnBadRequest("invalid user ID")
	}

	service, uow := h.serviceFactory()
	response, serviceErr := service.ResetUserPassword(middleware.ContextWithClaims(c), uow, userID)
	if serviceErr != nil {
		return serviceErr
	}

	return c.JSON(response)
}

// DeleteUser handles DELETE /admin/users/:id
func (h *AdminHandler) DeleteUser(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return handlererrors.ReturnBadRequest("invalid user ID")
	}

	service, uow := h.serviceFactory()
	if serviceErr := service.DeleteUser(middleware.ContextWithClaims(c), uow, userID); serviceErr != nil {
		return serviceErr
	}

	return c.SendStatus(fiber.StatusNoContent)
}
