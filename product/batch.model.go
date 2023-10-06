package product

import "time"

type BatchInput struct {
	ProductId int     `json:"productId,omitempty"`
	Quantity  float64 `json:"quantity"`
	UnitId    int     `json:"unitId"`
}

type BatchBase struct {
	Id          *int      `json:"id,omitempty"`
	WarehouseId *int      `json:"warehouseId,omitempty"`
	ProductId   int       `json:"productId"`
	Quantity    float64   `json:"quantity"`
	UnitId      int       `json:"unitId"`
	ExpiresAt   time.Time `json:"expiresAt"`
}

type Batch struct {
	Id          *int        `json:"id,omitempty"`
	WarehouseId *int        `json:"warehouseId,omitempty"`
	Product     ProductBase `json:"product"`
	Quantity    float64     `json:"quantity"`
	Unit        Unit        `json:"unit"`
	ExpiresAt   time.Time   `json:"expiresAt"`
}

func (b BatchBase) SetQuantity(quantity float64) BatchBase {
	b.Quantity = quantity
	return b
}

func (b BatchBase) SetExpiresAt(expiresAt time.Time) BatchBase {
	b.ExpiresAt = expiresAt
	return b
}
