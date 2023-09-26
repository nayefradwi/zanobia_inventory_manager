package warehouse

import (
	"context"
	"net/http"
	"strconv"
)

const WarehouseIdHeader = "X-Warehouse-Id"

type WarehouseIdKey struct{}

func SetWarehouseIdFromHeader(f http.Handler) http.Handler {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		warehouseIdValue := r.Header.Get(WarehouseIdHeader)
		warehouseId, _ := strconv.Atoi(warehouseIdValue)
		r = r.WithContext(SetWarehouseId(r.Context(), warehouseId))
		f.ServeHTTP(w, r)
	})
	return handler
}

func SetWarehouseId(ctx context.Context, warehouseId int) context.Context {
	return context.WithValue(ctx, WarehouseIdKey{}, warehouseId)
}

func GetWarehouseId(ctx context.Context) int {
	warehouseId := ctx.Value(WarehouseIdKey{})
	if warehouseId == nil {
		return 0
	}
	return warehouseId.(int)
}
