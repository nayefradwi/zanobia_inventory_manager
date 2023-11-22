package retailer

import (
	"github.com/jackc/pgx/v4/pgxpool"
)

type IRetailerRepository interface{}

type RetailerRepo struct {
	*pgxpool.Pool
}

func NewRetailerRepository(db *pgxpool.Pool) *RetailerRepo {
	return &RetailerRepo{
		db,
	}
}
