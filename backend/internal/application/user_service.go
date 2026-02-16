package application

import (
	"backend/internal/domain"
	"backend/internal/domain/repository"
	"backend/pkg/contracts"
	"backend/pkg/errors"
	"context"
)

type UserService struct {
	userRepository repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return UserService{
		userRepository: userRepo,
	}
}

func (s UserService) CreateLocalUser(ctx context.Context, request contracts.CreateLocalUser) *errors.Error {
	var (
		err  *errors.Error
		user domain.UserAggregate
	)
	_, err = s.userRepository.GetByEmail(ctx, request.Email)
	if err != nil {
		return err
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
		return err
	}

	err = s.userRepository.Create(ctx, user)
	if err != nil {
		return err
	}
	return nil
}
