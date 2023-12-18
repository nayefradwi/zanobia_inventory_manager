package transactions

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
)

type TransactionController struct {
	service ITransactionService
}

func NewTransactionController(service ITransactionService) TransactionController {
	return TransactionController{
		service,
	}
}

func (c TransactionController) CreateTransactionReason(w http.ResponseWriter, r *http.Request) {
	common.ParseBody[TransactionReason](w, r.Body, func(tr TransactionReason) {
		err := c.service.CreateTransactionReason(r.Context(), tr)
		common.WriteCreatedResponse(common.EmptyResult{
			Writer:  w,
			Error:   err,
			Message: "Transaction reason created successfully",
		})
	})
}

func (c TransactionController) GetTransactionReasons(w http.ResponseWriter, r *http.Request) {
	reasons, err := c.service.GetTransactionReasons(r.Context())
	common.WriteResponse[[]TransactionReason](common.Result[[]TransactionReason]{
		Writer: w,
		Error:  err,
		Data:   reasons,
	})
}

func (c TransactionController) GetTransactionsOfBatch(w http.ResponseWriter, r *http.Request) {
	id := common.GetIntURLParam(r, "id")
	transactions, err := c.service.GetTransactionsOfBatch(r.Context(), id)
	common.WriteResponse[[]Transaction](common.Result[[]Transaction]{
		Writer: w,
		Error:  err,
		Data:   transactions,
	})
}

func (c TransactionController) GetTransactionsOfSku(w http.ResponseWriter, r *http.Request) {
	sku := chi.URLParam(r, "sku")
	transactions, err := c.service.GetTransactionsOfSKU(r.Context(), sku)
	common.WriteResponse[[]Transaction](common.Result[[]Transaction]{
		Writer: w,
		Error:  err,
		Data:   transactions,
	})
}

func (c TransactionController) GetTransactionsOfRetailer(w http.ResponseWriter, r *http.Request) {
	id := common.GetIntURLParam(r, "id")
	transactions, err := c.service.GetTransactionsOfRetailer(r.Context(), id)
	common.WriteResponse[[]Transaction](common.Result[[]Transaction]{
		Writer: w,
		Error:  err,
		Data:   transactions,
	})
}

func (c TransactionController) GetTransactionsOfRetailerBatch(w http.ResponseWriter, r *http.Request) {
	id := common.GetIntURLParam(r, "id")
	batchId := common.GetIntURLParam(r, "batchId")
	transactions, err := c.service.GetTransactionsOfRetailerBatch(r.Context(), id, batchId)
	common.WriteResponse[[]Transaction](common.Result[[]Transaction]{
		Writer: w,
		Error:  err,
		Data:   transactions,
	})
}

func (c TransactionController) GetTransactionsOfMyWarehouse(w http.ResponseWriter, r *http.Request) {
	transactions, err := c.service.GetTransactionsOfWarehouse(r.Context())
	common.WriteResponse[[]Transaction](common.Result[[]Transaction]{
		Writer: w,
		Error:  err,
		Data:   transactions,
	})
}
