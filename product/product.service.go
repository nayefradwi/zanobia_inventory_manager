package product

import (
	"context"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
	"go.uber.org/zap"
)

type IProductService interface {
	CreateProduct(ctx context.Context, product ProductInput) error
	TranslateProduct(ctx context.Context, product ProductInput, languageCode string) error
	GetProducts(ctx context.Context, isArchive bool, isIngredient bool) (common.PaginatedResponse[ProductBase], error)
	GetProduct(ctx context.Context, id int) (Product, error)
	GetProductVariant(ctx context.Context, productVariantId int) (ProductVariant, error)
	AddProductVariant(ctx context.Context, input ProductVariantInput) error
	GetUnitIdOfProductVariantBySku(ctx context.Context, sku string) (int, error)
	GetProductVariantExpirationDateAndCost(ctx context.Context, sku string) (time.Time, float64, error)
	AddVariantOptionValue(ctx context.Context, input AddVariantValueInput) error
	UpdateProductVariantDetails(ctx context.Context, update ProductVariantUpdate) error
	DeleteProductVariant(ctx context.Context, id int) error
	UpdateProductVariantSku(ctx context.Context, input UpdateSkuInput) error
	GetOriginalUnitsBySkuList(ctx context.Context, skuList []string) (map[string]int, error)
	DeleteProduct(ctx context.Context, id int) error
	UpdateProductArchiveStatus(ctx context.Context, id int, isArchive bool) error
	UpdateProductVariantArchiveStatus(ctx context.Context, id int, isArchive bool) error
	SearchProductVariantByName(ctx context.Context, name string) (common.PaginatedResponse[ProductVariant], error)
	GetProductVariantBySku(ctx context.Context, sku string, getRecipe bool) (ProductVariant, error)
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
	recipes, totalCost := s.getRecipeOfProductVariant(ctx, productVariant.Sku)
	productVariant.Recipes = recipes
	productVariant.TotalCost = totalCost
	return productVariant, nil
}

func (s *ProductService) getRecipeOfProductVariant(ctx context.Context, sku string) ([]Recipe, float64) {
	var totalCost float64
	recipes, recipeErr := s.recipeService.GetRecipeOfProductVariantSku(ctx, sku)
	if recipeErr != nil {
		common.LoggerFromCtx(ctx).Error("failed to get recipe of product variant", zap.Error(recipeErr))
	} else if len(recipes) > 0 {
		totalCost, recipeErr = s.recipeService.GetTotalCostOfRecipes(ctx, recipes)
		if recipeErr != nil {
			common.LoggerFromCtx(ctx).Error("failed to get total cost of recipes", zap.Error(recipeErr))
		}
	}
	return recipes, totalCost
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

func (s *ProductService) GetProductVariantExpirationDateAndCost(ctx context.Context, sku string) (time.Time, float64, error) {
	return s.repo.GetProductVariantExpirationDateAndCost(ctx, sku)
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

func (s *ProductService) UpdateProductVariantDetails(ctx context.Context, update ProductVariantUpdate) error {
	if update.Id == 0 {
		return common.NewValidationError("invalid product variant", common.ErrorDetails{
			Message: "product variant id cannot be empty",
			Field:   "productVariantId",
		})
	}
	if update.Price < 0 {
		return common.NewValidationError("invalid product variant", common.ErrorDetails{
			Message: "price cannot be negative",
			Field:   "price",
		})
	}
	return s.repo.UpdateProductVariantDetails(ctx, update)
}

func (s *ProductService) DeleteProductVariant(ctx context.Context, id int) error {
	return common.RunWithTransaction(ctx, s.repo.(*ProductRepo).Pool, func(ctx context.Context, tx pgx.Tx) error {
		sku, isDefault, err := s.repo.(*ProductRepo).GetProductVariantSkuAndIsDefaultFromId(ctx, id)
		if err != nil {
			return err
		}
		if sku == "" {
			return common.NewNotFoundError("product variant not found")
		}
		if isDefault {
			return common.NewBadRequestFromMessage("cannot delete default variant")
		}
		if err := s.repo.DeleteProductVariant(ctx, id, sku); err != nil {
			return err
		}
		return nil
	})
}

func (s *ProductService) UpdateProductVariantSku(ctx context.Context, input UpdateSkuInput) error {
	if details := common.ValidateStringLength(input.NewSku, "sku", 10, 36); details.Message != "" {
		return common.NewValidationError("invalid sku", details)
	}
	return s.repo.UpdateProductVariantSku(ctx, input.OldSku, input.NewSku)
}

func (s *ProductService) GetOriginalUnitsBySkuList(ctx context.Context, skuList []string) (map[string]int, error) {
	return s.repo.GetOriginalUnitsBySkuList(ctx, skuList)
}

func (s *ProductService) DeleteProduct(ctx context.Context, id int) error {
	return common.RunWithTransaction(ctx, s.repo.(*ProductRepo).Pool, func(ctx context.Context, tx pgx.Tx) error {
		product, err := s.GetProduct(ctx, id)
		if err != nil {
			return err
		}
		if err := s.repo.DeleteProduct(ctx, product); err != nil {
			return err
		}
		return nil
	})
}

func (s *ProductService) UpdateProductArchiveStatus(ctx context.Context, id int, isArchive bool) error {
	return common.RunWithTransaction(ctx, s.repo.(*ProductRepo).Pool, func(ctx context.Context, tx pgx.Tx) error {
		if err := s.repo.UpdateProductArchiveStatus(ctx, id, isArchive); err != nil {
			return err
		}
		return nil
	})
}

func (s *ProductService) UpdateProductVariantArchiveStatus(ctx context.Context, id int, isArchive bool) error {
	_, isDefault, err := s.repo.GetProductVariantSkuAndIsDefaultFromId(ctx, id)
	if err != nil {
		return err
	}
	if isDefault {
		return common.NewBadRequestFromMessage("cannot archive default variant")
	}
	return s.repo.UpdateProductVariantArchiveStatus(ctx, id, isArchive)
}

func (s *ProductService) SearchProductVariantByName(
	ctx context.Context,
	name string,
) (common.PaginatedResponse[ProductVariant], error) {
	paginationParams := common.GetPaginationParams(ctx)
	productVariants, err := s.repo.SearchProductVariantsByName(ctx, paginationParams, name)
	if err != nil {
		return common.CreateEmptyPaginatedResponse[ProductVariant](paginationParams.PageSize), err
	}
	if len(productVariants) == 0 {
		return common.CreateEmptyPaginatedResponse[ProductVariant](paginationParams.PageSize), nil
	}
	first, last := productVariants[0], productVariants[len(productVariants)-1]
	return common.CreatePaginatedResponse[ProductVariant](
		paginationParams.PageSize,
		last,
		first,
		productVariants,
	), nil
}

func (s *ProductService) GetProductVariantBySku(ctx context.Context, sku string, withRecipe bool) (ProductVariant, error) {
	productVariant, err := s.repo.GetProductVariantBySku(ctx, sku)
	if err != nil {
		return ProductVariant{}, err
	}
	if productVariant.Id == nil {
		return ProductVariant{}, common.NewNotFoundError("product variant not found")
	}
	if withRecipe {
		recipes, totalCost := s.getRecipeOfProductVariant(ctx, sku)
		productVariant.Recipes = recipes
		productVariant.TotalCost = totalCost
	}
	return productVariant, nil
}
