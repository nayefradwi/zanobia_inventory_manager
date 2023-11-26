package transactions

import (
	"context"
	"log"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
)

type ITransactionRepository interface {
	CreateTransactionReason(ctx context.Context, reason TransactionReason) error
	GetTransactionReasons(ctx context.Context) ([]TransactionReason, error)
	InsertTransaction(ctx context.Context, input transactionInput) error
}

type TransactionRepository struct {
	*pgxpool.Pool
}

func NewTransactionRepository(db *pgxpool.Pool) *TransactionRepository {
	return &TransactionRepository{
		db,
	}
}

func (r *TransactionRepository) CreateTransactionReason(ctx context.Context, reason TransactionReason) error {
	sql := `INSERT INTO transaction_history_reasons (name, description, is_positive) VALUES ($1, $2, $3)`
	op := common.GetOperator(ctx, r.Pool)
	_, err := op.Exec(ctx, sql, reason.Name, reason.Description, reason.IsPositive)
	if err != nil {
		log.Printf("Failed to create transaction reason: %s", err.Error())
		return common.NewBadRequestFromMessage("Failed to create transaction reason")
	}
	return nil
}

func (r *TransactionRepository) GetTransactionReasons(ctx context.Context) ([]TransactionReason, error) {
	sql := `SELECT id, name, description, is_positive FROM transaction_history_reasons`
	op := common.GetOperator(ctx, r.Pool)
	rows, err := op.Query(ctx, sql)
	if err != nil {
		log.Printf("Failed to get transaction reasons: %s", err.Error())
		return nil, common.NewBadRequestFromMessage("Failed to get transaction reasons")
	}
	defer rows.Close()
	reasons := make([]TransactionReason, 0)
	for rows.Next() {
		reason := TransactionReason{}
		err := rows.Scan(&reason.Id, &reason.Name, &reason.Description, &reason.IsPositive)
		if err != nil {
			log.Printf("Failed to scan transaction reason: %s", err.Error())
			return nil, common.NewBadRequestFromMessage("Failed to get transaction reasons")
		}
		reasons = append(reasons, reason)
	}
	return reasons, nil
}

func (r *TransactionRepository) InsertTransaction(ctx context.Context, input transactionInput) error {
	sql := `
		INSERT INTO transaction_history (user_id, batch_id, retailer_batch_id, warehouse_id, retailer_id, 
		quantity, unit_id, amount, reason, comment, sku) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		`
	op := common.GetOperator(ctx, r.Pool)
	_, err := op.Exec(
		ctx, sql, input.UserId, input.BatchId, input.RetailerBatchId, input.WarehouseId, input.RetailerId,
		input.Quantity, input.UnitId, input.Amount, input.Reason, input.Comment, input.Sku,
	)
	if err != nil {
		log.Printf("Failed to insert transaction: %s", err.Error())
		return common.NewBadRequestFromMessage("Failed to insert transaction")
	}
	return nil
}
