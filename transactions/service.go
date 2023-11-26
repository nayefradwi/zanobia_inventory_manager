package transactions

import "context"

type ITransactionService interface {
	CreateTransactionReason(ctx context.Context, reason TransactionReason) error
	GetTransactionReasons(ctx context.Context) ([]TransactionReason, error)
	CreateWarehouseTransaction(ctx context.Context, command CreateWarehouseTransactionCommand) error
	CreateRetailerTransaction(ctx context.Context, command CreateRetailerTransactionCommand) error
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

func (s *TransactionService) CreateWarehouseTransaction(ctx context.Context, command CreateWarehouseTransactionCommand) error {
	input, err := ForWarehouseTransactions(ctx, command)
	if err != nil {
		return err
	}
	return s.repo.InsertTransaction(ctx, input)
}

func (s *TransactionService) CreateRetailerTransaction(ctx context.Context, command CreateRetailerTransactionCommand) error {
	input, err := ForRetailerTransactions(ctx, command)
	if err != nil {
		return err
	}
	return s.repo.InsertTransaction(ctx, input)
}
