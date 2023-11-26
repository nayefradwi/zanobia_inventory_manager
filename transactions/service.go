package transactions

type ITransactionService interface{}

type TransactionService struct {
	repo ITransactionRepository
}

func NewTransactionService(repo ITransactionRepository) *TransactionService {
	return &TransactionService{
		repo,
	}
}
