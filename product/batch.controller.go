package product

import (
	"net/http"

	"github.com/nayefradwi/zanobia_inventory_manager/common"
	"github.com/nayefradwi/zanobia_inventory_manager/transactions"
)

type BatchController struct {
	batchService IBatchService
}

const useMostExpiredKey = "useMostExpired"

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

func (c BatchController) IncrementBatchWithRecipe(w http.ResponseWriter, r *http.Request) {
	useMostExpired := r.URL.Query().Get(useMostExpiredKey)
	common.ParseBody[BatchInput](w, r.Body, func(input BatchInput) {
		ctx := common.SetBoolToContext(r.Context(), UseMostExpiredKey{}, useMostExpired)
		input.Reason = transactions.TransactionReasonTypeProduced
		err := c.batchService.IncrementBatchWithRecipe(ctx, input)
		common.WriteEmptyResponse(common.EmptyResult{
			Error:   err,
			Writer:  w,
			Message: "Batch incremented successfully with recipe",
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
		if len(inputs) > 100 {
			common.WriteEmptyResponse(common.EmptyResult{
				Error:   common.NewBadRequestFromMessage("batch input cannot be more than 100"),
				Writer:  w,
				Message: "Batch increment failed",
			})
			return
		}
		err := c.batchService.BulkIncrementBatch(r.Context(), inputs)
		common.WriteEmptyResponse(common.EmptyResult{
			Error:   err,
			Writer:  w,
			Message: "Batches incremented successfully",
		})
	})
}

func (c BatchController) BulkIncrementBatchWithRecipe(w http.ResponseWriter, r *http.Request) {
	useMostExpired := r.URL.Query().Get(useMostExpiredKey)
	common.ParseBody[[]BatchInput](w, r.Body, func(inputs []BatchInput) {
		if len(inputs) > 25 {
			common.WriteEmptyResponse(common.EmptyResult{
				Error:   common.NewBadRequestFromMessage("batch input cannot be more than 25"),
				Writer:  w,
				Message: "Batch increment with recipe failed",
			})
			return
		}
		ctx := common.SetBoolToContext(r.Context(), UseMostExpiredKey{}, useMostExpired)
		for i := range inputs {
			inputs[i].Reason = transactions.TransactionReasonTypeProduced
		}
		err := c.batchService.BulkIncrementWithRecipeBatch(ctx, inputs)
		common.WriteEmptyResponse(common.EmptyResult{
			Error:   err,
			Writer:  w,
			Message: "Batches incremented successfully with recipe",
		})
	})
}

func (c BatchController) BulkDecrementBatch(w http.ResponseWriter, r *http.Request) {
	common.ParseBody[[]BatchInput](w, r.Body, func(inputs []BatchInput) {
		if len(inputs) > 100 {
			common.WriteEmptyResponse(common.EmptyResult{
				Error:   common.NewBadRequestFromMessage("batch input cannot be more than 100"),
				Writer:  w,
				Message: "Batch decrement failed",
			})
			return
		}
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

func (c BatchController) SearchBatchesBySku(w http.ResponseWriter, r *http.Request) {
	sku := r.URL.Query().Get("sku")
	batchesPage, err := c.batchService.SearchBatchesBySku(r.Context(), sku)
	common.WriteResponse[common.PaginatedResponse[Batch]](
		common.Result[common.PaginatedResponse[Batch]]{
			Error:  err,
			Writer: w,
			Data:   batchesPage,
		},
	)
}
