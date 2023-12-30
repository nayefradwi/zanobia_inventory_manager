package product

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
	"github.com/nayefradwi/zanobia_inventory_manager/transactions"
)

/*
incase i need to refactor this code, it does the following in the simplest terms
1. get all that it needs to update the batch:
  - the original units of the product variants in the batch and recipes that are used
   by those batches
  - the original batch bases of those product variants
  - the original expiry dates of those product variants
  - the original cost of those product variants (cost per original unit)

2.  convert all batch inputs to the original units of the product variants
3. create batches if it needs to
4. update batches if it needs to
5. accumalate all recipe decrements and update batch bases of the accumaleted recipe decrements
   after converting them and accumulating total cost as well
5  create transaction history
*/
/*
Compare this snippet from product/batch.recipe.service.go:
psuedo code:
 1. get batch info from inputs, this includes batch bases and variant meta info
    of batches to update and create and recipe batch bases
    - the variant meta info includes original unit id, cost, and expires in days of the batch variant
    as well as the recipe variant
    - the batch bases includes the quantity of the batch variant
    - the recipe batch bases includes the quantity of the recipe variant
 2. create batch update request for batches to update
    - this includes the new value of the batch variant
    - this includes the modified by value of the batch variant
 3. create batch create request for batches to create
    - this includes the quantity of the batch variant
    - this includes the unit id of the batch variant
 4. create recipe update request for recipe batches
    - this includes the new value of the recipe variant
    which accumulates the modified by value of the recipe variant
    based on how many times the recipe is used (if 2 batches use the same recipe)
    - this includes the modified by value of the recipe variant
    - this includes the unit id of the recipe variant
 5. create transaction history for all batches
 6. create bulk batch update unit of work
 7. process bulk batch update unit of work
*/
func (s *BatchService) IncrementBatchWithRecipe(ctx context.Context, batchInput BatchInput) error {
	return s.BulkIncrementWithRecipeBatch(ctx, []BatchInput{batchInput})
}

func (s *BatchService) BulkIncrementWithRecipeBatch(ctx context.Context, inputs []BatchInput) error {
	if err := ValidateBatchInputsIncrement(inputs); err != nil {
		return err
	}
	bulkBatchUpdateInfo, err := s.batchRepo.GetBulkBatchUpdateInfoWithRecipe(ctx, inputs)
	if err != nil {
		if apiErr, ok := err.(*common.ApiError); ok {
			return apiErr
		}
		return common.NewBadRequestFromMessage("failed to process batch increment")
	}

	return common.RunWithTransaction(ctx, s.batchRepo.(*BatchRepository).Pool, func(ctx context.Context, tx pgx.Tx) error {
		return s.processBulkBatchIncrementWithRecipe(ctx, bulkBatchUpdateInfo)
	})
}

func (s *BatchService) processBulkBatchIncrementWithRecipe(
	ctx context.Context,
	bulkBatchUpdateInfo BulkBatchUpdateInfo,
) error {
	bulkBatchUpdateInfo, lockErr := s.lockBatchUpdateRequest(ctx, bulkBatchUpdateInfo)
	defer s.unlockBatchUpdateRequest(ctx, bulkBatchUpdateInfo)
	if lockErr != nil {
		return lockErr
	}
	batchUpdateRequestLookup, transactionHistory1, err := s.createIncrementBatchesUpdateRequest(ctx, bulkBatchUpdateInfo)
	if err != nil {
		return err
	}
	batchCreateRequestLookup, transactionHistory2, err := s.createBatchCreateRequest(ctx, bulkBatchUpdateInfo)
	if err != nil {
		return err
	}
	recipeBatchUpdateRequestLookup, recipeTransactions, err := s.createRecipeUpdateRequests(
		ctx,
		bulkBatchUpdateInfo,
		batchUpdateRequestLookup,
		batchCreateRequestLookup,
	)
	if err != nil {
		return err
	}
	batchUpdateRequestLookup = common.MergeMaps[string, BatchUpdateRequest](
		recipeBatchUpdateRequestLookup,
		batchUpdateRequestLookup,
	)
	transactionHistory := append(transactionHistory1, transactionHistory2...)
	transactionHistory = append(transactionHistory, recipeTransactions...)
	bulkBatchUpdateUnitOfWork := BulkBatchUpdateUnitOfWork{
		BatchUpdateRequestLookup: batchUpdateRequestLookup,
		BatchCreateRequestLookup: batchCreateRequestLookup,
		BatchTransactionHistory:  transactionHistory,
	}
	return s.processBulkBatchUnitOfWork(ctx, bulkBatchUpdateUnitOfWork)
}

