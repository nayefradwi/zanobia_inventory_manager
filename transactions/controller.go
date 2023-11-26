package transactions

import "net/http"

type TransactionController struct {
	service ITransactionService
}

func NewTransactionController(service ITransactionService) TransactionController {
	return TransactionController{
		service,
	}
}

func (c TransactionController) CreateTransactionReason(w http.ResponseWriter, r *http.Request) {
	// TODO fill
}

func (c TransactionController) GetTransactionsOfBatch(w http.ResponseWriter, r *http.Request) {
	// TODO fill
}

func (c TransactionController) GetTransactionsOfUser(w http.ResponseWriter, r *http.Request) {
	// TODO fill
}

func (c TransactionController) GetTransactionReasons(w http.ResponseWriter, r *http.Request) {
	// TODO fill
}

func (c TransactionController) GetTransactionsOfSku(w http.ResponseWriter, r *http.Request) {
	// TODO fill
}

func (c TransactionController) GetTransactionsOfRetailer(w http.ResponseWriter, r *http.Request) {
	// TODO fill
}
