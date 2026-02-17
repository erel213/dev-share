package application

import (
	"context"
	"time"

	"backend/internal/domain"
	"backend/internal/domain/repository"
	"backend/pkg/contracts"
	"backend/pkg/errors"
	"backend/pkg/validation"
)

type WorkspaceService struct {
	workspaceRepository repository.WorkspaceRepository
	validator           *validation.Service
}

func NewWorkspaceService(workspaceRepo repository.WorkspaceRepository, validator *validation.Service) WorkspaceService {
	return WorkspaceService{
		workspaceRepository: workspaceRepo,
		validator:           validator,
	}
}

// CreateWorkspace creates a new workspace with the provided details
func (s WorkspaceService) CreateWorkspace(ctx context.Context, request contracts.CreateWorkspace) (*domain.Workspace, *errors.Error) {
	if err := s.validator.Validate(request); err != nil {
		return nil, err
	}

	workspace := domain.NewWorkspace(request.Name, request.Description, request.AdminID)

	if err := s.workspaceRepository.Create(ctx, workspace); err != nil {
		return nil, err
	}

	return workspace, nil
}

// GetWorkspace retrieves a workspace by ID
func (s WorkspaceService) GetWorkspace(ctx context.Context, request contracts.GetWorkspace) (*domain.Workspace, *errors.Error) {
	if err := s.validator.Validate(request); err != nil {
		return nil, err
	}

	return s.workspaceRepository.GetByID(ctx, request.ID)
}

// GetWorkspacesByAdmin retrieves all workspaces for a given admin
func (s WorkspaceService) GetWorkspacesByAdmin(ctx context.Context, request contracts.GetWorkspacesByAdmin) ([]*domain.Workspace, *errors.Error) {
	if err := s.validator.Validate(request); err != nil {
		return nil, err
	}

	return s.workspaceRepository.GetByAdminID(ctx, request.AdminID)
}

// UpdateWorkspace updates an existing workspace
func (s WorkspaceService) UpdateWorkspace(ctx context.Context, request contracts.UpdateWorkspace) (*domain.Workspace, *errors.Error) {
	if err := s.validator.Validate(request); err != nil {
		return nil, err
	}

	// Get existing workspace
	workspace, err := s.workspaceRepository.GetByID(ctx, request.ID)
	if err != nil {
		return nil, err
	}

	// Update non-empty fields
	if request.Name != "" {
		workspace.Name = request.Name
	}
	if request.Description != "" {
		workspace.Description = request.Description
	}

	// Update timestamp
	workspace.UpdatedAt = time.Now()

	// Save changes
	if err := s.workspaceRepository.Update(ctx, workspace); err != nil {
		return nil, err
	}

	return workspace, nil
}

// DeleteWorkspace deletes a workspace by ID
func (s WorkspaceService) DeleteWorkspace(ctx context.Context, request contracts.DeleteWorkspace) *errors.Error {
	if err := s.validator.Validate(request); err != nil {
		return err
	}

	return s.workspaceRepository.Delete(ctx, request.ID)
}

// ListWorkspaces retrieves a paginated list of workspaces
func (s WorkspaceService) ListWorkspaces(ctx context.Context, request contracts.ListWorkspaces) ([]*domain.Workspace, *errors.Error) {
	if err := s.validator.Validate(request); err != nil {
		return nil, err
	}

	opts := repository.ListOptions{
		Limit:  request.Limit,
		Offset: request.Offset,
		SortBy: request.SortBy,
		Order:  request.Order,
	}

	// Apply defaults
	opts.ApplyDefaults()

	// Validate options
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	return s.workspaceRepository.List(ctx, opts)
}
