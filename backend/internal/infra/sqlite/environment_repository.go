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

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

var envColumns = []string{
	"id", "name", "created_at", "created_by", "description",
	"workspace_id", "template_id", "status", "last_applied_at",
	"last_operation", "last_error", "updated_at",
}

type environmentRepository struct {
	uow *UnitOfWork
}

func newEnvironmentRepository(uow *UnitOfWork) repository.EnvironmentRepository {
	return &environmentRepository{uow: uow}
}

// scanEnvironment scans a row into a domain.Environment using the standard column order.
func scanEnvironment(scanner interface{ Scan(dest ...any) error }) (*domain.Environment, error) {
	var env domain.Environment
	var cat, uat, lat TimestampDest
	var lastOp, lastErr sql.NullString
	var status string

	err := scanner.Scan(
		&env.ID,
		&env.Name,
		&cat,
		&env.CreatedBy,
		&env.Description,
		&env.WorkspaceID,
		&env.TemplateID,
		&status,
		&lat,
		&lastOp,
		&lastErr,
		&uat,
	)
	if err != nil {
		return nil, err
	}

	env.CreatedAt = cat.Time()
	env.UpdatedAt = uat.Time()
	env.Status = domain.EnvironmentStatus(status)

	if !lat.Time().IsZero() {
		t := lat.Time()
		env.LastAppliedAt = &t
	}
	if lastOp.Valid {
		env.LastOperation = lastOp.String
	}
	if lastErr.Valid {
		env.LastError = lastErr.String
	}

	return &env, nil
}

func (r *environmentRepository) Create(ctx context.Context, env *domain.Environment) error {
	if env.ID == uuid.Nil {
		env.ID = uuid.New()
	}

	query, args, err := builder.
		Insert("environments").
		Columns("id", "name", "description", "created_by", "workspace_id", "template_id", "status").
		Values(env.ID, env.Name, env.Description, env.CreatedBy, env.WorkspaceID, env.TemplateID, string(env.Status)).
		Suffix("RETURNING created_at, updated_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build insert query: %w", err)
	}

	var cat, uat TimestampDest
	err = r.uow.Querier().QueryRowContext(ctx, query, args...).Scan(&cat, &uat)
	if err != nil {
		return infraerrors.WrapSQLiteError(err, "create_environment")
	}

	env.CreatedAt = cat.Time()
	env.UpdatedAt = uat.Time()

	return nil
}

func (r *environmentRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Environment, error) {
	query, args, err := builder.
		Select(envColumns...).
		From("environments").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	env, scanErr := scanEnvironment(r.uow.Querier().QueryRowContext(ctx, query, args...))
	if scanErr != nil {
		if scanErr == sql.ErrNoRows {
			return nil, domainerrors.NotFound("Environment", id.String())
		}
		return nil, infraerrors.WrapSQLiteError(scanErr, "get_environment")
	}

	return env, nil
}

func (r *environmentRepository) queryMany(ctx context.Context, qb sq.SelectBuilder, op string) ([]*domain.Environment, error) {
	query, args, err := qb.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := r.uow.Querier().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, infraerrors.WrapSQLiteError(err, op)
	}
	defer rows.Close()

	var environments []*domain.Environment
	for rows.Next() {
		env, scanErr := scanEnvironment(rows)
		if scanErr != nil {
			return nil, infraerrors.WrapSQLiteError(scanErr, "scan_environment")
		}
		environments = append(environments, env)
	}

	if err := rows.Err(); err != nil {
		return nil, infraerrors.WrapSQLiteError(err, "iterate_environments")
	}

	return environments, nil
}

func (r *environmentRepository) GetByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) ([]*domain.Environment, error) {
	return r.queryMany(ctx, builder.
		Select(envColumns...).
		From("environments").
		Where(sq.Eq{"workspace_id": workspaceID}).
		OrderBy("created_at DESC"),
		"get_environments_by_workspace",
	)
}

func (r *environmentRepository) GetByCreatedBy(ctx context.Context, userID uuid.UUID) ([]*domain.Environment, error) {
	return r.queryMany(ctx, builder.
		Select(envColumns...).
		From("environments").
		Where(sq.Eq{"created_by": userID}).
		OrderBy("created_at DESC"),
		"get_environments_by_creator",
	)
}

