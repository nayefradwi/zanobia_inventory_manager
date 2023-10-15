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
	GetRecipeOfProductVariant(ctx context.Context, productVariantId int) ([]Recipe, error)
	DeleteRecipe(ctx context.Context, id int) error
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
	sql := `INSERT INTO recipes (product_variant_id, ingredient_id, quantity, unit_id) VALUES ($1, $2, $3, $4)`
	op := common.GetOperator(ctx, r.Pool)
	_, err := op.Exec(ctx, sql, recipe.ProductVariantId, recipe.IngredientId, recipe.Quantity, recipe.UnitId)
	if err != nil {
		log.Printf("failed to add ingredient to recipe: %s", err.Error())
		return common.NewBadRequestFromMessage("failed to add ingredient to recipe")
	}
	return nil
}

func (r *RecipeRepository) GetRecipeOfProductVariant(ctx context.Context, productId int) ([]Recipe, error) {
	// sql := `select r.id, r.product_id, ptx.name, r.quantity, utx.unit_id, utx.name,
	// utx.symbol, ingtx.ingredient_id, ingtx.name, ingtx.brand, ing.price,
	// ing.expires_in_days from recipes r
	// join product_translations ptx on r.product_id = ptx.product_id
	// join unit_translations utx on r.unit_id = utx.unit_id and ptx.language_code = utx.language_code
	// join ingredients ing on r.ingredient_id = ing.id
	// join ingredient_translations ingtx on ingtx.ingredient_id = r.ingredient_id and utx.language_code = ingtx.language_code
	// where ptx.product_id = $1 and ptx.language_code = $2;`
	// op := common.GetOperator(ctx, r.Pool)
	// languageCode := common.GetLanguageParam(ctx)
	// rows, err := op.Query(ctx, sql, productId, languageCode)
	// if err != nil {
	// 	log.Printf("failed to get recipe of product: %s", err.Error())
	// 	return nil, common.NewBadRequestFromMessage("failed to get recipe of product")
	// }
	// defer rows.Close()
	// recipes := make([]Recipe, 0)
	// for rows.Next() {
	// 	var recipe Recipe
	// 	var unit Unit
	// 	var ingredient Ingredient
	// 	err := rows.Scan(&recipe.Id, &recipe.ProductId, &recipe.ProductName, &recipe.Quantity,
	// 		&unit.Id, &unit.Name, &unit.Symbol, &ingredient.Id, &ingredient.Name,
	// 		&ingredient.Brand, &ingredient.Price, &ingredient.ExpiresInDays,
	// 	)
	// 	if err != nil {
	// 		log.Printf("failed to scan recipe: %s", err.Error())
	// 		return nil, common.NewInternalServerError()
	// 	}
	// 	recipe.Unit = unit
	// 	recipe.Ingredient = ingredient
	// 	recipes = append(recipes, recipe)
	// }
	// return recipes, nil
	return []Recipe{}, nil
}

func (r *RecipeRepository) DeleteRecipe(ctx context.Context, id int) error {
	sql := `DELETE FROM recipes WHERE id = $1`
	op := common.GetOperator(ctx, r.Pool)
	_, err := op.Exec(ctx, sql, id)
	if err != nil {
		log.Printf("failed to delete recipe: %s", err.Error())
		return common.NewBadRequestFromMessage("failed to delete recipe")
	}
	return nil
}
