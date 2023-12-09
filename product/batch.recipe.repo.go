package product

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
	"github.com/nayefradwi/zanobia_inventory_manager/warehouse"
	"go.uber.org/zap"
)

func (r *BatchRepository) GetBulkBatchUpdateInfoWithRecipe(
	ctx context.Context,
	inputs []BatchInput,
) (BulkBatchUpdateInfo, error) {
	pgxBatch := &pgx.Batch{}
	ids, skus, batchToUpdateLookup, batchToCreateLookup := r.extractBatchInfoFromInputs(inputs)
	r.getIfInputsIncludesRecipes(pgxBatch, skus)
	r.getBatchesBasedOnSkuListAndIds(ctx, pgxBatch, skus, ids)
	r.getProductMetaInfoAndRecipesFromSkuList(pgxBatch, skus)
	useMostExpired := common.GetBoolFromContext(ctx, UseMostExpiredKey{})
	if useMostExpired {
		common.GetLogger().Info("Using most expired batch")
		r.getMostExpiredRecipeBatchBases(ctx, pgxBatch, skus)
	} else {
		common.GetLogger().Info("Using least expired batch")
		r.getLeastExpiredRecipeBatchBases(ctx, pgxBatch, skus)
	}
	op := common.GetOperator(ctx, r.Pool)
	results := op.SendBatch(ctx, pgxBatch)
	defer results.Close()
	if r.parseIfRecipesIncludedFromResults(results) {
		return BulkBatchUpdateInfo{}, common.NewBadRequestFromMessage("bulk update with recipes cannot include recipes in inputs")
	}
	batchBasesLookup, err := r.parseBatchBasesLookupFromResults(results)
	if err != nil {
		return BulkBatchUpdateInfo{}, err
	}
	batchVariantMetaInfoLookup, recipeLookup, err := r.parseVariantInfoAndRecipe(results)
	if err != nil {
		return BulkBatchUpdateInfo{}, err
	}
	recipeBatchBasesLookup, err := r.parseBatchBasesLookupFromResults(results)
	if err != nil {
		return BulkBatchUpdateInfo{}, err
	}
	batchBasesLookup = common.MergeMaps[string, BatchBase](
		recipeBatchBasesLookup,
		batchBasesLookup,
	)
	return BulkBatchUpdateInfo{
		BatchBasesLookup:           batchBasesLookup,
		BatchVariantMetaInfoLookup: batchVariantMetaInfoLookup,
		BatchInputMapToUpdate:      batchToUpdateLookup,
		BatchInputMapToCreate:      batchToCreateLookup,
		SkuList:                    skus,
		Ids:                        ids,
		RecipeMap:                  recipeLookup,
	}, nil
}

func (r *BatchRepository) getProductMetaInfoAndRecipesFromSkuList(
	pgxBatch *pgx.Batch,
	skus []string,
) {
	pgxBatch.Queue(
		`
	select 
		pvar.sku,
		pvar.standard_unit_id,
		pvar.expires_in_days,
		pvar.price,
		r.id,
		r.result_variant_sku,
		r.recipe_variant_sku,
		r.quantity,
		r.unit_id,
		pvar_recipe.standard_unit_id,
		pvar_recipe.price
	from
		product_variants pvar
	left join 
		recipes r on r.result_variant_sku = pvar.sku
	left join 
		product_variants pvar_recipe on pvar_recipe.sku = r.recipe_variant_sku
	where
		pvar.sku = any($1)
		`,
		skus,
	)
}

