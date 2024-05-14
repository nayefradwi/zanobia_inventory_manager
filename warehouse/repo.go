package warehouse

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
	"go.uber.org/zap"
)

type IWarehouseRepository interface {
	CreateWarehouse(ctx context.Context, warehouse Warehouse) error
	GetWarehouses(ctx context.Context, userId int) ([]Warehouse, error)
	AddUserToWarehouse(ctx context.Context, input WarehouseUserInput) error
	GetWarehouseById(ctx context.Context, warehouseId, userId int) (Warehouse, error)
	UpdateWarehouse(ctx context.Context, warehouse Warehouse) error
}

type WarehouseRepository struct {
	*pgxpool.Pool
}

func NewWarehouseRepository(dbPool *pgxpool.Pool) IWarehouseRepository {
	return &WarehouseRepository{Pool: dbPool}
}

func (r *WarehouseRepository) CreateWarehouse(ctx context.Context, warehouse Warehouse) error {
	sql := `INSERT INTO warehouses (name, lat, lng) VALUES ($1, $2, $3)`
	_, err := r.Exec(ctx, sql, warehouse.Name, warehouse.Lat, warehouse.Lng)
	if err != nil {
		return common.NewBadRequestFromMessage("Failed to create warehouse")
	}
	return nil
}

func (r *WarehouseRepository) GetWarehouses(ctx context.Context, userId int) ([]Warehouse, error) {
	sql := `
	SELECT w.id, name, lat, lng FROM warehouses w
	join user_warehouses uw on uw.warehouse_id = w.id
	where uw.user_id = $1;
	`
	rows, err := r.Query(ctx, sql, userId)
	if err != nil {
		common.LoggerFromCtx(ctx).Error("Failed to get warehouses", zap.Error(err))
		return nil, common.NewBadRequestFromMessage("Failed to get warehouses")
	}
	defer rows.Close()
	warehouses := make([]Warehouse, 0)
	for rows.Next() {
		var warehouse Warehouse
		err := rows.Scan(&warehouse.Id, &warehouse.Name, &warehouse.Lat, &warehouse.Lng)
		if err != nil {
			common.LoggerFromCtx(ctx).Error("Failed to scan", zap.Error(err))
			return nil, common.NewBadRequestFromMessage("Failed to get warehouses")
		}
		warehouses = append(warehouses, warehouse)
	}
	return warehouses, nil
}

func (r *WarehouseRepository) AddUserToWarehouse(ctx context.Context, input WarehouseUserInput) error {
	sql := `INSERT INTO user_warehouses (warehouse_id, user_id) VALUES ($1, $2)`
	_, err := r.Exec(ctx, sql, input.WarehouseId, input.UserId)
	if err != nil {
		common.LoggerFromCtx(ctx).Error("Failed to add user to warehouse", zap.Error(err))
		return common.NewBadRequestFromMessage("Failed to add user to warehouse")
	}
	return nil
}

func (r *WarehouseRepository) GetWarehouseById(ctx context.Context, warehouseId, userId int) (Warehouse, error) {
	sql := `
	SELECT w.id, name, lat, lng FROM warehouses w
	join user_warehouses uw on uw.warehouse_id = w.id
	where uw.user_id = $1 and w.id = $2;
	`
	var warehouse Warehouse
	err := r.QueryRow(ctx, sql, userId, warehouseId).Scan(
		&warehouse.Id,
		&warehouse.Name,
		&warehouse.Lat,
		&warehouse.Lng,
	)
	if err != nil {
		common.LoggerFromCtx(ctx).Error("Failed to get warehouse", zap.Error(err))
		return warehouse, common.NewBadRequestFromMessage("Failed to get warehouse")
	}
	return warehouse, nil
}

func (r *WarehouseRepository) UpdateWarehouse(ctx context.Context, warehouse Warehouse) error {
	sql := `UPDATE warehouses SET name = $1, lat = $2, lng = $3 WHERE id = $4`
	_, err := r.Exec(ctx, sql, warehouse.Name, warehouse.Lat, warehouse.Lng, warehouse.Id)
	if err != nil {
		common.LoggerFromCtx(ctx).Error("Failed to update warehouse", zap.Error(err))
		return common.NewBadRequestFromMessage("Failed to update warehouse")
	}
	return nil
}
