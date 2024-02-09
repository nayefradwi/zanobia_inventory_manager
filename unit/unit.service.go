package unit

import (
	"context"
	"strconv"

	"github.com/nayefradwi/zanobia_inventory_manager/common"
	"go.uber.org/zap"
)

type IUnitService interface {
	InitiateAll(ctx context.Context) error
	CreateUnit(ctx context.Context, unit Unit) error
	GetAllUnits(ctx context.Context) ([]Unit, error)
	CreateConversion(ctx context.Context, conversion UnitConversion) error
	ConvertUnit(ctx context.Context, input ConvertUnitInput) (ConvertUnitOutput, error)
	GetUnitById(ctx context.Context, id *int) (Unit, error)
	TranslateUnit(ctx context.Context, unit Unit, languageCode string) error
	SetupUnitConversionsMap(ctx context.Context) error
	SetupUnitsMap(ctx context.Context) error
	GetUnitConversionKey(toUnitId int, fromUnitId int) string
	GetAllUnitConversions(ctx context.Context) map[string]UnitConversion
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

func (s UnitService) GetAllUnitConversions(ctx context.Context) map[string]UnitConversion {
	return s.unitConversionsMap
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
	common.LoggerFromCtx(ctx).Info(
		"adding unit to cached map",
		zap.String("unit_string", unit.Name),
	)
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
	common.LoggerFromCtx(ctx).Info(
		"adding unit conversion to cached map",
		zap.String("key", key),
	)
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
	common.LoggerFromCtx(ctx).Info("converting using cached unit conversions map")
	newQty := input.Quantity * unitConversion.ConversionFactor
	newUnit, err := s.GetUnitById(ctx, unitConversion.ToUnitId)
	if err != nil {
		return ConvertUnitOutput{}, err
	}
	return ConvertUnitOutput{Unit: newUnit, Quantity: newQty}, nil
}

func (s *UnitService) convertUsingDatabase(ctx context.Context, input ConvertUnitInput) (ConvertUnitOutput, error) {
	common.LoggerFromCtx(ctx).Info("converting using database; cache miss")
	unitConversion, err := s.repo.GetUnitConversionByUnitId(ctx, input.ToUnitId, input.FromUnitId)
	if err != nil {
		return ConvertUnitOutput{}, err
	}
	key := s.GetUnitConversionKey(*input.ToUnitId, *input.FromUnitId)
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
		common.LoggerFromCtx(ctx).Info(
			"unit found in cached map",
			zap.String("unit_string", unit.Name),
		)
		return unit, nil
	}
	common.LoggerFromCtx(ctx).Info("unit not found in cached map, fetching from database")
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
	common.LoggerFromCtx(ctx).Info(
		"unit conversions map setup",
		zap.Any("unit_conversions_map", s.unitConversionsMap),
	)
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
	common.LoggerFromCtx(ctx).Info(
		"units map setup",
		zap.Any("units_map", s.unitsMap),
	)
	return nil
}

func (s *UnitService) InitiateAll(ctx context.Context) error {
	s.initiateAllUnits(ctx)
	s.initiateAllUnitConversions(ctx)
	s.SetupUnitsMap(ctx)
	s.SetupUnitConversionsMap(ctx)
	return nil
}

func (s *UnitService) initiateAllUnits(ctx context.Context) error {
	for _, unit := range initialUnits {
		s.CreateUnit(ctx, unit)
	}
	return nil
}

func (s *UnitService) initiateAllUnitConversions(ctx context.Context) error {
	for _, input := range initialConversions {
		toUnit, err := s.repo.GetUnitFromName(ctx, input.ToUnitName)
		if err != nil {
			common.GetLogger().Error("failed to get unit from name", zap.String("unit", input.ToUnitName), zap.Error(err))
			continue
		}
		fromUnit, err := s.repo.GetUnitFromName(ctx, input.FromUnitName)
		if err != nil {
			common.GetLogger().Error("failed to get unit from name", zap.String("unit", input.FromUnitName), zap.Error(err))
			continue
		}
		unitConversion := UnitConversion{
			ToUnitId:         toUnit.Id,
			FromUnitId:       fromUnit.Id,
			ConversionFactor: input.ConversionFactor,
		}
		s.CreateConversion(ctx, unitConversion)
	}
	return nil
}
