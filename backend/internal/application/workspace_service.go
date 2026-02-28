package application

import (
	"context"
	"time"

	"backend/internal/application/handlers"
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
func (s WorkspaceService) CreateWorkspace(ctx context.Context, uow handlers.UnitOfWork, request contracts.CreateWorkspace) (*domain.Workspace, *errors.Error) {
	if err := s.validator.Validate(request); err != nil {
		return nil, err
	}

	if err := uow.Begin(); err != nil {
		return nil, err
	}
	defer uow.Rollback()

	workspace := domain.NewWorkspace(request.Name, request.Description, &request.AdminID)

	if err := s.workspaceRepository.Create(ctx, workspace); err != nil {
		return nil, err
	}

	return workspace, uow.Commit()
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
func (s WorkspaceService) UpdateWorkspace(ctx context.Context, uow handlers.UnitOfWork, request contracts.UpdateWorkspace) (*domain.Workspace, *errors.Error) {
	if err := s.validator.Validate(request); err != nil {
		return nil, err
	}

	if err := uow.Begin(); err != nil {
		return nil, err
	}
	defer uow.Rollback()

	workspace, err := s.workspaceRepository.GetByID(ctx, request.ID)
	if err != nil {
		return nil, err
	}

	if request.Name != "" {
		workspace.Name = request.Name
	}
	if request.Description != "" {
		workspace.Description = request.Description
	}

	workspace.UpdatedAt = time.Now()

	if err := s.workspaceRepository.Update(ctx, workspace); err != nil {
		return nil, err
	}

	return workspace, uow.Commit()
}

// DeleteWorkspace deletes a workspace by ID
func (s WorkspaceService) DeleteWorkspace(ctx context.Context, uow handlers.UnitOfWork, request contracts.DeleteWorkspace) *errors.Error {
	if err := s.validator.Validate(request); err != nil {
		return err
	}

	if err := uow.Begin(); err != nil {
		return err
	}
	defer uow.Rollback()

	if err := s.workspaceRepository.Delete(ctx, request.ID); err != nil {
		return err
	}

	return uow.Commit()
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

	opts.ApplyDefaults()

	if err := opts.Validate(); err != nil {
		return nil, err
	}

	return s.workspaceRepository.List(ctx, opts)
}
