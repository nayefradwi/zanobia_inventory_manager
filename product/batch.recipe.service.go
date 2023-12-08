package product

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
	"github.com/nayefradwi/zanobia_inventory_manager/transactions"
)

func (s *BatchService) IncrementBatchWithRecipe(ctx context.Context, batchInput BatchInput) error {
	return s.BulkIncrementBatch(ctx, []BatchInput{batchInput})
}

func (s *BatchService) BulkIncrementWithRecipeBatch(ctx context.Context, inputs []BatchInput) error {
	if err := ValidateBatchInputsIncrement(inputs); err != nil {
		return err
	}
	bulkBatchUpdateInfo, err := s.batchRepo.GetBulkBatchUpdateInfoWithRecipe(ctx, inputs)
	if err != nil {
		common.NewBadRequestFromMessage("failed to process batch increment")
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
	for _, recipe := range bulkUpdateBatchInfo.RecipeMap {
		recipeVariantMetaInfo, ok := bulkUpdateBatchInfo.BatchVariantMetaInfoLookup[recipe.RecipeVariantName]
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
			UnitId:   recipeVariantMetaInfo.UnitId,
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
			Quantity: convertedRecipeInput.Quantity,
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
	for _, recipe := range bulkUpdateBatchInfo.RecipeMap {
		recipeVariantMetaInfo, ok := bulkUpdateBatchInfo.BatchVariantMetaInfoLookup[recipe.RecipeVariantName]
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
			Quantity: convertedRecipeInput.Quantity,
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
