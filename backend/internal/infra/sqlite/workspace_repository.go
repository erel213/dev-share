package sqlite

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

type workspaceRepository struct {
	uow *UnitOfWork
}

func newWorkspaceRepository(uow *UnitOfWork) repository.WorkspaceRepository {
	return &workspaceRepository{uow: uow}
}

func (r *workspaceRepository) Create(ctx context.Context, workspace *domain.Workspace) *pkgerrors.Error {
	if workspace.ID == uuid.Nil {
		workspace.ID = uuid.New()
	}

	query, args, err := builder.
		Insert("workspaces").
		Columns("id", "name", "description", "admin_id").
		Values(workspace.ID, workspace.Name, workspace.Description, workspace.AdminID).
		Suffix("RETURNING created_at, updated_at").
		ToSql()
	if err != nil {
		return infraerrors.WrapSQLiteError(err, "create_workspace")
	}

	var cat, uat TimestampDest
	err = r.uow.Querier().QueryRowContext(ctx, query, args...).Scan(&cat, &uat)
	if err != nil {
		return infraerrors.WrapSQLiteError(err, "create_workspace")
	}

	workspace.CreatedAt = cat.Time()
	workspace.UpdatedAt = uat.Time()

	return nil
}

func (r *workspaceRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Workspace, *pkgerrors.Error) {
	query, args, err := builder.
		Select("id", "name", "description", "admin_id", "created_at", "updated_at").
		From("workspaces").
		Where(sq.Eq{"id": id}).
		Where("deleted_at IS NULL").
		ToSql()
	if err != nil {
		return nil, infraerrors.WrapSQLiteError(err, "get_workspace")
	}

	var workspace domain.Workspace
	var cat, uat TimestampDest
	err = r.uow.Querier().QueryRowContext(ctx, query, args...).Scan(
		&workspace.ID,
		&workspace.Name,
		&workspace.Description,
		&workspace.AdminID,
		&cat,
		&uat,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domainerrors.NotFound("Workspace", id.String())
		}
		return nil, infraerrors.WrapSQLiteError(err, "get_workspace")
	}

	workspace.CreatedAt = cat.Time()
	workspace.UpdatedAt = uat.Time()

	return &workspace, nil
}

func (r *workspaceRepository) GetByAdminID(ctx context.Context, adminID uuid.UUID) ([]*domain.Workspace, *pkgerrors.Error) {
	query, args, err := builder.
		Select("id", "name", "description", "admin_id", "created_at", "updated_at").
		From("workspaces").
		Where(sq.Eq{"admin_id": adminID}).
		Where("deleted_at IS NULL").
		OrderBy("created_at DESC").
		ToSql()
	if err != nil {
		return nil, infraerrors.WrapSQLiteError(err, "get_workspaces_by_admin")
	}

	rows, err := r.uow.Querier().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, infraerrors.WrapSQLiteError(err, "get_workspaces_by_admin")
	}
	defer rows.Close()

	var workspaces []*domain.Workspace
	for rows.Next() {
		var workspace domain.Workspace
		var cat, uat TimestampDest
		err := rows.Scan(
			&workspace.ID,
			&workspace.Name,
			&workspace.Description,
			&workspace.AdminID,
			&cat,
			&uat,
		)
		if err != nil {
			return nil, infraerrors.WrapSQLiteError(err, "scan_workspace")
		}
		workspace.CreatedAt = cat.Time()
		workspace.UpdatedAt = uat.Time()
		workspaces = append(workspaces, &workspace)
	}

	if err := rows.Err(); err != nil {
		return nil, infraerrors.WrapSQLiteError(err, "iterate_workspaces")
	}

	return workspaces, nil
}

