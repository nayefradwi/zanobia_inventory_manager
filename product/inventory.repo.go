package product

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
	"github.com/nayefradwi/zanobia_inventory_manager/warehouse"
)

type IInventoryRepository interface {
	GetInventoryBaseByIngredientId(ctx context.Context, ingredientId int) InventoryBase
	CreateInventory(ctx context.Context, input InventoryInput) error
	UpdateInventory(ctx context.Context, base InventoryBase) error
	GetInventories(ctx context.Context, params common.PaginationParams) ([]Inventory, error)
}

type InventoryRepository struct {
	*pgxpool.Pool
}

func NewInventoryRepository(dbPool *pgxpool.Pool) IInventoryRepository {
	return &InventoryRepository{Pool: dbPool}
}

func (r *InventoryRepository) UpdateInventory(ctx context.Context, base InventoryBase) error {
	currentTimestamp := time.Now().UTC()
	sql := `UPDATE inventories SET quantity = $1, updated_at = $2 WHERE id = $3 and warehouse_id = $4`
	op := common.GetOperator(ctx, r.Pool)
	_, err := op.Exec(ctx, sql, base.Quantity, currentTimestamp, base.Id, base.WarehouseId)
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

func (r *InventoryRepository) GetInventories(ctx context.Context, params common.PaginationParams) ([]Inventory, error) {
	rows, err := r.getInventoriesRowsDescending(ctx, params)
	if err != nil {
		log.Printf("Failed to get inventories: %s", err.Error())
		return nil, common.NewBadRequestFromMessage("Failed to get inventories")
	}
	var inventories []Inventory
	for rows.Next() {
		var inventory Inventory
		var unit Unit
		var ingredient Ingredient
		err := rows.Scan(
			&inventory.Id, &ingredient.Id, &unit.Id, &inventory.Quantity,
			&unit.Name, &unit.Symbol, &ingredient.Price, &ingredient.ExpiresInDays,
			&ingredient.Name, &ingredient.Brand, &inventory.UpdatedAt,
		)
		if err != nil {
			log.Printf("Failed to scan inventories: %s", err.Error())
			return nil, common.NewBadRequestFromMessage("Failed to scan inventories")
		}
		inventory.Unit = unit
		inventory.Ingredient = ingredient
		inventories = append(inventories, inventory)
	}
	return inventories, nil
}

func (r *InventoryRepository) getInventoriesRowsDescending(ctx context.Context, params common.PaginationParams) (pgx.Rows, error) {
	sqlBuilder := common.NewPaginationQueryBuilder(
		`
		select inv.id, ing.id ingredient_id,
		utx.unit_id, quantity, utx.name,
		utx.symbol, ing.price, expires_in_days,
		ingtx.name, ingtx.brand, inv.updated_at
		from inventories inv
		join ingredients ing on ing.id = inv.ingredient_id
		join ingredient_translations ingtx on ingtx.ingredient_id = inv.ingredient_id
		join unit_translations utx on utx.unit_id = inv.unit_id
		`,
		[]string{"inv.updated_at DESC", "inv.id ASC"},
	)
	q, sql := sqlBuilder.
		WithConditions(
			[]string{
				"utx.language_code = $1",
				"and",
				"warehouse_id = $2",
			},
		).
		WithCursor(params.EndCursor, params.PreviousCursor).
		WithCursorKeys([]string{"inv.id", "inv.updated_at"}).
		WithDirection(params.Direction).
		WithPageSize(params.PageSize).
		Build()
	op := common.GetOperator(ctx, r.Pool)
	languageCode := common.GetLanguageParam(ctx)
	warehouseId := warehouse.GetWarehouseId(ctx)
	return q.Query(ctx, op, sql, languageCode, warehouseId)
}
