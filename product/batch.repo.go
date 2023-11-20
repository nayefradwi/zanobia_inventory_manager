package product

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
	"github.com/nayefradwi/zanobia_inventory_manager/warehouse"
)

type IBatchRepository interface {
	CreateBatch(ctx context.Context, input BatchInput, expiresAt string) error
	UpdateBatch(ctx context.Context, base BatchBase) error
	GetBatchBaseById(ctx context.Context, id *int) (BatchBase, error)
	GetBatchBase(ctx context.Context, sku string, expirationDate string) (BatchBase, error)
	GetBatches(ctx context.Context, params common.PaginationParams) ([]Batch, error)
	SearchBatchesBySku(ctx context.Context, sku string, params common.PaginationParams) ([]Batch, error)
}

const baseBatchListingSql = `
select b.id, b.sku, b.quantity, b.expires_at, utx.unit_id, utx.name, utx.symbol,
pvartx.name, pvar.id, pvar.price, pvar.product_id, ptx.name from batches b
join unit_translations utx on utx.unit_id = b.unit_id
join product_variants pvar on pvar.sku = b.sku
join product_translations ptx on ptx.product_id = pvar.product_id
join product_variant_translations pvartx on pvartx.product_variant_id = pvar.id and utx.language_code = pvartx.language_code
`

type BatchRepository struct {
	*pgxpool.Pool
}

func NewBatchRepository(pool *pgxpool.Pool) *BatchRepository {
	return &BatchRepository{pool}
}

func (r *BatchRepository) CreateBatch(ctx context.Context, input BatchInput, expiresAt string) error {
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
	sql := `UPDATE batches SET quantity = $1, updated_at = $2 WHERE id = $3 and warehouse_id = $4`
	_, err := op.Exec(ctx, sql, base.Quantity, updatedAt, base.Id, base.WarehouseId)
	if err != nil {
		log.Printf("Failed to update batch: %s", err.Error())
		return common.NewBadRequestFromMessage("Failed to update batch")
	}
	return nil
}

func (r *BatchRepository) GetBatchBase(ctx context.Context, sku string, expirationDate string) (BatchBase, error) {
	sql := `SELECT id, warehouse_id, sku, quantity, unit_id, expires_at FROM batches WHERE sku = $1 AND expires_at = $2 and warehouse_id = $3`
	op := common.GetOperator(ctx, r.Pool)
	warehouseId := warehouse.GetWarehouseId(ctx)
	row := op.QueryRow(ctx, sql, sku, expirationDate, warehouseId)
	var batchBase BatchBase
	err := row.Scan(&batchBase.Id, &batchBase.WarehouseId, &batchBase.Sku, &batchBase.Quantity, &batchBase.UnitId, &batchBase.ExpiresAt)
	if err != nil {
		log.Printf("Failed to get batch base: %s", err.Error())
		return BatchBase{}, common.NewBadRequestFromMessage("Failed to get batch base")
	}
	return batchBase, nil
}

func (r *BatchRepository) GetBatchBaseById(ctx context.Context, id *int) (BatchBase, error) {
	if id == nil {
		return BatchBase{}, common.NewBadRequestFromMessage("Failed to get batch base")
	}
	sql := `SELECT id, warehouse_id, sku, quantity, unit_id, expires_at FROM batches WHERE id = $1 and warehouse_id = $2`
	op := common.GetOperator(ctx, r.Pool)
	warehouseId := warehouse.GetWarehouseId(ctx)
	row := op.QueryRow(ctx, sql, id, warehouseId)
	var batchBase BatchBase
	err := row.Scan(&batchBase.Id, &batchBase.WarehouseId, &batchBase.Sku, &batchBase.Quantity, &batchBase.UnitId, &batchBase.ExpiresAt)
	if err != nil {
		log.Printf("Failed to get batch base: %s", err.Error())
		return BatchBase{}, common.NewBadRequestFromMessage("Failed to get batch base")
	}
	return batchBase, nil
}

func (r *BatchRepository) GetBatches(ctx context.Context, params common.PaginationParams) ([]Batch, error) {
	sqlBuilder := common.NewPaginationQueryBuilder(
		baseBatchListingSql,
		[]string{"b.expires_at DESC", "b.id ASC"},
	)
	q, sql := sqlBuilder.
		WithConditions([]string{
			"utx.language_code = $1",
			"AND",
			"b.warehouse_id = $2",
		}).
		WithCursor(params.EndCursor, params.PreviousCursor).
		WithCursorKeys([]string{"b.id", "b.expires_at"}).
		WithPageSize(params.PageSize).
		WithDirection(params.Direction).
		Build()
	op := common.GetOperator(ctx, r.Pool)
	warehouseId := warehouse.GetWarehouseId(ctx)
	lang := common.GetLanguageParam(ctx)
	rows, err := q.Query(ctx, op, sql, lang, warehouseId)
	if err != nil {
		log.Printf("Failed to get batches: %s", err.Error())
		return []Batch{}, common.NewBadRequestFromMessage("Failed to get batches")
	}
	return r.parseBatchRows(rows)
}

func (r *BatchRepository) SearchBatchesBySku(ctx context.Context, sku string, params common.PaginationParams) ([]Batch, error) {
	sqlBuilder := common.NewPaginationQueryBuilder(
		baseBatchListingSql,
		[]string{"b.expires_at DESC", "b.id ASC"},
	)
	q, sql := sqlBuilder.
		WithConditions([]string{
			"utx.language_code = $1",
			"AND",
			"b.warehouse_id = $2",
			"AND",
			"b.sku = $3",
		}).
		WithCursor(params.EndCursor, params.PreviousCursor).
		WithCursorKeys([]string{"b.id", "b.expires_at"}).
		WithPageSize(params.PageSize).
		WithDirection(params.Direction).
		Build()
	op := common.GetOperator(ctx, r.Pool)
	warehouseId := warehouse.GetWarehouseId(ctx)
	lang := common.GetLanguageParam(ctx)
	rows, err := q.Query(ctx, op, sql, lang, warehouseId, sku)
	if err != nil {
		log.Printf("Failed to get batches: %s", err.Error())
		return []Batch{}, common.NewBadRequestFromMessage("Failed to search batches")
	}
	return r.parseBatchRows(rows)
}

func (r *BatchRepository) parseBatchRows(rows pgx.Rows) ([]Batch, error) {
	var batches []Batch
	for rows.Next() {
		var batch Batch
		var productVariantBase ProductVariantBase
		var unit Unit
		err := rows.Scan(
			&batch.Id, &batch.Sku, &batch.Quantity, &batch.ExpiresAt,
			&unit.Id, &unit.Name, &unit.Symbol,
			&productVariantBase.Name, &productVariantBase.Id, &productVariantBase.Price,
			&productVariantBase.ProductId, &batch.ProductName,
		)
		if err != nil {
			log.Printf("Failed to scan batches: %s", err.Error())
			return []Batch{}, common.NewBadRequestFromMessage("Failed to scan batches")
		}
		batch.Unit = unit
		batch.ProductVariantBase = &productVariantBase
		batches = append(batches, batch)
	}
	return batches, nil
}
