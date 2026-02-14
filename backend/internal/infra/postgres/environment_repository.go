package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"backend/internal/domain"
	"backend/internal/domain/repository"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

type environmentRepository struct {
	db *sql.DB
}

func NewEnvironmentRepository(db *sql.DB) repository.EnvironmentRepository {
	return &environmentRepository{db: db}
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

	err = r.db.QueryRowContext(ctx, query, args...).
		Scan(&env.ID, &env.CreatedAt, &env.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create environment: %w", err)
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
	err = r.db.QueryRowContext(ctx, query, args...).Scan(
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
			return nil, &repository.NotFoundError{EntityType: "Environment", ID: id}
		}
		return nil, fmt.Errorf("failed to get environment: %w", err)
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

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get environments by workspace: %w", err)
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
			return nil, fmt.Errorf("failed to scan environment: %w", err)
		}
		environments = append(environments, &env)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
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

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get environments by creator: %w", err)
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
			return nil, fmt.Errorf("failed to scan environment: %w", err)
		}
		environments = append(environments, &env)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
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

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get environments by template: %w", err)
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
			return nil, fmt.Errorf("failed to scan environment: %w", err)
		}
		environments = append(environments, &env)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
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

	err = r.db.QueryRowContext(ctx, query, args...).Scan(&env.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return &repository.NotFoundError{EntityType: "Environment", ID: env.ID}
		}
		return fmt.Errorf("failed to update environment: %w", err)
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

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to delete environment: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return &repository.NotFoundError{EntityType: "Environment", ID: id}
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

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list environments: %w", err)
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
			return nil, fmt.Errorf("failed to scan environment: %w", err)
		}
		environments = append(environments, &env)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return environments, nil
}
