package handlers

import (
	"backend/internal/application"
	"backend/internal/infra/http/middleware"
	"backend/pkg/contracts"
	"backend/pkg/jwt"

	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	userService application.UserService
	jwtService  *jwt.Service
	cookieCfg   jwt.CookieConfig
}

func NewUserHandler(userService application.UserService, jwtService *jwt.Service, cookieCfg jwt.CookieConfig) *UserHandler {
	return &UserHandler{
		userService: userService,
		jwtService:  jwtService,
		cookieCfg:   cookieCfg,
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

	user, serviceErr := h.userService.CreateLocalUser(c.Context(), request)
	if serviceErr != nil {
		return serviceErr
	}

	token, err := h.jwtService.GenerateToken(user.ID.String(), user.Name, user.WorkspaceID.String())
	if err != nil {
		return err
	}

	middleware.SetTokenCookie(c, token, h.cookieCfg)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User created successfully",
		"user_id": user.ID,
	})
}
