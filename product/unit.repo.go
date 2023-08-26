package product

import (
	"context"
	"log"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
	zimutils "github.com/nayefradwi/zanobia_inventory_manager/zim_utils"
)

type IUnitRepository interface {
	GetAllUnits(ctx context.Context) ([]Unit, error)
	CreateUnit(ctx context.Context, unit Unit) error
	GetUnitFromName(ctx context.Context, name string) (Unit, error)
	AddUnitConversion(ctx context.Context, conversion UnitConversion) error
}

type UnitRepository struct {
	*pgxpool.Pool
}

func NewUnitRepository(dbPool *pgxpool.Pool) *UnitRepository {
	return &UnitRepository{dbPool}
}

func (r *UnitRepository) CreateUnit(ctx context.Context, unit Unit) error {
	sql := `INSERT INTO units (name, symbol) VALUES ($1, $2)`
	c, err := r.Exec(ctx, sql, unit.Name, unit.Symbol)
	if err != nil {
		log.Printf("failed to create unit: %s", err.Error())
		return common.NewBadRequestError("Failed to create unit", zimutils.GetErrorCodeFromError(err))
	}
	if c.RowsAffected() == 0 {
		return common.NewInternalServerError()
	}
	return nil
}

func (r *UnitRepository) GetAllUnits(ctx context.Context) ([]Unit, error) {
	sql := `SELECT id, name, symbol FROM units`
	rows, err := r.Query(ctx, sql)
	if err != nil {
		log.Printf("failed to get units: %s", err.Error())
		return nil, common.NewInternalServerError()
	}
	defer rows.Close()
	var units []Unit
	for rows.Next() {
		var unit Unit
		err := rows.Scan(&unit.Id, &unit.Name, &unit.Symbol)
		if err != nil {
			log.Printf("failed to scan unit: %s", err.Error())
			return nil, common.NewInternalServerError()
		}
		units = append(units, unit)
	}
	return units, nil
}

func (r *UnitRepository) GetUnitFromName(ctx context.Context, name string) (Unit, error) {
	sql := `SELECT id, name, symbol FROM units WHERE name = $1`
	row := r.QueryRow(ctx, sql, name)
	var unit Unit
	err := row.Scan(&unit.Id, &unit.Name, &unit.Symbol)
	if err != nil {
		log.Printf("failed to scan unit: %s", err.Error())
		return Unit{}, common.NewBadRequestError("Failed to get unit", zimutils.GetErrorCodeFromError(err))
	}
	if unit.Id == nil {
		return Unit{}, common.NewNotFoundError("Unit not found")
	}
	return unit, nil
}

func (r *UnitRepository) AddUnitConversion(ctx context.Context, conversion UnitConversion) error {
	sql := `INSERT INTO unit_conversions (unit_id, conversion_unit_id, conversion_factor) VALUES ($1, $2, $3)`
	c, err := r.Exec(ctx, sql, conversion.UnitId, conversion.ConversionUnitId, conversion.ConversionFactor)
	if err != nil {
		log.Printf("failed to create unit conversion: %s", err.Error())
		return common.NewBadRequestError("Failed to create unit conversion", zimutils.GetErrorCodeFromError(err))
	}
	if c.RowsAffected() == 0 {
		return common.NewInternalServerError()
	}
	return nil
}
