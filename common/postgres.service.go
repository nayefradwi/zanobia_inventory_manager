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
		pageSize, pageIndex := getPaginationParams(r)
		ctx := r.Context()
		ctx = context.WithValue(ctx, pageSizeKey{}, pageSize)
		ctx = context.WithValue(ctx, pageIndexKey{}, pageIndex)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getPaginationParams(r *http.Request) (int, int) {
	pageSizeQuery := r.URL.Query().Get("pageSize")
	pageIndexQuery := r.URL.Query().Get("pageIndex")
	pageSize, _ := strconv.Atoi(pageSizeQuery)
	pageIndex, _ := strconv.Atoi(pageIndexQuery)
	if pageSize == 0 {
		pageSize = 10
	}
	return pageSize, pageIndex
}

func GetPageSize(ctx context.Context) int {
	pageSize := ctx.Value(pageSizeKey{})
	if pageSize == nil {
		return 10
	}
	return pageSize.(int)
}

func GetPageIndex(ctx context.Context) int {
	pageIndex := ctx.Value(pageIndexKey{})
	if pageIndex == nil {
		return 0
	}
	return pageIndex.(int)
}

func GetPaginatedResponse[T any](ctx context.Context, items []T, total int) PaginatedResponse[T] {
	pageSize := GetPageSize(ctx)
	pageIndex := GetPageIndex(ctx)
	return CreatePaginatedResponse[T](pageSize, pageIndex, total, items)
}
