package retailer

import (
	"context"
	"strconv"

	"github.com/nayefradwi/zanobia_inventory_manager/common"
	"github.com/nayefradwi/zanobia_inventory_manager/product"
	"github.com/nayefradwi/zanobia_inventory_manager/transactions"
	"github.com/nayefradwi/zanobia_inventory_manager/unit"
)

type IRetailerBatchService interface {
	IncrementBatch(ctx context.Context, batchInput RetailerBatchInput) error
	DecrementBatch(ctx context.Context, input RetailerBatchInput) error
	BulkIncrementBatch(ctx context.Context, inputs []RetailerBatchInput) error
	BulkDecrementBatch(ctx context.Context, inputs []RetailerBatchInput) error
	GetBatches(ctx context.Context, retailerId int) (common.PaginatedResponse[RetailerBatch], error)
	SearchBatchesBySku(ctx context.Context, retailerId int, sku string) (common.PaginatedResponse[RetailerBatch], error)
	DeleteBatchesOfRetailer(ctx context.Context, retailerId int) error
	MoveFromWarehouseToRetailer(ctx context.Context, moveInput RetailerBatchFromWarehouseInput) error
	ReturnBatchToWarehouse(ctx context.Context, moveInput RetailerBatchFromWarehouseInput) error
}

type RetailerBatchService struct {
	repo               IRetailerBatchRepository
	productService     product.IProductService
	lockingService     common.IDistributedLockingService
	unitService        unit.IUnitService
	transactionService transactions.ITransactionService
	batchService       product.IBatchService
}

func NewRetailerBatchService(
	repo IRetailerBatchRepository,
	productService product.IProductService,
	lockingService common.IDistributedLockingService,
	unitService unit.IUnitService,
	transactionService transactions.ITransactionService,
	batchService product.IBatchService,
) *RetailerBatchService {
	return &RetailerBatchService{
		repo,
		productService,
		lockingService,
		unitService,
		transactionService,
		batchService,
	}
}

func GenerateRetailerBatchLockKey(fieldValue string) string {
	return "retailer-batch:" + fieldValue + ":lock"
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

func (s *RetailerBatchService) convertBatchInput(
	ctx context.Context,
	batchInput RetailerBatchInput,
	batchVariantMetaInfo product.BatchVariantMetaInfo,
) (
	RetailerBatchInput,
	error,
) {
	convertInput := unit.ConvertUnitInput{
		ToUnitId:   &batchVariantMetaInfo.UnitId,
		Quantity:   batchInput.Quantity,
		FromUnitId: &batchInput.UnitId,
	}
	conversionOutput, err := s.unitService.ConvertUnit(ctx, convertInput)
	if err != nil {
		return RetailerBatchInput{}, err
	}
	batchInput.Quantity = conversionOutput.Quantity
	batchInput.UnitId = *conversionOutput.Unit.Id
	return batchInput, nil
}

func (s *RetailerBatchService) processBulkBatchUnitOfWork(
	ctx context.Context,
	bulkBatchUpdateUnitOfWork BulkRetailerBatchUpdateUnitOfWork,
) error {
	pgxBatch, err := s.transactionService.(*transactions.TransactionService).
		CreateTransactionHistoryBatches(
			ctx,
			bulkBatchUpdateUnitOfWork.BatchTransactionHistory,
		)
	if err != nil {
		return err
	}
	return s.repo.(*RetailerBatchRepository).
		processBulkBatchUnitOfWork(
			ctx,
			bulkBatchUpdateUnitOfWork,
			pgxBatch,
		)
}
