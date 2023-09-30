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
type endCursorKey struct{}
type sortKey struct{}

type DbOperatorKey struct{}

type TransactionFunc func(ctx context.Context, tx pgx.Tx) error
type PaginatedResponse[T any] struct {
	PageSize    int    `json:"pageSize"`
	EndCursor   string `json:"endCursor"`
	HasNext     bool   `json:"hasNext"`
	ItemsLength int    `json:"itemsLength"`
	Items       []T    `json:"items"`
}

func CreatePaginatedResponse[T any](pageSize int, endCursor string, items []T) PaginatedResponse[T] {
	return PaginatedResponse[T]{
		PageSize:    pageSize,
		EndCursor:   endCursor,
		HasNext:     len(items) >= pageSize,
		ItemsLength: len(items),
		Items:       items,
	}
}

func CreateEmptyPaginatedResponse[T any](pageSize int) PaginatedResponse[T] {
	return PaginatedResponse[T]{
		PageSize:    pageSize,
		HasNext:     false,
		ItemsLength: 0,
		Items:       []T{},
	}
}

func GetSortingQuery(sort int) string {
	if sort < 0 {
		return "DESC"
	}
	return "ASC"
}
