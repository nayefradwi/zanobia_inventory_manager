package warehouse

import "context"

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
	return s.repo.GetWarehouses(ctx)
}

func (s *WarehouseService) AddUserToWarehouse(ctx context.Context, input WarehouseUserInput) error {
	return s.repo.AddUserToWarehouse(ctx, input)
}
