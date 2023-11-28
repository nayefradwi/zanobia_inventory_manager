package product

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
)

type ProductController struct {
	service IProductService
}

func NewProductController(service IProductService) ProductController {
	return ProductController{
		service,
	}
}

func (c ProductController) CreateProduct(w http.ResponseWriter, r *http.Request) {
	common.ParseBody[ProductInput](w, r.Body, func(product ProductInput) {
		err := c.service.CreateProduct(r.Context(), product)
		common.WriteCreatedResponse(common.EmptyResult{
			Error:   err,
			Writer:  w,
			Message: "Product created successfully",
		})
	})
}

func (c ProductController) TranslateProduct(w http.ResponseWriter, r *http.Request) {
	common.GetTranslatedBody[ProductInput](w, r.Body, func(t common.Translation[ProductInput]) {
		err := c.service.TranslateProduct(r.Context(), t.Data, t.LanguageCode)
		common.WriteCreatedResponse(common.EmptyResult{
			Error:   err,
			Writer:  w,
			Message: "Product translated successfully",
		})
	})
}

func (c ProductController) GetProducts(w http.ResponseWriter, r *http.Request) {
	isArchive := r.URL.Query().Get("isArchive") == "true"
	isIngredient := r.URL.Query().Get("isIngredient") == "true"
	products, err := c.service.GetProducts(r.Context(), isArchive, isIngredient)
	common.WriteResponse(common.Result[common.PaginatedResponse[ProductBase]]{
		Error:  err,
		Writer: w,
		Data:   products,
	})
}

func (c ProductController) GetProduct(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, _ := strconv.Atoi(idParam)
	product, err := c.service.GetProduct(r.Context(), id)
	common.WriteResponse(common.Result[Product]{
		Error:  err,
		Writer: w,
		Data:   product,
	})
}

func (c ProductController) GetProductVariant(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, _ := strconv.Atoi(idParam)
	variant, err := c.service.GetProductVariant(r.Context(), id)
	common.WriteResponse(common.Result[ProductVariant]{
		Error:  err,
		Writer: w,
		Data:   variant,
	})
}

func (c ProductController) CreateProductVariant(w http.ResponseWriter, r *http.Request) {
	common.ParseBody[ProductVariantInput](w, r.Body, func(productVariant ProductVariantInput) {
		err := c.service.AddProductVariant(r.Context(), productVariant)
		common.WriteCreatedResponse(common.EmptyResult{
			Error:   err,
			Writer:  w,
			Message: "Product variant created successfully",
		})
	})
}

func (c ProductController) AddOptionValue(w http.ResponseWriter, r *http.Request) {
	common.ParseBody[AddVariantValueInput](w, r.Body, func(avvi AddVariantValueInput) {
		err := c.service.AddVariantOptionValue(r.Context(), avvi)
		common.WriteCreatedResponse(common.EmptyResult{
			Error:   err,
			Writer:  w,
			Message: "Product option value added successfully",
		})
	})
}

func (c ProductController) UpdateProductVariantDetails(w http.ResponseWriter, r *http.Request) {
	common.ParseBody[ProductVariantUpdate](w, r.Body, func(pvu ProductVariantUpdate) {
		err := c.service.UpdateProductVariantDetails(r.Context(), pvu)
		common.WriteCreatedResponse(common.EmptyResult{
			Error:   err,
			Writer:  w,
			Message: "Product variant updated successfully",
		})
	})
}

func (c ProductController) DeleteProductVariant(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, _ := strconv.Atoi(idParam)
	err := c.service.DeleteProductVariant(r.Context(), id)
	common.WriteEmptyResponse(
		common.EmptyResult{
			Error:   err,
			Writer:  w,
			Message: "Product variant deleted successfully",
		},
	)
}
