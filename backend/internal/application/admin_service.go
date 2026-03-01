package application

import (
	"context"

	"backend/internal/application/handlers"
	"backend/internal/domain"
	"backend/internal/domain/repository"
	"backend/pkg/contracts"
	"backend/pkg/errors"
	"backend/pkg/validation"
)

type AdminService struct {
	workspaceRepository repository.WorkspaceRepository
	userService         UserService
	userRepository      repository.UserRepository
	validator           *validation.Service
}

func NewAdminService(
	workspaceRepo repository.WorkspaceRepository,
	userService UserService,
	userRepo repository.UserRepository,
	validator *validation.Service,
) *AdminService {
	return &AdminService{
		workspaceRepository: workspaceRepo,
		userService:         userService,
		userRepository:      userRepo,
		validator:           validator,
	}
}

func (s *AdminService) InitializeSystem(
	ctx context.Context,
	uow handlers.UnitOfWork,
	request contracts.AdminInit,
) (*contracts.AdminInitResponse, *errors.Error) {
	if err := s.validator.Validate(request); err != nil {
		return nil, err
	}

	// Pre-flight check — no transaction needed
	count, err := s.userRepository.Count(ctx)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.WithCode(errors.CodeConflict, "System already initialized").WithHTTPStatus(409)
	}

	if err = uow.Begin(); err != nil {
		return nil, err
	}
	defer uow.Rollback()

	// Direct repo call: workspace created with nil adminID
	workspace := domain.NewWorkspace(request.WorkspaceName, request.WorkspaceDescription, nil)
	if err = s.workspaceRepository.Create(ctx, workspace); err != nil {
		return nil, err
	}

	adminUser, userErr := s.userService.CreateLocalUser(ctx, uow, contracts.CreateLocalUser{
		Name:        request.AdminName,
		Email:       request.AdminEmail,
		Password:    request.AdminPassword,
		WorkspaceID: workspace.ID,
	})
	if userErr != nil {
		return nil, userErr
	}

	// Direct repo call: link admin to workspace
	if err = s.workspaceRepository.UpdateAdminID(ctx, workspace.ID, adminUser.BaseUser.ID); err != nil {
		return nil, err
	}

	if err = uow.Commit(); err != nil { // depth→0, actual DB commit
		return nil, err
	}

	return &contracts.AdminInitResponse{
		Message:     "System initialized successfully",
		WorkspaceID: workspace.ID,
		AdminUserID: adminUser.BaseUser.ID,
	}, nil
}

func (s *AdminService) IsInitialized(ctx context.Context) (bool, error) {
	count, err := s.userRepository.Count(ctx)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
