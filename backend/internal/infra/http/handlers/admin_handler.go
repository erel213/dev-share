package handlers

import (
	"os"

	"backend/internal/application"
	apphandlers "backend/internal/application/handlers"
	"backend/pkg/contracts"

	"github.com/gofiber/fiber/v2"
)

type AdminHandler struct {
	serviceFactory func() (*application.AdminService, apphandlers.UnitOfWork)
}

func NewAdminHandler(serviceFactory func() (*application.AdminService, apphandlers.UnitOfWork)) *AdminHandler {
	return &AdminHandler{
		serviceFactory: serviceFactory,
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

	// Return 201 Created with response
	return c.Status(fiber.StatusCreated).JSON(response)
}
