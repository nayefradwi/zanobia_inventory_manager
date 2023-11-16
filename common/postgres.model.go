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

type paginationParamsKey struct{}

type DbOperatorKey struct{}

type TransactionFunc func(ctx context.Context, tx pgx.Tx) error
type PaginatedResponse[T any] struct {
	PageSize       int    `json:"pageSize"`
	EndCursor      string `json:"endCursor"`
	PreviousCursor string `json:"previousCursor"`
	HasNext        bool   `json:"hasNext"`
	HasPrevious    bool   `json:"hasPrevious"`
	ItemsLength    int    `json:"itemsLength"`
	Items          []T    `json:"items"`
}

func CreatePaginatedResponse[T any](
	pageSize int,
	endCursor,
	previousCursor string,
	items []T,
) PaginatedResponse[T] {
	return PaginatedResponse[T]{
		PageSize:       pageSize,
		EndCursor:      endCursor,
		PreviousCursor: previousCursor,
		HasNext:        len(items) >= pageSize,
		HasPrevious:    len(items) >= pageSize && previousCursor != "",
		ItemsLength:    len(items),
		Items:          items,
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

type paginationQueryBuilder struct {
	query *PaginationQuery
}
type PaginationQuery struct {
	BaseSql        string
	Conditions     []string
	Direction      int
	EndCursor      string
	PreviousCursor string
	CursorKey      *string
	OrderByQuery   string
	PageSize       int
}

type PaginationParams struct {
	PageSize       int
	EndCursor      string
	PreviousCursor string
	Direction      int
}

func NewPaginationQueryBuilder(baseSql string, orderByQuery string) *paginationQueryBuilder {
	return &paginationQueryBuilder{
		query: &PaginationQuery{
			BaseSql:      baseSql,
			OrderByQuery: orderByQuery,
			PageSize:     10,
			Direction:    1,
		},
	}
}

func (b *paginationQueryBuilder) WithDirection(direction int) *paginationQueryBuilder {
	b.query.Direction = direction
	return b
}

func (b *paginationQueryBuilder) WithCursor(endCursor string, previousCursor string) *paginationQueryBuilder {
	b.query.EndCursor = endCursor
	b.query.PreviousCursor = previousCursor
	return b
}

func (b *paginationQueryBuilder) WithCursorKey(cursorKey string) *paginationQueryBuilder {
	b.query.CursorKey = &cursorKey
	return b
}

func (b *paginationQueryBuilder) WithConditions(conditions []string) *paginationQueryBuilder {
	b.query.Conditions = conditions
	return b
}

func (b *paginationQueryBuilder) WithPageSize(pageSize int) *paginationQueryBuilder {
	b.query.PageSize = pageSize
	return b
}

func (b *paginationQueryBuilder) GetCurrentCursor() string {
	if b.query.Direction < 0 {
		return b.query.PreviousCursor
	}
	return b.query.EndCursor
}

func (b *paginationQueryBuilder) Build() string {
	if b.query == nil {
		panic("invalid pagination query")
	}
	if b.query.BaseSql == "" || b.query.PageSize <= 0 ||
		b.query.CursorKey == nil {
		panic("invalid pagination query")
	}
	return CreatePaginationQuery(*b.query)
}
