package product

import (
	"context"
	"strconv"

	"github.com/jackc/pgx/v4"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
)

type DecrementRecipeKey struct{}
type IBatchService interface {
	IncrementBatch(ctx context.Context, batchInput BatchInput) error
	DecrementBatch(ctx context.Context, input BatchInput) error
	BulkIncrementBatch(ctx context.Context, inputs []BatchInput) error
	BulkDecrementBatch(ctx context.Context, inputs []BatchInput) error
	GetBatches(ctx context.Context) (common.PaginatedResponse[Batch], error)
	SearchBatchesBySku(ctx context.Context, sku string) (common.PaginatedResponse[Batch], error)
}

type BatchService struct {
	batchRepo      IBatchRepository
	productService IProductService
	lockingService common.IDistributedLockingService
	unitService    IUnitService
	recipeService  IRecipeService
}

func NewBatchService(
	batchRepo IBatchRepository,
	productService IProductService,
	lockingService common.IDistributedLockingService,
	unitService IUnitService,
	recipeService IRecipeService,
) *BatchService {
	return &BatchService{
		batchRepo,
		productService,
		lockingService,
		unitService,
		recipeService,
	}
}

func GenerateBatchLockKey(sku string) string {
	return "batch:" + sku + ":lock"
}

func (s *BatchService) getConvertedBatch(ctx context.Context, input *BatchInput) (BatchBase, error) {
	var batchBase BatchBase
	if input.Id != nil {
		batchBase, _ = s.batchRepo.GetBatchBaseById(ctx, input.Id)
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

func (s *BatchService) IncrementBatch(ctx context.Context, batchInput BatchInput) error {
	if err := ValidateBatchInputIncrement(batchInput); err != nil {
		return err
	}
	err := common.RunWithTransaction(ctx, s.batchRepo.(*BatchRepository).Pool, func(ctx context.Context, tx pgx.Tx) error {
		ctx = common.SetOperator(ctx, tx)
		return s.tryToIncrementBatch(ctx, batchInput)
	})
	return err
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
	return s.lockingService.RunWithLock(ctx, GenerateBatchLockKey(input.Sku), func() error {
		expiresAt, err := s.productService.GetProductVariantExpirationDate(ctx, input.Sku)
		if err != nil {
			return err
		}
		return s.batchRepo.CreateBatch(ctx, input, common.GetUtcDateOnlyStringFromTime(expiresAt))
	})
}

func (s *BatchService) incrementBatch(ctx context.Context, batch BatchBase, input BatchInput) error {
	return s.lockingService.RunWithLock(ctx, GenerateBatchLockKey(strconv.Itoa(*input.Id)), func() error {
		batch = batch.SetQuantity(batch.Quantity + input.Quantity)
		err := s.batchRepo.UpdateBatch(ctx, batch)
		shouldDecrementRecipe := ctx.Value(DecrementRecipeKey{}).(bool)
		if shouldDecrementRecipe {
			err = s.tryToDecrementRecipe(ctx, input)
		}
		return err
	})
}

func (s *BatchService) tryToDecrementRecipe(ctx context.Context, input BatchInput) error {
	recipes, err := s.recipeService.GetRecipeOfProductVariantSku(ctx, input.Sku)
	if err != nil {
		return err
	}
	batchInputsFromRecipes := make([]BatchInput, len(recipes))
	for i, recipe := range recipes {
		batchInputsFromRecipes[i] = BatchInput{
			Id:       recipe.RecipeVariantId,
			Sku:      recipe.RecipeVariantSku,
			Quantity: recipe.Quantity * input.Quantity,
			UnitId:   *recipe.Unit.Id,
		}
	}
	return s.bulkDecrementBatch(ctx, batchInputsFromRecipes)
}

func (s *BatchService) DecrementBatch(ctx context.Context, input BatchInput) error {
	if err := ValidateBatchInputIncrement(input); err != nil {
		return err
	}
	err := common.RunWithTransaction(ctx, s.batchRepo.(*BatchRepository).Pool, func(ctx context.Context, tx pgx.Tx) error {
		ctx = common.SetOperator(ctx, tx)
		return s.tryToDecrementBatch(ctx, input)
	})
	return err
}

func (s *BatchService) tryToDecrementBatch(ctx context.Context, input BatchInput) error {
	batchBase, err := s.getConvertedBatch(ctx, &input)
	if err != nil {
		return err
	}
	if batchBase.Id == nil {
		return common.NewBadRequestFromMessage("batch not found")
	}
	return s.decrementBatch(ctx, batchBase, input)
}

func (s *BatchService) decrementBatch(ctx context.Context, batchBase BatchBase, input BatchInput) error {
	return s.lockingService.RunWithLock(ctx, GenerateBatchLockKey(strconv.Itoa(*input.Id)), func() error {
		newQty := batchBase.Quantity - input.Quantity
		if newQty < 0 {
			return common.NewBadRequestFromMessage("batch not enough")
		}
		batchBase = batchBase.SetQuantity(newQty)
		return s.batchRepo.UpdateBatch(ctx, batchBase)
	})
}

func (s *BatchService) BulkIncrementBatch(ctx context.Context, inputs []BatchInput) error {
	err := common.RunWithTransaction(ctx, s.batchRepo.(*BatchRepository).Pool, func(ctx context.Context, tx pgx.Tx) error {
		ctx = common.SetOperator(ctx, tx)
		return s.bulkIncrementBatch(ctx, inputs)
	})
	return err
}

func (s *BatchService) bulkIncrementBatch(ctx context.Context, inputs []BatchInput) error {
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

func (s *BatchService) BulkDecrementBatch(ctx context.Context, inputs []BatchInput) error {
	err := common.RunWithTransaction(ctx, s.batchRepo.(*BatchRepository).Pool, func(ctx context.Context, tx pgx.Tx) error {
		ctx = common.SetOperator(ctx, tx)
		return s.bulkDecrementBatch(ctx, inputs)
	})
	return err
}

func (s *BatchService) bulkDecrementBatch(ctx context.Context, inputs []BatchInput) error {
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

func (s *BatchService) GetBatches(ctx context.Context) (common.PaginatedResponse[Batch], error) {
	paginationParam := common.GetPaginationParams(ctx)
	batches, err := s.batchRepo.GetBatches(ctx, paginationParam)
	if err != nil {
		return common.PaginatedResponse[Batch]{}, err
	}
	return s.createBatchesPage(batches, paginationParam.PageSize), nil
}

func (s *BatchService) SearchBatchesBySku(ctx context.Context, sku string) (common.PaginatedResponse[Batch], error) {
	paginationParam := common.GetPaginationParams(ctx)
	batches, err := s.batchRepo.SearchBatchesBySku(ctx, sku, paginationParam)
	if err != nil {
		return common.PaginatedResponse[Batch]{}, err
	}
	return s.createBatchesPage(batches, paginationParam.PageSize), nil
}

func (s *BatchService) createBatchesPage(batches []Batch, pageSize int) common.PaginatedResponse[Batch] {
	if len(batches) == 0 {
		return common.CreateEmptyPaginatedResponse[Batch](pageSize)
	}
	first, last := batches[0], batches[len(batches)-1]
	res := common.CreatePaginatedResponse[Batch](
		pageSize,
		last,
		first,
		batches,
	)
	return res
}
