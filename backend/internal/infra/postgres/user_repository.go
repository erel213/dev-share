package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"backend/internal/domain"
	domainerrors "backend/internal/domain/errors"
	"backend/internal/domain/repository"
	infraerrors "backend/internal/infra/errors"
	pkgerrors "backend/pkg/errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) repository.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user domain.UserAggregate) *pkgerrors.Error {
	var oauthProvider, oauthID, password interface{}

	if user.ThirdPartyUser != nil {
		oauthProvider = user.ThirdPartyUser.OauthProvider
		oauthID = user.ThirdPartyUser.OauthID
		password = nil
	} else if user.LocalUser != nil {
		oauthProvider = nil
		oauthID = nil
		password = user.LocalUser.Password
	}

	query, args, err := StatementBuilder.
		Insert("users").
		Columns("name", "email", "workspace_id", "oauth_provider", "oauth_id", "password").
		Values(user.BaseUser.Name, user.BaseUser.Email, user.BaseUser.WorkspaceID, oauthProvider, oauthID, password).
		Suffix("RETURNING id, created_at, updated_at").
		ToSql()
	if err != nil {
		return infraerrors.WrapDatabaseError(err, "create_user")
	}

	err = r.db.QueryRowContext(ctx, query, args...).
		Scan(&user.BaseUser.ID, &user.BaseUser.CreatedAt, &user.BaseUser.UpdatedAt)
	if err != nil {
		return infraerrors.WrapDatabaseError(err, "create_user")
	}

	return nil
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.UserAggregate, *pkgerrors.Error) {
	query, args, err := StatementBuilder.
		Select("id", "oauth_provider", "oauth_id", "password", "name", "email", "workspace_id", "created_at", "updated_at").
		From("users").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, infraerrors.WrapDatabaseError(err, "get_user")
	}

	user, err := r.scanUser(r.db.QueryRowContext(ctx, query, args...))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domainerrors.NotFound("User", id.String())
		}
		return nil, infraerrors.WrapDatabaseError(err, "get_user")
	}

	return user, nil
}

// scanUser scans a database row into a UserAggregate
func (r *userRepository) scanUser(row *sql.Row) (*domain.UserAggregate, error) {
	var (
		id                               uuid.UUID
		oauthProvider, oauthID, password sql.NullString
		name, email                      string
		workspaceID                      uuid.UUID
		createdAt, updatedAt             sql.NullTime
	)

	err := row.Scan(
		&id,
		&oauthProvider,
		&oauthID,
		&password,
		&name,
		&email,
		&workspaceID,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, err
	}

	return buildUserAggregate(id, oauthProvider, oauthID, password, name, email, workspaceID, createdAt, updatedAt), nil
}

// scanUserFromRows scans a database row from sql.Rows into a UserAggregate
func (r *userRepository) scanUserFromRows(rows *sql.Rows) (*domain.UserAggregate, error) {
	var (
		id                               uuid.UUID
		oauthProvider, oauthID, password sql.NullString
		name, email                      string
		workspaceID                      uuid.UUID
		createdAt, updatedAt             sql.NullTime
	)

	err := rows.Scan(
		&id,
		&oauthProvider,
		&oauthID,
		&password,
		&name,
		&email,
		&workspaceID,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, err
	}

	return buildUserAggregate(id, oauthProvider, oauthID, password, name, email, workspaceID, createdAt, updatedAt), nil
}

// buildUserAggregate constructs a UserAggregate from database values
func buildUserAggregate(id uuid.UUID, oauthProvider, oauthID, password sql.NullString, name, email string, workspaceID uuid.UUID, createdAt, updatedAt sql.NullTime) *domain.UserAggregate {
	user := &domain.UserAggregate{
		BaseUser: domain.BaseUser{
			ID:          id,
			Name:        name,
			Email:       email,
			WorkspaceID: workspaceID,
			CreatedAt:   createdAt.Time,
			UpdatedAt:   updatedAt.Time,
		},
	}

	if oauthProvider.Valid && oauthID.Valid {
		user.ThirdPartyUser = &domain.ThirdPartyUser{
			OauthProvider: domain.OauthProvider(oauthProvider.String),
			OauthID:       oauthID.String,
		}
	} else if password.Valid {
		user.LocalUser = &domain.LocalUser{
			Password: password.String,
		}
	}

	return user
}

