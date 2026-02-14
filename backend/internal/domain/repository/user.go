package repository

import (
	"context"

	"backend/internal/domain"

	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetByOAuthID(ctx context.Context, provider domain.OauthProvider, oauthID string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) ([]*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, opts ListOptions) ([]*domain.User, error)
}
