package retailer

import (
	"context"
	"strconv"

	"github.com/jackc/pgx/v4"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
	"github.com/nayefradwi/zanobia_inventory_manager/product"
)

type IRetailerBatchService interface {
	IncrementBatch(ctx context.Context, batchInput RetailerBatchInput) error
	DecrementBatch(ctx context.Context, input RetailerBatchInput) error
	BulkIncrementBatch(ctx context.Context, inputs []RetailerBatchInput) error
	BulkDecrementBatch(ctx context.Context, inputs []RetailerBatchInput) error
	GetBatches(ctx context.Context, retailerId int) (common.PaginatedResponse[RetailerBatch], error)
	SearchBatchesBySku(ctx context.Context, retailerId int, sku string) (common.PaginatedResponse[RetailerBatch], error)
	DeleteBatchesOfRetailer(ctx context.Context, retailerId int) error
}

type RetailerBatchService struct {
	repo           IRetailerBatchRepository
	productService product.IProductService
	lockingService common.IDistributedLockingService
	unitService    product.IUnitService
}

func NewRetailerBatchService(
	repo IRetailerBatchRepository,
	productService product.IProductService,
	lockingService common.IDistributedLockingService,
	unitService product.IUnitService,
) *RetailerBatchService {
	return &RetailerBatchService{
		repo,
		productService,
		lockingService,
		unitService,
	}
}

func GenerateRetailerBatchLockKey(fieldValue string) string {
	return "retailer-batch:" + fieldValue + ":lock"
}

func (s *RetailerBatchService) getConvertedBatch(ctx context.Context, input *RetailerBatchInput) (RetailerBatchBase, error) {
	var batchBase RetailerBatchBase
	if input.Id != nil {
		batchBase, _ = s.repo.GetRetailerBatchBaseById(ctx, input.Id)
	}
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
	return batchBase, nil
}

func (s *RetailerBatchService) convertUnit(ctx context.Context, unitId int, input RetailerBatchInput) (float64, error) {
	if unitId == input.UnitId {
		return input.Quantity, nil
	}
	out, err := s.unitService.ConvertUnit(ctx, product.ConvertUnitInput{
		ToUnitId:   &unitId,
		FromUnitId: &input.UnitId,
		Quantity:   input.Quantity,
	})
	if err != nil {
		return 0, err
	}
	return out.Quantity, nil
}

func (s *RetailerBatchService) IncrementBatch(ctx context.Context, batchInput RetailerBatchInput) error {
	if err := ValidateBatchInputIncrement(batchInput); err != nil {
		return err
	}
	err := common.RunWithTransaction(ctx, s.repo.(*RetailerBatchRepository).Pool, func(ctx context.Context, tx pgx.Tx) error {
		ctx = common.SetOperator(ctx, tx)
		return s.tryToIncrementBatch(ctx, batchInput)
	})
	return err

}

func (s *RetailerBatchService) tryToIncrementBatch(ctx context.Context, input RetailerBatchInput) error {
	batchBase, err := s.getConvertedBatch(ctx, &input)
	if err != nil {
		return err
	}
	if batchBase.Id == nil {
		return s.tryToCreateBatch(ctx, input)
	}
	return s.incrementBatch(ctx, batchBase, input)
}

func (s *RetailerBatchService) tryToCreateBatch(ctx context.Context, input RetailerBatchInput) error {
	return s.lockingService.RunWithLock(ctx, GenerateRetailerBatchLockKey(input.Sku), func() error {
		expiresAt, err := s.productService.GetProductVariantExpirationDate(ctx, input.Sku)
		if err != nil {
			return err
		}
		return s.repo.CreateRetailerBatch(ctx, input, common.GetUtcDateOnlyStringFromTime(expiresAt))
	})
}

func (s *RetailerBatchService) incrementBatch(ctx context.Context, batch RetailerBatchBase, input RetailerBatchInput) error {
	return s.lockingService.RunWithLock(ctx, GenerateRetailerBatchLockKey(strconv.Itoa(*batch.Id)), func() error {
		batch = batch.SetQuantity(batch.Quantity + input.Quantity)
		err := s.repo.UpdateRetailerBatch(ctx, batch)
		// TODO check if should decrement warehouse
		return err
	})
}

func (s *RetailerBatchService) DecrementBatch(ctx context.Context, input RetailerBatchInput) error {
	if err := ValidateBatchInputDecrement(input); err != nil {
		return err
	}
	err := common.RunWithTransaction(ctx, s.repo.(*RetailerBatchRepository).Pool, func(ctx context.Context, tx pgx.Tx) error {
		ctx = common.SetOperator(ctx, tx)
		return s.tryToDecrementBatch(ctx, input)
	})
	return err
}

