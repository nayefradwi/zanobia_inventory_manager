package product

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
	"github.com/nayefradwi/zanobia_inventory_manager/warehouse"
)

type IInventoryRepository interface {
	GetInventoryBaseByIngredientId(ctx context.Context, ingredientId int) InventoryBase
	CreateInventory(ctx context.Context, input InventoryInput) error
	UpdateInventory(ctx context.Context, base InventoryBase) error
}

type InventoryRepository struct {
	*pgxpool.Pool
}

func NewInventoryRepository(dbPool *pgxpool.Pool) IInventoryRepository {
	return &InventoryRepository{Pool: dbPool}
}

func (r *InventoryRepository) UpdateInventory(ctx context.Context, base InventoryBase) error {
	currentTimestamp := time.Now().UTC()
	sql := `UPDATE inventories SET quantity = $1, updated_at = $2 WHERE id = $3`
	op := common.GetOperator(ctx, r.Pool)
	_, err := op.Exec(ctx, sql, base.Quantity, currentTimestamp, base.Id)
	if err != nil {
		log.Printf("Failed to increment inventory: %s", err.Error())
		return common.NewBadRequestFromMessage("Failed to increment inventory")
	}
	return nil
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

func (r *InventoryRepository) GetInventoryBaseByIngredientId(ctx context.Context, ingredientId int) InventoryBase {
	sql := `SELECT  inv.id, ingredient_id, warehouse_id, quantity, unit_id FROM inventories inv
			JOIN ingredients i ON i.id = inv.ingredient_id
			WHERE warehouse_id = $1 AND ingredient_id = $2`
	warehouseId := warehouse.GetWarehouseId(ctx)
	op := common.GetOperator(ctx, r.Pool)
	row := op.QueryRow(ctx, sql, warehouseId, ingredientId)
	var inventoryBase InventoryBase
	err := row.Scan(&inventoryBase.Id, &inventoryBase.IngredientId, &inventoryBase.WarehouseId, &inventoryBase.Quantity, &inventoryBase.UnitId)
	if err != nil {
		log.Printf("Failed to get inventory: %s", err.Error())
	}
	return inventoryBase
}

// func (r *InventoryRepository) GetInventories(ctx context.Context) ([]Inventory, error) {

// }