func (r *environmentRepository) GetByTemplateID(ctx context.Context, templateID uuid.UUID) ([]*domain.Environment, error) {
	return r.queryMany(ctx, builder.
		Select(envColumns...).
		From("environments").
		Where(sq.Eq{"template_id": templateID}).
		OrderBy("created_at DESC"),
		"get_environments_by_template",
	)
}

func (r *environmentRepository) Update(ctx context.Context, env *domain.Environment) error {
	qb := builder.
		Update("environments").
		Set("name", env.Name).
		Set("description", env.Description).
		Set("created_by", env.CreatedBy).
		Set("workspace_id", env.WorkspaceID).
		Set("template_id", env.TemplateID).
		Set("status", string(env.Status)).
		Set("last_operation", nilIfEmpty(env.LastOperation)).
		Set("last_error", nilIfEmpty(env.LastError)).
		Set("updated_at", sq.Expr("CURRENT_TIMESTAMP")).
		Where(sq.Eq{"id": env.ID}).
		Suffix("RETURNING updated_at")

	if env.LastAppliedAt != nil {
		qb = qb.Set("last_applied_at", env.LastAppliedAt.UTC().Format("2006-01-02 15:04:05"))
	}

	query, args, err := qb.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build update query: %w", err)
	}

	var uat TimestampDest
	err = r.uow.Querier().QueryRowContext(ctx, query, args...).Scan(&uat)
	if err != nil {
		if err == sql.ErrNoRows {
			return domainerrors.NotFound("Environment", env.ID.String())
		}
		return infraerrors.WrapSQLiteError(err, "update_environment")
	}

	env.UpdatedAt = uat.Time()

	return nil
}

func (r *environmentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query, args, err := builder.
		Delete("environments").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build delete query: %w", err)
	}

	result, err := r.uow.Querier().ExecContext(ctx, query, args...)
	if err != nil {
		return infraerrors.WrapSQLiteError(err, "delete_environment")
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return infraerrors.WrapSQLiteError(err, "get_rows_affected")
	}

	if affected == 0 {
		return domainerrors.NotFound("Environment", id.String())
	}

	return nil
}

func (r *environmentRepository) List(ctx context.Context, opts repository.ListOptions) ([]*domain.Environment, error) {
	opts.ApplyDefaults()
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	return r.queryMany(ctx, builder.
		Select(envColumns...).
		From("environments").
		OrderBy(fmt.Sprintf("%s %s", opts.SortBy, opts.Order)).
		Limit(uint64(opts.Limit)).
		Offset(uint64(opts.Offset)),
		"list_environments",
	)
}

// AcquireOperation atomically transitions the environment to newStatus only if
// the current status is not one of the blocking statuses. This acts as the
// concurrency mutex described in the research doc (Section 3).
func (r *environmentRepository) AcquireOperation(ctx context.Context, id uuid.UUID, newStatus domain.EnvironmentStatus) (*domain.Environment, error) {
	// Build the blocking status strings for the NOT IN clause.
	blocking := make([]string, len(domain.OperationBlockingStatuses))
	for i, s := range domain.OperationBlockingStatuses {
		blocking[i] = string(s)
	}

	query, args, err := builder.
		Update("environments").
		Set("status", string(newStatus)).
		Set("last_operation", domain.OperationFromStatus(newStatus)).
		Set("last_error", nil).
		Set("updated_at", sq.Expr("CURRENT_TIMESTAMP")).
		Where(sq.Eq{"id": id}).
		Where(sq.NotEq{"status": blocking}).
		Suffix("RETURNING " + joinColumns(envColumns)).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build acquire query: %w", err)
	}

	env, scanErr := scanEnvironment(r.uow.Querier().QueryRowContext(ctx, query, args...))
	if scanErr != nil {
		if scanErr == sql.ErrNoRows {
			// Either doesn't exist or is in a blocking status — check which.
			existing, getErr := r.GetByID(ctx, id)
			if getErr != nil {
				return nil, getErr
			}
			return nil, fmt.Errorf("environment is currently %s — cannot start %s", existing.Status, newStatus)
		}
		return nil, infraerrors.WrapSQLiteError(scanErr, "acquire_operation")
	}

	return env, nil
}
