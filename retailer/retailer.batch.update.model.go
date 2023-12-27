package retailer

import (
	"time"

	"github.com/nayefradwi/zanobia_inventory_manager/common"
	"github.com/nayefradwi/zanobia_inventory_manager/product"
	"github.com/nayefradwi/zanobia_inventory_manager/transactions"
)

type BulkRetailerBatchUpdateInfo struct {
	BatchBasesLookup           map[string]RetailerBatchBase
	BatchVariantMetaInfoLookup map[string]product.BatchVariantMetaInfo
	BatchInputMapToUpdate      map[string]RetailerBatchInput
	BatchInputMapToCreate      map[string]RetailerBatchInput
	SkuList                    []string
	Ids                        []int
	locks                      []common.Lock
}

type RetailerBatchUpdateRequest struct {
	BatchId    *int
	NewValue   float64
	Reason     string
	Sku        string
	ModifiedBy float64
}

type RetailerBatchCreateRequest struct {
	BatchSku   string
	RetailerId int
	Quantity   float64
	UnitId     int
	ExpiryDate time.Time
}

type BulkRetailerBatchUpdateUnitOfWork struct {
	BatchUpdateRequestLookup map[string]RetailerBatchUpdateRequest
	BatchCreateRequestLookup map[string]RetailerBatchCreateRequest
	BatchTransactionHistory  []transactions.CreateRetailerTransactionCommand
}
