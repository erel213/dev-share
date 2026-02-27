package postgres

import (
	"context"
	"database/sql"

	"backend/internal/application/handlers"
	"backend/pkg/errors"
)

type UnitOfWork struct {
	db     *sql.DB
	tx     *sql.Tx
	depth  int
	failed bool
}

type Querier interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

func NewUnitOfWork(db *sql.DB) handlers.UnitOfWork {
	return &UnitOfWork{db: db}
}

func (u *UnitOfWork) Begin() *errors.Error {
	if u.depth == 0 {
		tx, err := u.db.Begin()
		if err != nil {
			return errors.Wrap(err, "failed to begin transaction").
				WithCode(errors.CodeInternal).
				WithHTTPStatus(500)
		}
		u.tx = tx
	}
	u.depth++
	return nil
}

func (u *UnitOfWork) Commit() *errors.Error {
	if u.depth == 0 {
		return errors.New("no active transaction").
			WithCode(errors.CodeInternal).
			WithHTTPStatus(500)
	}
	u.depth--
	if u.depth > 0 {
		return nil // nested call — outer caller will commit
	}
	if u.failed {
		return u.doRollback()
	}
	err := u.tx.Commit()
	u.tx = nil
	if err != nil {
		return errors.Wrap(err, "failed to commit transaction").
			WithCode(errors.CodeInternal).
			WithHTTPStatus(500)
	}
	return nil
}

func (u *UnitOfWork) Rollback() *errors.Error {
	if u.depth == 0 {
		return nil // safe no-op after successful commit
	}
	u.failed = true
	u.depth = 0
	return u.doRollback()
}

func (u *UnitOfWork) doRollback() *errors.Error {
	err := u.tx.Rollback()
	u.tx = nil
	u.failed = false
	if err != nil {
		return errors.Wrap(err, "failed to rollback transaction").
			WithCode(errors.CodeInternal).
			WithHTTPStatus(500)
	}
	return nil
}

func (u *UnitOfWork) Querier() Querier {
	if u.tx != nil {
		return u.tx
	}
	return u.db
}