func (r *userRepository) GetByOAuthID(ctx context.Context, provider domain.OauthProvider, oauthID string) (*domain.UserAggregate, *pkgerrors.Error) {
	query, args, err := StatementBuilder.
		Select("id", "oauth_provider", "oauth_id", "password", "name", "email", "workspace_id", "created_at", "updated_at").
		From("users").
		Where(sq.Eq{
			"oauth_provider": provider,
			"oauth_id":       oauthID,
		}).
		ToSql()
	if err != nil {
		return nil, infraerrors.WrapDatabaseError(err, "get_user_by_oauth")
	}

	user, err := r.scanUser(r.db.QueryRowContext(ctx, query, args...))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domainerrors.NotFound("User", oauthID)
		}
		return nil, infraerrors.WrapDatabaseError(err, "get_user_by_oauth")
	}

	return user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.UserAggregate, *pkgerrors.Error) {
	query, args, err := StatementBuilder.
		Select("id", "oauth_provider", "oauth_id", "password", "name", "email", "workspace_id", "created_at", "updated_at").
		From("users").
		Where(sq.Eq{"email": email}).
		ToSql()
	if err != nil {
		return nil, infraerrors.WrapDatabaseError(err, "get_user_by_email")
	}

	user, err := r.scanUser(r.db.QueryRowContext(ctx, query, args...))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domainerrors.NotFound("User", email)
		}
		return nil, infraerrors.WrapDatabaseError(err, "get_user_by_email")
	}

	return user, nil
}

func (r *userRepository) GetByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) ([]*domain.UserAggregate, *pkgerrors.Error) {
	query, args, err := StatementBuilder.
		Select("id", "oauth_provider", "oauth_id", "password", "name", "email", "workspace_id", "created_at", "updated_at").
		From("users").
		Where(sq.Eq{"workspace_id": workspaceID}).
		OrderBy("created_at DESC").
		ToSql()
	if err != nil {
		return nil, infraerrors.WrapDatabaseError(err, "get_users_by_workspace")
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, infraerrors.WrapDatabaseError(err, "get_users_by_workspace")
	}
	defer rows.Close()

	var users []*domain.UserAggregate
	for rows.Next() {
		user, err := r.scanUserFromRows(rows)
		if err != nil {
			return nil, infraerrors.WrapDatabaseError(err, "scan_user")
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, infraerrors.WrapDatabaseError(err, "iterate_users")
	}

	return users, nil
}

func (r *userRepository) Update(ctx context.Context, user domain.UserAggregate) *pkgerrors.Error {
	builder := StatementBuilder.
		Update("users").
		Set("name", user.BaseUser.Name).
		Set("email", user.BaseUser.Email).
		Set("workspace_id", user.BaseUser.WorkspaceID).
		Set("updated_at", sq.Expr("CURRENT_TIMESTAMP"))

	if user.ThirdPartyUser != nil {
		builder = builder.
			Set("oauth_provider", user.ThirdPartyUser.OauthProvider).
			Set("oauth_id", user.ThirdPartyUser.OauthID).
			Set("password", nil)
	} else if user.LocalUser != nil {
		builder = builder.
			Set("oauth_provider", nil).
			Set("oauth_id", nil).
			Set("password", user.LocalUser.Password)
	}

	query, args, err := builder.
		Where(sq.Eq{"id": user.BaseUser.ID}).
		Suffix("RETURNING updated_at").
		ToSql()
	if err != nil {
		return infraerrors.WrapDatabaseError(err, "update_user")
	}

	err = r.db.QueryRowContext(ctx, query, args...).Scan(&user.BaseUser.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return domainerrors.NotFound("User", user.BaseUser.ID.String())
		}
		return infraerrors.WrapDatabaseError(err, "update_user")
	}

	return nil
}

func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) *pkgerrors.Error {
	query, args, err := StatementBuilder.
		Delete("users").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return infraerrors.WrapDatabaseError(err, "delete_user")
	}

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return infraerrors.WrapDatabaseError(err, "delete_user")
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return infraerrors.WrapDatabaseError(err, "get_rows_affected")
	}

	if rows == 0 {
		return domainerrors.NotFound("User", id.String())
	}

	return nil
}

func (r *userRepository) List(ctx context.Context, opts repository.ListOptions) ([]*domain.UserAggregate, *pkgerrors.Error) {
	opts.ApplyDefaults()
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	query, args, err := StatementBuilder.
		Select("id", "oauth_provider", "oauth_id", "password", "name", "email", "workspace_id", "created_at", "updated_at").
		From("users").
		OrderBy(fmt.Sprintf("%s %s", opts.SortBy, opts.Order)).
		Limit(uint64(opts.Limit)).
		Offset(uint64(opts.Offset)).
		ToSql()
	if err != nil {
		return nil, infraerrors.WrapDatabaseError(err, "list_users")
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, infraerrors.WrapDatabaseError(err, "list_users")
	}
	defer rows.Close()

	var users []*domain.UserAggregate
	for rows.Next() {
		user, err := r.scanUserFromRows(rows)
		if err != nil {
			return nil, infraerrors.WrapDatabaseError(err, "scan_user")
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, infraerrors.WrapDatabaseError(err, "iterate_users")
	}

	return users, nil
}
