package user

type IUserService interface {
	CleanUp()
}

type UserService struct {
	repository IUserRepository
}

func NewUserService(repository IUserRepository) IUserService {
	return &UserService{
		repository: repository,
	}
}

func (s *UserService) CleanUp() {
	s.repository.cleanUp()
}
