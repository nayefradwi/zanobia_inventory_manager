package product

type RecipeBase struct {
	Id              *int    `json:"id"`
	ResultVariantId *int    `json:"resultVariantId"`
	Quantity        float64 `json:"quantity"`
	UnitId          *int    `json:"unitId"`
	RecipeVariantId *int    `json:"recipeVariantId"`
}

type Recipe struct {
	Id                *int    `json:"id"`
	ResultVariantId   *int    `json:"resultVariantId,omitempty"`
	ResultVariantName string  `json:"resultVariantName,omitempty"`
	ProductName       string  `json:"productName,omitempty"`
	Quantity          float64 `json:"quantity"`
	Unit              Unit    `json:"unit"`
	RecipeVariantId   *int    `json:"recipeVariantId,omitempty"`
	RecipeVariantName string  `json:"recipeVariantName,omitempty"`
	IngredientCost    float64 `json:"ingredientCost,omitempty"`
}

func GetTotalCost(recipes []Recipe) float64 {
	var totalCost float64
	for _, recipe := range recipes {
		totalCost += recipe.IngredientCost
	}
	return totalCost
}
