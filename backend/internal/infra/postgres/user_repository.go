package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"backend/internal/domain"
	"backend/internal/domain/repository"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) repository.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	query, args, err := StatementBuilder.
		Insert("users").
		Columns("oauth_provider", "oauth_id", "name", "email", "workspace_id").
		Values(user.OauthProvider, user.OauthID, user.Name, user.Email, user.WorkspaceID).
		Suffix("RETURNING id, created_at, updated_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build insert query: %w", err)
	}

	err = r.db.QueryRowContext(ctx, query, args...).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return &repository.ConflictError{
				EntityType: "User",
				Field:      "oauth credentials",
				Value:      fmt.Sprintf("%s:%s", user.OauthProvider, user.OauthID),
			}
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query, args, err := StatementBuilder.
		Select("id", "oauth_provider", "oauth_id", "name", "email", "workspace_id", "created_at", "updated_at").
		From("users").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	var user domain.User
	err = r.db.QueryRowContext(ctx, query, args...).Scan(
		&user.ID,
		&user.OauthProvider,
		&user.OauthID,
		&user.Name,
		&user.Email,
		&user.WorkspaceID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &repository.NotFoundError{EntityType: "User", ID: id}
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (r *userRepository) GetByOAuthID(ctx context.Context, provider domain.OauthProvider, oauthID string) (*domain.User, error) {
	query, args, err := StatementBuilder.
		Select("id", "oauth_provider", "oauth_id", "name", "email", "workspace_id", "created_at", "updated_at").
		From("users").
		Where(sq.Eq{
			"oauth_provider": provider,
			"oauth_id":       oauthID,
		}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var user domain.User
	err = r.db.QueryRowContext(ctx, query, args...).Scan(
		&user.ID,
		&user.OauthProvider,
		&user.OauthID,
		&user.Name,
		&user.Email,
		&user.WorkspaceID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &repository.NotFoundError{EntityType: "User", ID: uuid.Nil}
		}
		return nil, fmt.Errorf("failed to get user by oauth: %w", err)
	}

	return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query, args, err := StatementBuilder.
		Select("id", "oauth_provider", "oauth_id", "name", "email", "workspace_id", "created_at", "updated_at").
		From("users").
		Where(sq.Eq{"email": email}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var user domain.User
	err = r.db.QueryRowContext(ctx, query, args...).Scan(
		&user.ID,
		&user.OauthProvider,
		&user.OauthID,
		&user.Name,
		&user.Email,
		&user.WorkspaceID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &repository.NotFoundError{EntityType: "User", ID: uuid.Nil}
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

func (r *userRepository) GetByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) ([]*domain.User, error) {
	query, args, err := StatementBuilder.
		Select("id", "oauth_provider", "oauth_id", "name", "email", "workspace_id", "created_at", "updated_at").
		From("users").
		Where(sq.Eq{"workspace_id": workspaceID}).
		OrderBy("created_at DESC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get users by workspace: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		var user domain.User
		err := rows.Scan(
			&user.ID,
			&user.OauthProvider,
			&user.OauthID,
			&user.Name,
			&user.Email,
			&user.WorkspaceID,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return users, nil
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	query, args, err := StatementBuilder.
		Update("users").
		Set("oauth_provider", user.OauthProvider).
		Set("oauth_id", user.OauthID).
		Set("name", user.Name).
		Set("email", user.Email).
		Set("workspace_id", user.WorkspaceID).
		Set("updated_at", sq.Expr("CURRENT_TIMESTAMP")).
		Where(sq.Eq{"id": user.ID}).
		Suffix("RETURNING updated_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build update query: %w", err)
	}

	err = r.db.QueryRowContext(ctx, query, args...).Scan(&user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return &repository.NotFoundError{EntityType: "User", ID: user.ID}
		}
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return &repository.ConflictError{
				EntityType: "User",
				Field:      "oauth credentials",
				Value:      fmt.Sprintf("%s:%s", user.OauthProvider, user.OauthID),
			}
		}
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query, args, err := StatementBuilder.
		Delete("users").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build delete query: %w", err)
	}

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return &repository.NotFoundError{EntityType: "User", ID: id}
	}

	return nil
}

func (r *userRepository) List(ctx context.Context, opts repository.ListOptions) ([]*domain.User, error) {
	opts.ApplyDefaults()
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	query, args, err := StatementBuilder.
		Select("id", "oauth_provider", "oauth_id", "name", "email", "workspace_id", "created_at", "updated_at").
		From("users").
		OrderBy(fmt.Sprintf("%s %s", opts.SortBy, opts.Order)).
		Limit(uint64(opts.Limit)).
		Offset(uint64(opts.Offset)).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build list query: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		var user domain.User
		err := rows.Scan(
			&user.ID,
			&user.OauthProvider,
			&user.OauthID,
			&user.Name,
			&user.Email,
			&user.WorkspaceID,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return users, nil
}
