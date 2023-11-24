package retailer

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
)

type IRetailerBatchRepository interface {
	CreateBatch(ctx context.Context, input RetailerBatchInput, expiresAt string) error
	UpdateBatch(ctx context.Context, base RetailerBatchBase) error
	GetBatchBaseById(ctx context.Context, id *int) (RetailerBatchBase, error)
	GetBatchBase(ctx context.Context, sku string, expirationDate string) (RetailerBatchBase, error)
	GetBatches(ctx context.Context, params common.PaginationParams) ([]RetailerBatch, error)
	SearchBatchesBySku(ctx context.Context, sku string, params common.PaginationParams) ([]RetailerBatch, error)
}

type RetailerBatchRepository struct {
	*pgxpool.Pool
}

func NewRetailerBatchRepository(db *pgxpool.Pool) *RetailerBatchRepository {
	return &RetailerBatchRepository{
		db,
	}
}

func (r *RetailerBatchRepository) CreateBatch(ctx context.Context, input RetailerBatchInput, expiresAt string) error {
	// TODO fill
	return nil
}

func (r *RetailerBatchRepository) UpdateBatch(ctx context.Context, base RetailerBatchBase) error {
	// TODO fill
	return nil
}

func (r *RetailerBatchRepository) GetBatchBaseById(ctx context.Context, id *int) (RetailerBatchBase, error) {
	// TODO fill
	return RetailerBatchBase{}, nil
}

func (r *RetailerBatchRepository) GetBatchBase(ctx context.Context, sku string, expirationDate string) (RetailerBatchBase, error) {
	// TODO fill
	return RetailerBatchBase{}, nil
}

func (r *RetailerBatchRepository) GetBatches(ctx context.Context, params common.PaginationParams) ([]RetailerBatch, error) {
	// TODO fill
	return nil, nil
}

func (r *RetailerBatchRepository) SearchBatchesBySku(ctx context.Context, sku string, params common.PaginationParams) ([]RetailerBatch, error) {
	// TODO fill
	return nil, nil
}
