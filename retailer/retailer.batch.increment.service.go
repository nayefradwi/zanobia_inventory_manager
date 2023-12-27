package retailer

import (
	"context"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
	"github.com/nayefradwi/zanobia_inventory_manager/transactions"
)

func (s *RetailerBatchService) IncrementBatch(ctx context.Context, batchInput RetailerBatchInput) error {
	return s.BulkIncrementBatch(ctx, []RetailerBatchInput{batchInput})
}

func (s *RetailerBatchService) BulkIncrementBatch(ctx context.Context, inputs []RetailerBatchInput) error {
	if err := ValidateBatchInputsIncrement(inputs); err != nil {
		return err
	}
	bulkBatchUpdateInfo, err := s.repo.GetBulkBatchUpdateInfo(ctx, inputs)
	if err != nil {
		return common.NewBadRequestFromMessage("failed to process batch increment")
	}
	return common.RunWithTransaction(ctx, s.repo.(*RetailerBatchRepository).Pool, func(ctx context.Context, tx pgx.Tx) error {
		return s.processBulkBatchIncrement(ctx, bulkBatchUpdateInfo)
	})
}

func (s *RetailerBatchService) processBulkBatchIncrement(
	ctx context.Context,
	bulkBatchUpdateInfo BulkRetailerBatchUpdateInfo,
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
	transactionHistory := append(transactionHistory1, transactionHistory2...)
	bulkBatchUpdateUnitOfWork := BulkRetailerBatchUpdateUnitOfWork{
		BatchUpdateRequestLookup: batchUpdateRequestLookup,
		BatchCreateRequestLookup: batchCreateRequestLookup,
		BatchTransactionHistory:  transactionHistory,
	}
	return s.processBulkBatchUnitOfWork(ctx, bulkBatchUpdateUnitOfWork)
}

func (s *RetailerBatchService) createIncrementBatchesUpdateRequest(
	ctx context.Context,
	bulkUpdateBatchInfo BulkRetailerBatchUpdateInfo,
) (
	map[string]RetailerBatchUpdateRequest,
	[]transactions.CreateWarehouseTransactionCommand,
	error,
) {
	batchUpdateRequestLookup := make(map[string]RetailerBatchUpdateRequest)
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
		updateValue := batchBase.Quantity + convertedBatchInput.Quantity
		batchUpdateRequestLookup[batchInput.Sku] = RetailerBatchUpdateRequest{
			BatchId:    batchInput.Id,
			NewValue:   updateValue,
			Reason:     batchInput.Reason,
			Sku:        batchInput.Sku,
			ModifiedBy: convertedBatchInput.Quantity,
		}
		// TODO: make the transaction command be retailer transaction command
		transactionCommand := transactions.CreateWarehouseTransactionCommand{
			BatchId:  *batchBase.Id,
			Quantity: convertedBatchInput.Quantity,
			UnitId:   batchVariantMetaInfo.UnitId,
			Reason:   batchInput.Reason,
			Comment:  batchInput.Comment,
			Cost:     totalCost,
			Sku:      batchInput.Sku,
		}
		transactionHistory = append(transactionHistory, transactionCommand)
	}
	return batchUpdateRequestLookup, transactionHistory, nil
}

func (s *RetailerBatchService) createBatchCreateRequest(
	ctx context.Context,
	bulkUpdateBatchInfo BulkRetailerBatchUpdateInfo,
) (
	map[string]RetailerBatchCreateRequest,
	[]transactions.CreateWarehouseTransactionCommand,
	error,
) {
	batchCreateRequestLookup := make(map[string]RetailerBatchCreateRequest, 0)
	transactionHistory := make([]transactions.CreateWarehouseTransactionCommand, 0)
	for _, batchInput := range bulkUpdateBatchInfo.BatchInputMapToCreate {
		batchVariantMetaInfo, ok := bulkUpdateBatchInfo.BatchVariantMetaInfoLookup[batchInput.Sku]
		if !ok {
			return nil, nil, common.NewBadRequestFromMessage("variant meta info not found")
		}
		convertedBatchInput, err := s.convertBatchInput(ctx, batchInput, batchVariantMetaInfo)
		if err != nil {
			return nil, nil, err
		}
		totalCost := batchVariantMetaInfo.Cost * convertedBatchInput.Quantity
		expiryDate := time.Now().AddDate(0, 0, batchVariantMetaInfo.ExpiresInDays)
		batchCreateRequestLookup[batchInput.Sku] = RetailerBatchCreateRequest{
			BatchSku:   batchInput.Sku,
			Quantity:   convertedBatchInput.Quantity,
			UnitId:     batchVariantMetaInfo.UnitId,
			ExpiryDate: expiryDate,
		}
		// TODO: make the transaction command be retailer transaction command
		transactionCommand := transactions.CreateWarehouseTransactionCommand{
			Quantity: convertedBatchInput.Quantity,
			UnitId:   batchVariantMetaInfo.UnitId,
			Reason:   batchInput.Reason,
			Comment:  batchInput.Comment,
			Cost:     totalCost,
			Sku:      batchInput.Sku,
		}
		transactionHistory = append(transactionHistory, transactionCommand)
	}
	return batchCreateRequestLookup, transactionHistory, nil
}
