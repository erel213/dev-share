package repository

import (
	"context"

	"backend/internal/domain"

	"github.com/google/uuid"
)

type WorkspaceRepository interface {
	Create(ctx context.Context, workspace *domain.Workspace) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Workspace, error)
	GetByAdminID(ctx context.Context, adminID uuid.UUID) ([]*domain.Workspace, error)
	Update(ctx context.Context, workspace *domain.Workspace) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, opts ListOptions) ([]*domain.Workspace, error)
}
