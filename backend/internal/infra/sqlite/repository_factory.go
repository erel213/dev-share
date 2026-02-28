package sqlite

import (
	apphandlers "backend/internal/application/handlers"
	"backend/internal/domain/repository"
)

type repositoryFactory struct{}

func NewRepositoryFactory() apphandlers.RepositoryFactory {
	return &repositoryFactory{}
}

func (f *repositoryFactory) CreateUserRepository(uow apphandlers.UnitOfWork) repository.UserRepository {
	return newUserRepository(uow.(*UnitOfWork))
}

func (f *repositoryFactory) CreateWorkspaceRepository(uow apphandlers.UnitOfWork) repository.WorkspaceRepository {
	return newWorkspaceRepository(uow.(*UnitOfWork))
}
