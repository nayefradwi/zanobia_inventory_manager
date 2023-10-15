package product

type RecipeBase struct {
	Id               *int    `json:"id"`
	ProductVariantId *int    `json:"productVariantId"`
	Quantity         float64 `json:"quantity"`
	UnitId           *int    `json:"unitId"`
	IngredientId     *int    `json:"ingredientId"`
}

type Recipe struct {
	Id                 *int       `json:"id"`
	ProductVariantId   *int       `json:"productVariantId,omitempty"`
	ProductVariantName string     `json:"productVariantName,omitempty"`
	ProductName        string     `json:"productName,omitempty"`
	Quantity           float64    `json:"quantity"`
	Unit               Unit       `json:"unit"`
	Ingredient         Ingredient `json:"ingredient"`
}
