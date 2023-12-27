package retailer

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
	"github.com/nayefradwi/zanobia_inventory_manager/product"
	"github.com/nayefradwi/zanobia_inventory_manager/warehouse"
	"go.uber.org/zap"
)

func (r *RetailerBatchRepository) GetBulkBatchUpdateInfo(
	ctx context.Context,
	inputs []RetailerBatchInput,
) (BulkRetailerBatchUpdateInfo, error) {
	pgxBatch := &pgx.Batch{}
	ids, skus, batchToUpdateLookup, batchToCreateLookup := r.extractBatchInfoFromInputs(inputs)
	r.getBatchesBasedOnSkuListAndIds(ctx, pgxBatch, skus, ids)
	r.getProductMetaInfoFromSkuList(pgxBatch, skus)
	op := common.GetOperator(ctx, r.Pool)
	results := op.SendBatch(ctx, pgxBatch)
	defer results.Close()
	batchBasesLookup, err := r.parseBatchBasesLookupFromResults(results)
	if err != nil {
		return BulkRetailerBatchUpdateInfo{}, err
	}
	batchVariantMetaInfoLookup, err := r.parseBatchVariantMetaInfoLookupFromResults(results)
	if err != nil {
		return BulkRetailerBatchUpdateInfo{}, err
	}
	return BulkRetailerBatchUpdateInfo{
		BatchBasesLookup:           batchBasesLookup,
		BatchVariantMetaInfoLookup: batchVariantMetaInfoLookup,
		BatchInputMapToUpdate:      batchToUpdateLookup,
		BatchInputMapToCreate:      batchToCreateLookup,
		SkuList:                    skus,
		Ids:                        ids,
	}, nil
}

func (r *RetailerBatchRepository) extractBatchInfoFromInputs(inputs []RetailerBatchInput) (
	ids []int,
	skus []string,
	batchToUpdateLookup map[string]RetailerBatchInput,
	batchToCreateLookup map[string]RetailerBatchInput,
) {
	ids = make([]int, 0)
	skus = make([]string, 0)
	batchToUpdateLookup = make(map[string]RetailerBatchInput)
	batchToCreateLookup = make(map[string]RetailerBatchInput)
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

func (r *RetailerBatchRepository) getBatchesBasedOnSkuListAndIds(
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

func (r *RetailerBatchRepository) getProductMetaInfoFromSkuList(
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

func (r *RetailerBatchRepository) parseBatchBasesLookupFromResults(
	results pgx.BatchResults,
) (
	map[string]RetailerBatchBase,
	error,
) {
	batchBasesLookup := make(map[string]RetailerBatchBase)
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
			batch := RetailerBatchBase{
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

func (r *RetailerBatchRepository) parseBatchVariantMetaInfoLookupFromResults(
	results pgx.BatchResults,
) (
	map[string]product.BatchVariantMetaInfo,
	error,
) {
	batchVariantMetaInfoLookup := make(map[string]product.BatchVariantMetaInfo)
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
			batchVariantMetaInfo := product.BatchVariantMetaInfo{
				UnitId:        *metaUnitId,
				ExpiresInDays: *metaExpiresInDays,
				Cost:          *metaCost,
			}
			batchVariantMetaInfoLookup[*metaSku] = batchVariantMetaInfo
		}
	}
	return batchVariantMetaInfoLookup, nil
}
