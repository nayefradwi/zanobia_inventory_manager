package product

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
)

func (r *BatchRepository) GetBulkBatchUpdateInfoWithRecipe(
	ctx context.Context,
	inputs []BatchInput,
) (BulkBatchUpdateInfo, error) {
	pgxBatch := &pgx.Batch{}
	ids, skus, batchToUpdateLookup, batchToCreateLookup := r.extractBatchInfoFromInputs(inputs)
	r.getBatchesBasedOnSkuListAndIds(ctx, pgxBatch, skus, ids)
	r.getProductMetaInfoAndRecipesFromSkuList(pgxBatch, skus)
	op := common.GetOperator(ctx, r.Pool)
	results := op.SendBatch(ctx, pgxBatch)
	defer results.Close()
	batchBasesLookup, err := r.parseBatchBasesLookupFromResults(results)
	if err != nil {
		return BulkBatchUpdateInfo{}, err
	}
	batchVariantMetaInfoLookup, recipeLookup, err := r.parseVariantInfoAndRecipe(results)
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
		r.unit_id
	from
		product_variants pvar
	left join 
		recipes r on r.result_variant_sku = pvar.sku
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
		err := rows.Scan(
			&metaSku, &metaUnitId, &metaExpiresInDays, &metaCost,
			&recipeId, &recipeResultVariantSku, &recipeRecipeVariantSku, &recipeQuantity, &recipeUnitId,
		)
		if err != nil {
			return batchVariantMetaInfoLookup,
				recipeLookup,
				common.NewBadRequestFromMessage("Failed to scan batch bases")
		}
		var recipeCost float64
		var ingredientUnitId int
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
			recipeCost = *metaCost
			ingredientUnitId = *metaUnitId
		}
		if recipeId != nil &&
			recipeResultVariantSku != nil &&
			recipeRecipeVariantSku != nil &&
			recipeQuantity != nil &&
			recipeUnitId != nil {
			recipe := Recipe{
				Id:                     recipeId,
				ResultVariantSku:       *recipeResultVariantSku,
				RecipeVariantSku:       *recipeRecipeVariantSku,
				Quantity:               *recipeQuantity,
				Unit:                   Unit{Id: recipeUnitId},
				IngredientCost:         recipeCost,
				IngredientStandardUnit: &Unit{Id: &ingredientUnitId},
			}
			recipeLookup[recipe.GetLookupKey()] = recipe
			// adding recipe variant meta info to batch variant meta info lookup
			// to be used for unit conversion when updating recipe variant
			batchVariantMetaInfoLookup[*recipeRecipeVariantSku] = BatchVariantMetaInfo{
				UnitId:        *recipeUnitId,
				ExpiresInDays: *metaExpiresInDays,
				Cost:          recipeCost,
			}
		}
	}
	return batchVariantMetaInfoLookup, recipeLookup, nil
}
