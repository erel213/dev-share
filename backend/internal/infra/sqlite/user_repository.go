package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"backend/internal/domain"
	domainerrors "backend/internal/domain/errors"
	"backend/internal/domain/repository"
	infraerrors "backend/internal/infra/errors"
	pkgerrors "backend/pkg/errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

type userRepository struct {
	uow *UnitOfWork
}

func newUserRepository(uow *UnitOfWork) repository.UserRepository {
	return &userRepository{uow: uow}
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

	if user.BaseUser.ID == uuid.Nil {
		user.BaseUser.ID = uuid.New()
	}

	query, args, err := builder.
		Insert("users").
		Columns("id", "name", "email", "workspace_id", "oauth_provider", "oauth_id", "password").
		Values(user.BaseUser.ID, user.BaseUser.Name, user.BaseUser.Email, user.BaseUser.WorkspaceID, oauthProvider, oauthID, password).
		Suffix("RETURNING created_at, updated_at").
		ToSql()
	if err != nil {
		return infraerrors.WrapSQLiteError(err, "create_user")
	}

	var cat, uat TimestampDest
	err = r.uow.Querier().QueryRowContext(ctx, query, args...).Scan(&cat, &uat)
	if err != nil {
		return infraerrors.WrapSQLiteError(err, "create_user")
	}

	user.BaseUser.CreatedAt = cat.Time()
	user.BaseUser.UpdatedAt = uat.Time()

	return nil
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.UserAggregate, *pkgerrors.Error) {
	query, args, err := builder.
		Select("id", "oauth_provider", "oauth_id", "password", "name", "email", "workspace_id", "created_at", "updated_at").
		From("users").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, infraerrors.WrapSQLiteError(err, "get_user")
	}

	user, err := r.scanUser(r.uow.Querier().QueryRowContext(ctx, query, args...))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domainerrors.NotFound("User", id.String())
		}
		return nil, infraerrors.WrapSQLiteError(err, "get_user")
	}

	return user, nil
}

func (r *userRepository) GetByOAuthID(ctx context.Context, provider domain.OauthProvider, oauthID string) (*domain.UserAggregate, *pkgerrors.Error) {
	query, args, err := builder.
		Select("id", "oauth_provider", "oauth_id", "password", "name", "email", "workspace_id", "created_at", "updated_at").
		From("users").
		Where(sq.Eq{
			"oauth_provider": provider,
			"oauth_id":       oauthID,
		}).
		ToSql()
	if err != nil {
		return nil, infraerrors.WrapSQLiteError(err, "get_user_by_oauth")
	}

	user, err := r.scanUser(r.uow.Querier().QueryRowContext(ctx, query, args...))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domainerrors.NotFound("User", oauthID)
		}
		return nil, infraerrors.WrapSQLiteError(err, "get_user_by_oauth")
	}

	return user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.UserAggregate, *pkgerrors.Error) {
	query, args, err := builder.
		Select("id", "oauth_provider", "oauth_id", "password", "name", "email", "workspace_id", "created_at", "updated_at").
		From("users").
		Where(sq.Eq{"email": email}).
		ToSql()
	if err != nil {
		return nil, infraerrors.WrapSQLiteError(err, "get_user_by_email")
	}

	user, err := r.scanUser(r.uow.Querier().QueryRowContext(ctx, query, args...))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domainerrors.NotFound("User", email)
		}
		return nil, infraerrors.WrapSQLiteError(err, "get_user_by_email")
	}

	return user, nil
}

func (r *userRepository) GetByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) ([]*domain.UserAggregate, *pkgerrors.Error) {
	query, args, err := builder.
		Select("id", "oauth_provider", "oauth_id", "password", "name", "email", "workspace_id", "created_at", "updated_at").
		From("users").
		Where(sq.Eq{"workspace_id": workspaceID}).
		OrderBy("created_at DESC").
		ToSql()
	if err != nil {
		return nil, infraerrors.WrapSQLiteError(err, "get_users_by_workspace")
	}

	rows, err := r.uow.Querier().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, infraerrors.WrapSQLiteError(err, "get_users_by_workspace")
	}
	defer rows.Close()

	var users []*domain.UserAggregate
	for rows.Next() {
		user, err := r.scanUserFromRows(rows)
		if err != nil {
			return nil, infraerrors.WrapSQLiteError(err, "scan_user")
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, infraerrors.WrapSQLiteError(err, "iterate_users")
	}

	return users, nil
}

