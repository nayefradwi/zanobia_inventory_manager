package product

import (
	"context"
)

type IUnitService interface {
	CreateUnit(ctx context.Context, unit Unit) error
	GetAllUnits(ctx context.Context) ([]Unit, error)
	CreateConversion(ctx context.Context, conversion UnitConversion) error
	CreateConversionFromName(ctx context.Context, input UnitConversionInput) error
	ConvertUnit(ctx context.Context, input ConvertUnitInput) (ConvertUnitOutput, error)
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

func (s *UnitService) CreateConversionFromName(ctx context.Context, input UnitConversionInput) error {
	unit, err := s.repo.GetUnitFromName(ctx, input.UnitName)
	if err != nil {
		return err
	}
	conversionUnit, err := s.repo.GetUnitFromName(ctx, input.ConversionUnitName)
	if err != nil {
		return err
	}
	conversion := UnitConversion{
		UnitId:           unit.Id,
		ConversionUnitId: conversionUnit.Id,
		ConversionFactor: input.ConversionFactor,
	}
	validationErr := ValidateUnitConversion(conversion)
	if validationErr != nil {
		return validationErr
	}
	return s.repo.AddUnitConversion(ctx, conversion)
}

func (s *UnitService) CreateConversion(ctx context.Context, conversion UnitConversion) error {
	validationErr := ValidateUnitConversion(conversion)
	if validationErr != nil {
		return validationErr
	}
	return s.repo.AddUnitConversion(ctx, conversion)
}

func (s *UnitService) ConvertUnit(ctx context.Context, input ConvertUnitInput) (ConvertUnitOutput, error) {
	unitConversion, err := s.repo.GetUnitConversionByUnitId(ctx, input.UnitId, input.ConversionUnitId)
	if err != nil {
		return ConvertUnitOutput{}, err
	}
	newQty := input.Quantity * unitConversion.ConversionFactor
	newUnit, err := s.repo.GetUnitById(ctx, unitConversion.ConversionUnitId)
	if err != nil {
		return ConvertUnitOutput{}, err
	}
	return ConvertUnitOutput{
		Unit:     newUnit,
		Quantity: newQty,
	}, nil
}
