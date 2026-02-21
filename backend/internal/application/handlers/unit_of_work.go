package handlers

import "backend/pkg/errors"

type UnitOfWork interface {
	Commit() *errors.Error
	Rollback() *errors.Error
	Begin() *errors.Error
}
