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

type templateVariableRepository struct {
	uow *UnitOfWork
}

func newTemplateVariableRepository(uow *UnitOfWork) repository.TemplateVariableRepository {
	return &templateVariableRepository{uow: uow}
}

var templateVariableCols = []string{
	"id", "template_id", "key", "description", "var_type", "default_value",
	"is_sensitive", "is_required", "validation_regex", "is_auto_parsed",
	"display_order", "created_at", "updated_at",
}

func (r *templateVariableRepository) scanVariable(row interface{ Scan(dest ...any) error }) (*domain.TemplateVariable, error) {
	var v domain.TemplateVariable
	var cat, uat TimestampDest
	var description, defaultValue, validationRegex sql.NullString
	err := row.Scan(
		&v.ID, &v.TemplateID, &v.Key, &description, &v.VarType, &defaultValue,
		&v.IsSensitive, &v.IsRequired, &validationRegex, &v.IsAutoParsed,
		&v.DisplayOrder, &cat, &uat,
	)
	if err != nil {
		return nil, err
	}
	v.Description = description.String
	v.DefaultValue = defaultValue.String
	v.ValidationRegex = validationRegex.String
	v.CreatedAt = cat.Time()
	v.UpdatedAt = uat.Time()
	return &v, nil
}

func (r *templateVariableRepository) Create(ctx context.Context, variable domain.TemplateVariable) *pkgerrors.Error {
	query, args, err := builder.
		Insert("template_variables").
		Columns("id", "template_id", "key", "description", "var_type", "default_value",
			"is_sensitive", "is_required", "validation_regex", "is_auto_parsed", "display_order").
		Values(variable.ID, variable.TemplateID, variable.Key,
			nullString(variable.Description), variable.VarType, nullString(variable.DefaultValue),
			variable.IsSensitive, variable.IsRequired, nullString(variable.ValidationRegex),
			variable.IsAutoParsed, variable.DisplayOrder).
		Suffix("RETURNING created_at, updated_at").
		ToSql()
	if err != nil {
		return infraerrors.WrapSQLiteError(err, "create_template_variable")
	}

	var cat, uat TimestampDest
	err = r.uow.Querier().QueryRowContext(ctx, query, args...).Scan(&cat, &uat)
	if err != nil {
		return infraerrors.WrapSQLiteError(err, "create_template_variable")
	}

	variable.CreatedAt = cat.Time()
	variable.UpdatedAt = uat.Time()

	return nil
}

func (r *templateVariableRepository) CreateBatch(ctx context.Context, variables []domain.TemplateVariable) *pkgerrors.Error {
	if len(variables) == 0 {
		return nil
	}

	q := builder.
		Insert("template_variables").
		Columns("id", "template_id", "key", "description", "var_type", "default_value",
			"is_sensitive", "is_required", "validation_regex", "is_auto_parsed", "display_order")

	for _, v := range variables {
		q = q.Values(v.ID, v.TemplateID, v.Key,
			nullString(v.Description), v.VarType, nullString(v.DefaultValue),
			v.IsSensitive, v.IsRequired, nullString(v.ValidationRegex),
			v.IsAutoParsed, v.DisplayOrder)
	}

	query, args, err := q.ToSql()
	if err != nil {
		return infraerrors.WrapSQLiteError(err, "create_batch_template_variables")
	}

	_, err = r.uow.Querier().ExecContext(ctx, query, args...)
	if err != nil {
		return infraerrors.WrapSQLiteError(err, "create_batch_template_variables")
	}

	return nil
}

func (r *templateVariableRepository) UpdateBatch(ctx context.Context, variables []domain.TemplateVariable) *pkgerrors.Error {
	if len(variables) == 0 {
		return nil
	}

	for _, v := range variables {
		query, args, err := builder.
			Update("template_variables").
			Set("description", nullString(v.Description)).
			Set("var_type", v.VarType).
			Set("default_value", nullString(v.DefaultValue)).
			Set("is_sensitive", v.IsSensitive).
			Set("is_required", v.IsRequired).
			Set("validation_regex", nullString(v.ValidationRegex)).
			Set("display_order", v.DisplayOrder).
			Set("updated_at", sq.Expr("CURRENT_TIMESTAMP")).
			Where(sq.Eq{"id": v.ID}).
			ToSql()
		if err != nil {
			return infraerrors.WrapSQLiteError(err, "update_batch_template_variables")
		}

		_, err = r.uow.Querier().ExecContext(ctx, query, args...)
		if err != nil {
			return infraerrors.WrapSQLiteError(err, "update_batch_template_variables")
		}
	}

	return nil
}

func (r *templateVariableRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.TemplateVariable, *pkgerrors.Error) {
	query, args, err := builder.
		Select(templateVariableCols...).
		From("template_variables").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, infraerrors.WrapSQLiteError(err, "get_template_variable")
	}

	v, scanErr := r.scanVariable(r.uow.Querier().QueryRowContext(ctx, query, args...))
	if scanErr != nil {
		if scanErr == sql.ErrNoRows {
			return nil, domainerrors.NotFound("TemplateVariable", id.String())
		}
		return nil, infraerrors.WrapSQLiteError(scanErr, "get_template_variable")
	}

	return v, nil
}

