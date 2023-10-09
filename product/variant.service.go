package product

import "context"

type IVariantService interface {
	CreateVariant(ctx context.Context, variant Variant) error
	AddVariantValues(ctx context.Context, variantId int, values []string) error
}

type VariantService struct {
	repo IVariantRepository
}

func NewVariantService(repo IVariantRepository) IVariantService {
	return &VariantService{repo}
}

func (s *VariantService) CreateVariant(ctx context.Context, variant Variant) error {
	validationErr := ValidateVariant(variant)
	if validationErr != nil {
		return validationErr
	}
	return s.repo.CreateVariant(ctx, variant)
}

func (s *VariantService) AddVariantValues(ctx context.Context, variantId int, values []string) error {
	validationErr := ValidateVariantValues(values)
	if validationErr != nil {
		return validationErr
	}
	return s.repo.AddVariantValues(ctx, variantId, values)
}
