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
	DeleteRecipe(ctx context.Context, id int) error
	GetRecipeOfProductVariantSku(ctx context.Context, sku string) ([]Recipe, error)
	GetRecipesLookUpMapFromSkus(ctx context.Context, skuList []string) (map[string]Recipe, []string, error)
}
type RecipeRepository struct {
	*pgxpool.Pool
}

func NewRecipeRepository(dbPool *pgxpool.Pool) IRecipeRepository {
	return &RecipeRepository{dbPool}
}

func (r *RecipeRepository) CreateRecipes(ctx context.Context, recipes []RecipeBase) error {
	err := common.RunWithTransaction(ctx, r.Pool, func(ctx context.Context, tx pgx.Tx) error {
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
	sql := `INSERT INTO recipes (result_variant_sku, recipe_variant_sku, quantity, unit_id) VALUES ($1, $2, $3, $4)`
	op := common.GetOperator(ctx, r.Pool)
	_, err := op.Exec(ctx, sql, recipe.ResultVariantSku, recipe.RecipeVariantSku, recipe.Quantity, recipe.UnitId)
	if err != nil {
		log.Printf("failed to add ingredient to recipe: %s", err.Error())
		return common.NewBadRequestFromMessage("failed to add ingredient to recipe")
	}
	return nil
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

func (r *RecipeRepository) GetRecipeOfProductVariantSku(ctx context.Context, sku string) ([]Recipe, error) {
	sql := `
	SELECT
    	r.id,
    	r.quantity,
		pvartx_result.product_variant_id AS result_variant_id,
		r.result_variant_sku,
    	pvartx_result.name AS result_variant_name,
		pvartx_recipe.product_variant_id AS recipe_variant_id,
		r.recipe_variant_sku,
   		pvartx_recipe.name AS recipe_variant_name,
		pvar_recipe.price AS recipe_price,
    	utx.unit_id as recipe_unit_id,
    	utx.name as recipe_unit_name,
    	utx.symbol as recipe_unit_symbol,
		ptx.name as product_name,
		orig_utx.name as ingredient_standard_unit_name,
		orig_utx.symbol as ingredient_standard_unit_symbol,
		orig_utx.unit_id as ingredient_standard_unit_id
	FROM
    	recipes r
	JOIN unit_translations utx ON r.unit_id = utx.unit_id
	JOIN product_variants pvar_result ON pvar_result.sku = r.result_variant_sku
	JOIN product_variants pvar_recipe ON pvar_recipe.sku = r.recipe_variant_sku
	JOIN product_variant_translations pvartx_result ON pvartx_result.product_variant_id = pvar_result.id
	JOIN product_variant_translations pvartx_recipe ON pvartx_recipe.product_variant_id = pvar_recipe.id
	JOIN product_translations ptx ON ptx.id = pvar_recipe.product_id
	JOIN unit_translations orig_utx ON orig_utx.unit_id = pvar_recipe.standard_unit_id AND orig_utx.language_code = utx.language_code
	WHERE
		pvar_result.sku = $1
    AND 
		utx.language_code = $2;
	`
	op := common.GetOperator(ctx, r.Pool)
	languageCode := common.GetLanguageParam(ctx)
	rows, err := op.Query(ctx, sql, sku, languageCode)
	if err != nil {
		log.Printf("failed to get recipe of product: %s", err.Error())
		return nil, common.NewBadRequestFromMessage("failed to get recipe of product")
	}
	defer rows.Close()
	recipes := make([]Recipe, 0)
	for rows.Next() {
		var recipe Recipe
		var unit, recipeStandardUnit Unit
		var productName string
		err := rows.Scan(
			&recipe.Id, &recipe.Quantity, &recipe.ResultVariantId, &recipe.ResultVariantSku, &recipe.ResultVariantName,
			&recipe.RecipeVariantId, &recipe.RecipeVariantSku, &recipe.RecipeVariantName, &recipe.IngredientCost,
			&unit.Id, &unit.Name, &unit.Symbol, &productName,
			&recipeStandardUnit.Name, &recipeStandardUnit.Symbol, &recipeStandardUnit.Id,
		)
		if err != nil {
			log.Printf("failed to scan recipe: %s", err.Error())
			return nil, common.NewInternalServerError()
		}
		recipe.Unit = unit
		recipe.ProductName = productName
		recipe.IngredientStandardUnit = &recipeStandardUnit
		recipes = append(recipes, recipe)
	}
	return recipes, nil
}

func (r *RecipeRepository) GetRecipesLookUpMapFromSkus(ctx context.Context, skuList []string) (map[string]Recipe, []string, error) {
	sql := `
	SELECT recipes.id, result_variant_sku, recipe_variant_sku, 
	quantity, unit_id, standard_unit_id, price
	JOIN product_variants pv ON pv.sku = recipes.recipe_variant_sku
	FROM recipes WHERE result_variant_sku = ANY($1)
	`
	op := common.GetOperator(ctx, r.Pool)
	rows, err := op.Query(ctx, sql, skuList)
	if err != nil {
		log.Printf("failed to get recipes: %s", err.Error())
		return nil, nil, common.NewInternalServerError()
	}
	defer rows.Close()
	recipeMap := make(map[string]Recipe)
	recipeSkuList := make([]string, 0)
	for rows.Next() {
		var recipe Recipe
		var unitId, standardUnitID int
		err := rows.Scan(
			&recipe.Id,
			&recipe.ResultVariantSku,
			&recipe.RecipeVariantSku,
			&recipe.Quantity,
			&unitId,
			&standardUnitID,
			&recipe.IngredientCost,
		)
		recipe.Unit = Unit{Id: &unitId}
		recipe.IngredientStandardUnit = &Unit{Id: &standardUnitID}
		if err != nil {
			log.Printf("failed to scan recipe: %s", err.Error())
			return nil, nil, common.NewInternalServerError()
		}
		recipeMap[recipe.RecipeVariantSku] = recipe
		recipeSkuList = append(recipeSkuList, recipe.RecipeVariantSku)
	}
	return recipeMap, recipeSkuList, nil

}
