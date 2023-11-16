package product

import (
	"context"
	"log"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
	zimutils "github.com/nayefradwi/zanobia_inventory_manager/zim_utils"
)

type IIngredientRepository interface {
	CreateIngredient(ctx context.Context, ingredientBase IngredientBase) error
	TranslateIngredient(ctx context.Context, ingredient IngredientBase, languageCode string) error
	GetIngredients(ctx context.Context, paginationParams common.PaginationParams) ([]Ingredient, error)
	GetUnitIdOfIngredient(ctx context.Context, ingredientId int) (int, error)
}

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
		ctx = common.SetOperator(ctx, tx)
		id, addErr := r.addIngredient(ctx, ingredientBase)
		if addErr != nil {
			return addErr
		}
		ingredientBase.Id = &id
		translationErr := r.insertTranslation(ctx, ingredientBase, common.DefaultLang)
		if translationErr != nil {
			return translationErr
		}
		return nil
	})
	return err
}

func (r *IngredientRepository) addIngredient(ctx context.Context, ingredient IngredientBase) (int, error) {
	sql := `INSERT INTO ingredients (price, standard_unit_id, expires_in_days, standard_quantity) VALUES ($1, $2, $3, $4) RETURNING id`
	var id int
	op := common.GetOperator(ctx, r.Pool)
	err := op.QueryRow(ctx, sql, ingredient.Price, ingredient.StandardUnitId, ingredient.ExpiresInDays, ingredient.StandardQty).Scan(&id)
	if err != nil {
		log.Printf("failed to create ingredient: %s", err.Error())
		return 0, common.NewBadRequestError("Failed to create ingredient", zimutils.GetErrorCodeFromError(err))
	}
	return id, nil
}

func (r *IngredientRepository) insertTranslation(ctx context.Context, ingredient IngredientBase, languageCode string) error {
	sql := `INSERT INTO ingredient_translations (ingredient_id, name, brand, language_code) VALUES ($1, $2, $3, $4)`
	op := common.GetOperator(ctx, r.Pool)
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
	return r.insertTranslation(ctx, ingredient, languageCode)
}

func (r *IngredientRepository) GetIngredients(ctx context.Context, paginationParams common.PaginationParams) ([]Ingredient, error) {
	sqlBuilder := common.NewPaginationQueryBuilder(
		`
		SELECT i.id, it.name, it.brand, i.price,
		i.expires_in_days, i.standard_quantity,
		ut.unit_id, ut.name, ut.symbol
		from ingredients i
		JOIN unit_translations ut on i.standard_unit_id = ut.unit_id
		JOIN ingredient_translations it ON it.ingredient_id = i.id AND it.language_code = ut.language_code
		`,
		"i.id ASC",
	)
	sql := sqlBuilder.
		WithConditions([]string{
			"it.language_code = $1",
		}).
		WithCursor(paginationParams.EndCursor, paginationParams.PreviousCursor).
		WithCursorKey(
			"i.id",
		).
		WithDirection(paginationParams.Direction).
		WithPageSize(paginationParams.PageSize).
		Build()
	languageCode := common.GetLanguageParam(ctx)
	/*

		" SELECT i.id, it.name,
		it.brand, i.price, i.expires_in_days, i.standard_quantity, ut.unit_id, ut.name, ut.symbol from ingredients i
		JOIN unit_translations ut on i.standard_unit_id = ut.unit_id JOIN ingredient_translations it ON
		it.ingredient_id = i.id AND it.language_code = ut.language_code  WHERE it.language_code = $1
		AND (i.id > $2 or $2 = $2) ORDER BY i.id ASC LIMIT 10;"
	*/
	rows, err := r.Query(ctx, sql, languageCode, sqlBuilder.GetCurrentCursor())
	if err != nil {
		log.Printf("failed to get ingredients: %s", err.Error())
		return nil, common.NewInternalServerError()
	}
	defer rows.Close()
	ingredients := make([]Ingredient, 0)
	for rows.Next() {
		unit := Unit{}
		ingredient := Ingredient{}
		err := rows.Scan(
			&ingredient.Id, &ingredient.Name,
			&ingredient.Brand, &ingredient.Price,
			&ingredient.ExpiresInDays, &ingredient.StandardQty,
			&unit.Id, &unit.Name, &unit.Symbol,
		)
		if err != nil {
			log.Printf("failed to get ingredients: %s", err.Error())
			return nil, common.NewInternalServerError()
		}
		ingredient.StandardUnit = &unit
		ingredients = append(ingredients, ingredient)
	}
	return ingredients, nil
}

func (r *IngredientRepository) GetUnitIdOfIngredient(ctx context.Context, ingredientId int) (int, error) {
	op := common.GetOperator(ctx, r.Pool)
	sql := `SELECT standard_unit_id FROM ingredients WHERE id = $1`
	var unitId int = -1
	err := op.QueryRow(ctx, sql, ingredientId).Scan(&unitId)
	if err != nil {
		log.Printf("failed to get unit id of ingredient: %s", err.Error())
		return 0, common.NewInternalServerError()
	}
	return unitId, nil
}

// get ingredient by id
// get ingredients by name
