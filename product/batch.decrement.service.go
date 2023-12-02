package product

import (
	"context"
	"log"

	"github.com/jackc/pgx/v4"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
)

func (s *BatchService) DecrementBatch(ctx context.Context, batchInput BatchInput) error {
	return s.BulkDecrementBatch(ctx, []BatchInput{batchInput})
}

func (s *BatchService) BulkDecrementBatch(ctx context.Context, inputs []BatchInput) error {
	if err := ValidateBatchInputsDecrement(inputs); err != nil {
		return err
	}
	return common.RunWithTransaction(ctx, s.batchRepo.(*BatchRepository).Pool, func(ctx context.Context, tx pgx.Tx) error {
		return s.bulkDecrementBatchesTransaction(ctx, inputs)
	})

}

func (s *BatchService) bulkDecrementBatchesTransaction(ctx context.Context, inputs []BatchInput) error {
	batchUpdateRequest, err := s.lockAndCreateUpdateRequestForDecrementBatches(ctx, inputs)
	defer s.unlockBatchesBySkus(ctx, batchUpdateRequest.SkuList)
	defer s.unlockBatchesByIds(ctx, batchUpdateRequest.BatchIds)
	if err != nil {
		return err
	}
	if len(batchUpdateRequest.BatchInputMapToUpdate) <= 0 {
		return common.NewBadRequestFromMessage("no batches to decrement")
	}
	decrementedQuantitis, err := s.bulkDecrementBatches(
		batchUpdateRequest.BatchBasesLookup,
		batchUpdateRequest.RecipeBatchInputMap,
	)
	if err != nil {
		return err
	}
	log.Printf("decrementedQuantitis: %v", decrementedQuantitis)
	// TODO: bulkUpdateBatches(ctx, decrementedQuantitis, batchUpdateRequest.BatchInputMapToCreate)
	// TODO: create transaction history
	return nil
}

func (s *BatchService) lockAndCreateUpdateRequestForDecrementBatches(ctx context.Context, inputs []BatchInput) (BulkBatchUpdateRequest, error) {
	batchUpdateRequest, err := s.createBulkUpdateRequestWithoutRecipe(ctx, inputs)
	if err != nil {
		return BulkBatchUpdateRequest{}, err
	}
	if err := s.lockBatchUpdateRequest(ctx, batchUpdateRequest); err != nil {
		return BulkBatchUpdateRequest{}, err
	}
	return batchUpdateRequest, nil
}

func (s *BatchService) bulkDecrementBatches(
	batchLookup map[string]BatchBase,
	batchInputMap map[string]BatchInput,
) (map[int]float64, error) {
	batchNewQuantities := make(map[int]float64, len(batchInputMap))
	for sku, batchInput := range batchInputMap {
		batchBase, found := batchLookup[sku]
		if !found {
			return nil, common.NewBadRequestFromMessage("invalid batch input")
		}
		newQuantity := batchBase.Quantity - batchInput.Quantity
		if newQuantity < 0 {
			return nil, common.NewBadRequestFromMessage("not enough quantity")
		}
		batchNewQuantities[*batchBase.Id] = newQuantity
	}
	return batchNewQuantities, nil
}
