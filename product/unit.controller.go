package product

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
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

func (c UnitController) CreateConversion(w http.ResponseWriter, r *http.Request) {
	common.ParseBody[UnitConversion](w, r.Body, func(conversion UnitConversion) {
		err := c.service.CreateConversion(r.Context(), conversion)
		common.WriteCreatedResponse(common.EmptyResult{
			Error:   err,
			Writer:  w,
			Message: "Created conversion successfully",
		})
	})
}

func (c UnitController) CreateConversionFromName(w http.ResponseWriter, r *http.Request) {
	common.ParseBody[UnitConversionInput](w, r.Body, func(input UnitConversionInput) {
		err := c.service.CreateConversionFromName(r.Context(), input)
		common.WriteCreatedResponse(common.EmptyResult{
			Error:   err,
			Writer:  w,
			Message: "Created conversion successfully",
		})
	})
}

func (c UnitController) ConvertUnit(w http.ResponseWriter, r *http.Request) {
	common.ParseBody[ConvertUnitInput](w, r.Body, func(input ConvertUnitInput) {
		result, err := c.service.ConvertUnit(r.Context(), input)
		common.WriteResponse(common.Result[ConvertUnitOutput]{
			Error:  err,
			Writer: w,
			Data:   result,
		})
	})
}

func (c UnitController) GetUnitById(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.Atoi(idStr)
	unit, err := c.service.GetUnitById(r.Context(), &id)
	common.WriteResponse[Unit](
		common.Result[Unit]{
			Data:   unit,
			Error:  err,
			Writer: w,
		})
}
