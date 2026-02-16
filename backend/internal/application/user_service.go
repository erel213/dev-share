package application

import (
	"backend/internal/domain"
	domainerrors "backend/internal/domain/errors"
	"backend/internal/domain/repository"
	"backend/pkg/contracts"
	"backend/pkg/errors"
	"context"

	"github.com/google/uuid"
)

type UserService struct {
	userRepository repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return UserService{
		userRepository: userRepo,
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

	err = s.userRepository.Create(ctx, user)
	if err != nil {
		return domain.UserAggregate{}, err
	}
	return user, nil
}
