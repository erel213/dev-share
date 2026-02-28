package repository

import (
	"context"

	"backend/internal/domain"
	"backend/pkg/errors"

	"github.com/google/uuid"
)

type WorkspaceRepository interface {
	Create(ctx context.Context, workspace *domain.Workspace) *errors.Error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Workspace, *errors.Error)
	GetByAdminID(ctx context.Context, adminID uuid.UUID) ([]*domain.Workspace, *errors.Error)
	Update(ctx context.Context, workspace *domain.Workspace) *errors.Error
	Delete(ctx context.Context, id uuid.UUID) *errors.Error
	List(ctx context.Context, opts ListOptions) ([]*domain.Workspace, *errors.Error)
	UpdateAdminID(ctx context.Context, workspaceID uuid.UUID, adminID uuid.UUID) *errors.Error
}
