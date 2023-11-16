package common

import (
	"context"
	"fmt"
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
	endCursor := r.URL.Query().Get("endCursor")
	previousCursor := r.URL.Query().Get("previousCursor")
	sortQuery := r.URL.Query().Get("sort")
	pageSize, _ := strconv.Atoi(pageSizeQuery)
	sort, _ := strconv.Atoi(sortQuery)
	if pageSize == 0 {
		pageSize = 10
	}
	if sort >= 0 {
		sort = 1
	} else {
		sort = -1
	}
	return PaginationParams{
		PageSize:       pageSize,
		Direction:      sort,
		EndCursor:      endCursor,
		PreviousCursor: previousCursor,
	}
}

func GetPaginationParams(ctx context.Context) PaginationParams {
	if params, ok := ctx.Value(paginationParamsKey{}).(PaginationParams); ok {
		return params
	}
	return PaginationParams{
		PageSize:       10,
		Direction:      1,
		EndCursor:      "",
		PreviousCursor: "",
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

func CreatePaginationQuery(sql PaginationQuery) string {
	unformattedSql := sql.BaseSql + " " + "WHERE %s %s ORDER BY %s LIMIT %s;"
	finalArgIndex, joinedConditions := getFinalArgAndJoinedConditions(sql.Conditions)
	unformattedPaginationConditonWithCursor := "AND %s %s $%d"
	unformattedPaginationConditionWithoutCursor := "AND $%d = $%d"
	directionString := getDirectionString(sql.Direction)
	var formattedPaginationCondition string
	if sql.EndCursor == "" && sql.PreviousCursor == "" {
		formattedPaginationCondition = fmt.Sprintf(
			unformattedPaginationConditionWithoutCursor,
			finalArgIndex,
			finalArgIndex,
		)
	} else {
		formattedPaginationCondition = fmt.Sprintf(
			unformattedPaginationConditonWithCursor,
			*sql.CursorKey,
			directionString,
			finalArgIndex,
		)
	}
	formattedSql := fmt.Sprintf(
		unformattedSql,
		joinedConditions,
		formattedPaginationCondition,
		sql.OrderByQuery,
		strconv.Itoa(sql.PageSize),
	)
	trimmedSql := strings.ReplaceAll(formattedSql, "\n", " ")
	trimmedSql = strings.ReplaceAll(trimmedSql, "\t", "")
	return trimmedSql
}

func getFinalArgAndJoinedConditions(conditions []string) (int, string) {
	finalArgIndex := 1
	for _, condition := range conditions {
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
		joinedConditions = strings.Join(conditions, " ")
	}
	return finalArgIndex, joinedConditions
}

func getDirectionString(direction int) string {
	if direction < 0 {
		return "<"
	}
	return ">"
}