func (r *BatchRepository) parseVariantInfoAndRecipe(
	results pgx.BatchResults,
) (
	map[string]BatchVariantMetaInfo,
	map[string]Recipe,
	error,
) {
	batchVariantMetaInfoLookup := make(map[string]BatchVariantMetaInfo)
	recipeLookup := make(map[string]Recipe)
	rows, err := results.Query()
	if err != nil {
		return batchVariantMetaInfoLookup,
			recipeLookup,
			common.NewBadRequestFromMessage("Failed to get batch bases")
	}
	defer rows.Close()
	for rows.Next() {
		var metaSku *string
		var metaUnitId *int
		var metaExpiresInDays *int
		var metaCost *float64
		var recipeId *int
		var recipeResultVariantSku *string
		var recipeRecipeVariantSku *string
		var recipeQuantity *float64
		var recipeUnitId *int
		var recipeStandardUnitId *int
		var recipeStandardUnitCost *float64
		err := rows.Scan(
			&metaSku, &metaUnitId, &metaExpiresInDays, &metaCost,
			&recipeId, &recipeResultVariantSku, &recipeRecipeVariantSku, &recipeQuantity, &recipeUnitId,
			&recipeStandardUnitId, &recipeStandardUnitCost,
		)
		if err != nil {
			common.GetLogger().Error("Failed to scan batch bases", zap.Error(err))
			return batchVariantMetaInfoLookup,
				recipeLookup,
				common.NewBadRequestFromMessage("Failed to scan batch bases")
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
		if recipeId != nil &&
			recipeResultVariantSku != nil &&
			recipeRecipeVariantSku != nil &&
			recipeQuantity != nil &&
			recipeUnitId != nil &&
			recipeStandardUnitId != nil &&
			recipeStandardUnitCost != nil {
			recipe := Recipe{
				Id:                     recipeId,
				ResultVariantSku:       *recipeResultVariantSku,
				RecipeVariantSku:       *recipeRecipeVariantSku,
				Quantity:               *recipeQuantity,
				Unit:                   Unit{Id: recipeUnitId},
				IngredientCost:         *recipeStandardUnitCost,
				IngredientStandardUnit: &Unit{Id: recipeStandardUnitId},
			}
			recipeLookup[recipe.GetLookupKey()] = recipe
			// adding recipe variant meta info to batch variant meta info lookup
			// to be used for unit conversion when updating recipe variant
			batchVariantMetaInfoLookup[*recipeRecipeVariantSku] = BatchVariantMetaInfo{
				UnitId:        *recipeStandardUnitId,
				ExpiresInDays: *metaExpiresInDays,
				Cost:          *recipeStandardUnitCost,
			}
		}
	}
	return batchVariantMetaInfoLookup, recipeLookup, nil
}

func (r *BatchRepository) getMostExpiredRecipeBatchBases(
	ctx context.Context,
	pgxBatch *pgx.Batch,
	skus []string,
) {
	warehouseId := warehouse.GetWarehouseId(ctx)
	pgxBatch.Queue(
		`
	select DISTINCT ON (batches.sku)
		batches.id as batch_id,
		batches.warehouse_id as warehouse_id,
		batches.sku as batch_sku,
		batches.quantity as batch_qty,
		batches.unit_id as batch_unit_id
	from
		batches
	join recipes on
		recipes.recipe_variant_sku = batches.sku
	where 
			recipes.result_variant_sku = any($1)
		and
			batches.warehouse_id = $2
		and
			expires_at >= NOW()
	ORDER BY batches.sku, batches.expires_at ASC
		`,
		skus,
		warehouseId,
	)
}

func (r *BatchRepository) getLeastExpiredRecipeBatchBases(
	ctx context.Context,
	pgxBatch *pgx.Batch,
	skus []string,
) {
	warehouseId := warehouse.GetWarehouseId(ctx)
	pgxBatch.Queue(
		`
	select DISTINCT ON (batches.sku)
		batches.id as batch_id,
		batches.warehouse_id as warehouse_id,
		batches.sku as batch_sku,
		batches.quantity as batch_qty,
		batches.unit_id as batch_unit_id	
	from
		batches
	join recipes on
		recipes.recipe_variant_sku = batches.sku
	where 
			recipes.result_variant_sku = any($1)
		and
			batches.warehouse_id = $2
		and
			expires_at >= NOW()
	ORDER BY batches.sku, batches.expires_at DESC
		`,
		skus,
		warehouseId,
	)
}

func (r *BatchRepository) getIfInputsIncludesRecipes(
	pgxBatch *pgx.Batch,
	skus []string,
) {
	pgxBatch.Queue(
		`
	select 
		recipes.recipe_variant_sku
	from
		recipes
	where
		recipes.recipe_variant_sku = any($1)
		`,
		skus,
	)
}

func (r *BatchRepository) parseIfRecipesIncludedFromResults(
	results pgx.BatchResults,
) bool {
	rows, err := results.Query()
	if err != nil {
		common.GetLogger().Error("Failed to get recipes using skus:", zap.Error(err))
		return true
	}
	defer rows.Close()
	for rows.Next() {
		var recipeVariantSku *string
		err := rows.Scan(&recipeVariantSku)
		if err != nil {
			common.GetLogger().Error("Failed to scan recipes", zap.Error(err))
			return true
		}
		if recipeVariantSku != nil {
			return true
		}
	}
	return false
}
