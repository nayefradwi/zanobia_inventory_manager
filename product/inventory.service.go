package product

import (
	"context"
	"strconv"

	"github.com/nayefradwi/zanobia_inventory_manager/common"
)

type IInventoryService interface{}

type InventoryService struct {
	inventoryRepo  IInventoryRepository
	unitService    IUnitService
	ingredientRepo IIngredientRepository
	lockingService common.IDistributedLockingService
}

func NewInventoryService(inventoryRepo IInventoryRepository, unitService IUnitService, ingredientRepo IIngredientRepository, lockingService common.IDistributedLockingService) IInventoryService {
	return &InventoryService{
		inventoryRepo:  inventoryRepo,
		unitService:    unitService,
		lockingService: lockingService,
		ingredientRepo: ingredientRepo,
	}
}

func (s *InventoryService) IncrementInventory(ctx context.Context, warehouseId int, inventoryInput InventoryInput) error {
	_, err := s.lockingService.Acquire(ctx, strconv.Itoa(inventoryInput.IngredientId))
	defer s.lockingService.Release(ctx, strconv.Itoa(inventoryInput.IngredientId))
	if err != nil {
		return common.NewBadRequestFromMessage("Failed to acquire lock")
	}
	validationErr := ValidateInventoryInput(inventoryInput)
	if validationErr != nil {
		return validationErr
	}
	invBase, unitId, err := s.inventoryRepo.GetInventoryBaseByIngredientId(ctx, warehouseId, inventoryInput.IngredientId)
	if err != nil {
		return err
	}
	convertedQty, err := s.convertUnit(ctx, unitId, inventoryInput)
	if err != nil {
		return err
	}
	inventoryInput.Quantity = convertedQty
	if invBase.Id == nil {
		return s.inventoryRepo.CreateInventory(ctx, s.inventoryRepo.(*InventoryRepository), warehouseId, inventoryInput)
	}
	return s.incrementInventory(ctx, s.inventoryRepo.(*InventoryRepository), invBase, inventoryInput)
}

func (s *InventoryService) convertUnit(ctx context.Context, unitId int, inventoryInput InventoryInput) (float64, error) {
	if unitId == inventoryInput.UnitId {
		return inventoryInput.Quantity, nil
	}
	out, err := s.unitService.ConvertUnit(ctx, ConvertUnitInput{
		UnitId:           &unitId,
		ConversionUnitId: &inventoryInput.UnitId,
		Quantity:         inventoryInput.Quantity,
	})
	if err != nil {
		return 0, err
	}
	return out.Quantity, nil
}

func (s *InventoryService) incrementInventory(ctx context.Context, op common.DbOperator, inventoryBase InventoryBase, inventoryInput InventoryInput) error {
	newQty := inventoryBase.Quantity + inventoryInput.Quantity
	inventoryBase.Quantity = newQty
	return s.inventoryRepo.IncrementInventory(ctx, op, inventoryBase)

}

func (s *InventoryService) DecrementInventory(ctx context.Context, inventoryInput InventoryInput) error {
	return nil
}

func (s *InventoryService) BulkIncrementInventory(ctx context.Context, inventoryInputs []InventoryInput) error {
	return nil
}

func (s *InventoryService) BulkDecrementInventory(ctx context.Context, inventoryInputs []InventoryInput) error {
	return nil
}

// TODO increment by recipe
// TODO decrement by recipe
