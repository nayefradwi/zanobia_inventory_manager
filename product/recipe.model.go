package product

type RecipeBase struct {
	Id               *int    `json:"id"`
	ProductVariantId *int    `json:"resultVariantId"`
	Quantity         float64 `json:"quantity"`
	UnitId           *int    `json:"unitId"`
	RecipeVariantId  *int    `json:"recipeVariantId"`
}

type Recipe struct {
	Id                 *int    `json:"id"`
	ProductVariantId   *int    `json:"resultVariantId,omitempty"`
	ProductVariantName string  `json:"resultVariantName,omitempty"`
	ProductName        string  `json:"productName,omitempty"`
	Quantity           float64 `json:"quantity"`
	Unit               Unit    `json:"unit"`
	RecipeVariantId    *int    `json:"recipeVariantId,omitempty"`
	RecipeVariantName  string  `json:"recipeVariantName,omitempty"`
}
