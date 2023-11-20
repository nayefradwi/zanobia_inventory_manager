package product

import (
	"context"

	"github.com/nayefradwi/zanobia_inventory_manager/common"
)

type IIngredientService interface {
	CreateIngredient(ctx context.Context, ingredientBase IngredientBase) error
	GetIngredients(ctx context.Context) (common.PaginatedResponse[Ingredient], error)
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
	validationErr := ValidateIngredient(ingredient)
	if validationErr != nil {
		return validationErr
	}
	return s.repo.TranslateIngredient(ctx, ingredient, languageCode)
}

func (s *IngredientService) GetIngredients(ctx context.Context) (common.PaginatedResponse[Ingredient], error) {
	paginationParams := common.GetPaginationParams(ctx)
	ingredients, err := s.repo.GetIngredients(ctx, paginationParams)
	if err != nil {
		return common.PaginatedResponse[Ingredient]{}, err
	}
	if len(ingredients) == 0 {
		return common.CreateEmptyPaginatedResponse[Ingredient](paginationParams.PageSize), nil
	}
	first, last := ingredients[0], ingredients[len(ingredients)-1]
	res := common.CreatePaginatedResponse[Ingredient](
		paginationParams.PageSize,
		last,
		first,
		ingredients,
	)
	return res, nil
}
