package product

import (
	"net/http"

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
	common.ParseBody[ProductBase](w, r.Body, func(product ProductBase) {
		err := c.service.CreateProduct(r.Context(), product)
		common.WriteCreatedResponse(common.EmptyResult{
			Error:   err,
			Writer:  w,
			Message: "Product created successfully",
		})
	})
}

func (c ProductController) TranslateProduct(w http.ResponseWriter, r *http.Request) {
	common.GetTranslatedBody[ProductBase](w, r.Body, func(t common.Translation[ProductBase]) {
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
	products, err := c.service.GetProducts(r.Context(), isArchive)
	common.WriteResponse(common.Result[common.PaginatedResponse[Product]]{
		Error:  err,
		Writer: w,
		Data:   products,
	})
}
