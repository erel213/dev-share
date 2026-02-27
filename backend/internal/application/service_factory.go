package application

import (
	apphandlers "backend/internal/application/handlers"
	"backend/pkg/validation"
)

type ServiceFactory struct {
	uowFactory  apphandlers.UnitOfWorkFactory
	repoFactory apphandlers.RepositoryFactory
	validator   *validation.Service
}

func NewServiceFactory(
	uowFactory apphandlers.UnitOfWorkFactory,
	repoFactory apphandlers.RepositoryFactory,
	validator *validation.Service,
) *ServiceFactory {
	return &ServiceFactory{uowFactory: uowFactory, repoFactory: repoFactory, validator: validator}
}

func (f *ServiceFactory) NewUserService() (UserService, apphandlers.UnitOfWork) {
	uow := f.uowFactory.Create()
	return NewUserService(f.repoFactory.CreateUserRepository(uow), f.validator), uow
}

func (f *ServiceFactory) NewWorkspaceService() (WorkspaceService, apphandlers.UnitOfWork) {
	uow := f.uowFactory.Create()
	return NewWorkspaceService(f.repoFactory.CreateWorkspaceRepository(uow), f.validator), uow
}

func (f *ServiceFactory) NewAdminService() (*AdminService, apphandlers.UnitOfWork) {
	uow := f.uowFactory.Create()
	userRepo := f.repoFactory.CreateUserRepository(uow)
	workspaceRepo := f.repoFactory.CreateWorkspaceRepository(uow)
	userService := NewUserService(userRepo, f.validator)
	return NewAdminService(workspaceRepo, userService, userRepo, f.validator), uow
}
