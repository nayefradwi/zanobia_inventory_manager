package common

import (
	"context"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type DbOperator interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, arguments ...interface{}) pgx.Row
	Query(ctx context.Context, sql string, arguments ...interface{}) (pgx.Rows, error)
}
type TransactionFunc func(ctx context.Context, tx pgx.Tx) error

func RunWithTransaction(ctx context.Context, pool *pgxpool.Pool, transaction TransactionFunc) error {
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return NewInternalServerError()
	}
	defer tx.Rollback(ctx)
	transactionErr := transaction(ctx, tx)
	if transactionErr != nil {
		if apiErr, ok := transactionErr.(*ApiError); ok {
			return apiErr
		}
		return NewInternalServerError()
	}
	tx.Commit(ctx)
	return nil
}
