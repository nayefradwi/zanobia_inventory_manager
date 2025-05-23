package unit

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
	zimutils "github.com/nayefradwi/zanobia_inventory_manager/zim_utils"
	"go.uber.org/zap"
)

type IUnitRepository interface {
	GetAllUnits(ctx context.Context) ([]Unit, error)
	CreateUnit(ctx context.Context, unit Unit) (int, error)
	GetUnitFromName(ctx context.Context, name string) (Unit, error)
	AddUnitConversion(ctx context.Context, conversion UnitConversion) error
	GetUnitById(ctx context.Context, id *int) (Unit, error)
	GetUnitConversionByUnitId(ctx context.Context, toUnitId *int, fromUnitId *int) (UnitConversion, error)
	TranslateUnit(ctx context.Context, unit Unit, languageCode string) error
	GetUnitConversions(ctx context.Context) ([]UnitConversion, error)
}

type UnitRepository struct {
	*pgxpool.Pool
}

func NewUnitRepository(dbPool *pgxpool.Pool) *UnitRepository {
	return &UnitRepository{dbPool}
}

func (r *UnitRepository) CreateUnit(ctx context.Context, unit Unit) (int, error) {
	var id int
	err := common.RunWithTransaction(ctx, r.Pool, func(ctx context.Context, tx pgx.Tx) error {
		var addErr error
		id, addErr = r.addUnit(ctx)
		if addErr != nil {
			return addErr
		}
		unit.Id = &id
		translationErr := r.insertTranslation(ctx, unit, common.DefaultLang)
		if translationErr != nil {
			return translationErr
		}
		return nil
	})
	return id, err
}

func (r *UnitRepository) TranslateUnit(ctx context.Context, unit Unit, languageCode string) error {
	return r.insertTranslation(ctx, unit, languageCode)
}

func (r *UnitRepository) insertTranslation(ctx context.Context, unit Unit, languageCode string) error {
	sql := `INSERT INTO unit_translations (unit_id, name, symbol, language_code) VALUES ($1, $2, $3, $4)`
	op := common.GetOperator(ctx, r.Pool)
	c, err := op.Exec(ctx, sql, unit.Id, unit.Name, unit.Symbol, languageCode)
	if err != nil {
		common.LoggerFromCtx(ctx).Error("failed to translate unit", zap.Error(err))
		return common.NewBadRequestError("Failed to translate unit", zimutils.GetErrorCodeFromError(err))
	}
	if c.RowsAffected() == 0 {
		return common.NewInternalServerError()
	}
	return nil
}

func (r *UnitRepository) addUnit(ctx context.Context) (int, error) {
	sql := `INSERT INTO units DEFAULT VALUES RETURNING id`
	op := common.GetOperator(ctx, r.Pool)
	row := op.QueryRow(ctx, sql)
	var id int
	err := row.Scan(&id)
	if err != nil {
		common.LoggerFromCtx(ctx).Error("failed to scan unit", zap.Error(err))
		return 0, common.NewInternalServerError()
	}
	return id, nil
}

func (r *UnitRepository) GetAllUnits(ctx context.Context) ([]Unit, error) {
	sql := `SELECT u.id, name, symbol FROM units u JOIN unit_translations utx on u.id = utx.unit_id where language_code = $1`
	languageCode := common.GetLanguageParam(ctx)
	rows, err := r.Query(ctx, sql, languageCode)
	if err != nil {
		common.LoggerFromCtx(ctx).Error("failed to get units", zap.Error(err))
		return nil, common.NewInternalServerError()
	}
	defer rows.Close()
	units := make([]Unit, 0)
	for rows.Next() {
		var unit Unit
		err := rows.Scan(&unit.Id, &unit.Name, &unit.Symbol)
		if err != nil {
			common.LoggerFromCtx(ctx).Error("failed to scan unit", zap.Error(err))
			return nil, common.NewInternalServerError()
		}
		units = append(units, unit)
	}
	return units, nil
}

