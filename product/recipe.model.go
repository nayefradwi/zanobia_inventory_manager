package product

type RecipeBase struct {
	Id               *int    `json:"id"`
	ResultVariantSku string  `json:"resultVariantSku"`
	Quantity         float64 `json:"quantity"`
	UnitId           *int    `json:"unitId"`
	RecipeVariantSku string  `json:"recipeVariantSku"`
}

type Recipe struct {
	Id                     *int    `json:"id"`
	ResultVariantId        *int    `json:"resultVariantId,omitempty"`
	ResultVariantName      string  `json:"resultVariantName,omitempty"`
	ResultVariantSku       string  `json:"resultVariantSku,omitempty"`
	ProductName            string  `json:"productName,omitempty"`
	Quantity               float64 `json:"quantity"`
	Unit                   Unit    `json:"unit"`
	RecipeVariantId        *int    `json:"recipeVariantId,omitempty"`
	RecipeVariantName      string  `json:"recipeVariantName,omitempty"`
	RecipeVariantSku       string  `json:"recipeVariantSku,omitempty"`
	IngredientCost         float64 `json:"ingredientCost,omitempty"`
	IngredientStandardUnit *Unit   `json:"ingredientStandardUnit,omitempty"`
}
