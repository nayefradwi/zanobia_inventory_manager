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

func (c RecipeController) AddIngredientToRecipe(w http.ResponseWriter, r *http.Request) {
	common.ParseBody[RecipeBase](w, r.Body, func(recipe RecipeBase) {
		err := c.service.AddIngredientToRecipe(r.Context(), recipe)
		common.WriteCreatedResponse(common.EmptyResult{
			Error:   err,
			Writer:  w,
			Message: "Ingredient added to recipe successfully",
		})
	})
}

func (c RecipeController) DeleteRecipe(w http.ResponseWriter, r *http.Request) {
	id := common.GetIntURLParam(r, "id")
	err := c.service.DeleteRecipe(r.Context(), id)
	common.WriteEmptyResponse(common.EmptyResult{
		Error:   err,
		Writer:  w,
		Message: "Recipe deleted successfully",
	})
}
