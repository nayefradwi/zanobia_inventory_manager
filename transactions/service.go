package transactions

import "context"

type ITransactionService interface {
	CreateTransactionReason(ctx context.Context, reason TransactionReason) error
	GetTransactionReasons(ctx context.Context) ([]TransactionReason, error)
}

type TransactionService struct {
	repo ITransactionRepository
}

func NewTransactionService(repo ITransactionRepository) *TransactionService {
	return &TransactionService{
		repo,
	}
}

func (s *TransactionService) CreateTransactionReason(ctx context.Context, reason TransactionReason) error {
	if err := ValidateTransactionReason(reason); err != nil {
		return err
	}
	return s.repo.CreateTransactionReason(ctx, reason)
}

func (s *TransactionService) GetTransactionReasons(ctx context.Context) ([]TransactionReason, error) {
	return s.repo.GetTransactionReasons(ctx)
}
