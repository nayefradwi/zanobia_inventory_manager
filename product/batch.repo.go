package product

import (
	"context"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
	"github.com/nayefradwi/zanobia_inventory_manager/unit"
	"github.com/nayefradwi/zanobia_inventory_manager/warehouse"
	"go.uber.org/zap"
)

type IBatchRepository interface {
	CreateBatch(ctx context.Context, input BatchInput, expiresAt string) (int, error)
	UpdateBatch(ctx context.Context, base BatchBase) error
	GetBatches(ctx context.Context, params common.PaginationParams) ([]Batch, error)
	SearchBatchesBySku(ctx context.Context, sku string, params common.PaginationParams) ([]Batch, error)
	GetBulkBatchUpdateInfo(ctx context.Context, inputs []BatchInput) (BulkBatchUpdateInfo, error)
	GetBulkBatchUpdateInfoWithRecipe(ctx context.Context, inputs []BatchInput) (BulkBatchUpdateInfo, error)
	GetBatchById(ctx context.Context, id int) (Batch, error)
}

const baseBatchListingSql = `
select b.id, b.sku, b.quantity, b.expires_at, utx.unit_id, utx.name, utx.symbol,
pvartx.name, pvar.id, pvar.price, pvar.product_id, ptx.name, p.is_ingredient from batches b
join unit_translations utx on utx.unit_id = b.unit_id
join product_variants pvar on pvar.sku = b.sku
join product_translations ptx on ptx.product_id = pvar.product_id
join products p on p.id = pvar.product_id
join product_variant_translations pvartx on pvartx.product_variant_id = pvar.id and utx.language_code = pvartx.language_code
`

type BatchRepository struct {
	*pgxpool.Pool
}

func NewBatchRepository(pool *pgxpool.Pool) IBatchRepository {
	return &BatchRepository{pool}
}

func (r *BatchRepository) CreateBatch(ctx context.Context, input BatchInput, expiresAt string) (int, error) {
	sql := `INSERT INTO batches (sku, warehouse_id, quantity, unit_id, expires_at) VALUES ($1, $2, $3, $4, $5) RETURNING id`
	op := common.GetOperator(ctx, r.Pool)
	warehouseId := warehouse.GetWarehouseId(ctx)
	row := op.QueryRow(ctx, sql, input.Sku, warehouseId, input.Quantity, input.UnitId, expiresAt)
	var id int
	err := row.Scan(&id)
	if err != nil {
		common.LoggerFromCtx(ctx).Error("Failed to create batch", zap.Error(err))
		return 0, common.NewBadRequestFromMessage("Failed to create batch")
	}
	return id, nil
}

func (r *BatchRepository) UpdateBatch(ctx context.Context, base BatchBase) error {
	updatedAt := time.Now().UTC()
	op := common.GetOperator(ctx, r.Pool)
	sql := `UPDATE batches SET quantity = $1, updated_at = $2 WHERE id = $3 and warehouse_id = $4`
	_, err := op.Exec(ctx, sql, base.Quantity, updatedAt, base.Id, base.WarehouseId)
	if err != nil {
		common.LoggerFromCtx(ctx).Error("Failed to update batch", zap.Error(err))
		return common.NewBadRequestFromMessage("Failed to update batch")
	}
	return nil
}

func (r *BatchRepository) GetBatches(ctx context.Context, params common.PaginationParams) ([]Batch, error) {
	warehouseId := warehouse.GetWarehouseId(ctx)
	op := common.GetOperator(ctx, r.Pool)
	lang := common.GetLanguageParam(ctx)
	sqlBuilder := common.NewPaginationQueryBuilder(
		baseBatchListingSql,
		[]string{"b.expires_at ASC", "b.id ASC"},
	)
	rows, err := sqlBuilder.
		WithOperator(op).
		WithConditions([]string{
			"utx.language_code = $1",
			"AND",
			"b.warehouse_id = $2",
		}).
		WithParams(params).
		WithCursorKeys([]string{"b.expires_at", "b.id"}).
		WithCompareSymbols(">", ">=", "<").
		Build().
		Query(ctx, lang, warehouseId)
	if err != nil {
		common.LoggerFromCtx(ctx).Error("Failed to get batches", zap.Error(err))
		return []Batch{}, common.NewBadRequestFromMessage("Failed to get batches")
	}
	defer rows.Close()
	return r.parseBatchRows(rows)
}

