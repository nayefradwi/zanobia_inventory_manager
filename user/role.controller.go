package user

import (
	"net/http"

	"github.com/nayefradwi/zanobia_inventory_manager/common"
)

type RoleController struct {
	service IRoleService
}

func NewRoleController(service IRoleService) *RoleController {
	return &RoleController{
		service: service,
	}
}

func (c RoleController) CreateRole(w http.ResponseWriter, r *http.Request) {
	common.ParseBody[RoleInput](w, r.Body, func(roleInput RoleInput) {
		err := c.service.CreateRole(r.Context(), roleInput)
		common.WriteCreatedResponse(common.EmptyResult{
			Writer:  w,
			Message: "role created successfully",
			Error:   err,
		})
	})
}
