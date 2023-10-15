package product

import "time"

type BatchInput struct {
	Sku      int     `json:"Sku,omitempty"`
	Quantity float64 `json:"quantity"`
	UnitId   int     `json:"unitId"`
}

type BatchBase struct {
	Id          *int      `json:"id,omitempty"`
	WarehouseId *int      `json:"warehouseId,omitempty"`
	Sku         string    `json:"Sku"`
	Quantity    float64   `json:"quantity"`
	UnitId      int       `json:"unitId"`
	ExpiresAt   time.Time `json:"expiresAt"`
}

type Batch struct {
	Id                 *int                `json:"id,omitempty"`
	WarehouseId        *int                `json:"warehouseId,omitempty"`
	ProductVariantBase *ProductVariantBase `json:"productVariantBase,omitempty"`
	Quantity           float64             `json:"quantity"`
	Unit               Unit                `json:"unit"`
	ExpiresAt          time.Time           `json:"expiresAt"`
}

func (b BatchBase) SetQuantity(quantity float64) BatchBase {
	b.Quantity = quantity
	return b
}

func (b BatchBase) SetExpiresAt(expiresAt time.Time) BatchBase {
	b.ExpiresAt = expiresAt
	return b
}
