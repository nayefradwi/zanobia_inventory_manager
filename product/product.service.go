package product

import "context"

type IProductService interface {
	CreateProduct(ctx context.Context, product ProductBase) error
	TranslateProduct(ctx context.Context, product ProductBase, languageCode string) error
}

type ProductService struct {
	repo IProductRepo
}

func NewProductService(repo IProductRepo) IProductService {
	return &ProductService{
		repo,
	}
}

func (s *ProductService) CreateProduct(ctx context.Context, product ProductBase) error {
	validationErr := ValidateProduct(product)
	if validationErr != nil {
		return validationErr
	}
	return s.repo.CreateProduct(ctx, product)
}

func (s *ProductService) TranslateProduct(ctx context.Context, product ProductBase, languageCode string) error {
	validationErr := ValidateProduct(product)
	if validationErr != nil {
		return validationErr
	}
	return s.repo.TranslateProduct(ctx, product, languageCode)
}