func (s *BatchService) createRecipeUpdateRequests(
	ctx context.Context,
	bulkUpdateBatchInfo BulkBatchUpdateInfo,
	batchUpdateRequestLookup map[string]BatchUpdateRequest,
	batchCreateRequestLookup map[string]BatchCreateRequest,
) (
	map[string]BatchUpdateRequest,
	[]transactions.CreateWarehouseTransactionCommand,
	error,
) {
	batchRecipeUpdateRequests1, transactionHistory1, err := s.createRecipeUpdateFromBatchUpdate(
		ctx,
		bulkUpdateBatchInfo,
		batchUpdateRequestLookup,
	)
	if err != nil {
		return nil, nil, err
	}
	batchRecipeUpdateRequests2, transactionHistory2, err := s.createRecipeUpdateFromBatchCreate(
		ctx,
		bulkUpdateBatchInfo,
		batchRecipeUpdateRequests1,
		batchCreateRequestLookup,
	)
	if err != nil {
		return nil, nil, err
	}
	merged := common.MergeMaps[string, BatchUpdateRequest](
		batchRecipeUpdateRequests1,
		batchRecipeUpdateRequests2,
	)
	mergedHistory := append(transactionHistory1, transactionHistory2...)
	return merged, mergedHistory, nil
}

func (s *BatchService) createRecipeUpdateFromBatchUpdate(
	ctx context.Context,
	// this will have previous decremented values of recipes
	bulkUpdateBatchInfo BulkBatchUpdateInfo,
	// this has converted units
	batchUpdateRequestLookup map[string]BatchUpdateRequest,
) (
	map[string]BatchUpdateRequest,
	[]transactions.CreateWarehouseTransactionCommand,
	error,
) {
	recipeTransactionHistory := make([]transactions.CreateWarehouseTransactionCommand, 0)
	if len(batchUpdateRequestLookup) == 0 {
		return batchUpdateRequestLookup, recipeTransactionHistory, nil
	}
	for _, recipe := range bulkUpdateBatchInfo.RecipeMap {
		recipeVariantMetaInfo, ok := bulkUpdateBatchInfo.BatchVariantMetaInfoLookup[recipe.RecipeVariantSku]
		if !ok {
			return nil, nil, common.NewBadRequestFromMessage("variant meta info not found")
		}
		recipeBatchBase, ok := bulkUpdateBatchInfo.BatchBasesLookup[recipe.RecipeVariantSku]
		if !ok {
			return nil, nil, common.NewBadRequestFromMessage("batch to update not found")
		}
		resultBatchUpdateRequest := batchUpdateRequestLookup[recipe.ResultVariantSku]
		recipeBatchInput := BatchInput{
			Id:       recipeBatchBase.Id,
			Sku:      recipe.RecipeVariantSku,
			Quantity: recipe.Quantity,
			UnitId:   *recipe.Unit.Id,
			Reason:   transactions.TransactionReasonTypeRecipeUse,
		}
		convertedRecipeInput, err := s.convertBatchInput(ctx, recipeBatchInput, recipeVariantMetaInfo)
		if err != nil {
			return nil, nil, err
		}
		recipeTotalModifyBy := convertedRecipeInput.Quantity * resultBatchUpdateRequest.ModifiedBy
		totalCost := recipeTotalModifyBy * recipeVariantMetaInfo.Cost
		var updatedValue float64
		if request, ok := batchUpdateRequestLookup[recipe.RecipeVariantSku]; ok {
			updatedValue = request.NewValue - recipeTotalModifyBy
		} else {
			updatedValue = recipeBatchBase.Quantity - recipeTotalModifyBy
		}
		if updatedValue < 0 {
			return nil, nil, common.NewBadRequestFromMessage("insufficient quantity")
		}
		batchUpdateRequestLookup[recipe.RecipeVariantSku] = BatchUpdateRequest{
			BatchId:    recipeBatchBase.Id,
			NewValue:   updatedValue,
			Reason:     transactions.TransactionReasonTypeRecipeUse,
			Sku:        recipe.RecipeVariantSku,
			ModifiedBy: recipeTotalModifyBy,
		}
		transactionCommand := transactions.CreateWarehouseTransactionCommand{
			BatchId:  *recipeBatchBase.Id,
			Quantity: recipeTotalModifyBy,
			UnitId:   recipeVariantMetaInfo.UnitId,
			Reason:   transactions.TransactionReasonTypeRecipeUse,
			Comment:  recipeBatchInput.Comment,
			Cost:     totalCost,
			Sku:      recipe.RecipeVariantSku,
		}
		recipeTransactionHistory = append(recipeTransactionHistory, transactionCommand)
	}
	return batchUpdateRequestLookup, recipeTransactionHistory, nil
}

