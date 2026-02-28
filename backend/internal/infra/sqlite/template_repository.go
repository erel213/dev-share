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

type templateRepository struct {
	uow *UnitOfWork
}

func newTemplateRepository(uow *UnitOfWork) repository.TemplateRepository {
	return &templateRepository{uow: uow}
}

func (r *templateRepository) Create(ctx context.Context, template domain.Template) *pkgerrors.Error {
	query, args, err := builder.
		Insert("templates").
		Columns("id", "name", "workspace_id", "path").
		Values(template.ID, template.Name, template.WorkspaceID, template.Path).
		Suffix("RETURNING created_at, updated_at").
		ToSql()
	if err != nil {
		return infraerrors.WrapSQLiteError(err, "create_template")
	}

	var cat, uat TimestampDest
	err = r.uow.Querier().QueryRowContext(ctx, query, args...).Scan(&cat, &uat)
	if err != nil {
		return infraerrors.WrapSQLiteError(err, "create_template")
	}

	template.CreatedAt = cat.Time()
	template.UpdatedAt = uat.Time()

	return nil
}

func (r *templateRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Template, *pkgerrors.Error) {
	query, args, err := builder.
		Select("id", "name", "workspace_id", "path", "created_at", "updated_at").
		From("templates").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, infraerrors.WrapSQLiteError(err, "get_template")
	}

	var template domain.Template
	var cat, uat TimestampDest
	err = r.uow.Querier().QueryRowContext(ctx, query, args...).Scan(
		&template.ID,
		&template.Name,
		&template.WorkspaceID,
		&template.Path,
		&cat,
		&uat,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domainerrors.NotFound("Template", id.String())
		}
		return nil, infraerrors.WrapSQLiteError(err, "get_template")
	}

	template.CreatedAt = cat.Time()
	template.UpdatedAt = uat.Time()

	return &template, nil
}

func (r *templateRepository) GetByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) ([]*domain.Template, *pkgerrors.Error) {
	query, args, err := builder.
		Select("id", "name", "workspace_id", "path", "created_at", "updated_at").
		From("templates").
		Where(sq.Eq{"workspace_id": workspaceID}).
		OrderBy("created_at DESC").
		ToSql()
	if err != nil {
		return nil, infraerrors.WrapSQLiteError(err, "get_templates_by_workspace")
	}

	rows, err := r.uow.Querier().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, infraerrors.WrapSQLiteError(err, "get_templates_by_workspace")
	}
	defer rows.Close()

	var templates []*domain.Template
	for rows.Next() {
		var template domain.Template
		var cat, uat TimestampDest
		err := rows.Scan(
			&template.ID,
			&template.Name,
			&template.WorkspaceID,
			&template.Path,
			&cat,
			&uat,
		)
		if err != nil {
			return nil, infraerrors.WrapSQLiteError(err, "scan_template")
		}
		template.CreatedAt = cat.Time()
		template.UpdatedAt = uat.Time()
		templates = append(templates, &template)
	}

	if err := rows.Err(); err != nil {
		return nil, infraerrors.WrapSQLiteError(err, "iterate_templates")
	}

	return templates, nil
}

func (r *templateRepository) Update(ctx context.Context, template domain.Template) *pkgerrors.Error {
	query, args, err := builder.
		Update("templates").
		Set("name", template.Name).
		Set("path", template.Path).
		Set("updated_at", sq.Expr("CURRENT_TIMESTAMP")).
		Where(sq.Eq{"id": template.ID}).
		Suffix("RETURNING updated_at").
		ToSql()
	if err != nil {
		return infraerrors.WrapSQLiteError(err, "update_template")
	}

	var uat TimestampDest
	err = r.uow.Querier().QueryRowContext(ctx, query, args...).Scan(&uat)
	if err != nil {
		if err == sql.ErrNoRows {
			return domainerrors.NotFound("Template", template.ID.String())
		}
		return infraerrors.WrapSQLiteError(err, "update_template")
	}

	template.UpdatedAt = uat.Time()

	return nil
}

func (r *templateRepository) Delete(ctx context.Context, id uuid.UUID) *pkgerrors.Error {
	query, args, err := builder.
		Delete("templates").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return infraerrors.WrapSQLiteError(err, "delete_template")
	}

	result, err := r.uow.Querier().ExecContext(ctx, query, args...)
	if err != nil {
		return infraerrors.WrapSQLiteError(err, "delete_template")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return infraerrors.WrapSQLiteError(err, "get_rows_affected")
	}

	if rowsAffected == 0 {
		return domainerrors.NotFound("Template", id.String())
	}

	return nil
}

func (r *templateRepository) List(ctx context.Context, opts repository.ListOptions) ([]*domain.Template, *pkgerrors.Error) {
	opts.ApplyDefaults()
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	query, args, err := builder.
		Select("id", "name", "workspace_id", "path", "created_at", "updated_at").
		From("templates").
		OrderBy(fmt.Sprintf("%s %s", opts.SortBy, opts.Order)).
		Limit(uint64(opts.Limit)).
		Offset(uint64(opts.Offset)).
		ToSql()
	if err != nil {
		return nil, infraerrors.WrapSQLiteError(err, "list_templates")
	}

	rows, err := r.uow.Querier().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, infraerrors.WrapSQLiteError(err, "list_templates")
	}
	defer rows.Close()

	var templates []*domain.Template
	for rows.Next() {
		var template domain.Template
		var cat, uat TimestampDest
		err := rows.Scan(
			&template.ID,
			&template.Name,
			&template.WorkspaceID,
			&template.Path,
			&cat,
			&uat,
		)
		if err != nil {
			return nil, infraerrors.WrapSQLiteError(err, "scan_template")
		}
		template.CreatedAt = cat.Time()
		template.UpdatedAt = uat.Time()
		templates = append(templates, &template)
	}

	if err := rows.Err(); err != nil {
		return nil, infraerrors.WrapSQLiteError(err, "iterate_templates")
	}

	return templates, nil
}
