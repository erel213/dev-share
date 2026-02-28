package sqlite

import (
	"database/sql"

	apphandlers "backend/internal/application/handlers"
)

type unitOfWorkFactory struct{ db *sql.DB }

func NewUnitOfWorkFactory(db *sql.DB) apphandlers.UnitOfWorkFactory {
	return &unitOfWorkFactory{db: db}
}

func (f *unitOfWorkFactory) Create() apphandlers.UnitOfWork {
	return NewUnitOfWork(f.db)
}
