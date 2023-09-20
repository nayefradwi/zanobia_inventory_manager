package product

import (
	"net/http"

	"github.com/nayefradwi/zanobia_inventory_manager/common"
)

type IngredientController struct {
	service IIngredientService
}

func NewIngredientController(service IIngredientService) IngredientController {
	return IngredientController{
		service: service,
	}
}

func (c IngredientController) CreateIngredient(w http.ResponseWriter, r *http.Request) {
	common.ParseBody[IngredientBase](w, r.Body, func(ingredient IngredientBase) {
		err := c.service.CreateIngredient(r.Context(), ingredient)
		common.WriteCreatedResponse(common.EmptyResult{
			Error:   err,
			Writer:  w,
			Message: "Ingredient created successfully",
		})
	})
}

func (c IngredientController) GetIngredients(w http.ResponseWriter, r *http.Request) {
	ingredients, err := c.service.GetIngredients(r.Context())
	common.WriteResponse[common.PaginatedResponse[Ingredient]](
		common.Result[common.PaginatedResponse[Ingredient]]{
			Error:  err,
			Writer: w,
			Data:   ingredients,
		},
	)
}
