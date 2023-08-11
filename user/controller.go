package user

import "net/http"

type UserController struct {
	service IUserService
}

func (c UserController) CreateUser(w http.ResponseWriter, r *http.Request) {
	// do something
}
