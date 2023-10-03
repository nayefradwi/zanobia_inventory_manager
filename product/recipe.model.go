package product

type RecipeBase struct {
	Id           *int    `json:"id"`
	ProductId    *int    `json:"productId"`
	Quantity     float64 `json:"quantity"`
	UnitId       *int    `json:"unitId"`
	IngredientId *int    `json:"ingredientId"`
}

type Recipe struct {
	Id          *int       `json:"id"`
	ProductId   *int       `json:"productId"`
	ProductName string     `json:"productName"`
	Quantity    int        `json:"quantity"`
	Unit        Unit       `json:"unit"`
	Ingredient  Ingredient `json:"ingredient"`
}