func (r *BatchRepository) SearchBatchesBySku(ctx context.Context, sku string, params common.PaginationParams) ([]Batch, error) {
	op := common.GetOperator(ctx, r.Pool)
	warehouseId := warehouse.GetWarehouseId(ctx)
	lang := common.GetLanguageParam(ctx)
	sqlBuilder := common.NewPaginationQueryBuilder(
		baseBatchListingSql,
		[]string{"b.expires_at ASC", "b.id ASC"},
	)
	rows, err := sqlBuilder.
		WithOperator(op).
		WithConditions([]string{
			"utx.language_code = $1",
			"AND",
			"b.warehouse_id = $2",
			"AND",
			"b.sku = $3",
		}).
		WithParams(params).
		WithCursorKeys([]string{"b.expires_at", "b.id"}).
		WithCompareSymbols(">", ">=", "<").
		Build().
		Query(ctx, lang, warehouseId, sku)
	if err != nil {
		common.LoggerFromCtx(ctx).Error("Failed to search batches", zap.Error(err))
		return []Batch{}, common.NewBadRequestFromMessage("Failed to search batches")
	}
	defer rows.Close()
	return r.parseBatchRows(rows)
}

func (r *BatchRepository) parseBatchRows(rows pgx.Rows) ([]Batch, error) {
	var batches []Batch
	for rows.Next() {
		var batch Batch
		var productVariantBase ProductVariantBase
		var unit unit.Unit
		err := rows.Scan(
			&batch.Id, &batch.Sku, &batch.Quantity, &batch.ExpiresAt,
			&unit.Id, &unit.Name, &unit.Symbol,
			&productVariantBase.Name, &productVariantBase.Id, &productVariantBase.Price,
			&productVariantBase.ProductId, &batch.ProductName, &batch.IsIngredient,
		)
		if err != nil {
			common.GetLogger().Error("Failed to scan batches", zap.Error(err))
			return []Batch{}, common.NewBadRequestFromMessage("Failed to scan batches")
		}
		batch.Unit = unit
		batch.ProductVariantBase = &productVariantBase
		batches = append(batches, batch)
	}
	return batches, nil
}

func (r *BatchRepository) processBulkBatchUnitOfWork(
	ctx context.Context,
	bulkBatchUpdateUnitOfWork BulkBatchUpdateUnitOfWork,
	transactionsBatch *pgx.Batch,
) error {
	// create update sql batches
	// create create sql batches
	op := common.GetOperator(ctx, r.Pool)
	warehouseId := warehouse.GetWarehouseId(ctx)
	for _, batchUpdateRequest := range bulkBatchUpdateUnitOfWork.BatchUpdateRequestLookup {
		transactionsBatch.Queue(
			"UPDATE batches SET quantity = $1 WHERE id = $2 and warehouse_id = $3",
			batchUpdateRequest.NewValue,
			batchUpdateRequest.BatchId,
			warehouseId,
		)
	}
	for _, batchCreateRequest := range bulkBatchUpdateUnitOfWork.BatchCreateRequestLookup {
		transactionsBatch.Queue(
			"INSERT INTO batches (sku, warehouse_id, quantity, unit_id, expires_at) VALUES ($1, $2, $3, $4, $5)",
			batchCreateRequest.BatchSku,
			warehouseId,
			batchCreateRequest.Quantity,
			batchCreateRequest.UnitId,
			common.GetUtcDateOnlyStringFromTime(batchCreateRequest.ExpiryDate),
		)
	}
	results := op.SendBatch(ctx, transactionsBatch)
	defer results.Close()
	for i := 0; i < transactionsBatch.Len(); i++ {
		_, err := results.Exec()
		if err != nil {
			common.LoggerFromCtx(ctx).Error("Failed to process bulk batch unit of work", zap.Error(err))
			return common.NewBadRequestFromMessage("Failed to process bulk batch unit of work")
		}
	}
	return nil
}

func (r *BatchRepository) GetBatchById(ctx context.Context, batchId int) (Batch, error) {
	sql := baseBatchListingSql + " WHERE b.id = $1 AND b.warehouse_id = $2 AND pvartx.language_code = $3"
	op := common.GetOperator(ctx, r.Pool)
	lang := common.GetLanguageParam(ctx)
	warehouseId := warehouse.GetWarehouseId(ctx)
	row := op.QueryRow(ctx, sql, batchId, warehouseId, lang)
	var batch Batch
	var productVariantBase ProductVariantBase
	var unit unit.Unit
	err := row.Scan(
		&batch.Id, &batch.Sku, &batch.Quantity, &batch.ExpiresAt,
		&unit.Id, &unit.Name, &unit.Symbol,
		&productVariantBase.Name, &productVariantBase.Id, &productVariantBase.Price,
		&productVariantBase.ProductId, &batch.ProductName, &batch.IsIngredient,
	)
	if err != nil {
		common.LoggerFromCtx(ctx).Error("Failed to get batch by id", zap.Error(err))
		return Batch{}, common.NewBadRequestFromMessage("Failed to get batch by id")
	}
	batch.Unit = unit
	batch.ProductVariantBase = &productVariantBase
	return batch, nil
}
