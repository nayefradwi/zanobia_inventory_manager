package warehouse

import (
	"context"
	"net/http"
	"strconv"

	"github.com/nayefradwi/zanobia_inventory_manager/common"
	"go.uber.org/zap"
)

const WarehouseIdHeader = "X-Warehouse-Id"

type WarehouseIdKey struct{}

func SetWarehouseIdFromHeader(f http.Handler) http.Handler {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		warehouseIdValue := r.Header.Get(WarehouseIdHeader)
		warehouseId, _ := strconv.Atoi(warehouseIdValue)
		ctx := SetWarehouseId(r.Context(), warehouseId)
		logger := common.LoggerFromCtx(ctx)
		logger = logger.With(zap.Int("warehouseId", warehouseId))
		ctx = common.SetLoggerToCtx(ctx, logger)
		r = r.WithContext(ctx)
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
