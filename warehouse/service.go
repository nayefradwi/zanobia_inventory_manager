package warehouse

import (
	"context"

	"github.com/nayefradwi/zanobia_inventory_manager/common"
)

type IWarehouseService interface {
	CreateWarehouse(ctx context.Context, warehouse Warehouse) error
	GetWarehouses(ctx context.Context) ([]Warehouse, error)
	AddUserToWarehouse(ctx context.Context, input WarehouseUserInput) error
}

type WarehouseService struct {
	repo IWarehouseRepository
}

func NewWarehouseService(warehouseRepo IWarehouseRepository) IWarehouseService {
	return &WarehouseService{
		repo: warehouseRepo,
	}
}

func (s *WarehouseService) CreateWarehouse(ctx context.Context, warehouse Warehouse) error {
	validationErr := ValidateWarehouse(warehouse)
	if validationErr != nil {
		return validationErr
	}
	return s.repo.CreateWarehouse(ctx, warehouse)
}

func (s *WarehouseService) GetWarehouses(ctx context.Context) ([]Warehouse, error) {
	userId := common.GetUserIdFromContext(ctx)
	return s.repo.GetWarehouses(ctx, userId)
}

func (s *WarehouseService) AddUserToWarehouse(ctx context.Context, input WarehouseUserInput) error {
	return s.repo.AddUserToWarehouse(ctx, input)
}
