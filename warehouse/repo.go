package warehouse

import (
	"context"
	"log"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
)

type IWarehouseRepository interface {
	CreateWarehouse(ctx context.Context, warehouse Warehouse) error
	GetWarehouses(ctx context.Context) ([]Warehouse, error)
	AddUserToWarehouse(ctx context.Context, input WarehouseUserInput) error
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

func (r *WarehouseRepository) GetWarehouses(ctx context.Context) ([]Warehouse, error) {
	sql := `SELECT id, name, lat, lng FROM warehouses`
	rows, err := r.Query(ctx, sql)
	if err != nil {
		log.Printf("Failed to get warehouses: %s", err.Error())
		return nil, common.NewBadRequestFromMessage("Failed to get warehouses")
	}
	defer rows.Close()
	warehouses := make([]Warehouse, 0)
	for rows.Next() {
		var warehouse Warehouse
		err := rows.Scan(&warehouse.Id, &warehouse.Name, &warehouse.Lat, &warehouse.Lng)
		if err != nil {
			log.Printf("Failed to scan: %s", err.Error())
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
		log.Printf("Failed to add user to warehouse: %s", err.Error())
		return common.NewBadRequestFromMessage("Failed to add user to warehouse")
	}
	return nil
}
