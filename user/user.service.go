package user

import (
	"context"
	"log"

	"github.com/nayefradwi/zanobia_inventory_manager/common"
)

type IUserService interface {
	Create(ctx context.Context, user UserInput) error
	InitiateSystemAdmin(ctx context.Context) error
	GetAllUsers(ctx context.Context) ([]User, error)
	LoginUser(ctx context.Context, input UserLoginInput) (common.Token, error)
	GetUserById(ctx context.Context, id int) (User, error)
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

func (s *UserService) GetAllUsers(ctx context.Context) ([]User, error) {
	return s.Repository.GetAllUsers(ctx)
}

func (s *UserService) LoginUser(ctx context.Context, input UserLoginInput) (common.Token, error) {
	user, _ := s.Repository.GetUserByEmail(ctx, input.Email)
	hash := user.Hash
	if hash == nil || *hash == "" {
		return common.Token{}, common.NewBadRequestError("User with this email is not found", "user_not_found")
	}
	match := common.CompareHash(input.Password, *hash)
	if !match {
		return common.Token{}, common.NewBadRequestError("Password is incorrect", "password_incorrect")
	}
	user.Hash = nil
	user.Email = nil
	userClaim, err := common.StructToMap(user)
	if err != nil {
		log.Printf("failed to convert user to map: %s", err.Error())
		return common.Token{}, common.NewInternalServerError()
	}
	return common.GenerateAccessToken(userClaim)
}

func (s *UserService) GetUserById(ctx context.Context, id int) (User, error) {
	return s.Repository.GetUserById(ctx, id)
}
