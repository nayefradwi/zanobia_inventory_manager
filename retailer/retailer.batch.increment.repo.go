package retailer

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
	"github.com/nayefradwi/zanobia_inventory_manager/product"
	"go.uber.org/zap"
)

func (r *RetailerBatchRepository) GetBulkBatchUpdateInfo(
	ctx context.Context,
	inputs []RetailerBatchInput,
) (BulkRetailerBatchUpdateInfo, error) {
	pgxBatch := &pgx.Batch{}
	ids, retailerIds, skus, batchToUpdateLookup, batchToCreateLookup := r.extractBatchInfoFromInputs(inputs)
	r.getBatchesBasedOnSkuListAndIds(pgxBatch, retailerIds, skus, ids)
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
	retailerIds []int,
	skus []string,
	batchToUpdateLookup map[string]RetailerBatchInput,
	batchToCreateLookup map[string]RetailerBatchInput,
) {
	ids = make([]int, 0)
	retailerIds = make([]int, 0)
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
		retailerIds = append(retailerIds, *input.RetailerId)
	}
	return ids, retailerIds, skus, batchToUpdateLookup, batchToCreateLookup
}

func (r *RetailerBatchRepository) getBatchesBasedOnSkuListAndIds(
	pgxBatch *pgx.Batch,
	retailerIds []int,
	skus []string,
	ids []int,
) {

	pgxBatch.Queue(
		`
	select
		batches.id as batch_id,
		batches.retailer_id as retailer_id,
		batches.sku as batch_sku,
		batches.quantity as batch_qty,
		batches.unit_id as batch_unit_id
	from
		retailer_batches as batches
	where
		batches.id = any($1)
	and
		batches.sku = any($2)
	and
		batches.retailer_id = any($3)
		`,
		ids,
		skus,
		retailerIds,
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
		var RetailerId *int
		var batchSku *string
		var batchQty *float64
		var batchUnitId *int
		err := rows.Scan(
			&batchId, &RetailerId, &batchSku, &batchQty, &batchUnitId,
		)
		if err != nil {
			common.GetLogger().Error("Failed to scan batch bases", zap.Error(err))
			return batchBasesLookup, common.NewBadRequestFromMessage("Failed to scan batch bases")
		}
		if batchId != nil &&
			batchSku != nil &&
			batchQty != nil &&
			batchUnitId != nil &&
			RetailerId != nil {
			batch := RetailerBatchBase{
				Id:         batchId,
				RetailerId: RetailerId,
				Sku:        *batchSku,
				Quantity:   *batchQty,
				UnitId:     *batchUnitId,
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

func (r *RetailerBatchRepository) processBulkBatchUnitOfWork(
	ctx context.Context,
	bulkBatchUpdateUnitOfWork BulkRetailerBatchUpdateUnitOfWork,
	transactionsBatch *pgx.Batch,
) error {
	// create update sql batches
	// create create sql batches
	op := common.GetOperator(ctx, r.Pool)
	for _, batchUpdateRequest := range bulkBatchUpdateUnitOfWork.BatchUpdateRequestLookup {
		transactionsBatch.Queue(
			"UPDATE retailer_batches SET quantity = $1 WHERE id = $2 and retailer_id = $3",
			batchUpdateRequest.NewValue,
			batchUpdateRequest.BatchId,
			batchUpdateRequest.RetailerId,
		)
	}
	for _, batchCreateRequest := range bulkBatchUpdateUnitOfWork.BatchCreateRequestLookup {
		transactionsBatch.Queue(
			"INSERT INTO retailer_batches (sku, retailer_id, quantity, unit_id, expires_at) VALUES ($1, $2, $3, $4, $5)",
			batchCreateRequest.BatchSku,
			batchCreateRequest.RetailerId,
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
