package product

import (
	"net/http"

	"github.com/nayefradwi/zanobia_inventory_manager/common"
)

type BatchController struct {
	batchService IBatchService
}

func NewBatchController(batchService IBatchService) BatchController {
	return BatchController{
		batchService,
	}
}

func (c BatchController) IncrementBatch(w http.ResponseWriter, r *http.Request) {
	common.ParseBody[BatchInput](w, r.Body, func(input BatchInput) {
		err := c.batchService.IncrementBatch(r.Context(), input)
		common.WriteEmptyResponse(common.EmptyResult{
			Error:   err,
			Writer:  w,
			Message: "Batch incremented successfully",
		})
	})
}

func (c BatchController) DecrementBatch(w http.ResponseWriter, r *http.Request) {
	common.ParseBody[BatchInput](w, r.Body, func(input BatchInput) {
		err := c.batchService.DecrementBatch(r.Context(), input)
		common.WriteEmptyResponse(common.EmptyResult{
			Error:   err,
			Writer:  w,
			Message: "Batch decremented successfully",
		})
	})
}

func (c BatchController) BulkIncrementBatch(w http.ResponseWriter, r *http.Request) {
	common.ParseBody[[]BatchInput](w, r.Body, func(inputs []BatchInput) {
		err := c.batchService.BulkIncrementBatch(r.Context(), inputs)
		common.WriteEmptyResponse(common.EmptyResult{
			Error:   err,
			Writer:  w,
			Message: "Batches incremented successfully",
		})
	})
}

func (c BatchController) BulkDecrementBatch(w http.ResponseWriter, r *http.Request) {
	common.ParseBody[[]BatchInput](w, r.Body, func(inputs []BatchInput) {
		err := c.batchService.BulkDecrementBatch(r.Context(), inputs)
		common.WriteEmptyResponse(common.EmptyResult{
			Error:   err,
			Writer:  w,
			Message: "Batches decremented successfully",
		})
	})
}

func (c BatchController) GetBatches(w http.ResponseWriter, r *http.Request) {
	batchesPage, err := c.batchService.GetBatches(r.Context())
	common.WriteResponse[common.PaginatedResponse[Batch]](
		common.Result[common.PaginatedResponse[Batch]]{
			Error:  err,
			Writer: w,
			Data:   batchesPage,
		},
	)
}
