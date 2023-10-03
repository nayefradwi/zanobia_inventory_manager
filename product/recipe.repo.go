package product

import (
	"context"
	"log"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
)

type IRecipeRepository interface {
	CreateRecipes(ctx context.Context, recipes []RecipeBase) error
	AddIngredientToRecipe(ctx context.Context, recipe RecipeBase) error
}

type RecipeRepository struct {
	*pgxpool.Pool
}

func NewRecipeRepository(dbPool *pgxpool.Pool) IRecipeRepository {
	return &RecipeRepository{dbPool}
}

func (r *RecipeRepository) CreateRecipes(ctx context.Context, recipes []RecipeBase) error {
	err := common.RunWithTransaction(ctx, r.Pool, func(ctx context.Context, tx pgx.Tx) error {
		ctx = common.SetOperator(ctx, tx)
		for _, recipe := range recipes {
			err := r.AddIngredientToRecipe(ctx, recipe)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func (r *RecipeRepository) AddIngredientToRecipe(ctx context.Context, recipe RecipeBase) error {
	sql := `INSERT INTO recipes (product_id, ingredient_id, quantity, unit_id) VALUES ($1, $2, $3, $4)`
	op := common.GetOperator(ctx, r.Pool)
	_, err := op.Exec(ctx, sql, recipe.ProductId, recipe.IngredientId, recipe.Quantity, recipe.UnitId)
	if err != nil {
		log.Printf("failed to add ingredient to recipe: %s", err.Error())
		return common.NewBadRequestFromMessage("failed to add ingredient to recipe")
	}
	return nil
}

func (r *RecipeRepository) RemoveIngredientFromRecipe(ctx context.Context, id, ingredientId int) error {
	return nil
}

func (r *RecipeRepository) GetRecipeOfProduct(ctx context.Context, productId int) ([]Recipe, error) {
	return nil, nil
}
