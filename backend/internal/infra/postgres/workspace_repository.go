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

type workspaceRepository struct {
	db *sql.DB
}

func NewWorkspaceRepository(db *sql.DB) repository.WorkspaceRepository {
	return &workspaceRepository{db: db}
}

func (r *workspaceRepository) Create(ctx context.Context, workspace *domain.Workspace) error {
	query, args, err := StatementBuilder.
		Insert("workspaces").
		Columns("name", "description", "admin_id").
		Values(workspace.Name, workspace.Description, workspace.AdminID).
		Suffix("RETURNING id, created_at, updated_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build insert query: %w", err)
	}

	err = r.db.QueryRowContext(ctx, query, args...).
		Scan(&workspace.ID, &workspace.CreatedAt, &workspace.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create workspace: %w", err)
	}

	return nil
}

func (r *workspaceRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Workspace, error) {
	query, args, err := StatementBuilder.
		Select("id", "name", "description", "admin_id", "created_at", "updated_at").
		From("workspaces").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	var workspace domain.Workspace
	err = r.db.QueryRowContext(ctx, query, args...).Scan(
		&workspace.ID,
		&workspace.Name,
		&workspace.Description,
		&workspace.AdminID,
		&workspace.CreatedAt,
		&workspace.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &repository.NotFoundError{EntityType: "Workspace", ID: id}
		}
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}

	return &workspace, nil
}

func (r *workspaceRepository) GetByAdminID(ctx context.Context, adminID uuid.UUID) ([]*domain.Workspace, error) {
	query, args, err := StatementBuilder.
		Select("id", "name", "description", "admin_id", "created_at", "updated_at").
		From("workspaces").
		Where(sq.Eq{"admin_id": adminID}).
		OrderBy("created_at DESC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspaces by admin: %w", err)
	}
	defer rows.Close()

	var workspaces []*domain.Workspace
	for rows.Next() {
		var workspace domain.Workspace
		err := rows.Scan(
			&workspace.ID,
			&workspace.Name,
			&workspace.Description,
			&workspace.AdminID,
			&workspace.CreatedAt,
			&workspace.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan workspace: %w", err)
		}
		workspaces = append(workspaces, &workspace)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return workspaces, nil
}

func (r *workspaceRepository) Update(ctx context.Context, workspace *domain.Workspace) error {
	query, args, err := StatementBuilder.
		Update("workspaces").
		Set("name", workspace.Name).
		Set("description", workspace.Description).
		Set("admin_id", workspace.AdminID).
		Set("updated_at", sq.Expr("CURRENT_TIMESTAMP")).
		Where(sq.Eq{"id": workspace.ID}).
		Suffix("RETURNING updated_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build update query: %w", err)
	}

	err = r.db.QueryRowContext(ctx, query, args...).Scan(&workspace.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return &repository.NotFoundError{EntityType: "Workspace", ID: workspace.ID}
		}
		return fmt.Errorf("failed to update workspace: %w", err)
	}

	return nil
}

func (r *workspaceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query, args, err := StatementBuilder.
		Delete("workspaces").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build delete query: %w", err)
	}

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to delete workspace: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return &repository.NotFoundError{EntityType: "Workspace", ID: id}
	}

	return nil
}

func (r *workspaceRepository) List(ctx context.Context, opts repository.ListOptions) ([]*domain.Workspace, error) {
	opts.ApplyDefaults()
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	query, args, err := StatementBuilder.
		Select("id", "name", "description", "admin_id", "created_at", "updated_at").
		From("workspaces").
		OrderBy(fmt.Sprintf("%s %s", opts.SortBy, opts.Order)).
		Limit(uint64(opts.Limit)).
		Offset(uint64(opts.Offset)).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build list query: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list workspaces: %w", err)
	}
	defer rows.Close()

	var workspaces []*domain.Workspace
	for rows.Next() {
		var workspace domain.Workspace
		err := rows.Scan(
			&workspace.ID,
			&workspace.Name,
			&workspace.Description,
			&workspace.AdminID,
			&workspace.CreatedAt,
			&workspace.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan workspace: %w", err)
		}
		workspaces = append(workspaces, &workspace)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return workspaces, nil
}
