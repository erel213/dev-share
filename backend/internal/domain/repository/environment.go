package repository

import (
	"context"

	"backend/internal/domain"
	"backend/pkg/contracts"
	"backend/pkg/errors"

	"github.com/google/uuid"
)

type EnvironmentRepository interface {
	Create(ctx context.Context, env *domain.Environment) *errors.Error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Environment, *errors.Error)
	GetByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) ([]*domain.Environment, *errors.Error)
	GetByCreatedBy(ctx context.Context, userID uuid.UUID) ([]*domain.Environment, *errors.Error)
	GetByTemplateID(ctx context.Context, templateID uuid.UUID) ([]*domain.Environment, *errors.Error)
	Update(ctx context.Context, env *domain.Environment) *errors.Error
	Delete(ctx context.Context, id uuid.UUID) *errors.Error
	List(ctx context.Context, opts ListOptions) ([]*domain.Environment, *errors.Error)

	// AcquireOperation atomically sets the environment status to newStatus
	// only if the current status is not one of the blocking statuses.
	// Returns the updated environment, or an error if the operation cannot proceed.
	AcquireOperation(ctx context.Context, id uuid.UUID, newStatus domain.EnvironmentStatus) (*domain.Environment, *errors.Error)

	// ListFiltered returns environments with enriched fields (created_by_name, template_name)
	// using JOINs, filtered by the provided options.
	ListFiltered(ctx context.Context, opts EnvironmentListOptions) ([]*contracts.EnvironmentResponse, *errors.Error)
}
