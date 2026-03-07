package handlers

import (
	"os"

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
}

func NewAdminHandler(serviceFactory func() (*application.AdminService, apphandlers.UnitOfWork)) *AdminHandler {
	jwtService, err := jwt.NewService()
	if err != nil {
		panic("failed to initialize JWT service: " + err.Error())
	}
	cookieCfg := jwt.DefaultCookieConfig()
	return &AdminHandler{
		serviceFactory: serviceFactory,
		jwtService:     jwtService,
		cookieCfg:      cookieCfg,
	}
}

// InitializeSystem handles POST /admin/init
func (h *AdminHandler) InitializeSystem(c *fiber.Ctx) error {
	// Check optional ADMIN_INIT_TOKEN
	// TODO: in the future we should consider temporary token approaches
	expectedToken := os.Getenv("ADMIN_INIT_TOKEN")
	if expectedToken != "" {
		providedToken := c.Get("X-Admin-Init-Token")
		if providedToken != expectedToken {
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
	token, err := h.jwtService.GenerateToken(response.AdminUserID.String(), response.UserName, response.WorkspaceID.String())
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
