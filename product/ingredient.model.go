package product

import "strconv"

type IngredientBase struct {
	Id             *int    `json:"id,omitempty"`
	Name           string  `json:"name"`
	Brand          string  `json:"brand"`
	Price          float64 `json:"price"`
	StandardUnitId *int    `json:"standardUnitId,omitempty"`
	ExpiresInDays  int     `json:"expiresInDays"`
	StandardQty    float64 `json:"standardQty,omitempty"`
}

type Ingredient struct {
	IngredientBase
	StandardUnit *Unit `json:"standardUnit,omitempty"`
}

func (b Ingredient) GetCursorValue() string {
	return strconv.Itoa(*b.Id)
}
