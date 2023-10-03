package product

import (
	"context"

	"github.com/nayefradwi/zanobia_inventory_manager/common"
)

type IRecipeService interface {
	CreateRecipes(ctx context.Context, recipes []RecipeBase) error
	GetRecipeOfProduct(ctx context.Context, productId int) ([]Recipe, error)
	AddIngredientToRecipe(ctx context.Context, recipe RecipeBase) error
	DeleteRecipe(ctx context.Context, id int) error
}

type RecipeService struct {
	repo IRecipeRepository
}

func NewRecipeService(repo IRecipeRepository) IRecipeService {
	return &RecipeService{
		repo,
	}
}

func (s *RecipeService) CreateRecipes(ctx context.Context, recipes []RecipeBase) error {
	validationErr := ValidateRecipes(recipes)
	if validationErr != nil {
		return validationErr
	}
	if len(recipes) == 0 {
		return common.NewBadRequestFromMessage("cannot create empty recipes")
	}
	return s.repo.CreateRecipes(ctx, recipes)
}

func (s *RecipeService) AddIngredientToRecipe(ctx context.Context, recipe RecipeBase) error {
	validationErr := ValidateRecipe(recipe)
	if validationErr != nil {
		return validationErr
	}
	return s.repo.AddIngredientToRecipe(ctx, recipe)
}

func (s *RecipeService) GetRecipeOfProduct(ctx context.Context, productId int) ([]Recipe, error) {
	return s.repo.GetRecipeOfProduct(ctx, productId)
}

func (s *RecipeService) DeleteRecipe(ctx context.Context, id int) error {
	return s.repo.DeleteRecipe(ctx, id)
}
