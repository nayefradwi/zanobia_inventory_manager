package product

import (
	"context"
	"log"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
	"github.com/nayefradwi/zanobia_inventory_manager/translation"
	zimutils "github.com/nayefradwi/zanobia_inventory_manager/zim_utils"
)

type IIngredientRepository interface{}

type IngredientRepository struct {
	*pgxpool.Pool
}

func NewIngredientRepository(dbPool *pgxpool.Pool) IIngredientRepository {
	return &IngredientRepository{
		dbPool,
	}
}

func (r *IngredientRepository) CreateIngredient(ctx context.Context, ingredientBase IngredientBase) error {
	err := common.RunWithTransaction(ctx, r.Pool, func(ctx context.Context, tx pgx.Tx) error {
		id, addErr := r.addIngredient(ctx, tx, ingredientBase)
		if addErr != nil {
			return addErr
		}
		ingredientBase.Id = &id
		translationErr := r.insertTranslation(ctx, tx, ingredientBase, translation.DefaultLang)
		if translationErr != nil {
			return translationErr
		}
		return nil
	})
	return err
}

func (r *IngredientRepository) addIngredient(ctx context.Context, tx pgx.Tx, ingredient IngredientBase) (int, error) {
	sql := `INSERT INTO ingredients (price, standard_unit_id, expires_in_days) VALUES ($1, $2, $3) RETURNING id`
	var id int
	err := tx.QueryRow(ctx, sql, ingredient.Price, ingredient.StandardUnitId, ingredient.ExpiresInDays).Scan(&id)
	if err != nil {
		log.Printf("failed to create ingredient: %s", err.Error())
		return 0, common.NewInternalServerError()
	}
	return id, nil
}

func (r *IngredientRepository) insertTranslation(ctx context.Context, op common.DbOperator, ingredient IngredientBase, languageCode string) error {
	sql := `INSERT INTO ingredient_translations (ingredient_id, name, brand, language_code) VALUES ($1, $2, $3, $4)`
	c, err := op.Exec(ctx, sql, ingredient.Id, ingredient.Name, ingredient.Brand, languageCode)
	if err != nil {
		log.Printf("failed to translate ingredient: %s", err.Error())
		return common.NewBadRequestError("Failed to translate ingredient", zimutils.GetErrorCodeFromError(err))
	}
	if c.RowsAffected() == 0 {
		return common.NewInternalServerError()
	}
	return err

}

func (r *IngredientRepository) TranslateIngredient(ctx context.Context, ingredient IngredientBase, languageCode string) error {
	return r.insertTranslation(ctx, r.Pool, ingredient, languageCode)
}

// get ingredient by id
// get ingredients by name
// get all ingredients
