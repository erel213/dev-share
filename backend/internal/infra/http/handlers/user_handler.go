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

func NewUserHandler(serviceFactory func() (application.UserService, apphandlers.UnitOfWork), jwtService *jwt.Service) *UserHandler {
	return &UserHandler{
		serviceFactory: serviceFactory,
		jwtService:     jwtService,
		cookieCfg:      jwt.DefaultCookieConfig(),
	}
}

func (h *UserHandler) RegisterRoutes(router fiber.Router) {
	router.Post("/users", h.CreateUser)
	router.Post("/login", h.Login)
}

func (h *UserHandler) RegisterProtectedRoutes(router fiber.Router) {
	router.Get("/me", h.Me)
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

	token, err := h.jwtService.GenerateToken(user.ID.String(), user.Name, string(user.Role), user.WorkspaceID.String())
	if err != nil {
		return err
	}

	middleware.SetTokenCookie(c, token, h.cookieCfg)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User created successfully",
		"user_id": user.ID,
	})
}

// Login handles POST /api/v1/login
func (h *UserHandler) Login(c *fiber.Ctx) error {
	var request contracts.LoginLocalUser

	if err := c.BodyParser(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	service, _ := h.serviceFactory()

	user, serviceErr := service.AuthenticateLocalUser(c.Context(), request)
	if serviceErr != nil {
		return serviceErr
	}

	token, err := h.jwtService.GenerateToken(user.UserID.String(), user.Name, user.Role, user.WorkspaceID.String())
	if err != nil {
		return err
	}

	middleware.SetTokenCookie(c, token, h.cookieCfg)

	return c.Status(fiber.StatusOK).JSON(user)
}

// Me handles GET /api/v1/me
func (h *UserHandler) Me(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "missing claims")
	}

	return c.JSON(fiber.Map{
		"user_id":      claims.ID,
		"name":         claims.Name,
		"role":         claims.Role,
		"workspace_id": claims.WorkspaceID,
	})
}
