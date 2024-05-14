package warehouse

import (
	"net/http"

	"github.com/nayefradwi/zanobia_inventory_manager/common"
)

type WarehouseController struct {
	service IWarehouseService
}

func NewWarehouseController(service IWarehouseService) WarehouseController {
	return WarehouseController{
		service: service,
	}
}

func (c WarehouseController) CreateWarehouse(w http.ResponseWriter, r *http.Request) {
	common.ParseBody[Warehouse](w, r.Body, func(warehouse Warehouse) {
		err := c.service.CreateWarehouse(r.Context(), warehouse)
		common.WriteCreatedResponse(common.EmptyResult{
			Error:   err,
			Writer:  w,
			Message: "warehouse created successfully",
		})
	})
}

func (c WarehouseController) GetWarehouses(w http.ResponseWriter, r *http.Request) {
	warehouses, err := c.service.GetWarehouses(r.Context())
	common.WriteResponse(common.Result[[]Warehouse]{
		Error:  err,
		Writer: w,
		Data:   warehouses,
	})
}

func (c WarehouseController) AddUserToWarehouse(w http.ResponseWriter, r *http.Request) {
	common.ParseBody[WarehouseUserInput](w, r.Body, func(input WarehouseUserInput) {
		err := c.service.AddUserToWarehouse(r.Context(), input)
		common.WriteCreatedResponse(common.EmptyResult{
			Error:   err,
			Message: "user added to warehouse successfully",
			Writer:  w,
		})
	})
}

func (c WarehouseController) GetCurrentWarehouse(w http.ResponseWriter, r *http.Request) {
	warehouse, err := c.service.GetMyCurrentWarehouse(r.Context())
	common.WriteResponse[Warehouse](common.Result[Warehouse]{
		Data:   warehouse,
		Error:  err,
		Writer: w,
	})
}

func (c WarehouseController) UpdateWarehouse(w http.ResponseWriter, r *http.Request) {
	common.ParseBody[Warehouse](w, r.Body, func(warehouse Warehouse) {
		err := c.service.UpdateWarehouse(r.Context(), warehouse)
		common.WriteCreatedResponse(common.EmptyResult{
			Error:   err,
			Message: "warehouse updated successfully",
			Writer:  w,
		})
	})
}