func (s *BatchService) createRecipeUpdateFromBatchCreate(
	ctx context.Context,
	bulkUpdateBatchInfo BulkBatchUpdateInfo,
	// this will have previous decremented values of recipes
	batchUpdateRequestLookup map[string]BatchUpdateRequest,
	// this has converted units
	batchCreateRequest map[string]BatchCreateRequest,
) (
	map[string]BatchUpdateRequest,
	[]transactions.CreateWarehouseTransactionCommand,
	error,
) {
	recipeTransactionHistory := make([]transactions.CreateWarehouseTransactionCommand, 0)
	if len(batchCreateRequest) == 0 {
		return batchUpdateRequestLookup, recipeTransactionHistory, nil
	}
	for _, recipe := range bulkUpdateBatchInfo.RecipeMap {
		recipeVariantMetaInfo, ok := bulkUpdateBatchInfo.BatchVariantMetaInfoLookup[recipe.RecipeVariantSku]
		if !ok {
			return nil, nil, common.NewBadRequestFromMessage("variant meta info not found")
		}
		recipeBatchBase, ok := bulkUpdateBatchInfo.BatchBasesLookup[recipe.RecipeVariantSku]
		if !ok {
			return nil, nil, common.NewBadRequestFromMessage("batch to update not found")
		}
		resultBatchCreateRequest := batchCreateRequest[recipe.ResultVariantSku]
		recipeBatchInput := BatchInput{
			Id:       recipeBatchBase.Id,
			Sku:      recipe.RecipeVariantSku,
			Quantity: recipe.Quantity,
			UnitId:   recipeVariantMetaInfo.UnitId,
			Reason:   transactions.TransactionReasonTypeRecipeUse,
		}
		convertedRecipeInput, err := s.convertBatchInput(ctx, recipeBatchInput, recipeVariantMetaInfo)
		if err != nil {
			return nil, nil, err
		}
		recipeTotalModifyBy := convertedRecipeInput.Quantity * resultBatchCreateRequest.Quantity
		totalCost := recipeTotalModifyBy * recipeVariantMetaInfo.Cost
		var updatedValue float64
		if request, ok := batchUpdateRequestLookup[recipe.RecipeVariantSku]; ok {
			updatedValue = request.NewValue - recipeTotalModifyBy
		} else {
			updatedValue = recipeBatchBase.Quantity - recipeTotalModifyBy
		}
		if updatedValue < 0 {
			return nil, nil, common.NewBadRequestFromMessage("insufficient quantity")
		}
		batchUpdateRequestLookup[recipe.RecipeVariantSku] = BatchUpdateRequest{
			BatchId:    recipeBatchBase.Id,
			NewValue:   updatedValue,
			Reason:     transactions.TransactionReasonTypeRecipeUse,
			Sku:        recipe.RecipeVariantSku,
			ModifiedBy: recipeTotalModifyBy,
		}
		transactionCommand := transactions.CreateWarehouseTransactionCommand{
			BatchId:  *recipeBatchBase.Id,
			Quantity: recipeTotalModifyBy,
			UnitId:   recipeVariantMetaInfo.UnitId,
			Reason:   transactions.TransactionReasonTypeRecipeUse,
			Comment:  recipeBatchInput.Comment,
			Cost:     totalCost,
			Sku:      recipe.RecipeVariantSku,
		}
		recipeTransactionHistory = append(recipeTransactionHistory, transactionCommand)
	}
	return batchUpdateRequestLookup, recipeTransactionHistory, nil
}
