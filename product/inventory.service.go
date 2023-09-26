package product

import (
	"context"
	"strconv"

	"github.com/nayefradwi/zanobia_inventory_manager/common"
)

type IInventoryService interface{}
type InventoryServiceWorkUnit struct {
	inventoryRepo  IInventoryRepository
	unitService    IUnitService
	ingredientRepo IIngredientRepository
	lockingService common.IDistributedLockingService
}

type InventoryService struct {
	inventoryRepo  IInventoryRepository
	unitService    IUnitService
	ingredientRepo IIngredientRepository
	lockingService common.IDistributedLockingService
}

func NewInventoryService(workUnit InventoryServiceWorkUnit) IInventoryService {
	return &InventoryService{
		inventoryRepo:  workUnit.inventoryRepo,
		unitService:    workUnit.unitService,
		lockingService: workUnit.lockingService,
		ingredientRepo: workUnit.ingredientRepo,
	}
}

func (s *InventoryService) IncrementInventory(ctx context.Context, inventoryInput InventoryInput) error {
	_, err := s.lockingService.Acquire(ctx, strconv.Itoa(inventoryInput.IngredientId))
	defer s.lockingService.Release(ctx, strconv.Itoa(inventoryInput.IngredientId))
	if err != nil {
		return common.NewBadRequestFromMessage("Failed to acquire lock")
	}
	validationErr := ValidateInventoryInput(inventoryInput)
	if validationErr != nil {
		return validationErr
	}
	InventoryBase, err := s.getConvertedInventory(ctx, inventoryInput)
	if err != nil {
		return err
	}
	if InventoryBase.Id == nil {
		return s.inventoryRepo.CreateInventory(ctx, inventoryInput)
	}
	return s.incrementInventory(ctx, InventoryBase, inventoryInput)
}

func (s *InventoryService) getConvertedInventory(ctx context.Context, inventoryInput InventoryInput) (InventoryBase, error) {
	invBase, unitId, err := s.inventoryRepo.GetInventoryBaseByIngredientId(ctx, inventoryInput.IngredientId)
	if err != nil {
		return invBase, err
	}
	convertedQty, err := s.convertUnit(ctx, unitId, inventoryInput)
	if err != nil {
		return invBase, err
	}
	inventoryInput.Quantity = convertedQty
	return invBase, nil
}

func (s *InventoryService) convertUnit(ctx context.Context, unitId int, inventoryInput InventoryInput) (float64, error) {
	if unitId == inventoryInput.UnitId {
		return inventoryInput.Quantity, nil
	}
	out, err := s.unitService.ConvertUnit(ctx, ConvertUnitInput{
		ToUnitId:   &unitId,
		FromUnitId: &inventoryInput.UnitId,
		Quantity:   inventoryInput.Quantity,
	})
	if err != nil {
		return 0, err
	}
	return out.Quantity, nil
}

func (s *InventoryService) incrementInventory(ctx context.Context, inventoryBase InventoryBase, inventoryInput InventoryInput) error {
	newQty := inventoryBase.Quantity + inventoryInput.Quantity
	inventoryBase.Quantity = newQty
	return s.inventoryRepo.IncrementInventory(ctx, inventoryBase)
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
