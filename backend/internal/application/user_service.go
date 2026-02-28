package application

import (
	"backend/internal/application/handlers"
	"backend/internal/domain"
	domainerrors "backend/internal/domain/errors"
	"backend/internal/domain/repository"
	"backend/pkg/contracts"
	"backend/pkg/errors"
	"backend/pkg/validation"
	"context"
)

type UserService struct {
	userRepository repository.UserRepository
	validator      *validation.Service
}

func NewUserService(userRepo repository.UserRepository, validator *validation.Service) UserService {
	return UserService{
		userRepository: userRepo,
		validator:      validator,
	}
}

// CreateLocalUser creates a new local (password-based) user.
// It accepts a UnitOfWork so it can participate in a caller-managed transaction.
// The caller is responsible for deferring uow.Rollback() when this method is
// the outermost transaction boundary.
func (s UserService) CreateLocalUser(ctx context.Context, uow handlers.UnitOfWork, request contracts.CreateLocalUser) (domain.UserAggregate, *errors.Error) {
	var (
		err  *errors.Error
		user domain.UserAggregate
	)

	if err = s.validator.Validate(request); err != nil {
		return domain.UserAggregate{}, err
	}

	// Check if user already exists with this email
	_, err = s.userRepository.GetByEmail(ctx, request.Email)
	if err != nil && err.HTTPStatus() != domainerrors.ErrNotFound.HTTPStatus() {
		return domain.UserAggregate{}, err
	}

	userFactory := domain.UserFactory{}
	user, err = userFactory.Create(
		nil,
		nil,
		request.Name,
		request.Email,
		&request.Password,
		request.WorkspaceID,
	)

	if err != nil {
		return domain.UserAggregate{}, err
	}

	if beginErr := uow.Begin(); beginErr != nil {
		return domain.UserAggregate{}, beginErr
	}

	// Persist user
	err = s.userRepository.Create(ctx, user)
	if err != nil {
		return domain.UserAggregate{}, err
	}

	if commitErr := uow.Commit(); commitErr != nil {
		return domain.UserAggregate{}, commitErr
	}

	return user, nil
}
