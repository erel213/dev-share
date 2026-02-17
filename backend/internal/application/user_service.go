package application

import (
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

func (s UserService) CreateLocalUser(ctx context.Context, request contracts.CreateLocalUser) (domain.UserAggregate, *errors.Error) {
	var (
		err  *errors.Error
		user domain.UserAggregate
	)

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

	// Persist user
	err = s.userRepository.Create(ctx, user)
	if err != nil {
		return domain.UserAggregate{}, err
	}
	return user, nil
}
