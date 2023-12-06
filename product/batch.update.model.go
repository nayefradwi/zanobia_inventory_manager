package product

import (
	"time"

	"github.com/nayefradwi/zanobia_inventory_manager/common"
	"github.com/nayefradwi/zanobia_inventory_manager/transactions"
)

type BatchVariantMetaInfo struct {
	UnitId        int
	ExpiresInDays int
	Cost          float64
}

type BulkBatchUpdateInfo struct {
	RecipeMap                  map[string]Recipe
	BatchBasesLookup           map[string]BatchBase
	BatchVariantMetaInfoLookup map[string]BatchVariantMetaInfo
	BatchInputMapToUpdate      map[string]BatchInput
	BatchInputMapToCreate      map[string]BatchInput
	SkuList                    []string
	Ids                        []int
	locks                      []common.Lock
}

type BatchUpdateRequest struct {
	BatchId    *int
	NewValue   float64
	Reason     string
	Sku        string
	ModifiedBy float64
}

type BatchCreateRequest struct {
	BatchSku   string
	Quantity   float64
	UnitId     int
	ExpiryDate time.Time
}

type BulkBatchUpdateUnitOfWork struct {
	BatchUpdateRequestLookup map[string]BatchUpdateRequest
	BatchCreateRequestLookup map[string]BatchCreateRequest
	BatchTransactionHistory  []transactions.CreateWarehouseTransactionCommand
}
