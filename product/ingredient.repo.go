package product

import "github.com/jackc/pgx/v4/pgxpool"

type IIngredientRepository interface{}

type IngredientRepository struct {
	*pgxpool.Pool
}

func NewIngredientRepository(dbPool *pgxpool.Pool) IIngredientRepository {
	return &IngredientRepository{
		dbPool,
	}
}
