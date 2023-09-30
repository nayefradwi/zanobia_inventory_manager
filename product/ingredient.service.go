package product

import (
	"context"

	"github.com/nayefradwi/zanobia_inventory_manager/common"
)

type IIngredientService interface {
	CreateIngredient(ctx context.Context, ingredientBase IngredientBase) error
	GetIngredients(ctx context.Context) (common.PaginatedResponse[Ingredient, int], error)
}

type IngredientService struct {
	repo   IIngredientRepository
	locker common.IDistributedLockingService
}

func NewIngredientService(repo IIngredientRepository, locker common.IDistributedLockingService) IIngredientService {
	return &IngredientService{
		repo:   repo,
		locker: locker,
	}
}

func (s *IngredientService) CreateIngredient(ctx context.Context, ingredientBase IngredientBase) error {
	validationErr := ValidateIngredient(ingredientBase)
	if validationErr != nil {
		return validationErr
	}
	return s.repo.CreateIngredient(ctx, ingredientBase)
}

func (s *IngredientService) TranslateIngredient(ctx context.Context, ingredient IngredientBase, languageCode string) error {
	return s.repo.TranslateIngredient(ctx, ingredient, languageCode)
}

func (s *IngredientService) GetIngredients(ctx context.Context) (common.PaginatedResponse[Ingredient, int], error) {
	pageSize, endCursor, _ := common.GetPaginationParams[int](ctx)
	ingredients, err := s.repo.GetIngredients(ctx, pageSize, endCursor)
	if err != nil {
		return common.PaginatedResponse[Ingredient, int]{}, err
	}
	if len(ingredients) == 0 {
		return common.CreateEmptyPaginatedResponse[Ingredient, int](pageSize), nil
	}
	last := ingredients[len(ingredients)-1]
	res := common.CreatePaginatedResponse[Ingredient](pageSize, *last.Id, ingredients)
	return res, nil
}
