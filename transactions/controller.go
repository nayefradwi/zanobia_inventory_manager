package transactions

import (
	"net/http"

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

func (c TransactionController) GetTransactionsOfBatch(w http.ResponseWriter, r *http.Request) {
	// TODO fill
}

func (c TransactionController) GetTransactionsOfUser(w http.ResponseWriter, r *http.Request) {
	// TODO fill
}

func (c TransactionController) GetTransactionReasons(w http.ResponseWriter, r *http.Request) {
	reasons, err := c.service.GetTransactionReasons(r.Context())
	common.WriteResponse[[]TransactionReason](common.Result[[]TransactionReason]{
		Writer: w,
		Error:  err,
		Data:   reasons,
	})
}

func (c TransactionController) GetTransactionsOfSku(w http.ResponseWriter, r *http.Request) {
	// TODO fill
}

func (c TransactionController) GetTransactionsOfRetailer(w http.ResponseWriter, r *http.Request) {
	// TODO fill
}
