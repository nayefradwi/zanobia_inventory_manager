package product

import "net/http"

type IngredientController struct {
	service IIngredientService
}

func NewIngredientController(service IIngredientService) IngredientController {
	return IngredientController{
		service: service,
	}
}

func (c IngredientController) GetIngredients(w http.ResponseWriter, r *http.Request) {

}
