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
	// TODO: add recipe batch update requests
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
	bulkUpdateBatchInfo BulkBatchUpdateInfo,
	batchUpdateRequest map[string]BatchUpdateRequest,
) (
	map[string]BatchUpdateRequest,
	[]transactions.CreateWarehouseTransactionCommand,
	error,
) {
	// TODO: create recipe batch update requests from batch update
	return nil, nil, nil
}

func (s *BatchService) createRecipeUpdateFromBatchCreate(
	ctx context.Context,
	bulkUpdateBatchInfo BulkBatchUpdateInfo,
	batchCreateRequest map[string]BatchCreateRequest,
) (
	map[string]BatchUpdateRequest,
	[]transactions.CreateWarehouseTransactionCommand,
	error,
) {
	// TODO: create recipe batch update requests from batch create
	return nil, nil, nil
}
