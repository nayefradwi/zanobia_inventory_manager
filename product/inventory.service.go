package product

import (
	"context"
	"strconv"

	"github.com/jackc/pgx/v4"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
)

type IInventoryService interface {
	IncrementInventory(ctx context.Context, inventoryInput InventoryInput) error
	DecrementInventory(ctx context.Context, inventoryInput InventoryInput) error
	BulkIncrementInventory(ctx context.Context, inventoryInputs []InventoryInput) error
	BulkDecrementInventory(ctx context.Context, inventoryInputs []InventoryInput) error
}
type InventoryServiceWorkUnit struct {
	InventoryRepo  IInventoryRepository
	UnitService    IUnitService
	IngredientRepo IIngredientRepository
	LockingService common.IDistributedLockingService
}

type InventoryService struct {
	inventoryRepo  IInventoryRepository
	unitService    IUnitService
	ingredientRepo IIngredientRepository
	lockingService common.IDistributedLockingService
}

func NewInventoryService(workUnit InventoryServiceWorkUnit) IInventoryService {
	return &InventoryService{
		inventoryRepo:  workUnit.InventoryRepo,
		unitService:    workUnit.UnitService,
		lockingService: workUnit.LockingService,
		ingredientRepo: workUnit.IngredientRepo,
	}
}

func (s *InventoryService) IncrementInventory(ctx context.Context, inventoryInput InventoryInput) error {
	_, lockErr := s.lockingService.Acquire(ctx, strconv.Itoa(inventoryInput.IngredientId))
	defer s.lockingService.Release(ctx, strconv.Itoa(inventoryInput.IngredientId))
	if lockErr != nil {
		return common.NewBadRequestFromMessage("Failed to acquire lock")
	}
	validationErr := ValidateInventoryInput(inventoryInput)
	if validationErr != nil {
		return validationErr
	}
	err := common.RunWithTransaction(ctx, s.inventoryRepo.(*InventoryRepository).Pool, func(ctx context.Context, tx pgx.Tx) error {
		ctx = common.SetOperator(ctx, tx)
		InventoryBase, err := s.getConvertedInventory(ctx, &inventoryInput)
		if err != nil {
			return err
		}
		if InventoryBase.Id == nil {
			return s.inventoryRepo.CreateInventory(ctx, inventoryInput)
		}
		return s.incrementInventory(ctx, InventoryBase, inventoryInput)
	})
	return err
}

func (s *InventoryService) getConvertedInventory(ctx context.Context, inventoryInput *InventoryInput) (InventoryBase, error) {
	invBase, unitId, err := s.inventoryRepo.GetInventoryBaseByIngredientId(ctx, inventoryInput.IngredientId)
	if err != nil {
		return invBase, err
	}
	convertedQty, err := s.convertUnit(ctx, unitId, *inventoryInput)
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
	inventoryBase = inventoryBase.SetQuantity(inventoryBase.Quantity + inventoryInput.Quantity)
	return s.inventoryRepo.IncrementInventory(ctx, inventoryBase)
}

func (s *InventoryService) DecrementInventory(ctx context.Context, inventoryInput InventoryInput) error {
	_, lockErr := s.lockingService.Acquire(ctx, strconv.Itoa(inventoryInput.IngredientId))
	defer s.lockingService.Release(ctx, strconv.Itoa(inventoryInput.IngredientId))
	if lockErr != nil {
		return common.NewBadRequestFromMessage("Failed to acquire lock")
	}
	validationErr := ValidateInventoryInput(inventoryInput)
	if validationErr != nil {
		return validationErr
	}
	err := common.RunWithTransaction(ctx, s.inventoryRepo.(*InventoryRepository).Pool, func(ctx context.Context, tx pgx.Tx) error {
		ctx = common.SetOperator(ctx, tx)
		InventoryBase, err := s.getConvertedInventory(ctx, &inventoryInput)
		if err != nil {
			return err
		}
		if InventoryBase.Id == nil {
			return common.NewBadRequestFromMessage("Inventory not found")
		}
		return s.decrementInventory(ctx, InventoryBase, inventoryInput)
	})
	return err
}

func (s *InventoryService) decrementInventory(ctx context.Context, inventoryBase InventoryBase, inventoryInput InventoryInput) error {
	newQty := inventoryBase.Quantity - inventoryInput.Quantity
	if newQty < 0 {
		return common.NewBadRequestFromMessage("Inventory not enough")
	}
	inventoryBase = inventoryBase.SetQuantity(newQty)
	return s.inventoryRepo.DecrementInventory(ctx, inventoryBase)
}

func (s *InventoryService) BulkIncrementInventory(ctx context.Context, inventoryInputs []InventoryInput) error {
	err := common.RunWithTransaction(ctx, s.inventoryRepo.(*InventoryRepository).Pool, func(ctx context.Context, tx pgx.Tx) error {
		ctx = common.SetOperator(ctx, tx)
		return s.bulkIncrementInventory(ctx, inventoryInputs)
	})
	return err
}

func (s *InventoryService) bulkIncrementInventory(ctx context.Context, inventoryInputs []InventoryInput) error {
	for _, input := range inventoryInputs {
		_, lockErr := s.lockingService.Acquire(ctx, strconv.Itoa(input.IngredientId))
		defer s.lockingService.Release(ctx, strconv.Itoa(input.IngredientId))
		if lockErr != nil {
			return common.NewBadRequestFromMessage("Failed to acquire lock")
		}
		validationErr := ValidateInventoryInput(input)
		if validationErr != nil {
			return validationErr
		}
		invBase, convErr := s.getConvertedInventory(ctx, &input)
		if convErr != nil {
			return convErr
		}
		err := s.incrementInventory(ctx, invBase, input)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *InventoryService) BulkDecrementInventory(ctx context.Context, inventoryInputs []InventoryInput) error {
	err := common.RunWithTransaction(ctx, s.inventoryRepo.(*InventoryRepository).Pool, func(ctx context.Context, tx pgx.Tx) error {
		ctx = common.SetOperator(ctx, tx)
		return s.bulkDecrementInventory(ctx, inventoryInputs)
	})
	return err
}

func (s *InventoryService) bulkDecrementInventory(ctx context.Context, inventoryInputs []InventoryInput) error {
	for _, input := range inventoryInputs {
		_, lockErr := s.lockingService.Acquire(ctx, strconv.Itoa(input.IngredientId))
		defer s.lockingService.Release(ctx, strconv.Itoa(input.IngredientId))
		if lockErr != nil {
			return common.NewBadRequestFromMessage("Failed to acquire lock")
		}
		validationErr := ValidateInventoryInput(input)
		if validationErr != nil {
			return validationErr
		}
		invBase, convErr := s.getConvertedInventory(ctx, &input)
		if convErr != nil {
			return convErr
		}
		err := s.decrementInventory(ctx, invBase, input)
		if err != nil {
			return err
		}
	}
	return nil
}

// TODO increment by recipe
// TODO decrement by recipe
