package product

import (
	"net/http"

	"github.com/nayefradwi/zanobia_inventory_manager/common"
)

type InventoryController struct {
	service IInventoryService
}

func NewInventoryController(service IInventoryService) InventoryController {
	return InventoryController{
		service: service,
	}
}

func (c InventoryController) IncrementInventory(w http.ResponseWriter, r *http.Request) {
	common.ParseBody[InventoryInput](w, r.Body, func(inventoryInput InventoryInput) {
		err := c.service.IncrementInventory(r.Context(), inventoryInput)
		common.WriteCreatedResponse(common.EmptyResult{
			Error:   err,
			Writer:  w,
			Message: "Inventory incremented successfully",
		})
	})
}

func (c InventoryController) DecrementInventory(w http.ResponseWriter, r *http.Request) {
	common.ParseBody[InventoryInput](w, r.Body, func(inventoryInput InventoryInput) {
		err := c.service.DecrementInventory(r.Context(), inventoryInput)
		common.WriteCreatedResponse(common.EmptyResult{
			Error:   err,
			Writer:  w,
			Message: "Inventory decremented successfully",
		})
	})
}

func (c InventoryController) BulkIncrementInventory(w http.ResponseWriter, r *http.Request) {
	common.ParseBody[[]InventoryInput](w, r.Body, func(inventoryInputs []InventoryInput) {
		err := c.service.BulkIncrementInventory(r.Context(), inventoryInputs)
		common.WriteCreatedResponse(common.EmptyResult{
			Error:   err,
			Writer:  w,
			Message: "Inventories incremented successfully",
		})
	})
}

func (c InventoryController) BulkDecrementInventory(w http.ResponseWriter, r *http.Request) {
	common.ParseBody[[]InventoryInput](w, r.Body, func(inputs []InventoryInput) {
		err := c.service.BulkDecrementInventory(r.Context(), inputs)
		common.WriteCreatedResponse(common.EmptyResult{
			Error:   err,
			Writer:  w,
			Message: "Inventories decremented successfully",
		})
	})
}
