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
		updateValue := batchBase.Quantity + convertedBatchInput.Quantity
		batchUpdateRequestLookup[convertedBatchInput.Sku] = RetailerBatchUpdateRequest{
			BatchId:    convertedBatchInput.Id,
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

func (s *RetailerBatchService) createBatchCreateRequest(
	ctx context.Context,
	bulkUpdateBatchInfo BulkRetailerBatchUpdateInfo,
) (
	map[string]RetailerBatchCreateRequest,
	[]transactions.CreateRetailerTransactionCommand,
	error,
) {
	batchCreateRequestLookup := make(map[string]RetailerBatchCreateRequest, 0)
	transactionHistory := make([]transactions.CreateRetailerTransactionCommand, 0)
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
		batchCreateRequestLookup[convertedBatchInput.Sku] = RetailerBatchCreateRequest{
			BatchSku:   convertedBatchInput.Sku,
			Quantity:   convertedBatchInput.Quantity,
			UnitId:     batchVariantMetaInfo.UnitId,
			ExpiryDate: expiryDate,
		}
		transactionCommand := transactions.CreateRetailerTransactionCommand{
			RetailerId: *convertedBatchInput.RetailerId,
			Quantity:   convertedBatchInput.Quantity,
			UnitId:     batchVariantMetaInfo.UnitId,
			Reason:     convertedBatchInput.Reason,
			Comment:    convertedBatchInput.Comment,
			Cost:       totalCost,
			Sku:        convertedBatchInput.Sku,
		}
		transactionHistory = append(transactionHistory, transactionCommand)
	}
	return batchCreateRequestLookup, transactionHistory, nil
}
