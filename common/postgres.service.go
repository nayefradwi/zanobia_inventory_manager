package common

import (
	"context"
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
		pageSize, endCursor := getPaginationParams(r)
		ctx := r.Context()
		ctx = context.WithValue(ctx, pageSizeKey{}, pageSize)
		ctx = context.WithValue(ctx, endCursorKey{}, endCursor)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getPaginationParams(r *http.Request) (int, int) {
	pageSizeQuery := r.URL.Query().Get("pageSize")
	endCursorQuery := r.URL.Query().Get("endCursor")
	pageSize, _ := strconv.Atoi(pageSizeQuery)
	endCursor, _ := strconv.Atoi(endCursorQuery)
	if pageSize == 0 {
		pageSize = 10
	}
	return pageSize, endCursor
}

func GetPageSize(ctx context.Context) int {
	pageSize := ctx.Value(pageSizeKey{})
	if pageSize == nil {
		return 10
	}
	return pageSize.(int)
}

func GetEndCursor(ctx context.Context) int {
	endCursor := ctx.Value(endCursorKey{})
	if endCursor == nil {
		return 0
	}
	return endCursor.(int)
}

func GetPaginationParams(ctx context.Context) (int, int) {
	return GetPageSize(ctx), GetEndCursor(ctx)
}
