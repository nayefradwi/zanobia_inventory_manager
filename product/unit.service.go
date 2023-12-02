package product

import (
	"context"
	"log"
	"strconv"

	"github.com/nayefradwi/zanobia_inventory_manager/common"
)

type IUnitService interface {
	CreateUnit(ctx context.Context, unit Unit) error
	GetAllUnits(ctx context.Context) ([]Unit, error)
	CreateConversion(ctx context.Context, conversion UnitConversion) error
	ConvertUnit(ctx context.Context, input ConvertUnitInput) (ConvertUnitOutput, error)
	GetUnitById(ctx context.Context, id *int) (Unit, error)
	TranslateUnit(ctx context.Context, unit Unit, languageCode string) error
	SetupUnitConversionsMap(ctx context.Context) error
	SetupUnitsMap(ctx context.Context) error
	GetUnitConversionKey(toUnitId int, fromUnitId int) string
}

type UnitService struct {
	repo               IUnitRepository
	unitConversionsMap map[string]UnitConversion
	unitsMap           map[int]Unit
}

func (s UnitService) GetUnitConversionKey(toUnitId int, fromUnitId int) string {
	return strconv.Itoa(toUnitId) + "-" + strconv.Itoa(fromUnitId)
}

func NewUnitService(repo IUnitRepository) IUnitService {
	return &UnitService{
		repo:               repo,
		unitConversionsMap: make(map[string]UnitConversion),
		unitsMap:           make(map[int]Unit),
	}
}

func (s *UnitService) CreateUnit(ctx context.Context, unit Unit) error {
	validationErr := ValidateUnit(unit)
	if validationErr != nil {
		return validationErr
	}
	id, err := s.repo.CreateUnit(ctx, unit)
	if err != nil {
		return err
	}
	unit.Id = &id
	s.unitsMap[id] = unit
	return nil
}

func (s *UnitService) TranslateUnit(ctx context.Context, unit Unit, languageCode string) error {
	validationErr := ValidateUnit(unit)
	if validationErr != nil {
		return validationErr
	}
	return s.repo.TranslateUnit(ctx, unit, languageCode)
}

func (s *UnitService) GetAllUnits(ctx context.Context) ([]Unit, error) {
	return s.repo.GetAllUnits(ctx)
}

func (s *UnitService) CreateConversion(ctx context.Context, conversion UnitConversion) error {
	validationErr := ValidateUnitConversion(conversion)
	if validationErr != nil {
		return validationErr
	}

	err := s.repo.AddUnitConversion(ctx, conversion)
	if err != nil {
		return err
	}
	key := s.GetUnitConversionKey(*conversion.ToUnitId, *conversion.FromUnitId)
	s.unitConversionsMap[key] = conversion
	return nil
}

func (s *UnitService) ConvertUnit(ctx context.Context, input ConvertUnitInput) (ConvertUnitOutput, error) {
	if (input.ToUnitId == nil && input.FromUnitId == nil) || input.Quantity == 0 {
		return ConvertUnitOutput{}, common.NewBadRequestFromMessage("Invalid unit conversion input")
	}
	if *input.ToUnitId == 0 || *input.FromUnitId == 0 {
		return ConvertUnitOutput{}, common.NewBadRequestFromMessage("Invalid unit conversion input")
	}
	if *input.ToUnitId == *input.FromUnitId {
		return s.getSameUnitOutput(ctx, *input.ToUnitId, input.Quantity)
	}
	return s.convertUsingMap(ctx, input)
}

func (s *UnitService) getSameUnitOutput(ctx context.Context, unitId int, quantity float64) (ConvertUnitOutput, error) {
	unit, err := s.GetUnitById(ctx, &unitId)
	if err != nil {
		return ConvertUnitOutput{}, err
	}
	return ConvertUnitOutput{Unit: unit, Quantity: quantity}, nil
}

func (s *UnitService) convertUsingMap(ctx context.Context, input ConvertUnitInput) (ConvertUnitOutput, error) {
	key := s.GetUnitConversionKey(*input.ToUnitId, *input.FromUnitId)
	unitConversion, ok := s.unitConversionsMap[key]
	if !ok {
		return s.convertUsingDatabase(ctx, input)
	}
	newQty := input.Quantity * unitConversion.ConversionFactor
	newUnit, err := s.GetUnitById(ctx, unitConversion.ToUnitId)
	if err != nil {
		return ConvertUnitOutput{}, err
	}
	return ConvertUnitOutput{Unit: newUnit, Quantity: newQty}, nil
}

func (s *UnitService) convertUsingDatabase(ctx context.Context, input ConvertUnitInput) (ConvertUnitOutput, error) {
	unitConversion, err := s.repo.GetUnitConversionByUnitId(ctx, input.ToUnitId, input.FromUnitId)
	if err != nil {
		return ConvertUnitOutput{}, err
	}
	key := s.GetUnitConversionKey(*unitConversion.ToUnitId, *unitConversion.FromUnitId)
	s.unitConversionsMap[key] = unitConversion
	newQty := input.Quantity * unitConversion.ConversionFactor
	newUnit, err := s.GetUnitById(ctx, unitConversion.ToUnitId)
	if err != nil {
		return ConvertUnitOutput{}, err
	}
	return ConvertUnitOutput{Unit: newUnit, Quantity: newQty}, nil
}

func (s *UnitService) GetUnitById(ctx context.Context, id *int) (Unit, error) {
	if id == nil {
		return Unit{}, common.NewBadRequestFromMessage("Invalid unit id")
	}
	if *id == 0 {
		return Unit{}, common.NewBadRequestFromMessage("Invalid unit id")
	}
	unit := s.unitsMap[*id]
	if unit.Id != nil {
		return unit, nil
	}
	unit, err := s.repo.GetUnitById(ctx, id)
	if err != nil {
		return Unit{}, err
	}
	s.unitsMap[*id] = unit
	return unit, nil
}

func (s *UnitService) SetupUnitConversionsMap(ctx context.Context) error {
	unitConversions, err := s.repo.GetUnitConversions(ctx)
	if err != nil {
		return err
	}
	unitConversionsMap := make(map[string]UnitConversion)
	for _, unitConversion := range unitConversions {
		key := s.GetUnitConversionKey(*unitConversion.ToUnitId, *unitConversion.FromUnitId)
		unitConversionsMap[key] = unitConversion
	}
	s.unitConversionsMap = unitConversionsMap
	log.Printf("unit conversions map: %v", s.unitConversionsMap)
	return nil
}

func (s *UnitService) SetupUnitsMap(ctx context.Context) error {
	units, err := s.repo.GetAllUnits(ctx)
	if err != nil {
		return err
	}
	unitsMap := make(map[int]Unit)
	for _, unit := range units {
		unitsMap[*unit.Id] = unit
	}
	s.unitsMap = unitsMap
	log.Printf("units map: %v", s.unitsMap)
	return nil
}
