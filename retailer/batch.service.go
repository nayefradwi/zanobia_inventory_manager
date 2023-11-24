package retailer

import (
	"context"

	"github.com/nayefradwi/zanobia_inventory_manager/common"
)

type IRetailerBatchService interface {
	IncrementBatch(ctx context.Context, batchInput RetailerBatchInput) error
	DecrementBatch(ctx context.Context, input RetailerBatchInput) error
	BulkIncrementBatch(ctx context.Context, inputs []RetailerBatchInput) error
	BulkDecrementBatch(ctx context.Context, inputs []RetailerBatchInput) error
	GetBatches(ctx context.Context) (common.PaginatedResponse[RetailerBatch], error)
	SearchBatchesBySku(ctx context.Context, sku string) (common.PaginatedResponse[RetailerBatch], error)
	DeleteBatchesOfRetailer(ctx context.Context, retailerId int) error
}

type RetailerBatchService struct {
	repo IRetailerBatchRepository
}

func NewRetailerBatchService(repo IRetailerBatchRepository) *RetailerBatchService {
	return &RetailerBatchService{
		repo,
	}
}

func (s *RetailerBatchService) IncrementBatch(ctx context.Context, batchInput RetailerBatchInput) error {
	// TODO fill
	return nil
}

func (s *RetailerBatchService) DecrementBatch(ctx context.Context, input RetailerBatchInput) error {
	// TODO fill
	return nil
}

func (s *RetailerBatchService) BulkIncrementBatch(ctx context.Context, inputs []RetailerBatchInput) error {
	// TODO fill
	return nil
}

func (s *RetailerBatchService) BulkDecrementBatch(ctx context.Context, inputs []RetailerBatchInput) error {
	// TODO fill
	return nil
}

func (s *RetailerBatchService) GetBatches(ctx context.Context) (common.PaginatedResponse[RetailerBatch], error) {
	// TODO fill
	return common.PaginatedResponse[RetailerBatch]{}, nil
}

func (s *RetailerBatchService) SearchBatchesBySku(ctx context.Context, sku string) (common.PaginatedResponse[RetailerBatch], error) {
	// TODO fill
	return common.PaginatedResponse[RetailerBatch]{}, nil
}

func (s *RetailerBatchService) DeleteBatchesOfRetailer(ctx context.Context, retailerId int) error {
	// TODO fill
	return nil
}
