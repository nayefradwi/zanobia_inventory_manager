package transactions

import (
	"github.com/jackc/pgx/v4/pgxpool"
)

type ITransactionRepository interface{}

type TransactionRepository struct {
	*pgxpool.Pool
}

func NewTransactionRepository(db *pgxpool.Pool) *TransactionRepository {
	return &TransactionRepository{
		db,
	}
}
