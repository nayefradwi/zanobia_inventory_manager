package product

import (
	"context"

	"github.com/nayefradwi/zanobia_inventory_manager/common"
)

type IProductService interface {
	CreateProduct(ctx context.Context, product ProductInput) error
	TranslateProduct(ctx context.Context, product ProductInput, languageCode string) error
	GetProducts(ctx context.Context, isArchive bool) (common.PaginatedResponse[Product], error)
	GetProduct(ctx context.Context, id int) (Product, error)
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

func (s *ProductService) GetProducts(ctx context.Context, isArchive bool) (common.PaginatedResponse[Product], error) {
	// 	size, cursor, _ := common.GetPaginationParams(ctx, "0")
	// 	products, err := s.repo.GetProducts(ctx, size, cursor, isArchive)
	// 	if err != nil {
	// 		return common.CreateEmptyPaginatedResponse[Product](size), err
	// 	}
	// 	if len(products) == 0 {
	// 		return common.CreateEmptyPaginatedResponse[Product](size), nil
	// 	}
	// 	lastId := products[len(products)-1].Id
	// 	return common.CreatePaginatedResponse[Product](size, strconv.Itoa(*lastId), products), nil
	return common.PaginatedResponse[Product]{}, nil
}

func (s *ProductService) GetProduct(ctx context.Context, id int) (Product, error) {
	// 	product, err := s.repo.GetProduct(ctx, id)
	// 	if err != nil {
	// 		return Product{}, err
	// 	}
	// 	recipe, err := s.recipeService.GetRecipeOfProduct(ctx, id)
	// 	if err != nil {
	// 		return Product{}, err
	// 	}
	// 	product.Recipe = recipe
	// 	return product, nil
	return Product{}, nil
}

// TODO get base product
