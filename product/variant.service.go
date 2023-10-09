package product

import (
	"context"

	"github.com/nayefradwi/zanobia_inventory_manager/common"
)

type IVariantService interface {
	CreateVariant(ctx context.Context, variant VariantInput) error
	AddVariantValues(ctx context.Context, variantId int, values []string) error
	GetVariant(ctx context.Context, variantId int) (Variant, error)
	UpdateVariantName(ctx context.Context, variantId int, newName string) error
}

type VariantService struct {
	repo IVariantRepository
}

func NewVariantService(repo IVariantRepository) IVariantService {
	return &VariantService{repo}
}

func (s *VariantService) CreateVariant(ctx context.Context, variant VariantInput) error {
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

func (s *VariantService) GetVariant(ctx context.Context, variantId int) (Variant, error) {
	variant, err := s.repo.GetVariant(ctx, variantId)
	if err != nil {
		return Variant{}, err
	}
	if variant.Id == nil {
		return Variant{}, common.NewNotFoundError("variant not found")
	}
	return variant, nil
}

func (s *VariantService) UpdateVariantName(ctx context.Context, variantId int, newName string) error {
	validationErr := common.ValidateAlphanuemericName(newName, "name")
	if validationErr.Message != "" {
		return common.NewValidationError("invalid name", validationErr)
	}
	return s.repo.UpdateVariantName(ctx, variantId, newName)
}
