package product

import (
	"net/http"

	"github.com/nayefradwi/zanobia_inventory_manager/common"
)

type UnitController struct {
	service IUnitService
}

func NewUnitController(service IUnitService) *UnitController {
	return &UnitController{service}
}

func (c UnitController) CreateUnit(w http.ResponseWriter, r *http.Request) {
	common.ParseBody[Unit](w, r.Body, func(unit Unit) {
		err := c.service.CreateUnit(r.Context(), unit)
		common.WriteCreatedResponse(common.EmptyResult{
			Error:   err,
			Writer:  w,
			Message: "Created unit successfully",
		})
	})
}

func (c UnitController) GetAllUnits(w http.ResponseWriter, r *http.Request) {
	units, err := c.service.GetAllUnits(r.Context())
	common.WriteResponse(common.Result[[]Unit]{
		Error:  err,
		Writer: w,
		Data:   units,
	})
}
