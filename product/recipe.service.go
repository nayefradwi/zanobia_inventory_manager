package product

import (
	"context"
	"log"

	"github.com/nayefradwi/zanobia_inventory_manager/common"
)

type IRecipeService interface {
	CreateRecipes(ctx context.Context, recipes []RecipeBase) error
	AddIngredientToRecipe(ctx context.Context, recipe RecipeBase) error
	DeleteRecipe(ctx context.Context, id int) error
	GetTotalCostOfRecipes(ctx context.Context, recipes []Recipe) (float64, error)
	GetRecipeOfProductVariantSku(ctx context.Context, sku string) ([]Recipe, error)
	GetRecipesLookUpMapFromSkus(ctx context.Context, skuList []string) (map[string]Recipe, []string, error)
}

type RecipeService struct {
	repo        IRecipeRepository
	unitService IUnitService
}

func NewRecipeService(repo IRecipeRepository, unitService IUnitService) IRecipeService {
	return &RecipeService{
		repo,
		unitService,
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

func (s *RecipeService) DeleteRecipe(ctx context.Context, id int) error {
	return s.repo.DeleteRecipe(ctx, id)
}
func (s *RecipeService) GetTotalCostOfRecipes(ctx context.Context, recipes []Recipe) (float64, error) {
	var totalCost float64
	for _, recipe := range recipes {
		cost, err := s.getCostOfRecipe(ctx, recipe)
		if err != nil {
			return 0, err
		}
		totalCost += cost
	}
	return totalCost, nil
}

func (s *RecipeService) getCostOfRecipe(ctx context.Context, recipe Recipe) (float64, error) {
	if recipe.IngredientStandardUnit == nil {
		return 0, common.NewBadRequestFromMessage("ingredient standard unit cannot be empty")
	}
	if recipe.Unit.Id == nil {
		return 0, common.NewBadRequestFromMessage("unit id cannot be empty")
	}
	if *recipe.Unit.Id == *recipe.IngredientStandardUnit.Id {
		return recipe.Quantity * recipe.IngredientCost, nil
	}
	// I need a way to bulk convert units
	newQty, err := s.unitService.ConvertUnit(ctx, ConvertUnitInput{
		FromUnitId: recipe.Unit.Id,
		ToUnitId:   recipe.IngredientStandardUnit.Id,
		Quantity:   recipe.Quantity,
	})
	if err != nil {
		log.Printf("failed to convert unit: %s", err.Error())
		return 0, err
	}
	return newQty.Quantity * recipe.IngredientCost, nil
}

func (s *RecipeService) GetRecipeOfProductVariantSku(ctx context.Context, sku string) ([]Recipe, error) {
	return s.repo.GetRecipeOfProductVariantSku(ctx, sku)
}

func (s *RecipeService) GetRecipesLookUpMapFromSkus(ctx context.Context, skuList []string) (map[string]Recipe, []string, error) {
	// TODO: fill
	return map[string]Recipe{}, []string{}, common.NewBadRequestFromMessage("not implemented: getBatchBasesFromIds")
}
