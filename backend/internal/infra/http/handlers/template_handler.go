package handlers

import (
	"backend/internal/application"
	"backend/internal/domain/storage"
	"backend/internal/infra/http/middleware"
	"backend/pkg/contracts"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type TemplateHandler struct {
	serviceFactory func() application.TemplateService
}

func NewTemplateHandler(serviceFactory func() application.TemplateService) *TemplateHandler {
	return &TemplateHandler{
		serviceFactory: serviceFactory,
	}
}

func (h *TemplateHandler) RegisterRoutes(router fiber.Router) {
	router.Post("/templates", h.CreateTemplate)
	router.Get("/templates/workspace/:workspace_id", h.GetTemplatesByWorkspace)
	router.Get("/templates/:id/files/:filename", h.GetTemplateFileContent)
	router.Get("/templates/:id/files", h.ListTemplateFiles)
	router.Get("/templates/:id", h.GetTemplate)
	router.Put("/templates/:id", h.UpdateTemplate)
	router.Delete("/templates/:id", h.DeleteTemplate)
	router.Get("/templates", h.ListTemplates)
}

// CreateTemplate handles POST /api/v1/templates
func (h *TemplateHandler) CreateTemplate(c *fiber.Ctx) error {
	var request contracts.CreateTemplate

	request.Name = c.FormValue("name")
	workspaceIDStr := c.FormValue("workspace_id")
	if workspaceIDStr != "" {
		wid, err := uuid.Parse(workspaceIDStr)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid workspace_id")
		}
		request.WorkspaceID = wid
	}

	// Parse uploaded files
	form, err := c.MultipartForm()
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid multipart form")
	}

	var fileInputs []storage.FileInput
	for _, fh := range form.File["files"] {
		f, err := fh.Open()
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Failed to read uploaded file: "+fh.Filename)
		}
		defer f.Close()

		fileInputs = append(fileInputs, storage.FileInput{
			Name:   fh.Filename,
			Reader: f,
			Size:   fh.Size,
		})
	}

	service := h.serviceFactory()
	template, serviceErr := service.CreateTemplate(middleware.ContextWithClaims(c), request, fileInputs)
	if serviceErr != nil {
		return serviceErr
	}

	return c.Status(fiber.StatusCreated).JSON(template)
}

// GetTemplate handles GET /api/v1/templates/:id
func (h *TemplateHandler) GetTemplate(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid template ID")
	}

	service := h.serviceFactory()
	template, serviceErr := service.GetTemplate(middleware.ContextWithClaims(c), contracts.GetTemplate{ID: id})
	if serviceErr != nil {
		return serviceErr
	}

	return c.JSON(template)
}

// GetTemplatesByWorkspace handles GET /api/v1/templates/workspace/:workspace_id
func (h *TemplateHandler) GetTemplatesByWorkspace(c *fiber.Ctx) error {
	workspaceID, err := uuid.Parse(c.Params("workspace_id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid workspace ID")
	}

	service := h.serviceFactory()
	templates, serviceErr := service.GetTemplatesByWorkspace(middleware.ContextWithClaims(c), contracts.GetTemplatesByWorkspace{WorkspaceID: workspaceID})
	if serviceErr != nil {
		return serviceErr
	}

	return c.JSON(templates)
}

// UpdateTemplate handles PUT /api/v1/templates/:id
func (h *TemplateHandler) UpdateTemplate(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid template ID")
	}

	var request contracts.UpdateTemplate
	request.Name = c.FormValue("name")
	request.ID = id

	// Parse uploaded files
	var fileInputs []storage.FileInput
	form, err := c.MultipartForm()
	if err == nil && form != nil {
		for _, fh := range form.File["files"] {
			f, err := fh.Open()
			if err != nil {
				return fiber.NewError(fiber.StatusBadRequest, "Failed to read uploaded file: "+fh.Filename)
			}
			defer f.Close()

			fileInputs = append(fileInputs, storage.FileInput{
				Name:   fh.Filename,
				Reader: f,
				Size:   fh.Size,
			})
		}
	}

	service := h.serviceFactory()
	template, serviceErr := service.UpdateTemplate(middleware.ContextWithClaims(c), request, fileInputs)
	if serviceErr != nil {
		return serviceErr
	}

	return c.JSON(template)
}

// DeleteTemplate handles DELETE /api/v1/templates/:id
func (h *TemplateHandler) DeleteTemplate(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid template ID")
	}

	service := h.serviceFactory()
	if serviceErr := service.DeleteTemplate(middleware.ContextWithClaims(c), contracts.DeleteTemplate{ID: id}); serviceErr != nil {
		return serviceErr
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ListTemplateFiles handles GET /api/v1/templates/:id/files
func (h *TemplateHandler) ListTemplateFiles(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid template ID")
	}

	service := h.serviceFactory()
	files, serviceErr := service.ListTemplateFiles(middleware.ContextWithClaims(c), contracts.ListTemplateFiles{ID: id})
	if serviceErr != nil {
		return serviceErr
	}

	return c.JSON(files)
}

// GetTemplateFileContent handles GET /api/v1/templates/:id/files/:filename
func (h *TemplateHandler) GetTemplateFileContent(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid template ID")
	}

	filename := c.Params("filename")
	if filename == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Filename is required")
	}

	service := h.serviceFactory()
	content, serviceErr := service.GetTemplateFileContent(middleware.ContextWithClaims(c), contracts.GetTemplateFileContent{ID: id, Filename: filename})
	if serviceErr != nil {
		return serviceErr
	}

	c.Set("Content-Type", "text/plain; charset=utf-8")
	return c.Send(content)
}

// ListTemplates handles GET /api/v1/templates
func (h *TemplateHandler) ListTemplates(c *fiber.Ctx) error {
	var request contracts.ListTemplates

	if err := c.QueryParser(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid query parameters")
	}

	service := h.serviceFactory()
	templates, serviceErr := service.ListTemplates(middleware.ContextWithClaims(c), request)
	if serviceErr != nil {
		return serviceErr
	}

	return c.JSON(templates)
}
