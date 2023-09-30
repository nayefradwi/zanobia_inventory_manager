package common

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

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

func SetPaginatedDataMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pageSize, endCursor, sort := getPaginationParams(r)
		ctx := r.Context()
		ctx = context.WithValue(ctx, pageSizeKey{}, pageSize)
		ctx = context.WithValue(ctx, endCursorKey{}, endCursor)
		ctx = context.WithValue(ctx, sortKey{}, sort)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getPaginationParams(r *http.Request) (pageSize int, endCursor int, sort int) {
	pageSizeQuery := r.URL.Query().Get("pageSize")
	endCursorQuery := r.URL.Query().Get("endCursor")
	sortQuery := r.URL.Query().Get("sort")
	pageSize, _ = strconv.Atoi(pageSizeQuery)
	endCursor, _ = strconv.Atoi(endCursorQuery)
	sort, _ = strconv.Atoi(sortQuery)
	if pageSize == 0 {
		pageSize = 10
	}
	if sort >= 0 {
		sort = 1
	} else {
		sort = -1
	}
	return pageSize, endCursor, sort
}

func GetPageSize(ctx context.Context) int {
	pageSize := ctx.Value(pageSizeKey{})
	if pageSize == nil {
		return 10
	}
	return pageSize.(int)
}

func GetEndCursor[P int | string](ctx context.Context) (empty P) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("recovered from panic: %s", r)
		}
	}()
	endCursor := ctx.Value(endCursorKey{})
	if endCursor == nil {
		return empty
	}
	return endCursor.(P)
}

func GetSort(ctx context.Context) int {
	sort := ctx.Value(sortKey{})
	if sort == nil {
		return 1
	}
	return sort.(int)
}

func GetPaginationParams[P int | string](ctx context.Context) (pageSize int, cursor P, sort int) {
	return GetPageSize(ctx), GetEndCursor[P](ctx), GetSort(ctx)
}

func SetOperator(ctx context.Context, op DbOperator) context.Context {
	return context.WithValue(ctx, DbOperatorKey{}, op)
}

func GetOperator(ctx context.Context, defaultOp DbOperator) DbOperator {
	op := ctx.Value(DbOperatorKey{})
	if op == nil {
		log.Printf("operator is nil, using default operator")
		return defaultOp
	}
	log.Printf("operator is of type %T", op)
	return op.(DbOperator)
}
