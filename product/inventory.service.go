package product

import (
	"context"
	"strconv"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
)

type IInventoryService interface {
	IncrementInventory(ctx context.Context, inventoryInput InventoryInput) error
	DecrementInventory(ctx context.Context, inventoryInput InventoryInput) error
	BulkIncrementInventory(ctx context.Context, inventoryInputs []InventoryInput) error
	BulkDecrementInventory(ctx context.Context, inventoryInputs []InventoryInput) error
	GetInventories(ctx context.Context) (common.PaginatedResponse[Inventory], error)

	// private methods for services to use from their own transaction
	bulkIncrementInventory(ctx context.Context, inventoryInputs []InventoryInput) error
	bulkDecrementInventory(ctx context.Context, inventoryInputs []InventoryInput) error
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
	return s.lockingService.RunWithLock(ctx, strconv.Itoa(inventoryInput.IngredientId), func() error {
		validationErr := ValidateInventoryInput(inventoryInput)
		if validationErr != nil {
			return validationErr
		}
		err := common.RunWithTransaction(ctx, s.inventoryRepo.(*InventoryRepository).Pool, func(ctx context.Context, tx pgx.Tx) error {
			ctx = common.SetOperator(ctx, tx)
			return s.tryToIncrementInventory(ctx, inventoryInput)
		})
		return err
	})

}

func (s *InventoryService) tryToIncrementInventory(ctx context.Context, inventoryInput InventoryInput) error {
	InventoryBase, err := s.getConvertedInventory(ctx, &inventoryInput)
	if err != nil {
		return err
	}
	if InventoryBase.Id == nil {
		return s.inventoryRepo.CreateInventory(ctx, inventoryInput)
	}
	return s.incrementInventory(ctx, InventoryBase, inventoryInput)
}

func (s *InventoryService) getConvertedInventory(ctx context.Context, inventoryInput *InventoryInput) (InventoryBase, error) {
	invBase := s.inventoryRepo.GetInventoryBaseByIngredientId(ctx, inventoryInput.IngredientId)
	unitId := invBase.UnitId
	if invBase.Id == nil {
		unitId, _ = s.ingredientRepo.GetUnitIdOfIngredient(ctx, inventoryInput.IngredientId)
	}
	if unitId != inventoryInput.UnitId {
		convertedQty, err := s.convertUnit(ctx, unitId, *inventoryInput)
		if err != nil {
			return invBase, err

		}
		inventoryInput.Quantity = convertedQty
		inventoryInput.UnitId = unitId
	}
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
	return s.inventoryRepo.UpdateInventory(ctx, inventoryBase)
}

func (s *InventoryService) DecrementInventory(ctx context.Context, inventoryInput InventoryInput) error {
	return s.lockingService.RunWithLock(ctx, strconv.Itoa(inventoryInput.IngredientId), func() error {
		validationErr := ValidateInventoryInput(inventoryInput)
		if validationErr != nil {
			return validationErr
		}
		err := common.RunWithTransaction(ctx, s.inventoryRepo.(*InventoryRepository).Pool, func(ctx context.Context, tx pgx.Tx) error {
			ctx = common.SetOperator(ctx, tx)
			return s.tryToDecrementInventory(ctx, inventoryInput)
		})
		return err
	})
}

func (s *InventoryService) tryToDecrementInventory(ctx context.Context, inventoryInput InventoryInput) error {
	InventoryBase, err := s.getConvertedInventory(ctx, &inventoryInput)
	if err != nil {
		return err
	}
	if InventoryBase.Id == nil {
		return common.NewBadRequestFromMessage("Inventory not found")
	}
	return s.decrementInventory(ctx, InventoryBase, inventoryInput)
}

func (s *InventoryService) decrementInventory(ctx context.Context, inventoryBase InventoryBase, inventoryInput InventoryInput) error {
	newQty := inventoryBase.Quantity - inventoryInput.Quantity
	if newQty < 0 {
		return common.NewBadRequestFromMessage("Inventory not enough")
	}
	inventoryBase = inventoryBase.SetQuantity(newQty)
	return s.inventoryRepo.UpdateInventory(ctx, inventoryBase)
}

func (s *InventoryService) BulkIncrementInventory(ctx context.Context, inventoryInputs []InventoryInput) error {
	err := common.RunWithTransaction(ctx, s.inventoryRepo.(*InventoryRepository).Pool, func(ctx context.Context, tx pgx.Tx) error {
		ctx = common.SetOperator(ctx, tx)
		return s.bulkIncrementInventory(ctx, inventoryInputs)
	})
	return err
}

func (s *InventoryService) bulkIncrementInventory(ctx context.Context, inventoryInputs []InventoryInput) error {
	locks := make([]common.Lock, 0)
	locksPtr := &locks
	defer s.lockingService.ReleaseMany(context.Background(), locksPtr)
	for _, input := range inventoryInputs {
		lock, lockErr := s.lockingService.Acquire(ctx, strconv.Itoa(input.IngredientId))
		if lockErr != nil {
			return common.NewBadRequestFromMessage("Failed to acquire lock")
		}
		locks = append(locks, lock)
		locksPtr = &locks
		validationErr := ValidateInventoryInput(input)
		if validationErr != nil {
			return validationErr
		}
		if err := s.tryToIncrementInventory(ctx, input); err != nil {
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
	locks := make([]common.Lock, 0)
	locksPtr := &locks
	defer s.lockingService.ReleaseMany(context.Background(), locksPtr)
	for _, input := range inventoryInputs {
		lock, lockErr := s.lockingService.Acquire(ctx, strconv.Itoa(input.IngredientId))
		if lockErr != nil {
			return common.NewBadRequestFromMessage("Failed to acquire lock")
		}
		locks = append(locks, lock)
		locksPtr = &locks
		validationErr := ValidateInventoryInput(input)
		if validationErr != nil {
			return validationErr
		}
		if err := s.tryToDecrementInventory(ctx, input); err != nil {
			return err
		}
	}
	return nil
}

func (s *InventoryService) GetInventories(ctx context.Context) (common.PaginatedResponse[Inventory], error) {
	size, cursor, sorting := common.GetPaginationParams(ctx, time.Now().Format(time.RFC3339))
	inventories, err := s.inventoryRepo.GetInventories(ctx, size, sorting, cursor)
	if err != nil {
		return common.PaginatedResponse[Inventory]{}, err
	}
	if len(inventories) == 0 {
		return common.CreateEmptyPaginatedResponse[Inventory](size), nil
	}
	last := inventories[len(inventories)-1]
	res := common.CreatePaginatedResponse[Inventory](size, last.UpdatedAt.Format(time.RFC3339), inventories)
	return res, nil
}

// TODO increment by recipe
// TODO decrement by recipe