func (r *workspaceRepository) Update(ctx context.Context, workspace *domain.Workspace) *pkgerrors.Error {
	query, args, err := builder.
		Update("workspaces").
		Set("name", workspace.Name).
		Set("description", workspace.Description).
		Set("admin_id", workspace.AdminID).
		Set("updated_at", sq.Expr("CURRENT_TIMESTAMP")).
		Where(sq.Eq{"id": workspace.ID}).
		Suffix("RETURNING updated_at").
		ToSql()
	if err != nil {
		return infraerrors.WrapSQLiteError(err, "update_workspace")
	}

	var uat TimestampDest
	err = r.uow.Querier().QueryRowContext(ctx, query, args...).Scan(&uat)
	if err != nil {
		if err == sql.ErrNoRows {
			return domainerrors.NotFound("Workspace", workspace.ID.String())
		}
		return infraerrors.WrapSQLiteError(err, "update_workspace")
	}

	workspace.UpdatedAt = uat.Time()

	return nil
}

func (r *workspaceRepository) Delete(ctx context.Context, id uuid.UUID) *pkgerrors.Error {
	query, args, err := builder.
		Update("workspaces").
		Set("deleted_at", sq.Expr("CURRENT_TIMESTAMP")).
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return infraerrors.WrapSQLiteError(err, "delete_workspace")
	}

	result, err := r.uow.Querier().ExecContext(ctx, query, args...)
	if err != nil {
		return infraerrors.WrapSQLiteError(err, "delete_workspace")
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return infraerrors.WrapSQLiteError(err, "get_rows_affected")
	}

	if rows == 0 {
		return domainerrors.NotFound("Workspace", id.String())
	}

	return nil
}

func (r *workspaceRepository) List(ctx context.Context, opts repository.ListOptions) ([]*domain.Workspace, *pkgerrors.Error) {
	opts.ApplyDefaults()
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	query, args, err := builder.
		Select("id", "name", "description", "admin_id", "created_at", "updated_at").
		From("workspaces").
		Where("deleted_at IS NULL").
		OrderBy(fmt.Sprintf("%s %s", opts.SortBy, opts.Order)).
		Limit(uint64(opts.Limit)).
		Offset(uint64(opts.Offset)).
		ToSql()
	if err != nil {
		return nil, infraerrors.WrapSQLiteError(err, "list_workspaces")
	}

	rows, err := r.uow.Querier().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, infraerrors.WrapSQLiteError(err, "list_workspaces")
	}
	defer rows.Close()

	var workspaces []*domain.Workspace
	for rows.Next() {
		var workspace domain.Workspace
		var cat, uat TimestampDest
		err := rows.Scan(
			&workspace.ID,
			&workspace.Name,
			&workspace.Description,
			&workspace.AdminID,
			&cat,
			&uat,
		)
		if err != nil {
			return nil, infraerrors.WrapSQLiteError(err, "scan_workspace")
		}
		workspace.CreatedAt = cat.Time()
		workspace.UpdatedAt = uat.Time()
		workspaces = append(workspaces, &workspace)
	}

	if err := rows.Err(); err != nil {
		return nil, infraerrors.WrapSQLiteError(err, "iterate_workspaces")
	}

	return workspaces, nil
}

func (r *workspaceRepository) UpdateAdminID(ctx context.Context, workspaceID uuid.UUID, adminID uuid.UUID) *pkgerrors.Error {
	query, args, err := builder.
		Update("workspaces").
		Set("admin_id", adminID).
		Set("updated_at", sq.Expr("CURRENT_TIMESTAMP")).
		Where(sq.Eq{"id": workspaceID}).
		ToSql()
	if err != nil {
		return infraerrors.WrapSQLiteError(err, "update_workspace_admin")
	}

	result, err := r.uow.Querier().ExecContext(ctx, query, args...)
	if err != nil {
		return infraerrors.WrapSQLiteError(err, "update_workspace_admin")
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return infraerrors.WrapSQLiteError(err, "get_rows_affected")
	}

	if rows == 0 {
		return domainerrors.NotFound("Workspace", workspaceID.String())
	}

	return nil
}
