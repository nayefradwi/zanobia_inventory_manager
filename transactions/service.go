package transactions

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
	"go.uber.org/zap"
)

type ITransactionService interface {
	InitiateAllReasons(ctx context.Context) error
	CreateTransactionReason(ctx context.Context, reason TransactionReason) error
	GetTransactionReasons(ctx context.Context) ([]TransactionReason, error)
	CreateWarehouseTransaction(ctx context.Context, command CreateWarehouseTransactionCommand) error
	CreateRetailerTransaction(ctx context.Context, command CreateRetailerTransactionCommand) error
	GetTransactionsOfRetailer(ctx context.Context, retailerId int) ([]Transaction, error)
	GetTransactionsOfRetailerBatch(ctx context.Context, retailerId, retailerBatchId int) ([]Transaction, error)
	GetTransactionsOfSKU(ctx context.Context, sku string) ([]Transaction, error)
	GetTransactionsOfBatch(ctx context.Context, batchId int) ([]Transaction, error)
	GetTransactionsOfWarehouse(ctx context.Context) ([]Transaction, error)
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

func (r *TransactionService) GetTransactionsOfRetailer(ctx context.Context, retailerId int) ([]Transaction, error) {
	return r.repo.GetTransactionsOfRetailer(ctx, retailerId)
}

func (r *TransactionService) GetTransactionsOfRetailerBatch(ctx context.Context, retailerId, retailerBatchId int) ([]Transaction, error) {
	return r.repo.GetTransactionsOfRetailerBatch(ctx, retailerId, retailerBatchId)
}

func (r *TransactionService) GetTransactionsOfSKU(ctx context.Context, sku string) ([]Transaction, error) {
	return r.repo.GetTransactionsOfSKU(ctx, sku)
}

func (r *TransactionService) GetTransactionsOfBatch(ctx context.Context, batchId int) ([]Transaction, error) {
	return r.repo.GetTransactionsOfBatch(ctx, batchId)
}

func (r *TransactionService) GetTransactionsOfWarehouse(ctx context.Context) ([]Transaction, error) {
	return r.repo.GetTransactionsOfWarehouse(ctx)
}

func (r *TransactionService) CreateTransactionHistoryBatches(
	ctx context.Context,
	transactionCommands []CreateWarehouseTransactionCommand,
) (*pgx.Batch, error) {
	batch := &pgx.Batch{}
	for _, command := range transactionCommands {
		input, err := ForWarehouseTransactions(ctx, command)
		if err != nil {
			return nil, err
		}
		r.repo.InsertTransactionToBatch(ctx, input, batch)
	}
	return batch, nil
}

func (r *TransactionService) CreateRetailerTransactionHistoryBatches(
	ctx context.Context,
	transactionCommands []CreateRetailerTransactionCommand,
) (*pgx.Batch, error) {
	batch := &pgx.Batch{}
	for _, command := range transactionCommands {
		input, err := ForRetailerTransactions(ctx, command)
		if err != nil {
			return nil, err
		}
		r.repo.InsertTransactionToBatch(ctx, input, batch)
	}
	return batch, nil
}

func (r *TransactionService) InitiateAllReasons(ctx context.Context) error {
	for _, reason := range initalTransactionReasons {
		if err := r.repo.CreateTransactionReason(ctx, reason); err != nil {
			common.LoggerFromCtx(ctx).Error(
				"failed to initiate reason",
				zap.String("reason", reason.Name),
				zap.Error(err),
			)
		}
	}
	return nil
}
