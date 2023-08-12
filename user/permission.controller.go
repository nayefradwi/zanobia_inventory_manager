package user

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
)

const (
	PermissionHandleParam = "permissionHandle"
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

func (c PermissionController) GetPermissionByHandle(w http.ResponseWriter, r *http.Request) {
	handle := chi.URLParam(r, PermissionHandleParam)
	permission, err := c.service.FindPermissionByHandle(r.Context(), handle)
	common.WriteResponse(common.Result[Permission]{
		Writer: w,
		Data:   permission,
		Error:  err,
	})
}
