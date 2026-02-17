package handlers

import (
	"backend/internal/application"
	"backend/pkg/contracts"

	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	userService application.UserService
}

func NewUserHandler(userService application.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) RegisterRoutes(router fiber.Router) {
	router.Post("/users", h.CreateUser)
}

// CreateUser handles POST /api/v1/users
func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	var request contracts.CreateLocalUser

	// Parse and validate request body
	if err := c.BodyParser(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	// Call userService.CreateLocalUser()
	user, serviceErr := h.userService.CreateLocalUser(c.Context(), request)
	if serviceErr != nil {
		return serviceErr
	}

	// Return 201 Created with success message and user ID
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User created successfully",
		"user_id": user.ID,
	})
}
