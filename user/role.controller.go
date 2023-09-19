package user

import (
	"net/http"

	"github.com/nayefradwi/zanobia_inventory_manager/common"
)

type RoleController struct {
	service IRoleService
}

func NewRoleController(service IRoleService) RoleController {
	return RoleController{
		service: service,
	}
}

func (c RoleController) CreateRole(w http.ResponseWriter, r *http.Request) {
	common.ParseBody[Role](w, r.Body, func(role Role) {
		err := c.service.CreateRole(r.Context(), role)
		common.WriteCreatedResponse(common.EmptyResult{
			Writer:  w,
			Message: "role created successfully",
			Error:   err,
		})
	})
}

func (c RoleController) GetRoles(w http.ResponseWriter, r *http.Request) {
	roles, err := c.service.GetRoles(r.Context())
	common.WriteResponse[[]Role](common.Result[[]Role]{
		Writer: w,
		Data:   roles,
		Error:  err,
	})
}
