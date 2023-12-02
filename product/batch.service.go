package product

import (
	"context"
	"strconv"

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

func GenerateBatchLockKey(batchInput BatchInput) string {
	if batchInput.Id != nil {
		return "batch:" + strconv.Itoa(*batchInput.Id) + ":lock"
	}
	return "batch:" + batchInput.Sku + ":lock"
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
