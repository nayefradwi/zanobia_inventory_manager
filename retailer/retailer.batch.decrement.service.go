package retailer

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
	"github.com/nayefradwi/zanobia_inventory_manager/transactions"
	"go.uber.org/zap"
)

func (s *RetailerBatchService) DecrementBatch(ctx context.Context, input RetailerBatchInput) error {
	return s.BulkDecrementBatch(ctx, []RetailerBatchInput{input})
}

func (s *RetailerBatchService) BulkDecrementBatch(ctx context.Context, inputs []RetailerBatchInput) error {
	if err := ValidateBatchInputsDecrement(inputs); err != nil {
		return err
	}
	bulkBatchUpdateInfo, err := s.repo.GetBulkBatchUpdateInfo(ctx, inputs)
	if err != nil {
		common.LoggerFromCtx(ctx).Error("failed to process batch decrement", zap.Error(err))
		return common.NewBadRequestFromMessage("failed to process batch decrement")
	}
	return common.RunWithTransaction(ctx, s.repo.(*RetailerBatchRepository).Pool, func(ctx context.Context, tx pgx.Tx) error {
		return s.processBulkBatchDecrement(ctx, bulkBatchUpdateInfo)
	})
}

func (s *RetailerBatchService) processBulkBatchDecrement(
	ctx context.Context,
	bulkBatchUpdateInfo BulkRetailerBatchUpdateInfo,
) error {
	bulkBatchUpdateInfo, lockErr := s.lockBatchUpdateRequest(ctx, bulkBatchUpdateInfo)
	defer s.unlockBatchUpdateRequest(ctx, bulkBatchUpdateInfo)
	if lockErr != nil {
		return lockErr
	}
	batchUpdateRequestLookup, transactionHistory, err := s.createDecrementBatchesUpdateRequest(ctx, bulkBatchUpdateInfo)
	if err != nil {
		return err
	}
	bulkBatchUpdateUnitOfWork := BulkRetailerBatchUpdateUnitOfWork{
		BatchUpdateRequestLookup: batchUpdateRequestLookup,
		BatchTransactionHistory:  transactionHistory,
	}
	return s.processBulkBatchUnitOfWork(ctx, bulkBatchUpdateUnitOfWork)
}

func (s *RetailerBatchService) createDecrementBatchesUpdateRequest(
	ctx context.Context,
	bulkUpdateBatchInfo BulkRetailerBatchUpdateInfo,
) (
	map[string]RetailerBatchUpdateRequest,
	[]transactions.CreateRetailerTransactionCommand,
	error,
) {
	batchUpdateRequestLookup := make(map[string]RetailerBatchUpdateRequest)
	transactionHistory := make([]transactions.CreateRetailerTransactionCommand, 0)
	for _, batchInput := range bulkUpdateBatchInfo.BatchInputMapToUpdate {
		batchBase, ok := bulkUpdateBatchInfo.BatchBasesLookup[batchInput.Sku]
		if !ok {
			return nil, nil, common.NewBadRequestFromMessage("batch to update not found")
		}
		batchVariantMetaInfo, ok := bulkUpdateBatchInfo.BatchVariantMetaInfoLookup[batchInput.Sku]
		if !ok {
			return nil, nil, common.NewBadRequestFromMessage("variant meta info not found")
		}
		convertedBatchInput, err := s.convertBatchInput(ctx, batchInput, batchVariantMetaInfo)
		if err != nil {
			return nil, nil, err
		}
		totalCost := batchVariantMetaInfo.Cost * convertedBatchInput.Quantity
		updateValue := batchBase.Quantity - convertedBatchInput.Quantity
		if updateValue < 0 {
			return nil, nil, common.NewBadRequestFromMessage("insufficient quantity")
		}
		batchUpdateRequestLookup[batchInput.Sku] = RetailerBatchUpdateRequest{
			BatchId:    convertedBatchInput.Id,
			RetailerId: convertedBatchInput.RetailerId,
			NewValue:   updateValue,
			Reason:     convertedBatchInput.Reason,
			Sku:        convertedBatchInput.Sku,
			ModifiedBy: convertedBatchInput.Quantity,
		}
		transactionCommand := transactions.CreateRetailerTransactionCommand{
			RetailerBatchId: *batchBase.Id,
			RetailerId:      *convertedBatchInput.RetailerId,
			Quantity:        convertedBatchInput.Quantity,
			UnitId:          batchVariantMetaInfo.UnitId,
			Reason:          convertedBatchInput.Reason,
			Comment:         convertedBatchInput.Comment,
			Cost:            totalCost,
			Sku:             convertedBatchInput.Sku,
		}
		transactionHistory = append(transactionHistory, transactionCommand)
	}
	return batchUpdateRequestLookup, transactionHistory, nil
}
