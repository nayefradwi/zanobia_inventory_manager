package product

type IBatchService interface{}

type BatchService struct {
	batchRepo        IBatchRepository
	inventoryService IInventoryService
	productService   IProductService
}

func NewBatchService(batchRepo IBatchRepository, inventoryService IInventoryService, productService IProductService) *BatchService {
	return &BatchService{
		batchRepo,
		inventoryService,
		productService,
	}
}

// Increment batch
// Decrement batch
// Bulk increment batches
// Bulk decrement batches
// Get batches paginated sorted by expiration date