func (r *userRepository) Update(ctx context.Context, user domain.UserAggregate) *pkgerrors.Error {
	b := builder.
		Update("users").
		Set("name", user.BaseUser.Name).
		Set("email", user.BaseUser.Email).
		Set("workspace_id", user.BaseUser.WorkspaceID).
		Set("updated_at", sq.Expr("CURRENT_TIMESTAMP"))

	if user.ThirdPartyUser != nil {
		b = b.
			Set("oauth_provider", user.ThirdPartyUser.OauthProvider).
			Set("oauth_id", user.ThirdPartyUser.OauthID).
			Set("password", nil)
	} else if user.LocalUser != nil {
		b = b.
			Set("oauth_provider", nil).
			Set("oauth_id", nil).
			Set("password", user.LocalUser.Password)
	}

	query, args, err := b.
		Where(sq.Eq{"id": user.BaseUser.ID}).
		Suffix("RETURNING updated_at").
		ToSql()
	if err != nil {
		return infraerrors.WrapSQLiteError(err, "update_user")
	}

	var uat TimestampDest
	err = r.uow.Querier().QueryRowContext(ctx, query, args...).Scan(&uat)
	if err != nil {
		if err == sql.ErrNoRows {
			return domainerrors.NotFound("User", user.BaseUser.ID.String())
		}
		return infraerrors.WrapSQLiteError(err, "update_user")
	}

	user.BaseUser.UpdatedAt = uat.Time()

	return nil
}

func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) *pkgerrors.Error {
	query, args, err := builder.
		Delete("users").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return infraerrors.WrapSQLiteError(err, "delete_user")
	}

	result, err := r.uow.Querier().ExecContext(ctx, query, args...)
	if err != nil {
		return infraerrors.WrapSQLiteError(err, "delete_user")
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return infraerrors.WrapSQLiteError(err, "get_rows_affected")
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

	query, args, err := builder.
		Select("id", "oauth_provider", "oauth_id", "password", "name", "email", "workspace_id", "created_at", "updated_at").
		From("users").
		OrderBy(fmt.Sprintf("%s %s", opts.SortBy, opts.Order)).
		Limit(uint64(opts.Limit)).
		Offset(uint64(opts.Offset)).
		ToSql()
	if err != nil {
		return nil, infraerrors.WrapSQLiteError(err, "list_users")
	}

	rows, err := r.uow.Querier().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, infraerrors.WrapSQLiteError(err, "list_users")
	}
	defer rows.Close()

	var users []*domain.UserAggregate
	for rows.Next() {
		user, err := r.scanUserFromRows(rows)
		if err != nil {
			return nil, infraerrors.WrapSQLiteError(err, "scan_user")
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, infraerrors.WrapSQLiteError(err, "iterate_users")
	}

	return users, nil
}

func (r *userRepository) Count(ctx context.Context) (int, *pkgerrors.Error) {
	query, args, err := builder.
		Select("COUNT(*)").
		From("users").
		ToSql()
	if err != nil {
		return 0, infraerrors.WrapSQLiteError(err, "count_users")
	}

	var count int
	err = r.uow.Querier().QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, infraerrors.WrapSQLiteError(err, "count_users")
	}

	return count, nil
}

func (r *userRepository) scanUser(row *sql.Row) (*domain.UserAggregate, error) {
	var (
		id                               uuid.UUID
		oauthProvider, oauthID, password sql.NullString
		name, email                      string
		workspaceID                      uuid.UUID
		cat, uat                         TimestampDest
	)

	err := row.Scan(
		&id,
		&oauthProvider,
		&oauthID,
		&password,
		&name,
		&email,
		&workspaceID,
		&cat,
		&uat,
	)
	if err != nil {
		return nil, err
	}

	return buildUserAggregate(id, oauthProvider, oauthID, password, name, email, workspaceID, cat.Time(), uat.Time()), nil
}

func (r *userRepository) scanUserFromRows(rows *sql.Rows) (*domain.UserAggregate, error) {
	var (
		id                               uuid.UUID
		oauthProvider, oauthID, password sql.NullString
		name, email                      string
		workspaceID                      uuid.UUID
		cat, uat                         TimestampDest
	)

	err := rows.Scan(
		&id,
		&oauthProvider,
		&oauthID,
		&password,
		&name,
		&email,
		&workspaceID,
		&cat,
		&uat,
	)
	if err != nil {
		return nil, err
	}

	return buildUserAggregate(id, oauthProvider, oauthID, password, name, email, workspaceID, cat.Time(), uat.Time()), nil
}

func buildUserAggregate(
	id uuid.UUID,
	oauthProvider, oauthID, password sql.NullString,
	name, email string,
	workspaceID uuid.UUID,
	createdAt, updatedAt time.Time,
) *domain.UserAggregate {
	user := &domain.UserAggregate{
		BaseUser: domain.BaseUser{
			ID:          id,
			Name:        name,
			Email:       email,
			WorkspaceID: workspaceID,
			CreatedAt:   createdAt,
			UpdatedAt:   updatedAt,
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
