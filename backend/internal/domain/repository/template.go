package repository

import (
	"context"

	"backend/internal/domain"
	"backend/pkg/errors"

	"github.com/google/uuid"
)

type TemplateRepository interface {
	Create(ctx context.Context, template domain.Template) *errors.Error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Template, *errors.Error)
	GetByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) ([]*domain.Template, *errors.Error)
	Update(ctx context.Context, template domain.Template) *errors.Error
	Delete(ctx context.Context, id uuid.UUID) *errors.Error
	List(ctx context.Context, opts ListOptions) ([]*domain.Template, *errors.Error)
}
