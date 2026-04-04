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

type GroupService struct {
	groupRepo repository.GroupRepository
	validator *validation.Service
}

func NewGroupService(groupRepo repository.GroupRepository, validator *validation.Service) GroupService {
	return GroupService{
		groupRepo: groupRepo,
		validator: validator,
	}
}

func (s GroupService) CreateGroup(ctx context.Context, request contracts.CreateGroup) (*domain.Group, *errors.Error) {
	claims, ok := jwt.ClaimsFromContext(ctx)
	if !ok {
		return nil, apperrors.ReturnUnauthorized("missing JWT claims in context")
	}

	if err := s.validator.Validate(request); err != nil {
		return nil, err
	}

	workspaceID, _ := uuid.Parse(claims.WorkspaceID)
	group := domain.NewGroup(request.Name, request.Description, workspaceID, request.AccessAllTemplates)

	if err := s.groupRepo.Create(ctx, group); err != nil {
		return nil, err
	}

	return group, nil
}

func (s GroupService) GetGroup(ctx context.Context, id uuid.UUID) (*domain.Group, *errors.Error) {
	claims, ok := jwt.ClaimsFromContext(ctx)
	if !ok {
		return nil, apperrors.ReturnUnauthorized("missing JWT claims in context")
	}

	group, err := s.groupRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if group.WorkspaceID.String() != claims.WorkspaceID {
		return nil, apperrors.ReturnForbidden("group does not belong to your workspace")
	}

	return group, nil
}

func (s GroupService) ListGroups(ctx context.Context) ([]*domain.Group, *errors.Error) {
	claims, ok := jwt.ClaimsFromContext(ctx)
	if !ok {
		return nil, apperrors.ReturnUnauthorized("missing JWT claims in context")
	}

	workspaceID, _ := uuid.Parse(claims.WorkspaceID)
	return s.groupRepo.GetByWorkspaceID(ctx, workspaceID)
}

func (s GroupService) UpdateGroup(ctx context.Context, id uuid.UUID, request contracts.UpdateGroup) (*domain.Group, *errors.Error) {
	claims, ok := jwt.ClaimsFromContext(ctx)
	if !ok {
		return nil, apperrors.ReturnUnauthorized("missing JWT claims in context")
	}

	if err := s.validator.Validate(request); err != nil {
		return nil, err
	}

	group, err := s.groupRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if group.WorkspaceID.String() != claims.WorkspaceID {
		return nil, apperrors.ReturnForbidden("group does not belong to your workspace")
	}

	if request.Name != "" {
		group.Name = request.Name
	}
	if request.Description != nil {
		group.Description = *request.Description
	}
	if request.AccessAllTemplates != nil {
		group.AccessAllTemplates = *request.AccessAllTemplates
	}
	group.UpdatedAt = time.Now()

	if err := s.groupRepo.Update(ctx, group); err != nil {
		return nil, err
	}

	return group, nil
}

func (s GroupService) DeleteGroup(ctx context.Context, id uuid.UUID) *errors.Error {
	claims, ok := jwt.ClaimsFromContext(ctx)
	if !ok {
		return apperrors.ReturnUnauthorized("missing JWT claims in context")
	}

	group, err := s.groupRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if group.WorkspaceID.String() != claims.WorkspaceID {
		return apperrors.ReturnForbidden("group does not belong to your workspace")
	}

	return s.groupRepo.Delete(ctx, id)
}

// --- Membership ---

func (s GroupService) AddMembers(ctx context.Context, groupID uuid.UUID, request contracts.AddGroupMembers) *errors.Error {
	if _, err := s.verifyGroupWorkspace(ctx, groupID); err != nil {
		return err
	}

	if valErr := s.validator.Validate(request); valErr != nil {
		return valErr
	}

	return s.groupRepo.AddMembers(ctx, groupID, request.UserIDs)
}

func (s GroupService) RemoveMember(ctx context.Context, groupID uuid.UUID, userID uuid.UUID) *errors.Error {
	if _, err := s.verifyGroupWorkspace(ctx, groupID); err != nil {
		return err
	}

	return s.groupRepo.RemoveMember(ctx, groupID, userID)
}

func (s GroupService) GetMembers(ctx context.Context, groupID uuid.UUID) ([]uuid.UUID, *errors.Error) {
	if _, err := s.verifyGroupWorkspace(ctx, groupID); err != nil {
		return nil, err
	}

	return s.groupRepo.GetMembers(ctx, groupID)
}

// --- Template access ---

func (s GroupService) AddTemplateAccess(ctx context.Context, groupID uuid.UUID, request contracts.AddGroupTemplateAccess) *errors.Error {
	if _, err := s.verifyGroupWorkspace(ctx, groupID); err != nil {
		return err
	}

	if valErr := s.validator.Validate(request); valErr != nil {
		return valErr
	}

	return s.groupRepo.AddTemplateAccess(ctx, groupID, request.TemplateIDs)
}

func (s GroupService) RemoveTemplateAccess(ctx context.Context, groupID uuid.UUID, templateID uuid.UUID) *errors.Error {
	if _, err := s.verifyGroupWorkspace(ctx, groupID); err != nil {
		return err
	}

	return s.groupRepo.RemoveTemplateAccess(ctx, groupID, templateID)
}

func (s GroupService) GetTemplateAccess(ctx context.Context, groupID uuid.UUID) ([]uuid.UUID, *errors.Error) {
	if _, err := s.verifyGroupWorkspace(ctx, groupID); err != nil {
		return nil, err
	}

	return s.groupRepo.GetTemplateAccess(ctx, groupID)
}

// --- Helpers ---

func (s GroupService) verifyGroupWorkspace(ctx context.Context, groupID uuid.UUID) (*domain.Group, *errors.Error) {
	claims, ok := jwt.ClaimsFromContext(ctx)
	if !ok {
		return nil, apperrors.ReturnUnauthorized("missing JWT claims in context")
	}

	group, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		return nil, err
	}

	if group.WorkspaceID.String() != claims.WorkspaceID {
		return nil, apperrors.ReturnForbidden("group does not belong to your workspace")
	}

	return group, nil
}