func (r *templateVariableRepository) GetByTemplateID(ctx context.Context, templateID uuid.UUID) ([]*domain.TemplateVariable, *pkgerrors.Error) {
	query, args, err := builder.
		Select(templateVariableCols...).
		From("template_variables").
		Where(sq.Eq{"template_id": templateID}).
		OrderBy("display_order ASC", "key ASC").
		ToSql()
	if err != nil {
		return nil, infraerrors.WrapSQLiteError(err, "get_template_variables_by_template")
	}

	rows, err := r.uow.Querier().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, infraerrors.WrapSQLiteError(err, "get_template_variables_by_template")
	}
	defer rows.Close()

	var variables []*domain.TemplateVariable
	for rows.Next() {
		v, scanErr := r.scanVariable(rows)
		if scanErr != nil {
			return nil, infraerrors.WrapSQLiteError(scanErr, "scan_template_variable")
		}
		variables = append(variables, v)
	}

	if err := rows.Err(); err != nil {
		return nil, infraerrors.WrapSQLiteError(err, "iterate_template_variables")
	}

	return variables, nil
}

func (r *templateVariableRepository) GetByTemplateIDAndKey(ctx context.Context, templateID uuid.UUID, key string) (*domain.TemplateVariable, *pkgerrors.Error) {
	query, args, err := builder.
		Select(templateVariableCols...).
		From("template_variables").
		Where(sq.Eq{"template_id": templateID, "key": key}).
		ToSql()
	if err != nil {
		return nil, infraerrors.WrapSQLiteError(err, "get_template_variable_by_key")
	}

	v, scanErr := r.scanVariable(r.uow.Querier().QueryRowContext(ctx, query, args...))
	if scanErr != nil {
		if scanErr == sql.ErrNoRows {
			return nil, domainerrors.NotFoundByField("TemplateVariable", "key", key)
		}
		return nil, infraerrors.WrapSQLiteError(scanErr, "get_template_variable_by_key")
	}

	return v, nil
}

func (r *templateVariableRepository) Update(ctx context.Context, variable domain.TemplateVariable) *pkgerrors.Error {
	query, args, err := builder.
		Update("template_variables").
		Set("description", nullString(variable.Description)).
		Set("var_type", variable.VarType).
		Set("default_value", nullString(variable.DefaultValue)).
		Set("is_sensitive", variable.IsSensitive).
		Set("is_required", variable.IsRequired).
		Set("validation_regex", nullString(variable.ValidationRegex)).
		Set("display_order", variable.DisplayOrder).
		Set("updated_at", sq.Expr("CURRENT_TIMESTAMP")).
		Where(sq.Eq{"id": variable.ID}).
		Suffix("RETURNING updated_at").
		ToSql()
	if err != nil {
		return infraerrors.WrapSQLiteError(err, "update_template_variable")
	}

	var uat TimestampDest
	err = r.uow.Querier().QueryRowContext(ctx, query, args...).Scan(&uat)
	if err != nil {
		if err == sql.ErrNoRows {
			return domainerrors.NotFound("TemplateVariable", variable.ID.String())
		}
		return infraerrors.WrapSQLiteError(err, "update_template_variable")
	}

	variable.UpdatedAt = uat.Time()

	return nil
}

func (r *templateVariableRepository) Delete(ctx context.Context, id uuid.UUID) *pkgerrors.Error {
	query, args, err := builder.
		Delete("template_variables").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return infraerrors.WrapSQLiteError(err, "delete_template_variable")
	}

	result, err := r.uow.Querier().ExecContext(ctx, query, args...)
	if err != nil {
		return infraerrors.WrapSQLiteError(err, "delete_template_variable")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return infraerrors.WrapSQLiteError(err, "get_rows_affected")
	}

	if rowsAffected == 0 {
		return domainerrors.NotFound("TemplateVariable", id.String())
	}

	return nil
}

func (r *templateVariableRepository) DeleteByTemplateIDAndKeys(ctx context.Context, templateID uuid.UUID, keys []string) *pkgerrors.Error {
	if len(keys) == 0 {
		return nil
	}

	query, args, err := builder.
		Delete("template_variables").
		Where(sq.Eq{"template_id": templateID, "key": keys}).
		ToSql()
	if err != nil {
		return infraerrors.WrapSQLiteError(err, "delete_template_variables_by_keys")
	}

	_, err = r.uow.Querier().ExecContext(ctx, query, args...)
	if err != nil {
		return infraerrors.WrapSQLiteError(err, "delete_template_variables_by_keys")
	}

	return nil
}

func nullString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// Ensure compile-time interface compliance
var _ repository.TemplateVariableRepository = (*templateVariableRepository)(nil)

func init() {
	// Verify columns slice matches expected count
	if len(templateVariableCols) != 13 {
		panic(fmt.Sprintf("templateVariableCols has %d columns, expected 13", len(templateVariableCols)))
	}
}
