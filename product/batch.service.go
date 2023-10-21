package product

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
)

type IBatchService interface{}

type BatchService struct {
	batchRepo        IBatchRepository
	inventoryService IInventoryService
	productService   IProductService
	lockingService   common.IDistributedLockingService
	unitService      IUnitService
}

func NewBatchService(
	batchRepo IBatchRepository,
	inventoryService IInventoryService,
	productService IProductService,
	lockingService common.IDistributedLockingService,
	unitService IUnitService,
) *BatchService {
	return &BatchService{
		batchRepo,
		inventoryService,
		productService,
		lockingService,
		unitService,
	}
}
func (s *BatchService) IncrementBatch(ctx context.Context, batchInput BatchInput) error {
	return s.lockingService.RunWithLock(ctx, batchInput.Sku, func() error {
		if err := ValidateBatchInput(batchInput); err != nil {
			return err
		}
		err := common.RunWithTransaction(ctx, s.batchRepo.(*BatchRepository).Pool, func(ctx context.Context, tx pgx.Tx) error {
			ctx = common.SetOperator(ctx, tx)
			return s.tryToIncrementBatch(ctx, batchInput)
		})
		return err
	})
}

func (s *BatchService) tryToIncrementBatch(ctx context.Context, input BatchInput) error {
	batchBase, err := s.getConvertedBatch(ctx, &input)
	if err != nil {
		return err
	}
	if batchBase.Id == nil {
		return s.tryToCreateBatch(ctx, input)
	}
	return s.incrementBatch(ctx, batchBase, input)
}

func (s *BatchService) tryToCreateBatch(ctx context.Context, input BatchInput) error {
	expiresAt, err := s.productService.GetProductVariantExpirationDate(ctx, input.Sku)
	if err != nil {
		return err
	}
	return s.batchRepo.CreateBatch(ctx, input, common.GetUtcDateOnlyStringFromTime(expiresAt))
}

func (s *BatchService) getConvertedBatch(ctx context.Context, input *BatchInput) (BatchBase, error) {
	batchBase, err := s.batchRepo.GetBatchBase(ctx, input.Sku, input.ExpiresAt)
	unitId := batchBase.UnitId
	if batchBase.Id == nil {
		unitId, _ = s.productService.GetUnitIdOfProductVariantBySku(ctx, input.Sku)
	}
	if unitId != input.UnitId {
		convertedQty, err := s.convertUnit(ctx, unitId, *input)
		if err != nil {
			return batchBase, err
		}
		input.Quantity = convertedQty
		input.UnitId = unitId
	}
	return batchBase, err
}

func (s *BatchService) convertUnit(ctx context.Context, unitId int, input BatchInput) (float64, error) {
	if unitId == input.UnitId {
		return input.Quantity, nil
	}
	out, err := s.unitService.ConvertUnit(ctx, ConvertUnitInput{
		ToUnitId:   &unitId,
		FromUnitId: &input.UnitId,
		Quantity:   input.Quantity,
	})
	if err != nil {
		return 0, err
	}
	return out.Quantity, nil
}

func (s *BatchService) incrementBatch(ctx context.Context, batch BatchBase, input BatchInput) error {
	batch = batch.SetQuantity(batch.Quantity + input.Quantity)
	err := s.batchRepo.UpdateBatch(ctx, batch)
	// TODO check if should decrement inventory
	return err
}

// Decrement batch
// Bulk increment batches
// Bulk decrement batches
// Get batches paginated sorted by expiration date
