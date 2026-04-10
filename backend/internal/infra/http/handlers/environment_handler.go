package handlers

import (
	"backend/internal/application"
	"backend/internal/infra/http/middleware"
	"backend/pkg/contracts"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type EnvironmentHandler struct {
	serviceFactory func() application.EnvironmentService
}

func NewEnvironmentHandler(serviceFactory func() application.EnvironmentService) *EnvironmentHandler {
	return &EnvironmentHandler{serviceFactory: serviceFactory}
}

func (h *EnvironmentHandler) RegisterRoutes(router fiber.Router) {
	router.Post("/environments", h.CreateEnvironment)
	router.Get("/environments/:id", h.GetEnvironment)
	router.Get("/environments/:id/outputs", h.GetEnvironmentOutputs)
	router.Get("/environments", h.ListEnvironments)
	router.Post("/environments/:id/plan", h.PlanEnvironment)
	router.Post("/environments/:id/apply", h.ApplyEnvironment)
	router.Post("/environments/:id/destroy", h.DestroyEnvironment)
	router.Delete("/environments/:id", h.DeleteEnvironment)
}

func (h *EnvironmentHandler) CreateEnvironment(c *fiber.Ctx) error {
	var request contracts.CreateEnvironment
	if err := c.BodyParser(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	service := h.serviceFactory()
	env, serviceErr := service.CreateEnvironment(middleware.ContextWithClaims(c), request)
	if serviceErr != nil {
		return serviceErr
	}

	return c.Status(fiber.StatusCreated).JSON(env)
}

func (h *EnvironmentHandler) GetEnvironment(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid environment ID")
	}

	service := h.serviceFactory()
	env, serviceErr := service.GetEnvironment(middleware.ContextWithClaims(c), contracts.GetEnvironment{ID: id})
	if serviceErr != nil {
		return serviceErr
	}

	return c.JSON(env)
}

func (h *EnvironmentHandler) ListEnvironments(c *fiber.Ctx) error {
	var request contracts.ListEnvironments
	if err := c.QueryParser(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid query parameters")
	}

	service := h.serviceFactory()
	envs, serviceErr := service.ListEnvironments(middleware.ContextWithClaims(c), request)
	if serviceErr != nil {
		return serviceErr
	}

	return c.JSON(envs)
}

func (h *EnvironmentHandler) PlanEnvironment(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid environment ID")
	}

	service := h.serviceFactory()
	env, serviceErr := service.PlanEnvironment(middleware.ContextWithClaims(c), contracts.PlanEnvironment{ID: id})
	if serviceErr != nil {
		return serviceErr
	}

	return c.Status(fiber.StatusAccepted).JSON(env)
}

func (h *EnvironmentHandler) ApplyEnvironment(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid environment ID")
	}

	service := h.serviceFactory()
	env, serviceErr := service.ApplyEnvironment(middleware.ContextWithClaims(c), contracts.ApplyEnvironment{ID: id})
	if serviceErr != nil {
		return serviceErr
	}

	return c.Status(fiber.StatusAccepted).JSON(env)
}

func (h *EnvironmentHandler) DestroyEnvironment(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid environment ID")
	}

	service := h.serviceFactory()
	env, serviceErr := service.DestroyEnvironment(middleware.ContextWithClaims(c), contracts.DestroyEnvironment{ID: id})
	if serviceErr != nil {
		return serviceErr
	}

	return c.Status(fiber.StatusAccepted).JSON(env)
}

func (h *EnvironmentHandler) GetEnvironmentOutputs(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid environment ID")
	}

	service := h.serviceFactory()
	outputs, serviceErr := service.GetEnvironmentOutputs(middleware.ContextWithClaims(c), contracts.GetEnvironmentOutputs{ID: id})
	if serviceErr != nil {
		return serviceErr
	}

	return c.JSON(outputs)
}

func (h *EnvironmentHandler) DeleteEnvironment(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid environment ID")
	}

	service := h.serviceFactory()
	if serviceErr := service.DeleteEnvironment(middleware.ContextWithClaims(c), contracts.DeleteEnvironment{ID: id}); serviceErr != nil {
		return serviceErr
	}

	return c.SendStatus(fiber.StatusNoContent)
}
