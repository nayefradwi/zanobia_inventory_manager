package product

import (
	"context"
	"log"

	"github.com/jackc/pgx/v4"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
)

func (s *BatchService) IncrementBatch(ctx context.Context, batchInput BatchInput) error {
	return s.BulkIncrementBatch(ctx, []BatchInput{batchInput})
}

func (s *BatchService) BulkIncrementBatch(ctx context.Context, inputs []BatchInput) error {
	if err := ValidateBatchInputsIncrement(inputs); err != nil {
		return err
	}
	return common.RunWithTransaction(ctx, s.batchRepo.(*BatchRepository).Pool, func(ctx context.Context, tx pgx.Tx) error {
		return s.bulkIncrementBatchesTransaction(ctx, inputs)
	})
}

func (s *BatchService) bulkIncrementBatchesTransaction(ctx context.Context, inputs []BatchInput) error {
	batchUpdateRequest, err := s.lockAndCreateUpdateRequestForIncrementBatches(ctx, inputs)
	defer s.unlockBatchesBySkus(ctx, batchUpdateRequest.SkuList)
	defer s.unlockBatchesByIds(ctx, batchUpdateRequest.BatchIds)
	if err != nil {
		return err
	}
	batchNewQuantities := make(map[int]float64)
	if len(batchUpdateRequest.BatchInputMapToUpdate) > 0 {
		incrementedQuantities, err := s.bulkIncrementBatches(
			batchUpdateRequest.BatchBasesLookup,
			batchUpdateRequest.BatchInputMapToUpdate,
		)
		if err != nil {
			return err
		}
		batchNewQuantities = common.MergeMaps[int, float64](batchNewQuantities, incrementedQuantities)
	}
	if len(batchUpdateRequest.RecipeBatchInputMap) > 0 {
		decrementedQuantitis, err := s.bulkDecrementBatches(
			batchUpdateRequest.BatchBasesLookup,
			batchUpdateRequest.RecipeBatchInputMap,
		)
		if err != nil {
			return err
		}
		batchNewQuantities = common.MergeMaps[int, float64](batchNewQuantities, decrementedQuantitis)
	}
	log.Printf("batchNewQuantities: %v", batchNewQuantities)
	// TODO: bulkUpdateBatches(ctx, batchNewQuantities, batchUpdateRequest.BatchInputMapToCreate)
	// TODO: create transaction history
	return nil
}

func (s *BatchService) lockAndCreateUpdateRequestForIncrementBatches(ctx context.Context, inputs []BatchInput) (BulkBatchUpdateRequest, error) {
	batchUpdateRequest, err := s.createBulkUpdateRequest(ctx, inputs)
	if err != nil {
		return BulkBatchUpdateRequest{}, err
	}
	if err := s.lockBatchUpdateRequest(ctx, batchUpdateRequest); err != nil {
		return BulkBatchUpdateRequest{}, err
	}
	return batchUpdateRequest, nil
}

func (s *BatchService) bulkIncrementBatches(
	batchLookup map[string]BatchBase,
	batchInputMap map[string]BatchInput,
) (map[int]float64, error) {
	batchNewQuantities := make(map[int]float64, len(batchInputMap))
	for sku, batchInput := range batchInputMap {
		batchBase, found := batchLookup[sku]
		if !found {
			return nil, common.NewBadRequestFromMessage("invalid batch input")
		}
		newQuantity := batchBase.Quantity + batchInput.Quantity
		batchNewQuantities[*batchBase.Id] = newQuantity
	}
	return batchNewQuantities, nil
}
