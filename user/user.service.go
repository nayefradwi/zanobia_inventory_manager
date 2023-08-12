package user

import (
	"context"
	"log"

	"github.com/nayefradwi/zanobia_inventory_manager/common"
)

type IUserService interface {
	Create(ctx context.Context, user UserInput) error
	InitiateSystemAdmin(ctx context.Context) error
}

type UserServiceInput struct {
	Repository       IUserRepository
	SysAdminEmail    string
	SysAdminPassword string
}

type UserService struct {
	UserServiceInput
}

func NewUserService(input UserServiceInput) IUserService {
	return &UserService{
		UserServiceInput: input,
	}
}

func (s *UserService) Create(ctx context.Context, user UserInput) error {
	validationError := ValidateUser(user)
	if validationError != nil {
		return validationError
	}
	hashPassword, hashError := common.Hash(user.Password)
	if hashError != nil {
		log.Printf("failed to hash: %s", hashError.Error())
		return common.NewInternalServerError()
	}
	user.Password = hashPassword
	creationError := s.Repository.Create(ctx, user)
	if creationError != nil {
		return creationError
	}
	return nil
}

func (s *UserService) InitiateSystemAdmin(ctx context.Context) error {
	userInput := UserInput{
		Email:             s.SysAdminEmail,
		Password:          s.SysAdminPassword,
		FirstName:         "System",
		LastName:          "Admin",
		PermissionHandles: []string{sysAdminPermissionHandle},
	}
	return s.Create(ctx, userInput)
}
