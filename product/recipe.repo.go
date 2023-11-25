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
	sql := `INSERT INTO recipes (result_variant_id, recipe_variant_id, quantity, unit_id) VALUES ($1, $2, $3, $4)`
	op := common.GetOperator(ctx, r.Pool)
	_, err := op.Exec(ctx, sql, recipe.ResultVariantId, recipe.RecipeVariantId, recipe.Quantity, recipe.UnitId)
	if err != nil {
		log.Printf("failed to add ingredient to recipe: %s", err.Error())
		return common.NewBadRequestFromMessage("failed to add ingredient to recipe")
	}
	return nil
}

func (r *RecipeRepository) GetRecipeOfProductVariant(ctx context.Context, resultVariantId int) ([]Recipe, error) {
	sql := `
	SELECT
    	r.id,
    	r.quantity,
    	pvartx_result.product_variant_id AS result_variant_id,
    	pvartx_result.name AS result_variant_name,
    	pvartx_recipe.product_variant_id AS recipe_variant_id,
   		pvartx_recipe.name AS recipe_variant_name,
    	pvar.price AS recipe_price,
    	utx.unit_id,
    	utx.name,
    	utx.symbol,
		ptx.name,
		orig_utx.name,
		orig_utx.symbol,
		orig_utx.unit_id
	FROM
    	recipes r
	JOIN unit_translations utx ON r.unit_id = utx.unit_id
	JOIN product_variant_translations pvartx_result ON pvartx_result.product_variant_id = r.result_variant_id
	JOIN product_variant_translations pvartx_recipe ON pvartx_recipe.product_variant_id = r.recipe_variant_id
	JOIN product_variants pvar ON pvar.id = r.recipe_variant_id
	JOIN product_translations ptx ON ptx.id = pvar.product_id
	JOIN unit_translations orig_utx ON orig_utx.unit_id = pvar.standard_unit_id AND orig_utx.language_code = utx.language_code
	WHERE
    	r.result_variant_id = $1
    AND 
		utx.language_code = $2;
	`
	op := common.GetOperator(ctx, r.Pool)
	languageCode := common.GetLanguageParam(ctx)
	rows, err := op.Query(ctx, sql, resultVariantId, languageCode)
	if err != nil {
		log.Printf("failed to get recipe of product: %s", err.Error())
		return nil, common.NewBadRequestFromMessage("failed to get recipe of product")
	}
	defer rows.Close()
	recipes := make([]Recipe, 0)
	var resultVariantName string
	for rows.Next() {
		var recipe Recipe
		var unit, variantUnit Unit
		var recipeVariantId int
		var recipeVariantName string
		var ingredientCost float64
		var productName string
		err := rows.Scan(
			&recipe.Id, &recipe.Quantity, &resultVariantId, &resultVariantName,
			&recipeVariantId, &recipeVariantName, &ingredientCost,
			&unit.Id, &unit.Name, &unit.Symbol, &productName,
			&variantUnit.Name, &variantUnit.Symbol, &variantUnit.Id,
		)
		if err != nil {
			log.Printf("failed to scan recipe: %s", err.Error())
			return nil, common.NewInternalServerError()
		}
		recipe.Unit = unit
		recipe.ResultVariantId = &resultVariantId
		recipe.ResultVariantName = resultVariantName
		recipe.RecipeVariantId = &recipeVariantId
		recipe.RecipeVariantName = recipeVariantName
		recipe.IngredientCost = ingredientCost
		recipe.ProductName = productName
		recipe.IngredientStandardUnit = &variantUnit
		recipes = append(recipes, recipe)
	}
	return recipes, nil
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
