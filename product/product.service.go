package product

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/nayefradwi/zanobia_inventory_manager/common"
)

type IProductService interface {
	CreateProduct(ctx context.Context, product ProductInput) error
	TranslateProduct(ctx context.Context, product ProductInput, languageCode string) error
	GetProducts(ctx context.Context, isArchive bool) (common.PaginatedResponse[ProductBase], error)
	GetProduct(ctx context.Context, id int) (Product, error)
	GetProductVariant(ctx context.Context, productVariantId int) (ProductVariant, error)
	AddProductVariant(ctx context.Context, input ProductVariantInput) error
	GetUnitIdOfProductVariantBySku(ctx context.Context, sku string) (int, error)
	GetProductVariantExpirationDate(ctx context.Context, sku string) (time.Time, error)
}

type ProductService struct {
	repo            IProductRepo
	recipeService   IRecipeService
	variantsService IVariantService
}

func NewProductService(repo IProductRepo, recipeService IRecipeService, variantService IVariantService) IProductService {
	return &ProductService{
		repo,
		recipeService,
		variantService,
	}
}

func (s *ProductService) CreateProduct(ctx context.Context, product ProductInput) error {
	validationErr := ValidateProduct(product)
	if validationErr != nil {
		return validationErr
	}
	variants, err := s.variantsService.GetVariantsFromListOfIds(ctx, product.Variants)
	if err != nil {
		return err
	}
	product.Variants = variants
	return s.repo.CreateProduct(ctx, product)
}

func (s *ProductService) TranslateProduct(ctx context.Context, product ProductInput, languageCode string) error {
	validationErr := ValidateProduct(product)
	if validationErr != nil {
		return validationErr
	}
	return s.repo.TranslateProduct(ctx, product, languageCode)
}

func (s *ProductService) GetProducts(ctx context.Context, isArchive bool) (common.PaginatedResponse[ProductBase], error) {
	size, cursor, _ := common.GetPaginationParams(ctx, "0")
	products, err := s.repo.GetProducts(ctx, size, cursor, isArchive)
	if err != nil {
		return common.CreateEmptyPaginatedResponse[ProductBase](size), err
	}
	if len(products) == 0 {
		return common.CreateEmptyPaginatedResponse[ProductBase](size), nil
	}
	lastId := products[len(products)-1].Id
	return common.CreatePaginatedResponse[ProductBase](size, strconv.Itoa(*lastId), products), nil
}

func (s *ProductService) GetProduct(ctx context.Context, id int) (Product, error) {
	product, err := s.repo.GetProduct(ctx, id)
	if err != nil {
		return Product{}, err
	}
	if product.Id == nil {
		return Product{}, common.NewNotFoundError("product not found")
	}
	productVariants, err := s.repo.GetProductVariantsOfProduct(ctx, *product.Id)
	if err != nil {
		return Product{}, err
	}
	product.ProductVariants = productVariants
	return product, nil
}

func (s *ProductService) GetProductVariant(ctx context.Context, productVariantId int) (ProductVariant, error) {
	productVariant, variantErr := s.repo.GetProductVariant(ctx, productVariantId)
	if variantErr != nil {
		return ProductVariant{}, variantErr
	}
	if productVariant.Id == nil {
		return ProductVariant{}, common.NewNotFoundError("product variant not found")
	}
	recipes, recipeErr := s.recipeService.GetRecipeOfProductVariant(ctx, *productVariant.Id)
	if recipeErr != nil {
		log.Printf("failed to get recipe of product variant: %s", recipeErr.Error())
	} else if len(recipes) > 0 {
		productVariant.Recipes = recipes
	}
	return productVariant, nil
}

func (s *ProductService) AddProductVariant(ctx context.Context, input ProductVariantInput) error {
	if input.ProductVariant.ProductId == nil {
		return common.NewValidationError("invalid product variant", common.ErrorDetails{
			Message: "product id cannot be empty",
			Field:   "productId",
		})
	}
	variants, variantErr := s.variantsService.GetProductOptions(ctx, *input.ProductVariant.ProductId)
	if variantErr != nil {
		return variantErr
	}
	validationErr := ValidateProductVariant(input, 1, len(variants))
	if validationErr != nil {
		return validationErr
	}
	variantValues, valuesErr := s.variantsService.GetProductSelectedValues(ctx, *input.ProductVariant.ProductId)
	if valuesErr != nil {
		return valuesErr
	}
	allAreProductSelectedValues := common.HasAllValues[VariantValue, int](
		input.VariantValues,
		variantValues,
		func(value VariantValue) int {
			return value.Id
		},
	)
	if !allAreProductSelectedValues {
		return common.NewValidationError("invalid product variant", common.ErrorDetails{
			Message: "invalid variant values",
			Field:   "variantValues",
		})
	}
	if input.ProductVariant.Sku == "" {
		sku, _ := common.GenerateUuid()
		input.ProductVariant.Sku = sku
	}
	input.ProductVariant.Name = GenerateName(input.VariantValues)
	return s.repo.AddProductVariant(ctx, input)
}

func (s *ProductService) GetUnitIdOfProductVariantBySku(ctx context.Context, sku string) (int, error) {
	return s.repo.GetUnitIdOfProductVariantBySku(ctx, sku)
}

func (s *ProductService) GetProductVariantExpirationDate(ctx context.Context, sku string) (time.Time, error) {
	return s.repo.GetProductVariantExpirationDate(ctx, sku)
}
