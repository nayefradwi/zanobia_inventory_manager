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
