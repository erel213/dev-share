package application

import (
	"context"
	"time"

	apperrors "backend/internal/application/errors"
	"backend/internal/domain"
	"backend/internal/domain/repository"
	"backend/pkg/contracts"
	"backend/pkg/errors"
	"backend/pkg/jwt"
	"backend/pkg/validation"

	"github.com/google/uuid"
)

type TemplateService struct {
	templateRepository  repository.TemplateRepository
	workspaceRepository repository.WorkspaceRepository
	validator           *validation.Service
}

func NewTemplateService(templateRepo repository.TemplateRepository, workspaceRepository repository.WorkspaceRepository, validator *validation.Service) TemplateService {
	return TemplateService{
		templateRepository:  templateRepo,
		workspaceRepository: workspaceRepository,
		validator:           validator,
	}
}

// CreateTemplate creates a new template with the provided details
func (s TemplateService) CreateTemplate(ctx context.Context, request contracts.CreateTemplate) (*domain.Template, *errors.Error) {
	claims, ok := jwt.ClaimsFromContext(ctx)
	if !ok {
		return nil, apperrors.ReturnUnauthorized("missing JWT claims in context")
	}
	if claims.WorkspaceID != request.WorkspaceID.String() {
		return nil, apperrors.ReturnForbidden("user does not belong to the specified workspace")
	}

	if err := s.validator.Validate(request); err != nil {
		return nil, err
	}

	template := domain.NewTemplate(request.Name, request.WorkspaceID, request.Path)

	if err := s.templateRepository.Create(ctx, *template); err != nil {
		return nil, err
	}

	return template, nil
}

// GetTemplate retrieves a template by ID
func (s TemplateService) GetTemplate(ctx context.Context, request contracts.GetTemplate) (*domain.Template, *errors.Error) {
	claims, ok := jwt.ClaimsFromContext(ctx)
	if !ok {
		return nil, apperrors.ReturnUnauthorized("missing JWT claims in context")
	}

	if err := s.validator.Validate(request); err != nil {
		return nil, err
	}

	template, err := s.templateRepository.GetByID(ctx, request.ID)
	if err != nil {
		return nil, err
	}

	// Verify the template belongs to the user's workspace
	if template.WorkspaceID.String() != claims.WorkspaceID {
		return nil, apperrors.ReturnForbidden("template does not belong to your workspace")
	}

	return template, nil
}

// GetTemplatesByWorkspace retrieves all templates for a given workspace
func (s TemplateService) GetTemplatesByWorkspace(ctx context.Context, request contracts.GetTemplatesByWorkspace) ([]*domain.Template, *errors.Error) {
	claims, ok := jwt.ClaimsFromContext(ctx)
	if !ok {
		return nil, apperrors.ReturnUnauthorized("missing JWT claims in context")
	}
	if claims.WorkspaceID != request.WorkspaceID.String() {
		return nil, apperrors.ReturnForbidden("cannot access templates from another workspace")
	}

	if err := s.validator.Validate(request); err != nil {
		return nil, err
	}

	return s.templateRepository.GetByWorkspaceID(ctx, request.WorkspaceID)
}

// UpdateTemplate updates an existing template
func (s TemplateService) UpdateTemplate(ctx context.Context, request contracts.UpdateTemplate) (*domain.Template, *errors.Error) {
	claims, ok := jwt.ClaimsFromContext(ctx)
	if !ok {
		return nil, apperrors.ReturnUnauthorized("missing JWT claims in context")
	}

	if err := s.validator.Validate(request); err != nil {
		return nil, err
	}

	// Get existing template
	template, err := s.templateRepository.GetByID(ctx, request.ID)
	if err != nil {
		return nil, err
	}

	// Verify the template belongs to the user's workspace
	if template.WorkspaceID.String() != claims.WorkspaceID {
		return nil, apperrors.ReturnForbidden("template does not belong to your workspace")
	}

	// Update non-empty fields
	if request.Name != "" {
		template.Name = request.Name
	}
	if request.Path != "" {
		template.Path = request.Path
	}

	// Update timestamp
	template.UpdatedAt = time.Now()

	// Save changes
	if err := s.templateRepository.Update(ctx, *template); err != nil {
		return nil, err
	}

	return template, nil
}

// DeleteTemplate deletes a template by ID
func (s TemplateService) DeleteTemplate(ctx context.Context, request contracts.DeleteTemplate) *errors.Error {
	claims, ok := jwt.ClaimsFromContext(ctx)
	if !ok {
		return apperrors.ReturnUnauthorized("missing JWT claims in context")
	}

	if err := s.validator.Validate(request); err != nil {
		return err
	}

	// Get existing template to verify ownership
	template, err := s.templateRepository.GetByID(ctx, request.ID)
	if err != nil {
		return err
	}

	// Verify the template belongs to the user's workspace
	if template.WorkspaceID.String() != claims.WorkspaceID {
		return apperrors.ReturnForbidden("template does not belong to your workspace")
	}

	return s.templateRepository.Delete(ctx, request.ID)
}

// ListTemplates retrieves a paginated list of templates for the user's workspace
func (s TemplateService) ListTemplates(ctx context.Context, request contracts.ListTemplates) ([]*domain.Template, *errors.Error) {
	claims, ok := jwt.ClaimsFromContext(ctx)
	if !ok {
		return nil, apperrors.ReturnUnauthorized("missing JWT claims in context")
	}

	if err := s.validator.Validate(request); err != nil {
		return nil, err
	}

	// Parse workspace ID from claims
	workspaceID, err := parseWorkspaceID(claims.WorkspaceID)
	if err != nil {
		return nil, apperrors.ReturnInternalError("invalid workspace ID in token")
	}

	// List templates filtered by the user's workspace
	return s.templateRepository.GetByWorkspaceID(ctx, workspaceID)
}

// parseWorkspaceID converts a workspace ID string to uuid.UUID
func parseWorkspaceID(workspaceIDStr string) (uuid.UUID, error) {
	return uuid.Parse(workspaceIDStr)
}
