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
