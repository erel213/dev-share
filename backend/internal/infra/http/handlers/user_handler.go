package handlers

import (
	"backend/internal/application"
	apphandlers "backend/internal/application/handlers"
	"backend/internal/infra/http/middleware"
	"backend/pkg/contracts"
	"backend/pkg/jwt"

	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	serviceFactory func() (application.UserService, apphandlers.UnitOfWork)
	jwtService     *jwt.Service
	cookieCfg      jwt.CookieConfig
}

func NewUserHandler(serviceFactory func() (application.UserService, apphandlers.UnitOfWork)) *UserHandler {
	jwtService, err := jwt.NewService()
	if err != nil {
		panic("failed to initialize JWT service: " + err.Error())
	}
	cookieCfg := jwt.DefaultCookieConfig()
	return &UserHandler{
		serviceFactory: serviceFactory,
		jwtService:     jwtService,
		cookieCfg:      cookieCfg,
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

	// UserService.CreateLocalUser does not defer rollback internally (it can be called
	// nested from AdminService), so the handler is responsible for cleanup.
	service, uow := h.serviceFactory()
	defer uow.Rollback()

	// Call userService.CreateLocalUser()
	user, serviceErr := service.CreateLocalUser(c.Context(), uow, request)
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
