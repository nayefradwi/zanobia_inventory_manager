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
}

func (s *BatchService) createRecipeUpdateFromBatchUpdate(
	ctx context.Context,
	batchUpdateRequest BatchUpdateRequest,
) (
	map[string]BatchUpdateRequest,
	[]transactions.CreateWarehouseTransactionCommand,
	error,
) {
	// TODO: create recipe batch update requests from batch update
}

func (s *BatchService) createRecipeUpdateFromBatchCreate(
	ctx context.Context,
	batchCreateRequest BatchCreateRequest,
) (
	map[string]BatchUpdateRequest,
	[]transactions.CreateWarehouseTransactionCommand,
	error,
) {
	// TODO: create recipe batch update requests from batch create
}
