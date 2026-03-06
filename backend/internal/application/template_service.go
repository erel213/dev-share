package application

import (
	"context"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	apperrors "backend/internal/application/errors"
	"backend/internal/domain"
	"backend/internal/domain/repository"
	"backend/internal/domain/storage"
	"backend/pkg/contracts"
	"backend/pkg/errors"
	"backend/pkg/jwt"
	"backend/pkg/validation"

	"github.com/google/uuid"
)

var allowedExtensions = map[string]bool{
	".tf":     true,
	".tfvars": true,
	".hcl":    true,
	".json":   true,
}

const maxFileSize int64 = 1 * 1024 * 1024 // 1MB

type TemplateService struct {
	templateRepository  repository.TemplateRepository
	workspaceRepository repository.WorkspaceRepository
	validator           *validation.Service
	fileStorage         storage.FileStorage
}

func NewTemplateService(templateRepo repository.TemplateRepository, workspaceRepository repository.WorkspaceRepository, validator *validation.Service, fileStorage storage.FileStorage) TemplateService {
	return TemplateService{
		templateRepository:  templateRepo,
		workspaceRepository: workspaceRepository,
		validator:           validator,
		fileStorage:         fileStorage,
	}
}

// CreateTemplate creates a new template with the provided details and uploaded files
func (s TemplateService) CreateTemplate(ctx context.Context, request contracts.CreateTemplate, files []storage.FileInput) (*domain.Template, *errors.Error) {
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

	// Validate files
	if len(files) == 0 {
		return nil, apperrors.ReturnBadRequest("at least one file is required")
	}

	for _, f := range files {
		// Check for path traversal
		if strings.Contains(f.Name, "..") || strings.Contains(f.Name, "/") || strings.Contains(f.Name, "\\") {
			return nil, apperrors.ReturnBadRequest("invalid file name: " + f.Name)
		}

		// Check allowed extensions
		ext := strings.ToLower(filepath.Ext(f.Name))
		if !allowedExtensions[ext] {
			return nil, apperrors.ReturnBadRequest("file extension not allowed: " + ext + " (allowed: .tf, .tfvars, .hcl, .json)")
		}

		// Check file size
		if f.Size > maxFileSize {
			return nil, apperrors.ReturnBadRequest("file too large: " + f.Name + " (max 1MB)")
		}
	}

	template := domain.NewTemplate(request.Name, request.WorkspaceID)

	// Save files to storage
	if err := s.fileStorage.SaveFiles(template.Path, files); err != nil {
		return nil, err
	}

	// Save to DB; on failure, cleanup files
	if err := s.templateRepository.Create(ctx, *template); err != nil {
		if cleanupErr := s.fileStorage.DeleteDir(template.Path); cleanupErr != nil {
			slog.Error("failed to cleanup files after DB error", "path", template.Path, "error", cleanupErr)
		}
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

// UpdateTemplate updates an existing template and optionally adds files
func (s TemplateService) UpdateTemplate(ctx context.Context, request contracts.UpdateTemplate, files []storage.FileInput) (*domain.Template, *errors.Error) {
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

	// Validate and save additional files
	for _, f := range files {
		if strings.Contains(f.Name, "..") || strings.Contains(f.Name, "/") || strings.Contains(f.Name, "\\") {
			return nil, apperrors.ReturnBadRequest("invalid file name: " + f.Name)
		}

		ext := strings.ToLower(filepath.Ext(f.Name))
		if !allowedExtensions[ext] {
			return nil, apperrors.ReturnBadRequest("file extension not allowed: " + ext + " (allowed: .tf, .tfvars, .hcl, .json)")
		}

		if f.Size > maxFileSize {
			return nil, apperrors.ReturnBadRequest("file too large: " + f.Name + " (max 1MB)")
		}
	}

	if len(files) > 0 {
		if err := s.fileStorage.SaveFiles(template.Path, files); err != nil {
			return nil, err
		}
	}

	// Update non-empty fields
	if request.Name != "" {
		template.Name = request.Name
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

	if err := s.templateRepository.Delete(ctx, request.ID); err != nil {
		return err
	}

	// Best-effort cleanup of files
	if cleanupErr := s.fileStorage.DeleteDir(template.Path); cleanupErr != nil {
		slog.Error("failed to cleanup template files after delete", "path", template.Path, "error", cleanupErr)
	}

	return nil
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

// ListTemplateFiles returns the list of files for a given template
func (s TemplateService) ListTemplateFiles(ctx context.Context, request contracts.ListTemplateFiles) ([]contracts.TemplateFileInfo, *errors.Error) {
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

	if template.WorkspaceID.String() != claims.WorkspaceID {
		return nil, apperrors.ReturnForbidden("template does not belong to your workspace")
	}

	files, err := s.fileStorage.ListFiles(template.Path)
	if err != nil {
		return nil, err
	}

	result := make([]contracts.TemplateFileInfo, len(files))
	for i, f := range files {
		result[i] = contracts.TemplateFileInfo{Name: f.Name, Size: f.Size}
	}

	return result, nil
}

// GetTemplateFileContent returns the content of a specific file within a template
func (s TemplateService) GetTemplateFileContent(ctx context.Context, request contracts.GetTemplateFileContent) ([]byte, *errors.Error) {
	claims, ok := jwt.ClaimsFromContext(ctx)
	if !ok {
		return nil, apperrors.ReturnUnauthorized("missing JWT claims in context")
	}

	if err := s.validator.Validate(request); err != nil {
		return nil, err
	}

	// Reject path traversal attempts
	if strings.Contains(request.Filename, "..") || strings.Contains(request.Filename, "/") || strings.Contains(request.Filename, "\\") {
		return nil, apperrors.ReturnBadRequest("invalid filename")
	}

	// Verify allowed extension
	ext := strings.ToLower(filepath.Ext(request.Filename))
	if !allowedExtensions[ext] {
		return nil, apperrors.ReturnBadRequest("file extension not allowed: " + ext)
	}

	template, err := s.templateRepository.GetByID(ctx, request.ID)
	if err != nil {
		return nil, err
	}

	if template.WorkspaceID.String() != claims.WorkspaceID {
		return nil, apperrors.ReturnForbidden("template does not belong to your workspace")
	}

	return s.fileStorage.ReadFile(filepath.Join(template.Path, request.Filename))
}

// parseWorkspaceID converts a workspace ID string to uuid.UUID
func parseWorkspaceID(workspaceIDStr string) (uuid.UUID, error) {
	return uuid.Parse(workspaceIDStr)
}
