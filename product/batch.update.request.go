package product

import (
	"context"

	"github.com/nayefradwi/zanobia_inventory_manager/common"
)

func (s *BatchService) createBulkUpdateRequest(
	ctx context.Context,
	input []BatchInput,
) (BulkBatchUpdateRequest, error) {
	shouldDecrementRecipe := common.GetBoolFromContext(ctx, DecrementRecipeKey{})
	if shouldDecrementRecipe {
		return s.createBulkUpdateRequestWithRecipe(ctx, input)
	}
	return s.createBulkUpdateRequestWithoutRecipe(ctx, input)
}

/*
complexity analysis:
- O(10n) time = O(n) time
- O(6n) space = O(n) space
- n = number of batch inputs
- O(4) database calls if recipe is used
*/
func (s *BatchService) createBulkUpdateRequestWithRecipe(
	ctx context.Context,
	input []BatchInput,
) (BulkBatchUpdateRequest, error) {
	// this will give you all the batches in their correct unit conversions
	batchUpdateRequest, err := s.createBulkUpdateRequestWithoutRecipe(ctx, input)
	if err != nil {
		return BulkBatchUpdateRequest{}, err
	}
	recipeLookUp, recipeVariantSkus, err := s.recipeService.GetRecipesLookUpMapFromSkus(ctx, batchUpdateRequest.SkuList)
	if err != nil {
		return BulkBatchUpdateRequest{}, err
	}
	if len(recipeVariantSkus) == 0 {
		// no recipes found
		return batchUpdateRequest, nil
	}
	recipeBatches, err := s.getRecipeBatchBases(ctx, recipeVariantSkus)
	if err != nil {
		return BulkBatchUpdateRequest{}, err
	}
	recipeBatchMap, recipeBatchBaseIds, err := s.createRecipeBatchInputs(ctx, recipeBatches, recipeLookUp, batchUpdateRequest.BatchInputLookUp)
	if err != nil {
		return BulkBatchUpdateRequest{}, err
	}
	batchUpdateRequest.RecipeBatchInputMap = recipeBatchMap
	batchUpdateRequest.RecipeMap = recipeLookUp
	batchUpdateRequest.SkuList = append(batchUpdateRequest.SkuList, recipeVariantSkus...)
	batchUpdateRequest.BatchIds = append(batchUpdateRequest.BatchIds, recipeBatchBaseIds...)
	batchUpdateRequest.BatchInputLookUp = common.MergeMaps[string, BatchInput](batchUpdateRequest.BatchInputLookUp, recipeBatchMap)
	return batchUpdateRequest, nil
}

/*
complexity analysis:
- O(6n) time = O(n) time
- O(2n) space = O(n) space
- n = number of batch inputs
- O(2) database calls
*/
func (s *BatchService) createBulkUpdateRequestWithoutRecipe(
	ctx context.Context,
	input []BatchInput,
) (BulkBatchUpdateRequest, error) {
	// TODO: check if cost is correct of batch
	batchInputMapToUpdate, batchInputMapToCreate, skuList := s.createBatchInputMapSkuListAndIdList(input)
	originalUnitsMap, err := s.productService.GetOriginalUnitsBySkuList(ctx, skuList)
	if err != nil {
		return BulkBatchUpdateRequest{}, err
	}
	batchInputMapToUpdate, err = s.convertUnitOfBatchInputMap(ctx, batchInputMapToUpdate, originalUnitsMap)
	if err != nil {
		return BulkBatchUpdateRequest{}, err
	}
	batchInputMapToCreate, err = s.convertUnitOfBatchInputMap(ctx, batchInputMapToCreate, originalUnitsMap)
	if err != nil {
		return BulkBatchUpdateRequest{}, err
	}
	batchIds := make([]int, len(batchInputMapToUpdate))
	i := 0
	for _, batchInput := range batchInputMapToUpdate {
		batchIds[i] = *batchInput.Id
		i++
	}
	batchBases, err := s.batchRepo.getBatchBasesFromIds(ctx, batchIds)
	if err != nil {
		return BulkBatchUpdateRequest{}, err
	}
	batchBaseMap := make(map[string]BatchBase, len(batchBases))
	for _, batchBase := range batchBases {
		if batchBase.Sku == "" {
			return BulkBatchUpdateRequest{}, common.NewBadRequestFromMessage("invalid batch input")
		}
		batchBaseMap[batchBase.Sku] = batchBase
	}
	batchInputLookup := common.MergeMaps[string, BatchInput](batchInputMapToUpdate, batchInputMapToCreate)
	return BulkBatchUpdateRequest{
		BatchInputMapToUpdate: batchInputMapToUpdate,
		BatchInputMapToCreate: batchInputMapToCreate,
		SkuList:               skuList,
		BatchBasesLookup:      batchBaseMap,
		BatchIds:              batchIds,
		// this will have all the converted unit batch inputs
		BatchInputLookUp: batchInputLookup,
	}, nil
}

func (s *BatchService) createBatchInputMapSkuListAndIdList(input []BatchInput) (map[string]BatchInput, map[string]BatchInput, []string) {
	batchInputMapToUpdate := make(map[string]BatchInput)
	batchInputMapToCreate := make(map[string]BatchInput)
	skuList := make([]string, 0)
	for _, batchInput := range input {
		if batchInput.Id != nil {
			batchInputMapToUpdate[batchInput.Sku] = batchInput
		} else {
			batchInputMapToCreate[batchInput.Sku] = batchInput
		}
		skuList = append(skuList, batchInput.Sku)
	}
	return batchInputMapToUpdate, batchInputMapToCreate, skuList
}

func (s *BatchService) convertUnitOfBatchInputMap(
	ctx context.Context,
	batchInputMap map[string]BatchInput,
	originalUnitsMap map[string]int,
) (map[string]BatchInput, error) {
	for sku, batchInput := range batchInputMap {
		originalUnitId := originalUnitsMap[sku]
		batchUnitId := batchInput.UnitId
		convertUnitInput := ConvertUnitInput{
			ToUnitId:   &originalUnitId,
			FromUnitId: &batchUnitId,
			Quantity:   batchInput.Quantity,
		}
		convertedOutput, err := s.unitService.ConvertUnit(ctx, convertUnitInput)
		if err != nil {
			return nil, err
		}
		batchInput.Quantity = convertedOutput.Quantity
		batchInput.UnitId = originalUnitId
		batchInputMap[sku] = batchInput
	}
	return batchInputMap, nil
}
