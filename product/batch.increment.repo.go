package product

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
	"github.com/nayefradwi/zanobia_inventory_manager/warehouse"
	"go.uber.org/zap"
)

func (r *BatchRepository) GetBulkBatchUpdateInfo(
	ctx context.Context,
	inputs []BatchInput,
) (BulkBatchUpdateInfo, error) {
	pgxBatch := &pgx.Batch{}
	ids, skus, batchToUpdateLookup, batchToCreateLookup := r.extractBatchInfoFromInputs(inputs)
	r.getBatchesBasedOnSkuListAndIds(ctx, pgxBatch, skus, ids)
	r.getProductMetaInfoFromSkuList(pgxBatch, skus)
	op := common.GetOperator(ctx, r.Pool)
	results := op.SendBatch(ctx, pgxBatch)
	defer results.Close()
	batchBasesLookup, err := r.parseBatchBasesLookupFromResults(results)
	if err != nil {
		return BulkBatchUpdateInfo{}, err
	}
	batchVariantMetaInfoLookup, err := r.parseBatchVariantMetaInfoLookupFromResults(results)
	if err != nil {
		return BulkBatchUpdateInfo{}, err
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

func (r *BatchRepository) extractBatchInfoFromInputs(inputs []BatchInput) (
	ids []int,
	skus []string,
	batchToUpdateLookup map[string]BatchInput,
	batchToCreateLookup map[string]BatchInput,
) {
	ids = make([]int, 0)
	skus = make([]string, 0)
	batchToUpdateLookup = make(map[string]BatchInput)
	batchToCreateLookup = make(map[string]BatchInput)
	for _, input := range inputs {
		if input.Id == nil {
			batchToCreateLookup[input.Sku] = input
		} else {
			ids = append(ids, *input.Id)
			batchToUpdateLookup[input.Sku] = input
		}
		skus = append(skus, input.Sku)
	}
	return ids, skus, batchToUpdateLookup, batchToCreateLookup
}

func (r *BatchRepository) getBatchesBasedOnSkuListAndIds(
	ctx context.Context,
	pgxBatch *pgx.Batch,
	skus []string,
	ids []int,
) {
	warehouseId := warehouse.GetWarehouseId(ctx)
	pgxBatch.Queue(
		`
	select
		batches.id as batch_id,
		batches.warehouse_id as warehouse_id,
		batches.sku as batch_sku,
		batches.quantity as batch_qty,
		batches.unit_id as batch_unit_id
	from
		batches
	where
		batches.id = any($1)
	and
		batches.sku = any($2)
	and
		batches.warehouse_id = $3
		`,
		ids,
		skus,
		warehouseId,
	)
}

func (r *BatchRepository) getProductMetaInfoFromSkuList(
	pgxBatch *pgx.Batch,
	skus []string,
) {
	pgxBatch.Queue(
		`
	select
		product_variants.sku as pvar_sku,
		product_variants.standard_unit_id as pvar_unit,
		product_variants.expires_in_days as pvar_expires_in,
		product_variants.price as pvar_price
	from
		product_variants
	where
		product_variants.sku = any($1)
		`,
		skus,
	)
}

func (r *BatchRepository) parseBatchBasesLookupFromResults(
	results pgx.BatchResults,
) (
	map[string]BatchBase,
	error,
) {
	batchBasesLookup := make(map[string]BatchBase)
	rows, err := results.Query()
	if err != nil {
		common.GetLogger().Error("Failed to get batch bases", zap.Error(err))
		return batchBasesLookup, common.NewBadRequestFromMessage("Failed to get batch bases")
	}
	defer rows.Close()
	for rows.Next() {
		var batchId *int
		var warehouseId *int
		var batchSku *string
		var batchQty *float64
		var batchUnitId *int
		err := rows.Scan(
			&batchId, &warehouseId, &batchSku, &batchQty, &batchUnitId,
		)
		if err != nil {
			common.GetLogger().Error("Failed to scan batch bases", zap.Error(err))
			return batchBasesLookup, common.NewBadRequestFromMessage("Failed to scan batch bases")
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
	}
	return batchBasesLookup, nil
}

func (r *BatchRepository) parseBatchVariantMetaInfoLookupFromResults(
	results pgx.BatchResults,
) (
	map[string]BatchVariantMetaInfo,
	error,
) {
	batchVariantMetaInfoLookup := make(map[string]BatchVariantMetaInfo)
	rows, err := results.Query()
	if err != nil {
		common.GetLogger().Error("Failed to get batch bases", zap.Error(err))
		return batchVariantMetaInfoLookup, common.NewBadRequestFromMessage("Failed to get batch bases")
	}
	defer rows.Close()
	for rows.Next() {
		var metaSku *string
		var metaUnitId *int
		var metaExpiresInDays *int
		var metaCost *float64
		err := rows.Scan(
			&metaSku, &metaUnitId, &metaExpiresInDays, &metaCost,
		)
		if err != nil {
			common.GetLogger().Error("Failed to scan batch bases", zap.Error(err))
			return batchVariantMetaInfoLookup, common.NewBadRequestFromMessage("Failed to scan batch bases")
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
	return batchVariantMetaInfoLookup, nil
}
