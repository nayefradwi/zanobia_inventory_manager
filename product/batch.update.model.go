package product

import "time"

type BulkBatchUpdateInfo struct {
	RecipeMap             map[string]Recipe
	BatchBasesLookup      map[string]BatchBase
	OriginalUnitsMap      map[string]int
	BatchInputMapToUpdate map[string]BatchInput
	BatchInputMapToCreate map[string]BatchInput
	SkuList               []string
	Ids                   []int
}

type BatchUpdateRequest struct {
	BatchId  *int
	NewValue float64
	Reason   string
}

type BatchCreateRequest struct {
	BatchSku   string
	Quantity   float64
	UnitId     int
	ExpiryDate time.Time
}

type BatchTransactionInfo struct {
	BatchId *int
	Amount  float64
	Reason  string
	UnitId  int
	Sku     string
}

type BulkBatchUpdateUnitOfWork struct {
	BatchUpdateRequestLookup map[string]BatchUpdateRequest
	BatchCreateRequestLookup map[string]BatchCreateRequest
	BatchTransactionInfo     []BatchTransactionInfo
}
