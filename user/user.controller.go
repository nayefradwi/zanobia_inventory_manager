package user

import (
	"net/http"

	"github.com/nayefradwi/zanobia_inventory_manager/common"
)

type UserController struct {
	service IUserService
}

func NewUserController(service IUserService) UserController {
	return UserController{
		service: service,
	}
}

func (c UserController) InitiateSysAdminUser(w http.ResponseWriter, r *http.Request) {
	result := common.EmptyResult{
		Writer:  w,
		Message: "System admin user created successfully",
		Error:   c.service.InitiateSystemAdmin(r.Context()),
	}
	common.WriteCreatedResponse(result)
}

func (c UserController) CreateUser(w http.ResponseWriter, r *http.Request) {
	common.ParseBody[UserInput](w, r.Body, func(data UserInput) {
		err := c.service.Create(r.Context(), data)
		common.WriteCreatedResponse(common.EmptyResult{
			Message: "User created successfully",
			Writer:  w,
			Error:   err,
		})
	})
}

func (c UserController) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := c.service.GetAllUsers(r.Context())
	common.WriteResponse(common.Result[[]User]{
		Writer: w,
		Data:   users,
		Error:  err,
	})
}

func (c UserController) LoginUser(w http.ResponseWriter, r *http.Request) {
	common.ParseBody[UserLoginInput](w, r.Body, func(data UserLoginInput) {
		token, err := c.service.LoginUser(r.Context(), data)
		common.WriteResponse(common.Result[common.Token]{
			Writer: w,
			Data:   token,
			Error:  err,
		})
	})
}

func (c UserController) GetUserByContext(w http.ResponseWriter, r *http.Request) {
	user, err := c.service.GetUserByContext(r.Context())
	common.WriteResponse(common.Result[User]{
		Writer: w,
		Data:   user,
		Error:  err,
	})
}
