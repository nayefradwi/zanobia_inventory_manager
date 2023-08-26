package product

import "context"

type IUnitService interface {
	CreateUnit(ctx context.Context, unit Unit) error
	GetAllUnits(ctx context.Context) ([]Unit, error)
}

type UnitService struct {
	repo IUnitRepository
}

func NewUnitService(repo IUnitRepository) *UnitService {
	return &UnitService{repo}
}

func (s *UnitService) CreateUnit(ctx context.Context, unit Unit) error {
	validationErr := ValidateUnit(unit)
	if validationErr != nil {
		return validationErr
	}
	return s.repo.CreateUnit(ctx, unit)
}

func (s *UnitService) GetAllUnits(ctx context.Context) ([]Unit, error) {
	return s.repo.GetAllUnits(ctx)
}
