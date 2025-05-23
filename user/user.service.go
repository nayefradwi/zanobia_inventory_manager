package user

import (
	"context"

	"github.com/nayefradwi/zanobia_inventory_manager/common"
	"go.uber.org/zap"
)

type IUserService interface {
	Create(ctx context.Context, user UserInput) error
	InitiateSystemAdmin(ctx context.Context) error
	GetAllUsers(ctx context.Context) ([]User, error)
	LoginUser(ctx context.Context, input UserLoginInput) (common.Token, error)
	GetUserById(ctx context.Context, id int) (User, error)
	GetUserByContext(ctx context.Context) (User, error)
	BanUser(ctx context.Context, id int) error
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
		common.LoggerFromCtx(ctx).Error("failed to hash", zap.Error(hashError))
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
		PermissionHandles: []string{SysAdminPermissionHandle},
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
		common.LoggerFromCtx(ctx).Error("User with this email is not found")
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
		common.LoggerFromCtx(ctx).Error("failed to convert user to map", zap.Error(err))
		return common.Token{}, common.NewInternalServerError()
	}
	return common.GenerateAccessToken(userClaim)
}

func (s *UserService) GetUserById(ctx context.Context, id int) (User, error) {
	user, err := s.Repository.GetUserById(ctx, id)
	if err != nil {
		return User{}, err
	}
	user.Hash = nil
	user.Email = nil
	return user, nil
}

func (s *UserService) GetUserByContext(ctx context.Context) (User, error) {
	userVal := ctx.Value(common.UserKey{})
	if userVal != nil {
		user := userVal.(User)
		user.Hash = nil
		return user, nil
	}
	return User{}, common.NewUnAuthorizedError("Aunauthorized user")
}

func (s *UserService) BanUser(ctx context.Context, id int) error {
	currentUser := ctx.Value(common.UserKey{}).(User)
	if currentUser.Id == id {
		return common.NewBadRequestError("You can't ban yourself", "ban_self")
	}
	return s.Repository.BanUser(ctx, id)
}
