package user

import (
	"net/http"

	"github.com/nayefradwi/zanobia_inventory_manager/common"
)

type PermissionController struct {
	service IPermissionService
}

func NewPermissionController(service IPermissionService) PermissionController {
	return PermissionController{
		service: service,
	}
}

func (c PermissionController) InitiateInitialPermissions(w http.ResponseWriter, r *http.Request) {
	err := c.service.InitiateInitialPermissions(r.Context())
	common.WriteCreatedResponse(common.EmptyResult{
		Message: "Initial permissions created successfully",
		Writer:  w,
		Error:   err,
	})
}

func (c PermissionController) CreatePermission(w http.ResponseWriter, r *http.Request) {
	common.ParseBody[Permission](w, r.Body, func(data Permission) {
		err := c.service.CreatePermission(r.Context(), data)
		common.WriteCreatedResponse(common.EmptyResult{
			Message: "Permission created successfully",
			Writer:  w,
			Error:   err,
		})
	})
}
