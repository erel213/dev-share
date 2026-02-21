package postgres

import (
	"backend/internal/application/handlers"
	"backend/pkg/errors"
	"database/sql"
)

type UnitOfWork struct {
	db *sql.DB
	tx *sql.Tx
}

func NewUnitOfWork(db *sql.DB) handlers.UnitOfWork {
	return &UnitOfWork{db: db}
}

func (u *UnitOfWork) Begin() *errors.Error {
	if u.tx != nil {
		return nil
	}
	tx, err := u.db.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction").
			WithCode(errors.CodeInternal).
			WithHTTPStatus(500)
	}
	u.tx = tx
	return nil
}

func (u *UnitOfWork) Commit() *errors.Error {
	if u.tx == nil {
		return errors.New("no transaction to commit").
			WithCode(errors.CodeInternal).
			WithHTTPStatus(500)
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
	if u.tx == nil {
		return errors.New("no transaction to rollback").
			WithCode(errors.CodeInternal).
			WithHTTPStatus(500)
	}
	err := u.tx.Rollback()
	u.tx = nil
	if err != nil {
		return errors.Wrap(err, "failed to rollback transaction").
			WithCode(errors.CodeInternal).
			WithHTTPStatus(500)
	}
	return nil
}
