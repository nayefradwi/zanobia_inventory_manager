package product

type InventoryBase struct {
	Id           *int    `json:"id,omitempty"`
	IngredientId int     `json:"ingredientId,omitempty"`
	WarehouseId  int     `json:"warehouseId,omitempty"`
	Quantity     float64 `json:"quantity"`
	UnitId       int     `json:"unitId,omitempty"`
}

type Inventory struct {
	Id          *int       `json:"id,omitempty"`
	WarehouseId *int       `json:"warehouse"`
	Ingredient  Ingredient `json:"ingredient"`
	Quantity    float64    `json:"quantity"`
	Unit        Unit       `json:"unit"`
}

type InventoryInput struct {
	IngredientId int     `json:"ingredientId"`
	Quantity     float64 `json:"quantity"`
	UnitId       int     `json:"unitId"`
}

func (b InventoryBase) SetQuantity(quantity float64) InventoryBase {
	b.Quantity = quantity
	return b
}
