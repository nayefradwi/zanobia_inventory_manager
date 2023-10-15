package product

import (
	"context"
	"strconv"

	"github.com/nayefradwi/zanobia_inventory_manager/common"
)

type IProductService interface {
	CreateProduct(ctx context.Context, product ProductInput) error
	TranslateProduct(ctx context.Context, product ProductInput, languageCode string) error
	GetProducts(ctx context.Context, isArchive bool) (common.PaginatedResponse[ProductBase], error)
	GetProduct(ctx context.Context, id int) (Product, error)
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
