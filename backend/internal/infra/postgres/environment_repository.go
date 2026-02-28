package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"backend/internal/domain"
	domainerrors "backend/internal/domain/errors"
	"backend/internal/domain/repository"
	infraerrors "backend/internal/infra/errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

type environmentRepository struct {
	uow *UnitOfWork
}

func NewEnvironmentRepository(uow *UnitOfWork) repository.EnvironmentRepository {
	return &environmentRepository{
		uow: uow,
	}
}

func (r *environmentRepository) Create(ctx context.Context, env *domain.Environment) error {
	query, args, err := StatementBuilder.
		Insert("environments").
		Columns("name", "description", "created_by", "workspace_id", "template_id").
		Values(env.Name, env.Description, env.CreatedBy, env.WorkspaceID, env.TemplateID).
		Suffix("RETURNING id, created_at, updated_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build insert query: %w", err)
	}

	err = r.uow.Querier().QueryRowContext(ctx, query, args...).
		Scan(&env.ID, &env.CreatedAt, &env.UpdatedAt)
	if err != nil {
		return infraerrors.WrapDatabaseError(err, "create_environment")
	}

	return nil
}

func (r *environmentRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Environment, error) {
	query, args, err := StatementBuilder.
		Select("id", "name", "created_at", "created_by", "description", "workspace_id", "template_id", "updated_at").
		From("environments").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	var env domain.Environment
	err = r.uow.Querier().QueryRowContext(ctx, query, args...).Scan(
		&env.ID,
		&env.Name,
		&env.CreatedAt,
		&env.CreatedBy,
		&env.Description,
		&env.WorkspaceID,
		&env.TemplateID,
		&env.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domainerrors.NotFound("Environment", id.String())
		}
		return nil, infraerrors.WrapDatabaseError(err, "get_environment")
	}

	return &env, nil
}

func (r *environmentRepository) GetByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) ([]*domain.Environment, error) {
	query, args, err := StatementBuilder.
		Select("id", "name", "created_at", "created_by", "description", "workspace_id", "template_id", "updated_at").
		From("environments").
		Where(sq.Eq{"workspace_id": workspaceID}).
		OrderBy("created_at DESC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := r.uow.Querier().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, infraerrors.WrapDatabaseError(err, "get_environments_by_workspace")
	}
	defer rows.Close()

	var environments []*domain.Environment
	for rows.Next() {
		var env domain.Environment
		err := rows.Scan(
			&env.ID,
			&env.Name,
			&env.CreatedAt,
			&env.CreatedBy,
			&env.Description,
			&env.WorkspaceID,
			&env.TemplateID,
			&env.UpdatedAt,
		)
		if err != nil {
			return nil, infraerrors.WrapDatabaseError(err, "scan_environment")
		}
		environments = append(environments, &env)
	}

	if err := rows.Err(); err != nil {
		return nil, infraerrors.WrapDatabaseError(err, "iterate_environments")
	}

	return environments, nil
}

func (r *environmentRepository) GetByCreatedBy(ctx context.Context, userID uuid.UUID) ([]*domain.Environment, error) {
	query, args, err := StatementBuilder.
		Select("id", "name", "created_at", "created_by", "description", "workspace_id", "template_id", "updated_at").
		From("environments").
		Where(sq.Eq{"created_by": userID}).
		OrderBy("created_at DESC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := r.uow.Querier().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, infraerrors.WrapDatabaseError(err, "get_environments_by_creator")
	}
	defer rows.Close()

	var environments []*domain.Environment
	for rows.Next() {
		var env domain.Environment
		err := rows.Scan(
			&env.ID,
			&env.Name,
			&env.CreatedAt,
			&env.CreatedBy,
			&env.Description,
			&env.WorkspaceID,
			&env.TemplateID,
			&env.UpdatedAt,
		)
		if err != nil {
			return nil, infraerrors.WrapDatabaseError(err, "scan_environment")
		}
		environments = append(environments, &env)
	}

	if err := rows.Err(); err != nil {
		return nil, infraerrors.WrapDatabaseError(err, "iterate_environments")
	}

	return environments, nil
}

