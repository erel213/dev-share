package repository

import (
	"context"

	"backend/internal/domain"
	"backend/pkg/errors"

	"github.com/google/uuid"
)

type TemplateVariableRepository interface {
	Create(ctx context.Context, variable domain.TemplateVariable) *errors.Error
	CreateBatch(ctx context.Context, variables []domain.TemplateVariable) *errors.Error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.TemplateVariable, *errors.Error)
	GetByTemplateID(ctx context.Context, templateID uuid.UUID) ([]*domain.TemplateVariable, *errors.Error)
	GetByTemplateIDAndKey(ctx context.Context, templateID uuid.UUID, key string) (*domain.TemplateVariable, *errors.Error)
	Update(ctx context.Context, variable domain.TemplateVariable) *errors.Error
	UpdateBatch(ctx context.Context, variables []domain.TemplateVariable) *errors.Error
	Delete(ctx context.Context, id uuid.UUID) *errors.Error
	DeleteByTemplateIDAndKeys(ctx context.Context, templateID uuid.UUID, keys []string) *errors.Error
}
