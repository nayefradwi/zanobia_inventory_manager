package user

type IUserService interface{}

type UserService struct {
	repository IUserRepository
}

func NewUserService(repository IUserRepository) IUserService {
	return &UserService{
		repository: repository,
	}
}
