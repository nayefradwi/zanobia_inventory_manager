package retailer

import (
	"github.com/jackc/pgx/v4/pgxpool"
)

type IRetailerBatchRepository interface{}

type RetailerBatchRepository struct {
	*pgxpool.Pool
}

func NewRetailerBatchRepository(db *pgxpool.Pool) *RetailerBatchRepository {
	return &RetailerBatchRepository{
		db,
	}
}
