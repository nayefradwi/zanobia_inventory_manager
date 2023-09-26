package product

import (
	"context"
	"log"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
	"github.com/nayefradwi/zanobia_inventory_manager/warehouse"
)

type IInventoryRepository interface {
	GetInventoryBaseByIngredientId(ctx context.Context, ingredientId int) (InventoryBase, int, error)
	CreateInventory(ctx context.Context, input InventoryInput) error
	IncrementInventory(ctx context.Context, base InventoryBase) error
	DecrementInventory(ctx context.Context, base InventoryBase) error
}

type InventoryRepository struct {
	*pgxpool.Pool
}

func NewInventoryRepository(dbPool *pgxpool.Pool) IInventoryRepository {
	return &InventoryRepository{Pool: dbPool}
}

func (r *InventoryRepository) IncrementInventory(ctx context.Context, base InventoryBase) error {
	return common.NewInternalServerError()
}

func (r *InventoryRepository) DecrementInventory(ctx context.Context, base InventoryBase) error {
	return common.NewInternalServerError()
}

func (r *InventoryRepository) CreateInventory(ctx context.Context, input InventoryInput) error {
	sql := `INSERT INTO inventories (ingredient_id, warehouse_id, quantity, unit_id) VALUES ($1, $2, $3, $4)`
	op := common.GetOperator(ctx, r.Pool)
	warehouseId := warehouse.GetWarehouseId(ctx)
	_, err := op.Exec(ctx, sql, input.IngredientId, warehouseId, input.Quantity, input.UnitId)
	if err != nil {
		log.Printf("Failed to create inventory: %s", err.Error())
		return common.NewBadRequestFromMessage("Failed to create inventory")
	}
	return nil
}

func (r *InventoryRepository) GetInventoryBaseByIngredientId(ctx context.Context, ingredientId int) (InventoryBase, int, error) {
	sql := `SELECT u.id, inv.id, ingredient_id, warehouse_id, quantity, unit_id FROM inventories inv
			JOIN ingredients i ON i.id = inv.ingredient_id
			JOIN units u ON u.id = i.unit_id
			WHERE warehouse_id = $1 AND ingredient_id = $2`
	warehouseId := warehouse.GetWarehouseId(ctx)
	op := common.GetOperator(ctx, r.Pool)
	row := op.QueryRow(ctx, sql, warehouseId, ingredientId)
	var inventoryBase InventoryBase
	var unitId int
	err := row.Scan(&unitId, &inventoryBase.Id, &inventoryBase.IngredientId, &inventoryBase.WarehouseId, &inventoryBase.Quantity, &inventoryBase.UnitId)
	if err != nil {
		log.Printf("Failed to get inventory: %s", err.Error())
		return InventoryBase{}, 0, common.NewBadRequestFromMessage("Failed to get inventory")
	}
	return inventoryBase, unitId, nil
}

// func (r *InventoryRepository) GetInventories(ctx context.Context) ([]Inventory, error) {

// }
