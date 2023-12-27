package retailer

import (
	"net/http"

	"github.com/nayefradwi/zanobia_inventory_manager/common"
)

type RetailerBatchController struct {
	service IRetailerBatchService
}
type DecrementWarehouseKey struct{}

func NewRetailerBatchController(service IRetailerBatchService) *RetailerBatchController {
	return &RetailerBatchController{
		service,
	}
}
func (c RetailerBatchController) IncrementBatch(w http.ResponseWriter, r *http.Request) {
	common.ParseBody[RetailerBatchInput](w, r.Body, func(input RetailerBatchInput) {
		err := c.service.IncrementBatch(r.Context(), input)
		common.WriteEmptyResponse(common.EmptyResult{
			Error:   err,
			Writer:  w,
			Message: "Batch incremented successfully",
		})
	})
}

func (c RetailerBatchController) DecrementBatch(w http.ResponseWriter, r *http.Request) {
	common.ParseBody[RetailerBatchInput](w, r.Body, func(input RetailerBatchInput) {
		err := c.service.DecrementBatch(r.Context(), input)
		common.WriteEmptyResponse(common.EmptyResult{
			Error:   err,
			Writer:  w,
			Message: "Batch decremented successfully",
		})
	})
}

func (c RetailerBatchController) BulkIncrementBatch(w http.ResponseWriter, r *http.Request) {
	common.ParseBody[[]RetailerBatchInput](w, r.Body, func(inputs []RetailerBatchInput) {
		err := c.service.BulkIncrementBatch(r.Context(), inputs)
		common.WriteEmptyResponse(common.EmptyResult{
			Error:   err,
			Writer:  w,
			Message: "Batches incremented successfully",
		})
	})
}

func (c RetailerBatchController) BulkDecrementBatch(w http.ResponseWriter, r *http.Request) {
	common.ParseBody[[]RetailerBatchInput](w, r.Body, func(inputs []RetailerBatchInput) {
		err := c.service.BulkDecrementBatch(r.Context(), inputs)
		common.WriteEmptyResponse(common.EmptyResult{
			Error:   err,
			Writer:  w,
			Message: "Batches decremented successfully",
		})
	})
}

func (c RetailerBatchController) GetBatchesOfRetailer(w http.ResponseWriter, r *http.Request) {
	id := common.GetIntURLParam(r, "id")
	batchesPage, err := c.service.GetBatchesOfRetailer(r.Context(), id)
	common.WriteResponse[common.PaginatedResponse[RetailerBatch]](
		common.Result[common.PaginatedResponse[RetailerBatch]]{
			Error:  err,
			Writer: w,
			Data:   batchesPage,
		},
	)
}

func (c RetailerBatchController) SearchBatchesBySku(w http.ResponseWriter, r *http.Request) {
	sku := r.URL.Query().Get("sku")
	id := common.GetIntURLParam(r, "id")
	batchesPage, err := c.service.SearchBatchesBySku(r.Context(), id, sku)
	common.WriteResponse[common.PaginatedResponse[RetailerBatch]](
		common.Result[common.PaginatedResponse[RetailerBatch]]{
			Error:  err,
			Writer: w,
			Data:   batchesPage,
		},
	)
}

func (c RetailerBatchController) GetBatches(w http.ResponseWriter, r *http.Request) {
	batchesPage, err := c.service.GetBatches(r.Context())
	common.WriteResponse[common.PaginatedResponse[RetailerBatch]](
		common.Result[common.PaginatedResponse[RetailerBatch]]{
			Error:  err,
			Writer: w,
			Data:   batchesPage,
		},
	)
}
