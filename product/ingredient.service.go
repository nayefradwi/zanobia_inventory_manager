package product

import "context"

type IIngredientService interface {
	CreateIngredient(ctx context.Context, ingredientBase IngredientBase) error
}

type IngredientService struct {
	repo IIngredientRepository
}

func NewIngredientService(repo IIngredientRepository) IIngredientService {
	return &IngredientService{
		repo,
	}
}

func (s *IngredientService) CreateIngredient(ctx context.Context, ingredientBase IngredientBase) error {
	validationErr := ValidateIngredient(ingredientBase)
	if validationErr != nil {
		return validationErr
	}
	return s.repo.CreateIngredient(ctx, ingredientBase)
}
