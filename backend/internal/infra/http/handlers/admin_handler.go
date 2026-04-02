package handlers

import (
	"backend/internal/application"
	apphandlers "backend/internal/application/handlers"
	"backend/internal/infra/http/middleware"
	"backend/pkg/contracts"
	"backend/pkg/jwt"

	"github.com/gofiber/fiber/v2"
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
