package product

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
	"github.com/nayefradwi/zanobia_inventory_manager/transactions"
	"go.uber.org/zap"
)

func (s *BatchService) DecrementBatch(ctx context.Context, input BatchInput) error {
	return s.BulkDecrementBatch(ctx, []BatchInput{input})
}

func (s *BatchService) BulkDecrementBatch(ctx context.Context, inputs []BatchInput) error {
	if err := ValidateBatchInputsDecrement(inputs); err != nil {
		return err
	}
	bulkBatchUpdateInfo, err := s.batchRepo.GetBulkBatchUpdateInfo(ctx, inputs)
	if err != nil {
		common.LoggerFromCtx(ctx).Error("failed to process batch decrement", zap.Error(err))
		return common.NewBadRequestFromMessage("failed to process batch decrement")
	}
	return common.RunWithTransaction(ctx, s.batchRepo.(*BatchRepository).Pool, func(ctx context.Context, tx pgx.Tx) error {
		return s.processBulkBatchDecrement(ctx, bulkBatchUpdateInfo)
	})
}

func (s *BatchService) processBulkBatchDecrement(
	ctx context.Context,
	bulkBatchUpdateInfo BulkBatchUpdateInfo,
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
	bulkBatchUpdateUnitOfWork := BulkBatchUpdateUnitOfWork{
		BatchUpdateRequestLookup: batchUpdateRequestLookup,
		BatchTransactionHistory:  transactionHistory,
	}
	return s.processBulkBatchUnitOfWork(ctx, bulkBatchUpdateUnitOfWork)
}

func (s *BatchService) createDecrementBatchesUpdateRequest(
	ctx context.Context,
	bulkUpdateBatchInfo BulkBatchUpdateInfo,
) (
	map[string]BatchUpdateRequest,
	[]transactions.CreateWarehouseTransactionCommand,
	error,
) {
	batchUpdateRequestLookup := make(map[string]BatchUpdateRequest)
	transactionHistory := make([]transactions.CreateWarehouseTransactionCommand, 0)
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
		batchUpdateRequestLookup[convertedBatchInput.Sku] = BatchUpdateRequest{
			BatchId:    convertedBatchInput.Id,
			NewValue:   updateValue,
			Reason:     convertedBatchInput.Reason,
			Sku:        convertedBatchInput.Sku,
			ModifiedBy: convertedBatchInput.Quantity,
		}
		transactionCommand := transactions.CreateWarehouseTransactionCommand{
			BatchId:  *batchBase.Id,
			Quantity: convertedBatchInput.Quantity,
			UnitId:   batchVariantMetaInfo.UnitId,
			Reason:   convertedBatchInput.Reason,
			Comment:  convertedBatchInput.Comment,
			Cost:     totalCost,
			Sku:      convertedBatchInput.Sku,
		}
		transactionHistory = append(transactionHistory, transactionCommand)
	}
	return batchUpdateRequestLookup, transactionHistory, nil
}
