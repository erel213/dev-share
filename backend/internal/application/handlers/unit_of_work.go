package handlers

import (
	"backend/pkg/errors"
)

type UnitOfWork interface {
	Begin() *errors.Error
	Commit() *errors.Error
	Rollback() *errors.Error
}
