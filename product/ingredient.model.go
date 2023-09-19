package product

type IngredientBase struct {
	Id             *int   `json:"id,omitempty"`
	Name           string `json:"name"`
	Brand          string `json:"brand"`
	StandardUnitId *int   `json:"standard_unit_id,omitempty"`
}

type Ingredient struct {
	IngredientBase
	StandardUnit *Unit `json:"standard_unit"`
}
