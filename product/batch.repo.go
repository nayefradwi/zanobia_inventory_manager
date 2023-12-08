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
	CreateBatch(ctx context.Context, input BatchInput, expiresAt string) (int, error)
	UpdateBatch(ctx context.Context, base BatchBase) error
	GetBatches(ctx context.Context, params common.PaginationParams) ([]Batch, error)
	SearchBatchesBySku(ctx context.Context, sku string, params common.PaginationParams) ([]Batch, error)
	GetBulkBatchUpdateInfo(ctx context.Context, inputs []BatchInput) (BulkBatchUpdateInfo, error)
	GetBulkBatchUpdateInfoWithRecipe(ctx context.Context, inputs []BatchInput) (BulkBatchUpdateInfo, error)
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
		log.Printf("Failed to create batch: %s", err.Error())
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
		log.Printf("Failed to update batch: %s", err.Error())
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
		[]string{"b.expires_at DESC", "b.id DESC"},
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
		WithCompareSymbols("<", "<=", ">").
		Build().
		Query(ctx, lang, warehouseId)
	if err != nil {
		log.Printf("Failed to get batches: %s", err.Error())
		return []Batch{}, common.NewBadRequestFromMessage("Failed to get batches")
	}
	return r.parseBatchRows(rows)
}

func (r *BatchRepository) SearchBatchesBySku(ctx context.Context, sku string, params common.PaginationParams) ([]Batch, error) {
	op := common.GetOperator(ctx, r.Pool)
	warehouseId := warehouse.GetWarehouseId(ctx)
	lang := common.GetLanguageParam(ctx)
	sqlBuilder := common.NewPaginationQueryBuilder(
		baseBatchListingSql,
		[]string{"b.expires_at DESC", "b.id DESC"},
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
		WithCompareSymbols("<", "<=", ">").
		Build().
		Query(ctx, lang, warehouseId, sku)
	if err != nil {
		log.Printf("Failed to search batches: %s", err.Error())
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

func (r *BatchRepository) GetBulkBatchUpdateInfo(
	ctx context.Context,
	inputs []BatchInput,
) (BulkBatchUpdateInfo, error) {
	ids := make([]int, 0)
	skus := make([]string, 0)
	batchToUpdateLookup := make(map[string]BatchInput)
	batchToCreateLookup := make(map[string]BatchInput)
	for _, input := range inputs {
		if input.Id == nil {
			batchToCreateLookup[input.Sku] = input
		} else {
			ids = append(ids, *input.Id)
			batchToUpdateLookup[input.Sku] = input
		}
		skus = append(skus, input.Sku)
	}
	sql := `
	select 
		batches.id as batch_id,
		batches.warehouse_id as warehouse_id,
		batches.sku as batch_sku,
		batches.quantity as batch_qty,
		batches.unit_id as batch_unit_id,
		null as pvar_sku,
		null as pvar_unit,
		null as pvar_expires_in,
		null as pvar_price
	from 
		batches
	where
		batches.id = any($1)
	and
		batches.sku = any($2)
	and 
		batches.warehouse_id = $3
	union all
	select
		null as batch_id,
		null as warehouse_id,
		null as batch_sku,
		null as batch_qty,
		null as batch_unit_id,
		product_variants.sku as pvar_sku,
		product_variants.standard_unit_id as pvar_unit,
		product_variants.expires_in_days as pvar_expires_in,
		product_variants.price as pvar_price
	from
		product_variants
	where
		product_variants.sku = any($2)
	`
	op := common.GetOperator(ctx, r.Pool)
	warehouseId := warehouse.GetWarehouseId(ctx)
	rows, err := op.Query(ctx, sql, ids, skus, warehouseId)
	if err != nil {
		log.Printf("Failed to get bulk batch update info: %s", err.Error())
		return BulkBatchUpdateInfo{}, common.NewBadRequestFromMessage("Failed to get bulk batch update info")
	}
	defer rows.Close()
	batchBasesLookup := make(map[string]BatchBase)
	batchVariantMetaInfoLookup := make(map[string]BatchVariantMetaInfo)
	for rows.Next() {
		var batchId *int
		var warehouseId *int
		var batchSku *string
		var batchQty *float64
		var batchUnitId *int
		var metaSku *string
		var metaUnitId *int
		var metaExpiresInDays *int
		var metaCost *float64
		err := rows.Scan(
			&batchId, &warehouseId, &batchSku, &batchQty, &batchUnitId, &metaSku,
			&metaUnitId, &metaExpiresInDays, &metaCost,
		)
		if err != nil {
			log.Printf("Failed to scan bulk batch update info: %s", err.Error())
			return BulkBatchUpdateInfo{}, common.NewBadRequestFromMessage("Failed to scan bulk batch update info")
		}
		if batchId != nil &&
			batchSku != nil &&
			batchQty != nil &&
			batchUnitId != nil &&
			warehouseId != nil {
			batch := BatchBase{
				Id:          batchId,
				WarehouseId: warehouseId,
				Sku:         *batchSku,
				Quantity:    *batchQty,
				UnitId:      *batchUnitId,
			}
			batchBasesLookup[batch.Sku] = batch
		}
		if metaSku != nil &&
			metaUnitId != nil &&
			metaExpiresInDays != nil &&
			metaCost != nil {
			batchVariantMetaInfo := BatchVariantMetaInfo{
				UnitId:        *metaUnitId,
				ExpiresInDays: *metaExpiresInDays,
				Cost:          *metaCost,
			}
			batchVariantMetaInfoLookup[*metaSku] = batchVariantMetaInfo
		}
	}
	return BulkBatchUpdateInfo{
		BatchBasesLookup:           batchBasesLookup,
		BatchVariantMetaInfoLookup: batchVariantMetaInfoLookup,
		BatchInputMapToUpdate:      batchToUpdateLookup,
		BatchInputMapToCreate:      batchToCreateLookup,
		SkuList:                    skus,
		Ids:                        ids,
	}, nil
}

func (r *BatchRepository) GetBulkBatchUpdateInfoWithRecipe(
	ctx context.Context,
	inputs []BatchInput,
) (BulkBatchUpdateInfo, error) {
	// TODO: fill
	return BulkBatchUpdateInfo{}, nil
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
			log.Printf("Failed to process bulk batch unit of work: %s", err.Error())
			return common.NewBadRequestFromMessage("Failed to process bulk batch unit of work")
		}
	}
	return nil
}
