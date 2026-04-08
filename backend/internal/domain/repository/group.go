package repository

import (
	"context"

	"backend/internal/domain"
	"backend/pkg/errors"

	"github.com/google/uuid"
)

type GroupRepository interface {
	// Group CRUD
	Create(ctx context.Context, group *domain.Group) *errors.Error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Group, *errors.Error)
	GetByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) ([]*domain.Group, *errors.Error)
	Update(ctx context.Context, group *domain.Group) *errors.Error
	Delete(ctx context.Context, id uuid.UUID) *errors.Error

	// Membership
	AddMembers(ctx context.Context, groupID uuid.UUID, userIDs []uuid.UUID) *errors.Error
	RemoveMember(ctx context.Context, groupID uuid.UUID, userID uuid.UUID) *errors.Error
	GetMembers(ctx context.Context, groupID uuid.UUID) ([]uuid.UUID, *errors.Error)
	GetGroupIDsForUser(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, *errors.Error)

	// Template access
	AddTemplateAccess(ctx context.Context, groupID uuid.UUID, templateIDs []uuid.UUID) *errors.Error
	RemoveTemplateAccess(ctx context.Context, groupID uuid.UUID, templateID uuid.UUID) *errors.Error
	GetTemplateAccess(ctx context.Context, groupID uuid.UUID) ([]uuid.UUID, *errors.Error)

	// Access query — returns the set of template IDs accessible to a user via their groups.
	// If hasAccessAll is true, the user belongs to a group with access_all_templates=true.
	GetAccessibleTemplateIDs(ctx context.Context, userID uuid.UUID, workspaceID uuid.UUID) (templateIDs []uuid.UUID, hasAccessAll bool, err *errors.Error)

	// GetCoMemberUserIDs returns the distinct user IDs of all users who share
	// at least one group with the given user in the specified workspace.
	GetCoMemberUserIDs(ctx context.Context, userID uuid.UUID, workspaceID uuid.UUID) ([]uuid.UUID, *errors.Error)
}
