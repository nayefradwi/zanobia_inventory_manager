package product

import (
	"time"
)

type InventoryBase struct {
	Id           *int    `json:"id,omitempty"`
	IngredientId int     `json:"ingredientId"`
	WarehouseId  *int    `json:"warehouseId,omitempty"`
	Quantity     float64 `json:"quantity"`
	UnitId       int     `json:"unitId,omitempty"`
}

type Inventory struct {
	Id          *int       `json:"id,omitempty"`
	WarehouseId *int       `json:"warehouse,omitempty"`
	Ingredient  Ingredient `json:"ingredient"`
	Quantity    float64    `json:"quantity"`
	Unit        Unit       `json:"unit"`
	UpdatedAt   time.Time  `json:"updatedAt,omitempty"`
}

type InventoryInput struct {
	IngredientId int     `json:"ingredientId,omitempty"`
	Quantity     float64 `json:"quantity"`
	UnitId       int     `json:"unitId"`
}

func (b InventoryBase) SetQuantity(quantity float64) InventoryBase {
	b.Quantity = quantity
	return b
}

func (b Inventory) GetCursorValue() string {
	return b.UpdatedAt.Format(time.RFC3339)
}