func (s *RetailerBatchService) tryToDecrementBatch(ctx context.Context, input RetailerBatchInput) error {
	batchBase, err := s.getConvertedBatch(ctx, &input)
	if err != nil {
		return err
	}
	if batchBase.Id == nil {
		return common.NewBadRequestFromMessage("retailer batch not found")
	}
	return s.decrementBatch(ctx, batchBase, input)
}

func (s *RetailerBatchService) decrementBatch(ctx context.Context, batchBase RetailerBatchBase, input RetailerBatchInput) error {
	return s.lockingService.RunWithLock(ctx, GenerateRetailerBatchLockKey(strconv.Itoa(*batchBase.Id)), func() error {
		newQty := batchBase.Quantity - input.Quantity
		if newQty < 0 {
			return common.NewBadRequestFromMessage("retailer batch not enough")
		}
		batchBase = batchBase.SetQuantity(newQty)
		return s.repo.UpdateRetailerBatch(ctx, batchBase)
	})
}

func (s *RetailerBatchService) BulkIncrementBatch(ctx context.Context, inputs []RetailerBatchInput) error {
	err := common.RunWithTransaction(ctx, s.repo.(*RetailerBatchRepository).Pool, func(ctx context.Context, tx pgx.Tx) error {
		ctx = common.SetOperator(ctx, tx)
		return s.bulkIncrementBatch(ctx, inputs)
	})
	return err
}
func (s *RetailerBatchService) bulkIncrementBatch(ctx context.Context, inputs []RetailerBatchInput) error {
	for _, input := range inputs {
		if err := ValidateBatchInputIncrement(input); err != nil {
			return err
		}
		if err := s.tryToIncrementBatch(ctx, input); err != nil {
			return err
		}
	}
	return nil
}

func (s *RetailerBatchService) BulkDecrementBatch(ctx context.Context, inputs []RetailerBatchInput) error {
	err := common.RunWithTransaction(ctx, s.repo.(*RetailerBatchRepository).Pool, func(ctx context.Context, tx pgx.Tx) error {
		ctx = common.SetOperator(ctx, tx)
		return s.bulkDecrementBatch(ctx, inputs)
	})
	return err
}

func (s *RetailerBatchService) bulkDecrementBatch(ctx context.Context, inputs []RetailerBatchInput) error {
	for _, input := range inputs {
		if err := ValidateBatchInputDecrement(input); err != nil {
			return err
		}
		if err := s.tryToDecrementBatch(ctx, input); err != nil {
			return err
		}
	}
	return nil
}

func (s *RetailerBatchService) GetBatches(ctx context.Context, retailerId int) (common.PaginatedResponse[RetailerBatch], error) {
	paginationParam := common.GetPaginationParams(ctx)
	batches, err := s.repo.GetRetailerBatches(ctx, retailerId, paginationParam)
	if err != nil {
		return common.PaginatedResponse[RetailerBatch]{}, err
	}
	return s.createBatchesPage(batches, paginationParam.PageSize), nil
}

func (s *RetailerBatchService) SearchBatchesBySku(ctx context.Context, retailerId int, sku string) (common.PaginatedResponse[RetailerBatch], error) {
	paginationParam := common.GetPaginationParams(ctx)
	batches, err := s.repo.SearchRetailerBatchesBySku(ctx, retailerId, sku, paginationParam)
	if err != nil {
		return common.PaginatedResponse[RetailerBatch]{}, err
	}
	return s.createBatchesPage(batches, paginationParam.PageSize), nil
}

func (s *RetailerBatchService) createBatchesPage(batches []RetailerBatch, pageSize int) common.PaginatedResponse[RetailerBatch] {
	if len(batches) == 0 {
		return common.CreateEmptyPaginatedResponse[RetailerBatch](pageSize)
	}
	first, last := batches[0], batches[len(batches)-1]
	res := common.CreatePaginatedResponse[RetailerBatch](
		pageSize,
		last,
		first,
		batches,
	)
	return res
}

func (s *RetailerBatchService) DeleteBatchesOfRetailer(ctx context.Context, retailerId int) error {
	return s.lockingService.RunWithLock(ctx, GenerateRetailerBatchLockKey(strconv.Itoa(retailerId)), func() error {
		return s.repo.DeleteBatchesOfRetailer(ctx, retailerId)
	})
}
