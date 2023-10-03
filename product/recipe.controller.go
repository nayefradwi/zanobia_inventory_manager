package product

import (
	"net/http"

	"github.com/nayefradwi/zanobia_inventory_manager/common"
)

type RecipeController struct {
	service IRecipeService
}

func NewRecipeController(service IRecipeService) RecipeController {
	return RecipeController{
		service,
	}
}

func (c RecipeController) CreateRecipe(w http.ResponseWriter, r *http.Request) {
	common.ParseBody[[]RecipeBase](w, r.Body, func(recipes []RecipeBase) {
		err := c.service.CreateRecipes(r.Context(), recipes)
		common.WriteCreatedResponse(common.EmptyResult{
			Error:   err,
			Writer:  w,
			Message: "Recipe created successfully",
		})
	})
}
