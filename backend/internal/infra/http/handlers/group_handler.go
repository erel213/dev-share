package handlers

import (
	"backend/internal/application"
	"backend/internal/infra/http/middleware"
	"backend/pkg/contracts"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type GroupHandler struct {
	serviceFactory func() application.GroupService
}

func NewGroupHandler(serviceFactory func() application.GroupService) *GroupHandler {
	return &GroupHandler{serviceFactory: serviceFactory}
}

func (h *GroupHandler) RegisterRoutes(router fiber.Router) {
	router.Post("/groups", h.CreateGroup)
	router.Get("/groups", h.ListGroups)
	router.Get("/groups/:id", h.GetGroup)
	router.Put("/groups/:id", h.UpdateGroup)
	router.Delete("/groups/:id", h.DeleteGroup)
	router.Post("/groups/:id/members", h.AddMembers)
	router.Get("/groups/:id/members", h.GetMembers)
	router.Delete("/groups/:id/members/:user_id", h.RemoveMember)
	router.Post("/groups/:id/templates", h.AddTemplateAccess)
	router.Get("/groups/:id/templates", h.GetTemplateAccess)
	router.Delete("/groups/:id/templates/:template_id", h.RemoveTemplateAccess)
}

func (h *GroupHandler) CreateGroup(c *fiber.Ctx) error {
	var request contracts.CreateGroup
	if err := c.BodyParser(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	service := h.serviceFactory()
	group, serviceErr := service.CreateGroup(middleware.ContextWithClaims(c), request)
	if serviceErr != nil {
		return serviceErr
	}

	return c.Status(fiber.StatusCreated).JSON(group)
}

func (h *GroupHandler) GetGroup(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid group ID")
	}

	service := h.serviceFactory()
	group, serviceErr := service.GetGroup(middleware.ContextWithClaims(c), id)
	if serviceErr != nil {
		return serviceErr
	}

	return c.JSON(group)
}

func (h *GroupHandler) ListGroups(c *fiber.Ctx) error {
	service := h.serviceFactory()
	groups, serviceErr := service.ListGroups(middleware.ContextWithClaims(c))
	if serviceErr != nil {
		return serviceErr
	}

	return c.JSON(groups)
}

func (h *GroupHandler) UpdateGroup(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid group ID")
	}

	var request contracts.UpdateGroup
	if err := c.BodyParser(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}
	request.ID = id

	service := h.serviceFactory()
	group, serviceErr := service.UpdateGroup(middleware.ContextWithClaims(c), id, request)
	if serviceErr != nil {
		return serviceErr
	}

	return c.JSON(group)
}

func (h *GroupHandler) DeleteGroup(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid group ID")
	}

	service := h.serviceFactory()
	if serviceErr := service.DeleteGroup(middleware.ContextWithClaims(c), id); serviceErr != nil {
		return serviceErr
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *GroupHandler) AddMembers(c *fiber.Ctx) error {
	groupID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid group ID")
	}

	var request contracts.AddGroupMembers
	if err := c.BodyParser(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	service := h.serviceFactory()
	if serviceErr := service.AddMembers(middleware.ContextWithClaims(c), groupID, request); serviceErr != nil {
		return serviceErr
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *GroupHandler) GetMembers(c *fiber.Ctx) error {
	groupID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid group ID")
	}

	service := h.serviceFactory()
	members, serviceErr := service.GetMembers(middleware.ContextWithClaims(c), groupID)
	if serviceErr != nil {
		return serviceErr
	}

	return c.JSON(members)
}

func (h *GroupHandler) RemoveMember(c *fiber.Ctx) error {
	groupID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid group ID")
	}

	userID, err := uuid.Parse(c.Params("user_id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid user ID")
	}

	service := h.serviceFactory()
	if serviceErr := service.RemoveMember(middleware.ContextWithClaims(c), groupID, userID); serviceErr != nil {
		return serviceErr
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *GroupHandler) AddTemplateAccess(c *fiber.Ctx) error {
	groupID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid group ID")
	}

	var request contracts.AddGroupTemplateAccess
	if err := c.BodyParser(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	service := h.serviceFactory()
	if serviceErr := service.AddTemplateAccess(middleware.ContextWithClaims(c), groupID, request); serviceErr != nil {
		return serviceErr
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *GroupHandler) GetTemplateAccess(c *fiber.Ctx) error {
	groupID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid group ID")
	}

	service := h.serviceFactory()
	templateIDs, serviceErr := service.GetTemplateAccess(middleware.ContextWithClaims(c), groupID)
	if serviceErr != nil {
		return serviceErr
	}

	return c.JSON(templateIDs)
}

func (h *GroupHandler) RemoveTemplateAccess(c *fiber.Ctx) error {
	groupID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid group ID")
	}

	templateID, err := uuid.Parse(c.Params("template_id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid template ID")
	}

	service := h.serviceFactory()
	if serviceErr := service.RemoveTemplateAccess(middleware.ContextWithClaims(c), groupID, templateID); serviceErr != nil {
		return serviceErr
	}

	return c.SendStatus(fiber.StatusNoContent)
}
