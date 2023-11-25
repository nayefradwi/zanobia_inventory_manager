package product

import (
	"context"
	"log"
	"time"

	"github.com/nayefradwi/zanobia_inventory_manager/common"
)

type IProductService interface {
	CreateProduct(ctx context.Context, product ProductInput) error
	TranslateProduct(ctx context.Context, product ProductInput, languageCode string) error
	GetProducts(ctx context.Context, isArchive bool, isIngredient bool) (common.PaginatedResponse[ProductBase], error)
	GetProduct(ctx context.Context, id int) (Product, error)
	GetProductVariant(ctx context.Context, productVariantId int) (ProductVariant, error)
	AddProductVariant(ctx context.Context, input ProductVariantInput) error
	GetUnitIdOfProductVariantBySku(ctx context.Context, sku string) (int, error)
	GetProductVariantExpirationDate(ctx context.Context, sku string) (time.Time, error)
	AddVariantOptionValue(ctx context.Context, input AddVariantValueInput) error
}

type ProductService struct {
	repo          IProductRepo
	recipeService IRecipeService
}

func NewProductService(repo IProductRepo, recipeService IRecipeService) IProductService {
	return &ProductService{
		repo,
		recipeService,
	}
}

func (s *ProductService) CreateProduct(ctx context.Context, product ProductInput) error {
	validationErr := ValidateProduct(product)
	if validationErr != nil {
		return validationErr
	}
	return s.repo.CreateProduct(ctx, product)
}

func (s *ProductService) TranslateProduct(ctx context.Context, product ProductInput, languageCode string) error {
	validationErr := ValidateProduct(product)
	if validationErr != nil {
		return validationErr
	}
	return s.repo.TranslateProduct(ctx, product, languageCode)
}

func (s *ProductService) GetProducts(ctx context.Context, isArchive bool, isIngredient bool) (common.PaginatedResponse[ProductBase], error) {
	paginationParams := common.GetPaginationParams(ctx)
	products, err := s.repo.GetProducts(ctx, paginationParams, isArchive, isIngredient)
	if err != nil {
		return common.CreateEmptyPaginatedResponse[ProductBase](paginationParams.PageSize), err
	}
	if len(products) == 0 {
		return common.CreateEmptyPaginatedResponse[ProductBase](paginationParams.PageSize), nil
	}
	first, last := products[0], products[len(products)-1]
	return common.CreatePaginatedResponse[ProductBase](
		paginationParams.PageSize,
		last,
		first,
		products,
	), nil
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
		TotalCost, err := s.recipeService.GetTotalCostOfRecipes(ctx, recipes)
		if err != nil {
			log.Printf("failed to get total cost of recipes: %s", err.Error())
		} else {
			productVariant.TotalCost = TotalCost
		}
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
	options, optionsErr := s.repo.GetProductOptions(ctx, *input.ProductVariant.ProductId)
	if optionsErr != nil {
		return optionsErr
	}
	if len(options) == 0 {
		return common.NewBadRequestFromMessage("product has no options")
	}
	validationErr := ValidateProductVariant(input, len(options), len(options))
	if validationErr != nil {
		return validationErr
	}
	productOptionValuesMap, valuesErr := s.repo.GetProductSelectedValues(ctx, *input.ProductVariant.ProductId, input.OptionValueIds)
	if valuesErr != nil {
		return valuesErr
	}
	optionValues := common.GetValues[string, ProductOptionValue](productOptionValuesMap)
	if len(optionValues) != len(options) {
		return common.NewValidationError("invalid product variant", common.ErrorDetails{
			Message: "invalid variant values",
			Field:   "variantValues",
		})
	}
	if input.ProductVariant.Sku == "" {
		sku, _ := common.GenerateUuid()
		input.ProductVariant.Sku = sku
	}
	input.ProductVariant.Name = GenerateName(optionValues)
	input.OptionValues = optionValues
	input.ProductVariant.IsDefault = false
	return s.repo.AddProductVariant(ctx, input)
}

func (s *ProductService) GetUnitIdOfProductVariantBySku(ctx context.Context, sku string) (int, error) {
	return s.repo.GetUnitIdOfProductVariantBySku(ctx, sku)
}

func (s *ProductService) GetProductVariantExpirationDate(ctx context.Context, sku string) (time.Time, error) {
	return s.repo.GetProductVariantExpirationDate(ctx, sku)
}

func (s *ProductService) AddVariantOptionValue(ctx context.Context, input AddVariantValueInput) error {
	if len(input.Value) < 1 || len(input.Value) > 50 {
		return common.NewValidationError("invalid product option value", common.ErrorDetails{
			Message: "option value must be between 1 and 50 characters",
			Field:   "options",
		})
	}
	_, err := s.repo.InsertProductOptionValue(ctx, input.ProductOptionId, input.ToProductOptionValue())
	return err
}
