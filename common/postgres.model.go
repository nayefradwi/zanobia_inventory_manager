package common

import (
	"context"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

type DbOperator interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, arguments ...interface{}) pgx.Row
	Query(ctx context.Context, sql string, arguments ...interface{}) (pgx.Rows, error)
}

type pageSizeKey struct{}
type pageIndexKey struct{}

type TransactionFunc func(ctx context.Context, tx pgx.Tx) error
type PaginatedResponse[T any] struct {
	PageSize  int  `json:"pageSize"`
	PageIndex int  `json:"pageIndex"`
	Total     int  `json:"total"`
	HasNext   bool `json:"hasNext"`
	Items     []T  `json:"items"`
}

func CreatePaginatedResponse[T any](pageSize int, pageIndex int, total int, items []T) PaginatedResponse[T] {
	return PaginatedResponse[T]{
		PageSize:  pageSize,
		PageIndex: pageIndex,
		Total:     total,
		HasNext:   (pageIndex+1)*pageSize < total,
		Items:     items,
	}
}
