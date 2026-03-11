package sqlite

import (
	"context"

	"backend/internal/domain"
	"backend/internal/domain/repository"
	infraerrors "backend/internal/infra/errors"
	pkgerrors "backend/pkg/errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

type environmentVariableValueRepository struct {
	uow *UnitOfWork
}

func newEnvironmentVariableValueRepository(uow *UnitOfWork) repository.EnvironmentVariableValueRepository {
	return &environmentVariableValueRepository{uow: uow}
}

func (r *environmentVariableValueRepository) Create(ctx context.Context, value domain.EnvironmentVariableValue) *pkgerrors.Error {
	query, args, err := builder.
		Insert("environment_variable_values").
		Columns("id", "environment_id", "template_variable_id", "value").
		Values(value.ID, value.EnvironmentID, value.TemplateVariableID, value.Value).
		Suffix("RETURNING created_at, updated_at").
		ToSql()
	if err != nil {
		return infraerrors.WrapSQLiteError(err, "create_env_var_value")
	}

	var cat, uat TimestampDest
	err = r.uow.Querier().QueryRowContext(ctx, query, args...).Scan(&cat, &uat)
	if err != nil {
		return infraerrors.WrapSQLiteError(err, "create_env_var_value")
	}

	value.CreatedAt = cat.Time()
	value.UpdatedAt = uat.Time()

	return nil
}

func (r *environmentVariableValueRepository) GetByEnvironmentID(ctx context.Context, environmentID uuid.UUID) ([]*domain.EnvironmentVariableValue, *pkgerrors.Error) {
	query, args, err := builder.
		Select("id", "environment_id", "template_variable_id", "value", "created_at", "updated_at").
		From("environment_variable_values").
		Where(sq.Eq{"environment_id": environmentID}).
		ToSql()
	if err != nil {
		return nil, infraerrors.WrapSQLiteError(err, "get_env_var_values")
	}

	rows, err := r.uow.Querier().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, infraerrors.WrapSQLiteError(err, "get_env_var_values")
	}
	defer rows.Close()

	var values []*domain.EnvironmentVariableValue
	for rows.Next() {
		var v domain.EnvironmentVariableValue
		var cat, uat TimestampDest
		err := rows.Scan(&v.ID, &v.EnvironmentID, &v.TemplateVariableID, &v.Value, &cat, &uat)
		if err != nil {
			return nil, infraerrors.WrapSQLiteError(err, "scan_env_var_value")
		}
		v.CreatedAt = cat.Time()
		v.UpdatedAt = uat.Time()
		values = append(values, &v)
	}

	if err := rows.Err(); err != nil {
		return nil, infraerrors.WrapSQLiteError(err, "iterate_env_var_values")
	}

	return values, nil
}

func (r *environmentVariableValueRepository) UpsertBatch(ctx context.Context, values []domain.EnvironmentVariableValue) *pkgerrors.Error {
	for _, v := range values {
		query := `INSERT INTO environment_variable_values (id, environment_id, template_variable_id, value)
			VALUES (?, ?, ?, ?)
			ON CONFLICT(environment_id, template_variable_id) DO UPDATE SET
				value = excluded.value,
				updated_at = strftime('%Y-%m-%d %H:%M:%S', 'now')`

		_, err := r.uow.Querier().ExecContext(ctx, query, v.ID, v.EnvironmentID, v.TemplateVariableID, v.Value)
		if err != nil {
			return infraerrors.WrapSQLiteError(err, "upsert_env_var_value")
		}
	}

	return nil
}

func (r *environmentVariableValueRepository) DeleteByEnvironmentID(ctx context.Context, environmentID uuid.UUID) *pkgerrors.Error {
	query, args, err := builder.
		Delete("environment_variable_values").
		Where(sq.Eq{"environment_id": environmentID}).
		ToSql()
	if err != nil {
		return infraerrors.WrapSQLiteError(err, "delete_env_var_values")
	}

	_, err = r.uow.Querier().ExecContext(ctx, query, args...)
	if err != nil {
		return infraerrors.WrapSQLiteError(err, "delete_env_var_values")
	}

	return nil
}

var _ repository.EnvironmentVariableValueRepository = (*environmentVariableValueRepository)(nil)
