package sqlite

import (
	"context"
	"database/sql"
	"time"

	"backend/internal/domain"
	"backend/internal/domain/repository"
	infraerrors "backend/internal/infra/errors"
	pkgerrors "backend/pkg/errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

const timestampFormat = "2006-01-02 15:04:05"

type teardownQueueRepository struct {
	uow *UnitOfWork
}

func newTeardownQueueRepository(uow *UnitOfWork) repository.TeardownQueueRepository {
	return &teardownQueueRepository{uow: uow}
}

func (r *teardownQueueRepository) Enqueue(ctx context.Context, entry *domain.TeardownEntry) *pkgerrors.Error {
	query, args, err := builder.
		Insert("teardown_queue").
		Columns("environment_id", "teardown_at", "status").
		Values(entry.EnvironmentID, entry.TeardownAt.UTC().Format(timestampFormat), string(entry.Status)).
		Suffix("RETURNING created_at, updated_at").
		ToSql()
	if err != nil {
		return infraerrors.WrapSQLiteError(err, "enqueue_teardown")
	}

	var cat, uat TimestampDest
	err = r.uow.Querier().QueryRowContext(ctx, query, args...).Scan(&cat, &uat)
	if err != nil {
		return infraerrors.WrapSQLiteError(err, "enqueue_teardown")
	}

	entry.CreatedAt = cat.Time()
	entry.UpdatedAt = uat.Time()

	return nil
}

func (r *teardownQueueRepository) FindDue(ctx context.Context, now time.Time) (*domain.TeardownEntry, *pkgerrors.Error) {
	query, args, err := builder.
		Select("environment_id", "teardown_at", "status", "created_at", "updated_at").
		From("teardown_queue").
		Where(sq.Eq{"status": string(domain.TeardownStatusPending)}).
		Where(sq.LtOrEq{"teardown_at": now.UTC().Format(timestampFormat)}).
		OrderBy("teardown_at ASC").
		Limit(1).
		ToSql()
	if err != nil {
		return nil, infraerrors.WrapSQLiteError(err, "find_due_teardown")
	}

	var entry domain.TeardownEntry
	var teardownAt, cat, uat TimestampDest
	var status string

	err = r.uow.Querier().QueryRowContext(ctx, query, args...).Scan(
		&entry.EnvironmentID,
		&teardownAt,
		&status,
		&cat,
		&uat,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, infraerrors.WrapSQLiteError(err, "find_due_teardown")
	}

	entry.TeardownAt = teardownAt.Time()
	entry.Status = domain.TeardownStatus(status)
	entry.CreatedAt = cat.Time()
	entry.UpdatedAt = uat.Time()

	return &entry, nil
}

func (r *teardownQueueRepository) UpdateStatus(ctx context.Context, envID uuid.UUID, status domain.TeardownStatus) *pkgerrors.Error {
	query, args, err := builder.
		Update("teardown_queue").
		Set("status", string(status)).
		Set("updated_at", sq.Expr("CURRENT_TIMESTAMP")).
		Where(sq.Eq{"environment_id": envID}).
		ToSql()
	if err != nil {
		return infraerrors.WrapSQLiteError(err, "update_teardown_status")
	}

	_, err = r.uow.Querier().ExecContext(ctx, query, args...)
	if err != nil {
		return infraerrors.WrapSQLiteError(err, "update_teardown_status")
	}

	return nil
}

func (r *teardownQueueRepository) ResetProcessing(ctx context.Context) *pkgerrors.Error {
	query, args, err := builder.
		Update("teardown_queue").
		Set("status", string(domain.TeardownStatusPending)).
		Set("updated_at", sq.Expr("CURRENT_TIMESTAMP")).
		Where(sq.Eq{"status": string(domain.TeardownStatusProcessing)}).
		ToSql()
	if err != nil {
		return infraerrors.WrapSQLiteError(err, "reset_processing_teardowns")
	}

	_, err = r.uow.Querier().ExecContext(ctx, query, args...)
	if err != nil {
		return infraerrors.WrapSQLiteError(err, "reset_processing_teardowns")
	}

	return nil
}
