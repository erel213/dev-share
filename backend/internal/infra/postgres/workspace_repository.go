package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"backend/internal/domain"
	domainerrors "backend/internal/domain/errors"
	"backend/internal/domain/repository"
	infraerrors "backend/internal/infra/errors"
	"backend/pkg/errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

type workspaceRepository struct {
	db *sql.DB
}

func NewWorkspaceRepository(db *sql.DB) repository.WorkspaceRepository {
	return &workspaceRepository{db: db}
}

func (r *workspaceRepository) Create(ctx context.Context, workspace *domain.Workspace) *errors.Error {
	query, args, err := StatementBuilder.
		Insert("workspaces").
		Columns("name", "description", "admin_id").
		Values(workspace.Name, workspace.Description, workspace.AdminID).
		Suffix("RETURNING id, created_at, updated_at").
		ToSql()
	if err != nil {
		return errors.Wrap(err, "failed to build insert query")
	}

	err = r.db.QueryRowContext(ctx, query, args...).
		Scan(&workspace.ID, &workspace.CreatedAt, &workspace.UpdatedAt)
	if err != nil {
		return infraerrors.WrapDatabaseError(err, "create_workspace")
	}

	return nil
}

func (r *workspaceRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Workspace, *errors.Error) {
	query, args, err := StatementBuilder.
		Select("id", "name", "description", "admin_id", "created_at", "updated_at").
		From("workspaces").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build select query")
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
			return nil, domainerrors.NotFound("Workspace", id.String())
		}
		return nil, infraerrors.WrapDatabaseError(err, "get_workspace")
	}

	return &workspace, nil
}

func (r *workspaceRepository) GetByAdminID(ctx context.Context, adminID uuid.UUID) ([]*domain.Workspace, *errors.Error) {
	query, args, err := StatementBuilder.
		Select("id", "name", "description", "admin_id", "created_at", "updated_at").
		From("workspaces").
		Where(sq.Eq{"admin_id": adminID}).
		OrderBy("created_at DESC").
		ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build query")
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, infraerrors.WrapDatabaseError(err, "get_workspaces_by_admin")
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
			return nil, infraerrors.WrapDatabaseError(err, "scan_workspace")
		}
		workspaces = append(workspaces, &workspace)
	}

	if err := rows.Err(); err != nil {
		return nil, infraerrors.WrapDatabaseError(err, "iterate_workspaces")
	}

	return workspaces, nil
}

func (r *workspaceRepository) Update(ctx context.Context, workspace *domain.Workspace) *errors.Error {
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
		return errors.Wrap(err, "failed to build update query")
	}

	err = r.db.QueryRowContext(ctx, query, args...).Scan(&workspace.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return domainerrors.NotFound("Workspace", workspace.ID.String())
		}
		return infraerrors.WrapDatabaseError(err, "update_workspace")
	}

	return nil
}

func (r *workspaceRepository) Delete(ctx context.Context, id uuid.UUID) *errors.Error {
	query, args, err := StatementBuilder.
		Update("workspaces").
		Set("deleted_at", sq.Expr("CURRENT_TIMESTAMP")).
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return errors.Wrap(err, "failed to build delete query")
	}

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return infraerrors.WrapDatabaseError(err, "delete_workspace")
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return infraerrors.WrapDatabaseError(err, "get_rows_affected")
	}

	if rows == 0 {
		return domainerrors.NotFound("Workspace", id.String())
	}

	return nil
}

func (r *workspaceRepository) List(ctx context.Context, opts repository.ListOptions) ([]*domain.Workspace, *errors.Error) {
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
		return nil, errors.Wrap(err, "failed to build list query")
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, infraerrors.WrapDatabaseError(err, "list_workspaces")
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
			return nil, infraerrors.WrapDatabaseError(err, "scan_workspace")
		}
		workspaces = append(workspaces, &workspace)
	}

	if err := rows.Err(); err != nil {
		return nil, infraerrors.WrapDatabaseError(err, "iterate_workspaces")
	}

	return workspaces, nil
}
