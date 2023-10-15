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
	ProductVariantId   *int       `json:"productVariantId"`
	ProductVariantName string     `json:"productVariantName"`
	ProductName        string     `json:"productName"`
	Quantity           float64    `json:"quantity"`
	Unit               Unit       `json:"unit"`
	Ingredient         Ingredient `json:"ingredient"`
}
