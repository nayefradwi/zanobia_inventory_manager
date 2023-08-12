package user

import "net/http"

type UserController struct {
	service IUserService
}

func NewUserController(service IUserService) UserController {
	return UserController{
		service: service,
	}
}

func (c UserController) InitiateSysAdminUser() {
	// do something
}

func (c UserController) CreateUser(w http.ResponseWriter, r *http.Request) {
	// do something
}
