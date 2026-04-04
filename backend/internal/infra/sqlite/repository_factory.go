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

func (f *repositoryFactory) CreateTemplateRepository(uow apphandlers.UnitOfWork) repository.TemplateRepository {
	return newTemplateRepository(uow.(*UnitOfWork))
}

func (f *repositoryFactory) CreateEnvironmentRepository(uow apphandlers.UnitOfWork) repository.EnvironmentRepository {
	return newEnvironmentRepository(uow.(*UnitOfWork))
}

func (f *repositoryFactory) CreateTemplateVariableRepository(uow apphandlers.UnitOfWork) repository.TemplateVariableRepository {
	return newTemplateVariableRepository(uow.(*UnitOfWork))
}

func (f *repositoryFactory) CreateEnvironmentVariableValueRepository(uow apphandlers.UnitOfWork) repository.EnvironmentVariableValueRepository {
	return newEnvironmentVariableValueRepository(uow.(*UnitOfWork))
}

func (f *repositoryFactory) CreateTeardownQueueRepository(uow apphandlers.UnitOfWork) repository.TeardownQueueRepository {
	return newTeardownQueueRepository(uow.(*UnitOfWork))
}

func (f *repositoryFactory) CreateGroupRepository(uow apphandlers.UnitOfWork) repository.GroupRepository {
	return newGroupRepository(uow.(*UnitOfWork))
}
