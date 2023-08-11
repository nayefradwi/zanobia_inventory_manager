package user

import "context"

type IUserService interface {
	Create(ctx context.Context, user UserInput) error
}

type UserService struct {
	repository IUserRepository
}

func NewUserService(repository IUserRepository) IUserService {
	return &UserService{
		repository: repository,
	}
}

func (s *UserService) CreateSysAdmin() error {
	return nil
}

func (s *UserService) Create(ctx context.Context, user UserInput) error {
	validationError := ValidateUser(user)
	if validationError != nil {
		return validationError
	}
	creationError := s.repository.Create(ctx, user)
	if creationError != nil {
		return creationError
	}
	return nil
}
