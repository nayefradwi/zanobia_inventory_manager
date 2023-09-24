package product

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
)

type IInventoryRepository interface {
	GetInventoryBaseByIngredientId(ctx context.Context, warehouseId, ingredientId int) (InventoryBase, int, error)
	CreateInventory(ctx context.Context, op common.DbOperator, warehouseId int, input InventoryInput) error
	IncrementInventory(ctx context.Context, op common.DbOperator, base InventoryBase) error
}

type InventoryRepository struct {
	*pgxpool.Pool
}

func NewInventoryRepository(dbPool *pgxpool.Pool) IInventoryRepository {
	return &InventoryRepository{Pool: dbPool}
}

func (s *InventoryRepository) IncrementInventory(ctx context.Context, op common.DbOperator, base InventoryBase) error {
	return nil
}

func (s *InventoryRepository) DecrementInventory(ctx context.Context, op common.DbOperator, warehouseId int, input InventoryInput) error {
	return nil
}

func (s *InventoryRepository) CreateInventory(ctx context.Context, op common.DbOperator, warehouseId int, input InventoryInput) error {
	return nil
}

func (s *InventoryRepository) GetInventoryBaseByIngredientId(ctx context.Context, warehouseId, ingredientId int) (InventoryBase, int, error) {
	sql := `SELECT u.id, inv.id, ingredient_id, warehouse_id, quantity, unit_id FROM inventories inv
			JOIN ingredients i ON i.id = inv.ingredient_id
			JOIN units u ON u.id = i.unit_id
			WHERE warehouse_id = $1 AND ingredient_id = $2`
	row := s.QueryRow(ctx, sql, warehouseId, ingredientId)
	var inventoryBase InventoryBase
	var unitId int
	err := row.Scan(&unitId, &inventoryBase.Id, &inventoryBase.IngredientId, &inventoryBase.WarehouseId, &inventoryBase.Quantity, &inventoryBase.UnitId)
	if err != nil {
		return InventoryBase{}, 0, common.NewBadRequestFromMessage("Failed to get inventory")
	}
	return inventoryBase, unitId, nil
}
