package transactions

import (
	"context"
	"log"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
	"github.com/nayefradwi/zanobia_inventory_manager/warehouse"
)

type ITransactionRepository interface {
	CreateTransactionReason(ctx context.Context, reason TransactionReason) error
	GetTransactionReasons(ctx context.Context) ([]TransactionReason, error)
	InsertTransaction(ctx context.Context, input transactionInput) error
	GetTransactionsOfRetailer(ctx context.Context, retailerId int) ([]Transaction, error)
	GetTransactionsOfRetailerBatch(ctx context.Context, retailerId, retailerBatchId int) ([]Transaction, error)
	GetTransactionsOfSKU(ctx context.Context, sku string) ([]Transaction, error)
	GetTransactionsOfBatch(ctx context.Context, batchId int) ([]Transaction, error)
	GetTransactionsOfWarehouse(ctx context.Context) ([]Transaction, error)
	InsertTransactionToBatch(ctx context.Context, input transactionInput, batch *pgx.Batch)
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

func (r *TransactionRepository) GetTransactionsOfRetailer(ctx context.Context, retailerId int) ([]Transaction, error) {
	sql := `
	SELECT transaction_history.id, user_id, batch_id, retailer_batch_id, warehouse_id, retailer_id, quantity, unit_id, 
	amount, comment, sku, transaction_history.created_at, name, is_positive
	FROM transaction_history
	JOIN transaction_history_reasons ON transaction_history.reason = transaction_history_reasons.name
	WHERE retailer_id = $1
	AND transaction_history.created_at >= CURRENT_DATE - INTERVAL '30 days';
`
	op := common.GetOperator(ctx, r.Pool)
	rows, err := op.Query(ctx, sql, retailerId)
	if err != nil {
		log.Printf("Failed to get transactions of retailer: %s", err.Error())
		return nil, common.NewBadRequestFromMessage("Failed to get transactions of retailer")
	}
	defer rows.Close()
	return r.parseRows(rows)
}

func (r *TransactionRepository) GetTransactionsOfRetailerBatch(ctx context.Context, retailerId, retailerBatchId int) ([]Transaction, error) {
	sql := `
	SELECT transaction_history.id, user_id, batch_id, retailer_batch_id, warehouse_id, retailer_id, quantity, unit_id, 
	amount, comment, sku, transaction_history.created_at, name, is_positive
	FROM transaction_history
	JOIN transaction_history_reasons ON transaction_history.reason = transaction_history_reasons.name
	WHERE retailer_batch_id = $1 AND retailer_id = $2
	AND transaction_history.created_at >= CURRENT_DATE - INTERVAL '30 days';
`
	op := common.GetOperator(ctx, r.Pool)
	rows, err := op.Query(ctx, sql, retailerBatchId, retailerId)
	if err != nil {
		log.Printf("Failed to get transactions of retailer batch: %s", err.Error())
		return nil, common.NewBadRequestFromMessage("Failed to get transactions of retailer batch")
	}
	defer rows.Close()
	return r.parseRows(rows)
}

func (r *TransactionRepository) GetTransactionsOfSKU(ctx context.Context, sku string) ([]Transaction, error) {
	sql := `
	SELECT transaction_history.id, user_id, batch_id, retailer_batch_id, warehouse_id, retailer_id, quantity, unit_id, 
	amount, comment, sku, transaction_history.created_at, name, is_positive
	FROM transaction_history
	JOIN transaction_history_reasons ON transaction_history.reason = transaction_history_reasons.name
	WHERE sku = $1 AND (warehouse_id = $2 OR warehouse_id IS NULL)
	AND transaction_history.created_at >= CURRENT_DATE - INTERVAL '30 days';
`
	op := common.GetOperator(ctx, r.Pool)
	rows, err := op.Query(ctx, sql, sku, warehouse.GetWarehouseId(ctx))
	if err != nil {
		log.Printf("Failed to get transactions for SKU: %s", err.Error())
		return nil, common.NewBadRequestFromMessage("Failed to get transactions for SKU")
	}
	defer rows.Close()
	return r.parseRows(rows)
}

func (r *TransactionRepository) GetTransactionsOfBatch(ctx context.Context, batchId int) ([]Transaction, error) {
	sql := `
	SELECT transaction_history.id, user_id, batch_id, retailer_batch_id, warehouse_id, retailer_id, quantity, unit_id, 
	amount, comment, sku, transaction_history.created_at, name, is_positive
	FROM transaction_history
	JOIN transaction_history_reasons ON transaction_history.reason = transaction_history_reasons.name
	WHERE batch_id = $1 AND warehouse_id = $2
	AND transaction_history.created_at >= CURRENT_DATE - INTERVAL '30 days';
`
	op := common.GetOperator(ctx, r.Pool)
	rows, err := op.Query(ctx, sql, batchId, warehouse.GetWarehouseId(ctx))
	if err != nil {
		log.Printf("Failed to get transactions for batch: %s", err.Error())
		return nil, common.NewBadRequestFromMessage("Failed to get transactions for batch")
	}
	defer rows.Close()
	return r.parseRows(rows)
}

func (r *TransactionRepository) GetTransactionsOfWarehouse(ctx context.Context) ([]Transaction, error) {
	sql := `
	SELECT transaction_history.id, user_id, batch_id, retailer_batch_id, warehouse_id, retailer_id, quantity, unit_id, 
	amount, comment, sku, transaction_history.created_at, name, is_positive
	FROM transaction_history
	JOIN transaction_history_reasons ON transaction_history.reason = transaction_history_reasons.name
	WHERE warehouse_id = $1
	AND transaction_history.created_at >= CURRENT_DATE - INTERVAL '30 days';
`
	op := common.GetOperator(ctx, r.Pool)
	rows, err := op.Query(ctx, sql, warehouse.GetWarehouseId(ctx))
	if err != nil {
		log.Printf("Failed to get transactions for warehouse: %s", err.Error())
		return nil, common.NewBadRequestFromMessage("Failed to get transactions for warehouse")
	}
	defer rows.Close()
	return r.parseRows(rows)
}

func (r *TransactionRepository) parseRows(rows pgx.Rows) ([]Transaction, error) {
	transactions := make([]Transaction, 0)
	for rows.Next() {
		transaction := Transaction{}
		transactionReason := TransactionReason{}
		err := rows.Scan(&transaction.Id, &transaction.UserId, &transaction.BatchId,
			&transaction.RetailerBatchId, &transaction.WarehouseId, &transaction.RetailerId,
			&transaction.Quantity, &transaction.UnitId, &transaction.Amount,
			&transaction.Comment, &transaction.Sku,
			&transaction.CreatedAt, &transactionReason.Name, &transactionReason.IsPositive,
		)
		if err != nil {
			log.Printf("Failed to scan transaction: %s", err.Error())
			return nil, common.NewBadRequestFromMessage("Failed to get transactions")
		}
		transaction.Reason = transactionReason
		transactions = append(transactions, transaction)
	}
	return transactions, nil
}

func (r *TransactionRepository) InsertTransactionToBatch(
	ctx context.Context,
	input transactionInput,
	batch *pgx.Batch,
) {
	batch.Queue(
		`INSERT INTO transaction_history (user_id, batch_id, retailer_batch_id, warehouse_id, retailer_id, 
		quantity, unit_id, amount, reason, comment, sku) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		input.UserId, input.BatchId, input.RetailerBatchId, input.WarehouseId, input.RetailerId,
		input.Quantity, input.UnitId, input.Amount, input.Reason, input.Comment, input.Sku,
	)
}
