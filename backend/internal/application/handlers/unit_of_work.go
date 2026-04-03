package handlers

import (
	"backend/internal/domain/repository"
	"backend/pkg/errors"
)

type (
	UnitOfWorkFactory interface {
		Create() UnitOfWork
	}
	UnitOfWork interface {
		Begin() *errors.Error
		Commit() *errors.Error
		Rollback() *errors.Error
	}
	RepositoryFactory interface {
		CreateUserRepository(uow UnitOfWork) repository.UserRepository
		CreateWorkspaceRepository(uow UnitOfWork) repository.WorkspaceRepository
		CreateTemplateRepository(uow UnitOfWork) repository.TemplateRepository
		CreateEnvironmentRepository(uow UnitOfWork) repository.EnvironmentRepository
		CreateTemplateVariableRepository(uow UnitOfWork) repository.TemplateVariableRepository
		CreateEnvironmentVariableValueRepository(uow UnitOfWork) repository.EnvironmentVariableValueRepository
		CreateTeardownQueueRepository(uow UnitOfWork) repository.TeardownQueueRepository
	}
)
