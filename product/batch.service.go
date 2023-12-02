package product

import (
	"context"
	"log"

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

func (s *BatchService) IncrementBatch(ctx context.Context, batchInput BatchInput) error {
	return s.BulkIncrementBatch(ctx, []BatchInput{batchInput})
}

func (s *BatchService) BulkIncrementBatch(ctx context.Context, inputs []BatchInput) error {
	if err := ValidateBatchInputsIncrement(inputs); err != nil {
		return err
	}
	return common.RunWithTransaction(ctx, s.batchRepo.(*BatchRepository).Pool, func(ctx context.Context, tx pgx.Tx) error {
		ctx = common.SetOperator(ctx, tx)
		return s.bulkIncrementBatchesTransaction(ctx, inputs)
	})
}

func (s *BatchService) bulkIncrementBatchesTransaction(ctx context.Context, inputs []BatchInput) error {
	// TODO: lock all the batches
	batchUpdateRequest, err := s.createBulkUpdateRequest(ctx, inputs)
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
	// TODO: bulkUpdateBatches(ctx, batchNewQuantities, batchUpdateRequest.BatchInputMapToCreate)
	// TODO: create transaction history
	return nil
}

func (s *BatchService) DecrementBatch(ctx context.Context, batchInput BatchInput) error {
	return s.BulkDecrementBatch(ctx, []BatchInput{batchInput})
}

func (s *BatchService) BulkDecrementBatch(ctx context.Context, inputs []BatchInput) error {
	if err := ValidateBatchInputsDecrement(inputs); err != nil {
		return err
	}
	return common.RunWithTransaction(ctx, s.batchRepo.(*BatchRepository).Pool, func(ctx context.Context, tx pgx.Tx) error {
		ctx = common.SetOperator(ctx, tx)
		return s.bulkDecrementBatchesTransaction(ctx, inputs)
	})

}

func (s *BatchService) bulkDecrementBatchesTransaction(ctx context.Context, inputs []BatchInput) error {
	// TODO: lock all the batches

	batchUpdateRequest, err := s.createBulkUpdateRequestWithoutRecipe(ctx, inputs)
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

func (s *BatchService) createBulkUpdateRequest(
	ctx context.Context,
	input []BatchInput,
) (BulkBatchUpdateRequest, error) {
	shouldDecrementRecipe := common.GetBoolFromContext(ctx, DecrementRecipeKey{})
	if shouldDecrementRecipe {
		return s.createBulkUpdateRequestWithRecipe(ctx, input)
	}
	return s.createBulkUpdateRequestWithoutRecipe(ctx, input)
}

/*
complexity analysis:
- O(9n) time = O(n) time
- O(6n) space = O(n) space
- n = number of batch inputs
- O(4) database calls if recipe is used
*/
func (s *BatchService) createBulkUpdateRequestWithRecipe(
	ctx context.Context,
	input []BatchInput,
) (BulkBatchUpdateRequest, error) {
	// this will give you all the batches in their correct unit conversions
	batchUpdateRequest, err := s.createBulkUpdateRequest(ctx, input)
	if err != nil {
		return BulkBatchUpdateRequest{}, err
	}
	recipeLookUp, recipeVariantSkus, err := s.recipeService.GetRecipesLookUpMapFromSkus(ctx, batchUpdateRequest.SkuList)
	if err != nil {
		return BulkBatchUpdateRequest{}, err
	}
	if len(recipeVariantSkus) == 0 {
		// no recipes found
		return batchUpdateRequest, nil
	}
	recipeBatches, err := s.getRecipeBatchBases(ctx, recipeVariantSkus)
	if err != nil {
		return BulkBatchUpdateRequest{}, err
	}
	recipeBatchMap := make(map[string]BatchInput, len(recipeBatches))
	recipeBatchBaseIds := make([]int, len(recipeBatches))
	for _, recipeBatchBase := range recipeBatches {
		if recipeBatchBase.Id == nil && recipeBatchBase.Sku == "" {
			return BulkBatchUpdateRequest{}, common.NewBadRequestFromMessage("invalid recipe batch")
		}
		recipe, recipeFound := recipeLookUp[recipeBatchBase.Sku]
		if !recipeFound {
			return BulkBatchUpdateRequest{}, common.NewBadRequestFromMessage("invalid recipe batch")
		}
		resultBatch, resultFound := batchUpdateRequest.BatchInputLookUp[recipe.ResultVariantSku]
		if !resultFound {
			return BulkBatchUpdateRequest{}, common.NewBadRequestFromMessage("invalid recipe batch")
		}
		recipeBatchInput := BatchInput{
			Id: recipeBatchBase.Id,
			// calculate the quantity of the recipe batch
			Quantity:   resultBatch.Quantity * recipe.Quantity,
			UnitId:     recipeBatchBase.UnitId,
			Sku:        recipe.RecipeVariantSku,
			Reason:     transactions.TransactionReasonTypeRecipeUse,
			CostPerQty: recipe.IngredientCost * recipe.Quantity * resultBatch.Quantity,
		}
		recipeBatchMap[recipe.RecipeVariantSku] = recipeBatchInput
		recipeBatchBaseIds = append(recipeBatchBaseIds, *recipeBatchBase.Id)
	}
	batchUpdateRequest.RecipeBatchInputMap = recipeBatchMap
	batchUpdateRequest.RecipeMap = recipeLookUp
	batchUpdateRequest.SkuList = append(batchUpdateRequest.SkuList, recipeVariantSkus...)
	batchUpdateRequest.BatchIds = append(batchUpdateRequest.BatchIds, recipeBatchBaseIds...)
	batchUpdateRequest.BatchInputLookUp = common.MergeMaps[string, BatchInput](batchUpdateRequest.BatchInputLookUp, recipeBatchMap)
	return batchUpdateRequest, nil
}

func (s *BatchService) getRecipeBatchBases(ctx context.Context, recipeSkus []string) ([]BatchBase, error) {
	useMostExpired := common.GetBoolFromContext(ctx, UseMostExpiredKey{})
	if useMostExpired {
		return s.batchRepo.getMostExpiredBatchBasesFromSkus(ctx, recipeSkus)
	}
	return s.batchRepo.getLeastExpiredBatchBasesFromSkus(ctx, recipeSkus)
}

/*
complexity analysis:
- O(6n) time = O(n) time
- O(2n) space = O(n) space
- n = number of batch inputs
- O(2) database calls
*/
func (s *BatchService) createBulkUpdateRequestWithoutRecipe(
	ctx context.Context,
	input []BatchInput,
) (BulkBatchUpdateRequest, error) {
	batchInputMapToUpdate, batchInputMapToCreate, skuList := s.createBatchInputMapSkuListAndIdList(input)
	originalUnitsMap, err := s.productService.GetOriginalUnitsBySkuList(ctx, skuList)
	if err != nil {
		return BulkBatchUpdateRequest{}, err
	}
	batchInputMapToUpdate, err = s.convertUnitOfBatchInputMap(ctx, batchInputMapToUpdate, originalUnitsMap)
	if err != nil {
		return BulkBatchUpdateRequest{}, err
	}
	batchInputMapToCreate, err = s.convertUnitOfBatchInputMap(ctx, batchInputMapToCreate, originalUnitsMap)
	if err != nil {
		return BulkBatchUpdateRequest{}, err
	}
	batchIds := make([]int, len(batchInputMapToUpdate))
	i := 0
	for _, batchInput := range batchInputMapToUpdate {
		batchIds[i] = *batchInput.Id
		i++
	}
	batchBases, err := s.batchRepo.getBatchBasesFromIds(ctx, batchIds)
	if err != nil {
		return BulkBatchUpdateRequest{}, err
	}
	batchBaseMap := make(map[string]BatchBase, len(batchBases))
	for _, batchBase := range batchBases {
		if batchBase.Sku == "" {
			return BulkBatchUpdateRequest{}, common.NewBadRequestFromMessage("invalid batch input")
		}
		batchBaseMap[batchBase.Sku] = batchBase
	}
	batchInputLookup := common.MergeMaps[string, BatchInput](batchInputMapToUpdate, batchInputMapToCreate)
	return BulkBatchUpdateRequest{
		BatchInputMapToUpdate: batchInputMapToUpdate,
		BatchInputMapToCreate: batchInputMapToCreate,
		SkuList:               skuList,
		BatchBasesLookup:      batchBaseMap,
		BatchIds:              batchIds,
		// this will have all the converted unit batch inputs
		BatchInputLookUp: batchInputLookup,
	}, nil
}

func (s *BatchService) createBatchInputMapSkuListAndIdList(input []BatchInput) (map[string]BatchInput, map[string]BatchInput, []string) {
	batchInputMapToUpdate := make(map[string]BatchInput)
	batchInputMapToCreate := make(map[string]BatchInput)
	skuList := make([]string, 0)
	for _, batchInput := range input {
		if batchInput.Id != nil {
			batchInputMapToUpdate[batchInput.Sku] = batchInput
		} else {
			batchInputMapToCreate[batchInput.Sku] = batchInput
		}
		skuList = append(skuList, batchInput.Sku)
	}
	return batchInputMapToUpdate, batchInputMapToCreate, skuList
}

func (s *BatchService) convertUnitOfBatchInputMap(
	ctx context.Context,
	batchInputMap map[string]BatchInput,
	originalUnitsMap map[string]int,
) (map[string]BatchInput, error) {
	for sku, batchInput := range batchInputMap {
		originalUnitId, _ := originalUnitsMap[sku]
		batchUnitId := batchInput.UnitId
		convertUnitInput := ConvertUnitInput{
			ToUnitId:   &originalUnitId,
			FromUnitId: &batchUnitId,
			Quantity:   batchInput.Quantity,
		}
		convertedOutput, err := s.unitService.ConvertUnit(ctx, convertUnitInput)
		if err != nil {
			return nil, err
		}
		batchInput.Quantity = convertedOutput.Quantity
		batchInput.UnitId = originalUnitId
		batchInputMap[sku] = batchInput
	}
	return batchInputMap, nil
}
