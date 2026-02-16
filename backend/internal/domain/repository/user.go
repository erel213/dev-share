package repository

import (
	"context"

	"backend/internal/domain"
	"backend/pkg/errors"

	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, user domain.UserAggregate) *errors.Error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.UserAggregate, *errors.Error)
	GetByOAuthID(ctx context.Context, provider domain.OauthProvider, oauthID string) (*domain.UserAggregate, *errors.Error)
	GetByEmail(ctx context.Context, email string) (*domain.UserAggregate, *errors.Error)
	GetByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) ([]*domain.UserAggregate, *errors.Error)
	Update(ctx context.Context, user domain.UserAggregate) *errors.Error
	Delete(ctx context.Context, id uuid.UUID) *errors.Error
	List(ctx context.Context, opts ListOptions) ([]*domain.UserAggregate, *errors.Error)
}
