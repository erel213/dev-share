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

type templateRepository struct {
	uow *UnitOfWork
}

func NewTemplateRepository(uow *UnitOfWork) repository.TemplateRepository {
	return &templateRepository{uow: uow}
}

func (r *templateRepository) Create(ctx context.Context, template domain.Template) *errors.Error {
	query, args, err := StatementBuilder.
		Insert("templates").
		Columns("name", "workspace_id", "path").
		Values(template.Name, template.WorkspaceID, template.Path).
		Suffix("RETURNING id, created_at, updated_at").
		ToSql()
	if err != nil {
		return errors.Wrap(err, "failed to build insert query")
	}

	err = r.uow.Querier().QueryRowContext(ctx, query, args...).
		Scan(&template.ID, &template.CreatedAt, &template.UpdatedAt)
	if err != nil {
		return infraerrors.WrapDatabaseError(err, "create_template")
	}

	return nil
}

func (r *templateRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Template, *errors.Error) {
	query, args, err := StatementBuilder.
		Select("id", "name", "workspace_id", "path", "created_at", "updated_at").
		From("templates").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build select query")
	}

	var template domain.Template
	err = r.uow.Querier().QueryRowContext(ctx, query, args...).Scan(
		&template.ID,
		&template.Name,
		&template.WorkspaceID,
		&template.Path,
		&template.CreatedAt,
		&template.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domainerrors.NotFound("Template", id.String())
		}
		return nil, infraerrors.WrapDatabaseError(err, "get_template")
	}

	return &template, nil
}

func (r *templateRepository) GetByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) ([]*domain.Template, *errors.Error) {
	query, args, err := StatementBuilder.
		Select("id", "name", "workspace_id", "path", "created_at", "updated_at").
		From("templates").
		Where(sq.Eq{"workspace_id": workspaceID}).
		OrderBy("created_at DESC").
		ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build query")
	}

	rows, err := r.uow.Querier().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, infraerrors.WrapDatabaseError(err, "get_templates_by_workspace")
	}
	defer rows.Close()

	var templates []*domain.Template
	for rows.Next() {
		var template domain.Template
		err := rows.Scan(
			&template.ID,
			&template.Name,
			&template.WorkspaceID,
			&template.Path,
			&template.CreatedAt,
			&template.UpdatedAt,
		)
		if err != nil {
			return nil, infraerrors.WrapDatabaseError(err, "scan_template")
		}
		templates = append(templates, &template)
	}

	if err := rows.Err(); err != nil {
		return nil, infraerrors.WrapDatabaseError(err, "iterate_templates")
	}

	return templates, nil
}

func (r *templateRepository) Update(ctx context.Context, template domain.Template) *errors.Error {
	query, args, err := StatementBuilder.
		Update("templates").
		Set("name", template.Name).
		Set("path", template.Path).
		Set("updated_at", sq.Expr("CURRENT_TIMESTAMP")).
		Where(sq.Eq{"id": template.ID}).
		Suffix("RETURNING updated_at").
		ToSql()
	if err != nil {
		return errors.Wrap(err, "failed to build update query")
	}

	err = r.uow.Querier().QueryRowContext(ctx, query, args...).Scan(&template.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return domainerrors.NotFound("Template", template.ID.String())
		}
		return infraerrors.WrapDatabaseError(err, "update_template")
	}

	return nil
}

func (r *templateRepository) Delete(ctx context.Context, id uuid.UUID) *errors.Error {
	query, args, err := StatementBuilder.
		Delete("templates").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return errors.Wrap(err, "failed to build delete query")
	}

	result, err := r.uow.Querier().ExecContext(ctx, query, args...)
	if err != nil {
		return infraerrors.WrapDatabaseError(err, "delete_template")
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return infraerrors.WrapDatabaseError(err, "get_rows_affected")
	}

	if rows == 0 {
		return domainerrors.NotFound("Template", id.String())
	}

	return nil
}

func (r *templateRepository) List(ctx context.Context, opts repository.ListOptions) ([]*domain.Template, *errors.Error) {
	opts.ApplyDefaults()
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	query, args, err := StatementBuilder.
		Select("id", "name", "workspace_id", "path", "created_at", "updated_at").
		From("templates").
		OrderBy(fmt.Sprintf("%s %s", opts.SortBy, opts.Order)).
		Limit(uint64(opts.Limit)).
		Offset(uint64(opts.Offset)).
		ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build list query")
	}

	rows, err := r.uow.Querier().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, infraerrors.WrapDatabaseError(err, "list_templates")
	}
	defer rows.Close()

	var templates []*domain.Template
	for rows.Next() {
		var template domain.Template
		err := rows.Scan(
			&template.ID,
			&template.Name,
			&template.WorkspaceID,
			&template.Path,
			&template.CreatedAt,
			&template.UpdatedAt,
		)
		if err != nil {
			return nil, infraerrors.WrapDatabaseError(err, "scan_template")
		}
		templates = append(templates, &template)
	}

	if err := rows.Err(); err != nil {
		return nil, infraerrors.WrapDatabaseError(err, "iterate_templates")
	}

	return templates, nil
}
