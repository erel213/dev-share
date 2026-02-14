package repository

import (
	"context"

	"backend/internal/domain"

	"github.com/google/uuid"
)

type EnvironmentRepository interface {
	Create(ctx context.Context, env *domain.Environment) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Environment, error)
	GetByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) ([]*domain.Environment, error)
	GetByCreatedBy(ctx context.Context, userID uuid.UUID) ([]*domain.Environment, error)
	GetByTemplateID(ctx context.Context, templateID uuid.UUID) ([]*domain.Environment, error)
	Update(ctx context.Context, env *domain.Environment) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, opts ListOptions) ([]*domain.Environment, error)
}
