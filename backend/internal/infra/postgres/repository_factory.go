package postgres

import (
	apphandlers "backend/internal/application/handlers"
	"backend/internal/domain/repository"
)

type repositoryFactory struct{}

func NewRepositoryFactory() apphandlers.RepositoryFactory {
	return &repositoryFactory{}
}

func (f *repositoryFactory) CreateUserRepository(uow apphandlers.UnitOfWork) repository.UserRepository {
	return NewUserRepository(uow.(*UnitOfWork))
}

func (f *repositoryFactory) CreateWorkspaceRepository(uow apphandlers.UnitOfWork) repository.WorkspaceRepository {
	return NewWorkspaceRepository(uow.(*UnitOfWork))
}

func (f *repositoryFactory) CreateTemplateRepository(uow apphandlers.UnitOfWork) repository.TemplateRepository {
	return NewTemplateRepository(uow.(*UnitOfWork))
}
