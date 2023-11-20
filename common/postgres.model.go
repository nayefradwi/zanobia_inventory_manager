package common

import (
	"context"
	"fmt"
	"log"
	"strings"

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
type Cursorable interface {
	GetCursorValue() string
}

func CreatePaginatedResponse[T any](
	pageSize int,
	endCursor,
	previousCursor Cursorable,
	items []T,
) PaginatedResponse[T] {
	endCursorValue := endCursor.GetCursorValue()
	previousCursorValue := previousCursor.GetCursorValue()
	return PaginatedResponse[T]{
		PageSize:       pageSize,
		EndCursor:      endCursorValue,
		PreviousCursor: previousCursorValue,
		HasNext:        len(items) >= pageSize,
		HasPrevious:    len(items) >= pageSize && previousCursorValue != "",
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

type paginationQueryBuilder struct {
	query *PaginationQuery
}
type PaginationQuery struct {
	BaseSql        string
	Conditions     []string
	Direction      int
	EndCursor      string
	PreviousCursor string
	CursorKeys     []string
	OrderByQuery   []string
	PageSize       int
}

type PaginationParams struct {
	PageSize       int
	EndCursor      string
	PreviousCursor string
	Direction      int
}

func NewPaginationQueryBuilder(baseSql string, orderByQuery []string) *paginationQueryBuilder {
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

func (b *paginationQueryBuilder) WithCursorKeys(cursorKeys []string) *paginationQueryBuilder {
	b.query.CursorKeys = cursorKeys
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

func (b *paginationQueryBuilder) Build() (PaginationQuery, string) {
	if b.query == nil {
		panic("invalid pagination query")
	}
	if b.query.BaseSql == "" || b.query.PageSize <= 0 ||
		b.query.CursorKeys == nil || b.query.OrderByQuery == nil ||
		len(b.query.CursorKeys) != len(b.query.OrderByQuery) ||
		len(b.query.OrderByQuery) == 0 || len(b.query.CursorKeys) == 0 {
		panic("invalid pagination query")
	}
	return *b.query, CreatePaginationQuery(*b.query)
}

func (q PaginationQuery) GetCurrentCursor() []string {
	if q.Direction < 0 {
		return strings.Split(q.PreviousCursor, ",")
	}
	return strings.Split(q.EndCursor, ",")
}

func (q *PaginationQuery) getFinalArgAndJoinedConditions() (int, string) {
	finalArgIndex := 1
	for _, condition := range q.Conditions {
		if condition != "" &&
			condition != " " &&
			condition != "AND" &&
			condition != "OR" &&
			condition != "and" &&
			condition != "or" {
			finalArgIndex++
		}
	}
	var joinedConditions string
	if finalArgIndex != 1 {
		joinedConditions = strings.Join(q.Conditions, " ")
	}
	return finalArgIndex, joinedConditions
}

func (q *PaginationQuery) getFormatedPaginationConditionQuery(finalArgIndex int) string {
	// this will be like so AND (cursorKey1, cursorKey2, cursorKey3) > ($finalArgIndex1, $finalArgIndex2, $finalArgIndex3)
	unformattedPaginationConditonWithCursor := "AND %s %s %s"
	if q.EndCursor == "" && q.PreviousCursor == "" {
		return ""
	}
	cursorKeys := q.getCursorKeysJoined()
	argsForCursors := q.getArgsForCursorsJoined(finalArgIndex)
	return fmt.Sprintf(
		unformattedPaginationConditonWithCursor,
		cursorKeys,
		q.getDirectionString(),
		argsForCursors,
	)
}

func (q *PaginationQuery) getDirectionString() string {
	if q.Direction == 0 {
		return ">="
	} else if q.Direction > 0 {
		return ">"
	}
	return "<"
}

func (q *PaginationQuery) getFormattedOrderByQuery() string {
	return strings.Join(q.OrderByQuery, ", ")
}

func (q *PaginationQuery) getCursorKeysJoined() string {
	return "(" + strings.Join(q.CursorKeys, ", ") + " )"
}

func (q *PaginationQuery) getArgsForCursorsJoined(finalArgIndex int) string {
	return "(" + strings.Join(q.getArgsForCursors(finalArgIndex), ", ") + " )"
}

func (q *PaginationQuery) getArgsForCursors(finalArgIndex int) []string {
	args := make([]string, 0)
	for i := 0; i < len(q.CursorKeys); i++ {
		args = append(args, fmt.Sprintf("$%d", finalArgIndex+i))
	}
	return args
}

func (q PaginationQuery) Query(ctx context.Context, op DbOperator, sql string, arguments ...interface{}) (pgx.Rows, error) {
	// TODO: remove this or use logger
	log.Print(sql)
	if q.EndCursor == "" && q.PreviousCursor == "" {
		return op.Query(ctx, sql, arguments...)
	}
	cursors := q.GetCurrentCursor()
	if len(cursors) == 0 || cursors[0] == "" {
		return op.Query(ctx, sql, arguments...)
	}
	args := make([]interface{}, 0)
	args = append(args, arguments...)
	for _, cursor := range cursors {
		args = append(args, cursor)
	}
	return op.Query(ctx, sql, args...)
}
