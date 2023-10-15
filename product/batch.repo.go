package product

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
	"github.com/nayefradwi/zanobia_inventory_manager/warehouse"
)

type IBatchRepository interface{}

type BatchRepository struct {
	*pgxpool.Pool
}

func NewBatchRepository(pool *pgxpool.Pool) *BatchRepository {
	return &BatchRepository{pool}
}

func (r *BatchRepository) CreateBatch(ctx context.Context, input BatchInput, expiresAt time.Time) error {
	sql := `INSERT INTO batches (sku, warehouse_id, quantity, unit_id, expires_at) VALUES ($1, $2, $3, $4, $5)`
	op := common.GetOperator(ctx, r.Pool)
	warehouseId := warehouse.GetWarehouseId(ctx)
	_, err := op.Exec(ctx, sql, input.Sku, warehouseId, input.Quantity, input.UnitId, expiresAt)
	if err != nil {
		log.Printf("Failed to create batch: %s", err.Error())
		return common.NewBadRequestFromMessage("Failed to create batch")
	}
	return nil
}

func (r *BatchRepository) UpdateBatch(ctx context.Context, base BatchBase) error {
	updatedAt := time.Now().UTC()
	op := common.GetOperator(ctx, r.Pool)
	sql := `UPDATE batches SET quantity = $1, updated_at = $2 WHERE id = $3`
	_, err := op.Exec(ctx, sql, base.Quantity, updatedAt, base.Id)
	if err != nil {
		log.Printf("Failed to update batch: %s", err.Error())
		return common.NewBadRequestFromMessage("Failed to update batch")
	}
	return nil
}

func (r *BatchRepository) GetBatchBase(ctx context.Context, productId int, expirationDate string) (Batch, error) {
	// Get batch base by product id and expiration date
	return Batch{}, nil
}

func (r *BatchRepository) GetBatches(ctx context.Context, pageSize int, cursor string) ([]Batch, error) {
	// Get batches paginated sorted by expiration date
	return []Batch{}, nil
}