func (r *environmentRepository) GetByTemplateID(ctx context.Context, templateID uuid.UUID) ([]*domain.Environment, error) {
	query, args, err := StatementBuilder.
		Select("id", "name", "created_at", "created_by", "description", "workspace_id", "template_id", "updated_at").
		From("environments").
		Where(sq.Eq{"template_id": templateID}).
		OrderBy("created_at DESC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := r.uow.Querier().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, infraerrors.WrapDatabaseError(err, "get_environments_by_template")
	}
	defer rows.Close()

	var environments []*domain.Environment
	for rows.Next() {
		var env domain.Environment
		err := rows.Scan(
			&env.ID,
			&env.Name,
			&env.CreatedAt,
			&env.CreatedBy,
			&env.Description,
			&env.WorkspaceID,
			&env.TemplateID,
			&env.UpdatedAt,
		)
		if err != nil {
			return nil, infraerrors.WrapDatabaseError(err, "scan_environment")
		}
		environments = append(environments, &env)
	}

	if err := rows.Err(); err != nil {
		return nil, infraerrors.WrapDatabaseError(err, "iterate_environments")
	}

	return environments, nil
}

func (r *environmentRepository) Update(ctx context.Context, env *domain.Environment) error {
	query, args, err := StatementBuilder.
		Update("environments").
		Set("name", env.Name).
		Set("description", env.Description).
		Set("created_by", env.CreatedBy).
		Set("workspace_id", env.WorkspaceID).
		Set("template_id", env.TemplateID).
		Set("updated_at", sq.Expr("CURRENT_TIMESTAMP")).
		Where(sq.Eq{"id": env.ID}).
		Suffix("RETURNING updated_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build update query: %w", err)
	}

	err = r.uow.Querier().QueryRowContext(ctx, query, args...).Scan(&env.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return domainerrors.NotFound("Environment", env.ID.String())
		}
		return infraerrors.WrapDatabaseError(err, "update_environment")
	}

	return nil
}

func (r *environmentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query, args, err := StatementBuilder.
		Delete("environments").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build delete query: %w", err)
	}

	result, err := r.uow.Querier().ExecContext(ctx, query, args...)
	if err != nil {
		return infraerrors.WrapDatabaseError(err, "delete_environment")
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return infraerrors.WrapDatabaseError(err, "get_rows_affected")
	}

	if rows == 0 {
		return domainerrors.NotFound("Environment", id.String())
	}

	return nil
}

func (r *environmentRepository) List(ctx context.Context, opts repository.ListOptions) ([]*domain.Environment, error) {
	opts.ApplyDefaults()
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	query, args, err := StatementBuilder.
		Select("id", "name", "created_at", "created_by", "description", "workspace_id", "template_id", "updated_at").
		From("environments").
		OrderBy(fmt.Sprintf("%s %s", opts.SortBy, opts.Order)).
		Limit(uint64(opts.Limit)).
		Offset(uint64(opts.Offset)).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build list query: %w", err)
	}

	rows, err := r.uow.Querier().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, infraerrors.WrapDatabaseError(err, "list_environments")
	}
	defer rows.Close()

	var environments []*domain.Environment
	for rows.Next() {
		var env domain.Environment
		err := rows.Scan(
			&env.ID,
			&env.Name,
			&env.CreatedAt,
			&env.CreatedBy,
			&env.Description,
			&env.WorkspaceID,
			&env.TemplateID,
			&env.UpdatedAt,
		)
		if err != nil {
			return nil, infraerrors.WrapDatabaseError(err, "scan_environment")
		}
		environments = append(environments, &env)
	}

	if err := rows.Err(); err != nil {
		return nil, infraerrors.WrapDatabaseError(err, "iterate_environments")
	}

	return environments, nil
}
