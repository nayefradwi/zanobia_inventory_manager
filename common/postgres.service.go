package common

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"strings"

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
		paginationParam := getPaginationParams(r)
		ctx := r.Context()
		ctx = context.WithValue(ctx, paginationParamsKey{}, paginationParam)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getPaginationParams(r *http.Request) PaginationParams {
	pageSizeQuery := r.URL.Query().Get("pageSize")
	cursor := r.URL.Query().Get("cursor")
	directionQuery := r.URL.Query().Get("direction")
	pageSize, _ := strconv.Atoi(pageSizeQuery)
	direction, _ := strconv.Atoi(directionQuery)
	if pageSize == 0 {
		pageSize = 10
	}
	if direction > 0 {
		direction = 1
	} else if direction < 0 {
		direction = -1
	}
	return PaginationParams{
		PageSize:  pageSize,
		Direction: direction,
		Cursor:    cursor,
	}
}

func GetPaginationParams(ctx context.Context) PaginationParams {
	if params, ok := ctx.Value(paginationParamsKey{}).(PaginationParams); ok {
		return params
	}
	return PaginationParams{
		PageSize:  10,
		Direction: 1,
		Cursor:    "",
	}
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

func createCursorValue(c Cursorable) string {
	cursorColumns := c.GetCursorValue()
	if len(cursorColumns) == 0 {
		return ""
	}
	if len(cursorColumns) == 1 {
		return cursorColumns[0]
	}
	cursorValue := strings.Join(cursorColumns, ",")
	return Base64Encode(cursorValue)
}
