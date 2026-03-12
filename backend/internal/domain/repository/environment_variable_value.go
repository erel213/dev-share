package repository

import (
	"context"

	"backend/internal/domain"
	"backend/pkg/errors"

	"github.com/google/uuid"
)

type EnvironmentVariableValueRepository interface {
	Create(ctx context.Context, value domain.EnvironmentVariableValue) *errors.Error
	GetByEnvironmentID(ctx context.Context, environmentID uuid.UUID) ([]*domain.EnvironmentVariableValue, *errors.Error)
	UpsertBatch(ctx context.Context, values []domain.EnvironmentVariableValue) *errors.Error
	DeleteByEnvironmentID(ctx context.Context, environmentID uuid.UUID) *errors.Error
}
