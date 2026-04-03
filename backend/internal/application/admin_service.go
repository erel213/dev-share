package application

import (
	"context"

	"backend/internal/application/handlers"
	"backend/internal/domain"
	domainerrors "backend/internal/domain/errors"
	"backend/internal/domain/repository"
	"backend/pkg/contracts"
	"backend/pkg/errors"
	"backend/pkg/jwt"
	"backend/pkg/validation"

	"github.com/google/uuid"
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

	// Mark the admin user
	adminUser.Role = domain.RoleAdmin
	if err = s.userRepository.Update(ctx, adminUser); err != nil {
		return nil, err
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
		UserName:    adminUser.Name,
	}, nil
}

func (s *AdminService) IsInitialized(ctx context.Context) (bool, error) {
	count, err := s.userRepository.Count(ctx)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *AdminService) InviteUser(
	ctx context.Context,
	uow handlers.UnitOfWork,
	request contracts.InviteUser,
) (*contracts.InviteUserResponse, *errors.Error) {
	if err := s.validator.Validate(request); err != nil {
		return nil, err
	}
	context, ok := jwt.ClaimsFromContext(ctx)
	if !ok {
		return nil, errors.WithCode(errors.CodeUnauthorized, "missing JWT claims in context").WithHTTPStatus(401)
	}
	callerWorkspaceID, prsErr := uuid.Parse(context.WorkspaceID)
	if prsErr != nil {
		return nil, errors.WithCode(errors.CodeUnauthorized, "invalid workspace ID in JWT claims").WithHTTPStatus(401)
	}

	// Generate random password
	plainPassword, genErr := domain.GenerateRandomPassword(16)
	if genErr != nil {
		return nil, errors.Wrap(genErr, "failed to generate password").WithHTTPStatus(500)
	}

	if beginErr := uow.Begin(); beginErr != nil {
		return nil, beginErr
	}
	defer uow.Rollback()

	user, createErr := s.userService.CreateLocalUser(ctx, uow, contracts.CreateLocalUser{
		Name:        request.Name,
		Email:       request.Email,
		Password:    plainPassword,
		WorkspaceID: callerWorkspaceID,
	})
	if createErr != nil {
		return nil, createErr
	}

	// Set the requested role (CreateLocalUser always sets RoleUser)
	user.Role = domain.Role(request.Role)
	if err := s.userRepository.Update(ctx, user); err != nil {
		return nil, err
	}

	if err := uow.Commit(); err != nil {
		return nil, err
	}

	return &contracts.InviteUserResponse{
		UserID:   user.ID,
		Name:     user.Name,
		Email:    user.Email,
		Role:     request.Role,
		Password: plainPassword,
	}, nil
}

func (s *AdminService) ResetUserPassword(
	ctx context.Context,
	uow handlers.UnitOfWork,
	userID uuid.UUID,
) (*contracts.ResetPasswordResponse, *errors.Error) {
	user, err := s.userRepository.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	claims, ok := jwt.ClaimsFromContext(ctx)
	if !ok {
		return nil, errors.WithCode(errors.CodeUnauthorized, "missing JWT claims in context").WithHTTPStatus(401)
	}

	if claims.WorkspaceID != user.WorkspaceID.String() {
		return nil, errors.WithCode(errors.CodeForbidden, "Forbidden").WithHTTPStatus(403)
	}

	if user.LocalUser == nil {
		return nil, domainerrors.InvalidInput("user", "cannot reset password for OAuth user")
	}

	plainPassword, genErr := domain.GenerateRandomPassword(16)
	if genErr != nil {
		return nil, errors.Wrap(genErr, "failed to generate password").WithHTTPStatus(500)
	}

	localUser, hashErr := domain.NewLocalUser(plainPassword)
	if hashErr != nil {
		return nil, hashErr
	}

	if beginErr := uow.Begin(); beginErr != nil {
		return nil, beginErr
	}
	defer uow.Rollback()

	user.LocalUser.Password = localUser.Password
	if err := s.userRepository.Update(ctx, *user); err != nil {
		return nil, err
	}

	if err := uow.Commit(); err != nil {
		return nil, err
	}

	return &contracts.ResetPasswordResponse{
		UserID:   userID,
		Password: plainPassword,
	}, nil
}

func (s *AdminService) ListUsers(ctx context.Context) ([]*contracts.AdminUserResponse, *errors.Error) {
	claims, ok := jwt.ClaimsFromContext(ctx)
	if !ok {
		return nil, errors.WithCode(errors.CodeUnauthorized, "missing JWT claims in context").WithHTTPStatus(401)
	}
	filterBy := make(map[string]string)
	filterBy["workspace_id"] = claims.WorkspaceID
	users, err := s.userRepository.List(ctx, repository.ListOptions{
		Limit:    1000,
		SortBy:   "created_at",
		FilterBy: filterBy,
		Order:    "DESC",
	})
	if err != nil {
		return nil, err
	}

	result := make([]*contracts.AdminUserResponse, len(users))
	for i, u := range users {
		result[i] = &contracts.AdminUserResponse{
			ID:          u.ID,
			Name:        u.Name,
			Email:       u.Email,
			Role:        string(u.Role),
			WorkspaceID: u.WorkspaceID,
			CreatedAt:   u.CreatedAt,
			UpdatedAt:   u.UpdatedAt,
		}
	}
	return result, nil
}

func (s *AdminService) DeleteUser(
	ctx context.Context,
	uow handlers.UnitOfWork,
	userID uuid.UUID,
) *errors.Error {
	claims, ok := jwt.ClaimsFromContext(ctx)
	if !ok {
		return errors.WithCode(errors.CodeUnauthorized, "missing JWT claims in context").WithHTTPStatus(401)
	}
	callerId, prsErr := uuid.Parse(claims.ID)
	if prsErr != nil {
		return errors.WithCode(errors.CodeUnauthorized, "invalid user ID in JWT claims").WithHTTPStatus(401)
	}
	if userID == callerId {
		return domainerrors.InvalidInput("user_id", "cannot delete yourself")
	}

	callerUser, err := s.userRepository.GetByID(ctx, callerId)
	if err != nil {
		return err
	}
	if callerUser.WorkspaceID.String() != claims.WorkspaceID {
		return errors.WithCode(errors.CodeForbidden, "Forbidden").WithHTTPStatus(403)
	}

	// Verify user exists
	if _, err := s.userRepository.GetByID(ctx, userID); err != nil {
		return err
	}

	if err := uow.Begin(); err != nil {
		return err
	}
	defer uow.Rollback()

	if err := s.userRepository.Delete(ctx, userID); err != nil {
		return err
	}

	if err := uow.Commit(); err != nil {
		return err
	}

	return nil
}
