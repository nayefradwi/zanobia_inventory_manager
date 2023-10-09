package product

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
)

type VariantController struct {
	service IVariantService
}

func NewVariantController(service IVariantService) VariantController {
	return VariantController{service}
}

func (c VariantController) CreateVariant(w http.ResponseWriter, r *http.Request) {
	common.ParseBody[Variant](w, r.Body, func(v Variant) {
		err := c.service.CreateVariant(r.Context(), v)
		common.WriteCreatedResponse(common.EmptyResult{
			Error:   err,
			Writer:  w,
			Message: "Variant created successfully",
		})
	})
}

func (c VariantController) AddVariantValues(w http.ResponseWriter, r *http.Request) {
	variantIdParam := chi.URLParam(r, "id")
	variantId, _ := strconv.Atoi(variantIdParam)
	common.ParseBody[[]string](w, r.Body, func(values []string) {
		err := c.service.AddVariantValues(r.Context(), variantId, values)
		common.WriteCreatedResponse(common.EmptyResult{
			Error:   err,
			Writer:  w,
			Message: "Variant values added successfully",
		})
	})
}
