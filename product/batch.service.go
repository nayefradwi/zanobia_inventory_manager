package product

import (
	"context"
	"strconv"

	"github.com/jackc/pgx/v4"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
	"github.com/nayefradwi/zanobia_inventory_manager/transactions"
)

type DecrementRecipeKey struct{}
type UseMostExpiredKey struct{}
type IBatchService interface {
	IncrementBatch(ctx context.Context, batchInput BatchInput) error
	DecrementBatch(ctx context.Context, input BatchInput) error
	BulkIncrementBatch(ctx context.Context, inputs []BatchInput) error
	BulkDecrementBatch(ctx context.Context, inputs []BatchInput) error
	GetBatches(ctx context.Context) (common.PaginatedResponse[Batch], error)
	SearchBatchesBySku(ctx context.Context, sku string) (common.PaginatedResponse[Batch], error)
	DecrementForRetailer(ctx context.Context, input BatchInput) error
	ReturnToWarehouse(ctx context.Context, input BatchInput) error
}

type BatchService struct {
	batchRepo          IBatchRepository
	productService     IProductService
	lockingService     common.IDistributedLockingService
	unitService        IUnitService
	recipeService      IRecipeService
	transactionService transactions.ITransactionService
}

func NewBatchService(
	batchRepo IBatchRepository,
	productService IProductService,
	lockingService common.IDistributedLockingService,
	unitService IUnitService,
	recipeService IRecipeService,
	transactionService transactions.ITransactionService,
) *BatchService {
	return &BatchService{
		batchRepo,
		productService,
		lockingService,
		unitService,
		recipeService,
		transactionService,
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
		expiresAt, price, err := s.productService.GetProductVariantExpirationDateAndCost(ctx, input.Sku)
		if err != nil {
			return err
		}
		id, err := s.batchRepo.CreateBatch(ctx, input, common.GetUtcDateOnlyStringFromTime(expiresAt))
		if err != nil {
			return err
		}
		shouldDecrementRecipe := ctx.Value(DecrementRecipeKey{}).(bool)
		if shouldDecrementRecipe {
			err = s.tryToDecrementRecipe(ctx, input)
		}
		if err != nil {
			return err
		}
		transactionCommand := transactions.CreateWarehouseTransactionCommand{
			BatchId:    id,
			Quantity:   input.Quantity,
			UnitId:     input.UnitId,
			Reason:     input.Reason,
			Sku:        input.Sku,
			CostPerQty: price,
			Comment:    "New batch created",
		}
		return s.transactionService.CreateWarehouseTransaction(ctx, transactionCommand)
	})
}

func (s *BatchService) incrementBatch(ctx context.Context, batch BatchBase, input BatchInput) error {
	return s.lockingService.RunWithLock(ctx, GenerateBatchLockKey(strconv.Itoa(*input.Id)), func() error {
		batch = batch.SetQuantity(batch.Quantity + input.Quantity)
		err := s.batchRepo.UpdateBatch(ctx, batch)
		if err != nil {
			return err
		}
		shouldDecrementRecipe := ctx.Value(DecrementRecipeKey{}).(bool)
		if shouldDecrementRecipe {
			err = s.tryToDecrementRecipe(ctx, input)
		}
		if err != nil {
			return err
		}
		_, price, err := s.productService.GetProductVariantExpirationDateAndCost(ctx, input.Sku)
		if err != nil {
			return err
		}
		transactionCommand := transactions.CreateWarehouseTransactionCommand{
			BatchId:    *batch.Id,
			Quantity:   input.Quantity,
			UnitId:     input.UnitId,
			Reason:     input.Reason,
			Sku:        input.Sku,
			CostPerQty: price,
			Comment:    "New batch incremented",
		}
		return s.transactionService.CreateWarehouseTransaction(ctx, transactionCommand)
	})
}

func (s *BatchService) tryToDecrementRecipe(ctx context.Context, input BatchInput) error {
	recipes, err := s.recipeService.GetRecipeOfProductVariantSku(ctx, input.Sku)
	if err != nil {
		return err
	}
	batchInputsFromRecipes := make([]BatchInput, len(recipes))
	for i, recipe := range recipes {
		batchId, err := s.getBatchIdForRecipeBasedOnCtx(ctx, recipe.RecipeVariantSku)
		if err != nil {
			return err
		}
		batchInputsFromRecipes[i] = BatchInput{
			Id:         &batchId,
			Sku:        recipe.RecipeVariantSku,
			Quantity:   recipe.Quantity * input.Quantity,
			UnitId:     *recipe.Unit.Id,
			Reason:     transactions.TransactionReasonTypeRecipeUse,
			CostPerQty: recipe.IngredientCost,
		}
	}
	return s.bulkDecrementBatch(ctx, batchInputsFromRecipes)
}

func (s *BatchService) getBatchIdForRecipeBasedOnCtx(ctx context.Context, sku string) (int, error) {
	shouldUseMostExpired := ctx.Value(UseMostExpiredKey{}).(bool)
	if shouldUseMostExpired {
		return s.batchRepo.GetMostExpiredBatchId(ctx, sku)
	}
	return s.batchRepo.GetLeastExpiredBatchId(ctx, sku)
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
		if input.CostPerQty == 0 {
			_, costPerQty, err := s.productService.GetProductVariantExpirationDateAndCost(ctx, input.Sku)
			if err != nil {
				return err
			}
			input.CostPerQty = costPerQty
		}
		if err := s.batchRepo.UpdateBatch(ctx, batchBase); err != nil {
			return err
		}
		transactionCommand := transactions.CreateWarehouseTransactionCommand{
			BatchId:    *input.Id,
			Quantity:   input.Quantity,
			UnitId:     input.UnitId,
			Reason:     input.Reason,
			Sku:        input.Sku,
			CostPerQty: input.CostPerQty,
			Comment:    "New batch decremented",
		}
		return s.transactionService.CreateWarehouseTransaction(ctx, transactionCommand)
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

func (s *BatchService) DecrementForRetailer(ctx context.Context, input BatchInput) error {
	return s.tryToDecrementBatch(ctx, input)
}

func (s *BatchService) ReturnToWarehouse(ctx context.Context, input BatchInput) error {
	return s.tryToIncrementBatch(ctx, input)
}