func (r *UnitRepository) GetUnitFromName(ctx context.Context, name string) (Unit, error) {
	sql := `SELECT u.id, name, symbol FROM units u JOIN unit_translations utx on u.id = utx.unit_id WHERE name = $1 and language_code = $2`
	languageCode := common.GetLanguageParam(ctx)
	row := r.QueryRow(ctx, sql, name, languageCode)
	var unit Unit
	err := row.Scan(&unit.Id, &unit.Name, &unit.Symbol)
	if err != nil {
		common.LoggerFromCtx(ctx).Error("failed to scan unit", zap.Error(err))
		return Unit{}, common.NewBadRequestError("Failed to get unit", zimutils.GetErrorCodeFromError(err))
	}
	if unit.Id == nil {
		return Unit{}, common.NewNotFoundError("Unit not found")
	}
	return unit, nil
}

func (r *UnitRepository) AddUnitConversion(ctx context.Context, conversion UnitConversion) error {
	sql := `INSERT INTO unit_conversions (to_unit_id, from_unit_id, conversion_factor) VALUES ($1, $2, $3)`
	c, err := r.Exec(ctx, sql, conversion.ToUnitId, conversion.FromUnitId, conversion.ConversionFactor)
	if err != nil {
		common.LoggerFromCtx(ctx).Error("failed to create unit conversion", zap.Error(err))
		return common.NewBadRequestError("Failed to create unit conversion", zimutils.GetErrorCodeFromError(err))
	}
	if c.RowsAffected() == 0 {
		return common.NewInternalServerError()
	}
	return nil
}

func (r *UnitRepository) GetUnitById(ctx context.Context, id *int) (Unit, error) {
	sql := `SELECT u.id, name, symbol FROM units u JOIN unit_translations utx on u.id = utx.unit_id WHERE u.id = $1 AND language_code = $2`
	languageCode := common.GetLanguageParam(ctx)
	row := r.QueryRow(ctx, sql, id, languageCode)
	var unit Unit
	err := row.Scan(&unit.Id, &unit.Name, &unit.Symbol)
	if err != nil {
		common.LoggerFromCtx(ctx).Error("failed to scan unit", zap.Error(err))
		return Unit{}, common.NewBadRequestError("Failed to get unit", zimutils.GetErrorCodeFromError(err))
	}
	if unit.Id == nil {
		return Unit{}, common.NewNotFoundError("Unit not found")
	}
	return unit, nil
}

func (r *UnitRepository) GetUnitConversionByUnitId(ctx context.Context, toUnitId *int, fromUnitId *int) (UnitConversion, error) {
	sql := `SELECT id, to_unit_id, from_unit_id, conversion_factor FROM unit_conversions WHERE to_unit_id = $1 AND from_unit_id = $2`
	op := common.GetOperator(ctx, r.Pool)
	row := op.QueryRow(ctx, sql, toUnitId, fromUnitId)
	var conversion UnitConversion
	err := row.Scan(&conversion.Id, &conversion.ToUnitId, &conversion.FromUnitId, &conversion.ConversionFactor)
	if err != nil {
		common.LoggerFromCtx(ctx).Error("failed to scan unit conversion", zap.Error(err))
		return UnitConversion{}, common.NewBadRequestError("Failed to get unit conversion", zimutils.GetErrorCodeFromError(err))
	}
	if conversion.Id == nil {
		return UnitConversion{}, common.NewNotFoundError("Unit conversion not found")
	}
	return conversion, nil
}

func (r *UnitRepository) GetUnitConversions(ctx context.Context) ([]UnitConversion, error) {
	sql := `SELECT id, to_unit_id, from_unit_id, conversion_factor FROM unit_conversions`
	op := common.GetOperator(ctx, r.Pool)
	rows, err := op.Query(ctx, sql)
	if err != nil {
		common.LoggerFromCtx(ctx).Error("failed to get unit conversions", zap.Error(err))
		return nil, common.NewInternalServerError()
	}
	defer rows.Close()
	conversions := make([]UnitConversion, 0)
	for rows.Next() {
		var conversion UnitConversion
		err := rows.Scan(&conversion.Id, &conversion.ToUnitId, &conversion.FromUnitId, &conversion.ConversionFactor)
		if err != nil {
			common.LoggerFromCtx(ctx).Error("failed to scan unit conversion", zap.Error(err))
			return nil, common.NewInternalServerError()
		}
		conversions = append(conversions, conversion)
	}
	return conversions, nil
}
