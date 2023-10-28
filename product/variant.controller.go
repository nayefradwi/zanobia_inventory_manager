package product

import (
	"net/http"
)

type VariantController struct {
}

func NewVariantController() VariantController {
	return VariantController{}
}

func (c VariantController) CreateVariant(w http.ResponseWriter, r *http.Request) {
	// common.ParseBody[VariantInput](w, r.Body, func(v VariantInput) {
	// 	err := c.service.CreateVariant(r.Context(), v)
	// 	common.WriteCreatedResponse(common.EmptyResult{
	// 		Error:   err,
	// 		Writer:  w,
	// 		Message: "Variant created successfully",
	// 	})
	// })
}

func (c VariantController) AddVariantValues(w http.ResponseWriter, r *http.Request) {
	// variantIdParam := chi.URLParam(r, "id")
	// variantId, _ := strconv.Atoi(variantIdParam)
	// common.ParseBody[[]string](w, r.Body, func(values []string) {
	// 	err := c.service.AddVariantValues(r.Context(), variantId, values)
	// 	common.WriteCreatedResponse(common.EmptyResult{
	// 		Error:   err,
	// 		Writer:  w,
	// 		Message: "Variant values added successfully",
	// 	})
	// })
}

func (c VariantController) GetVariant(w http.ResponseWriter, r *http.Request) {
	// variantIdParam := chi.URLParam(r, "id")
	// variantId, _ := strconv.Atoi(variantIdParam)
	// variant, err := c.service.GetVariant(r.Context(), variantId)
	// common.WriteResponse(common.Result[Variant]{
	// 	Error:  err,
	// 	Writer: w,
	// 	Data:   variant,
	// })
}

func (c VariantController) UpdateVariantName(w http.ResponseWriter, r *http.Request) {
	// variantIdParam := chi.URLParam(r, "id")
	// variantId, _ := strconv.Atoi(variantIdParam)
	// common.ParseBody[VariantInput](w, r.Body, func(v VariantInput) {
	// 	err := c.service.UpdateVariantName(r.Context(), variantId, v.Name)
	// 	common.WriteEmptyResponse(common.EmptyResult{
	// 		Error:   err,
	// 		Writer:  w,
	// 		Message: "Variant name updated successfully",
	// 	})
	// })
}
