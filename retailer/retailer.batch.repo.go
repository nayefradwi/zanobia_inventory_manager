package retailer

import (
	"context"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
	"github.com/nayefradwi/zanobia_inventory_manager/product"
	"github.com/nayefradwi/zanobia_inventory_manager/unit"
	"go.uber.org/zap"
)

type IRetailerBatchRepository interface {
	CreateRetailerBatch(ctx context.Context, input RetailerBatchInput, expiresAt string) (int, error)
	UpdateRetailerBatch(ctx context.Context, base RetailerBatchBase) error
	GetRetailerBatchBaseById(ctx context.Context, id *int, retailerId int) (RetailerBatchBase, error)
	GetRetailerBatchBase(ctx context.Context, retailerId int, sku string, expirationDate string) (RetailerBatchBase, error)
	GetRetailerBatches(ctx context.Context, retailerId int, params common.PaginationParams) ([]RetailerBatch, error)
	SearchRetailerBatchesBySku(ctx context.Context, retailerId int, sku string, params common.PaginationParams) ([]RetailerBatch, error)
	DeleteBatchesOfRetailer(ctx context.Context, retailerId int) error
}

type RetailerBatchRepository struct {
	*pgxpool.Pool
}

func NewRetailerBatchRepository(db *pgxpool.Pool) *RetailerBatchRepository {
	return &RetailerBatchRepository{
		db,
	}
}

func (r *RetailerBatchRepository) CreateRetailerBatch(ctx context.Context, input RetailerBatchInput, expiresAt string) (int, error) {
	sql := `INSERT INTO retailer_batches (sku, retailer_id, quantity, unit_id, expires_at) VALUES ($1, $2, $3, $4, $5) RETURNING id`
	op := common.GetOperator(ctx, r.Pool)
	row := op.QueryRow(ctx, sql, input.Sku, input.RetailerId, input.Quantity, input.UnitId, expiresAt)
	var id int
	err := row.Scan(&id)
	if err != nil {
		common.LoggerFromCtx(ctx).Error("Failed to create retailer batch", zap.Error(err))
		return 0, common.NewBadRequestFromMessage("Failed to create retailer batch")
	}
	return id, nil
}

func (r *RetailerBatchRepository) UpdateRetailerBatch(ctx context.Context, base RetailerBatchBase) error {
	updatedAt := time.Now().UTC()
	op := common.GetOperator(ctx, r.Pool)
	sql := `UPDATE retailer_batches SET quantity = $1, updated_at = $2 WHERE id = $3 and retailer_id = $4`
	_, err := op.Exec(ctx, sql, base.Quantity, updatedAt, base.Id, base.RetailerId)
	if err != nil {
		common.LoggerFromCtx(ctx).Error("Failed to update retailer batch", zap.Error(err))
		return common.NewBadRequestFromMessage("Failed to update retailer batch")
	}
	return nil
}

func (r *RetailerBatchRepository) GetRetailerBatchBaseById(ctx context.Context, id *int, retailerId int) (RetailerBatchBase, error) {
	op := common.GetOperator(ctx, r.Pool)
	sql := `SELECT id, sku, quantity, unit_id, expires_at, retailer_id FROM retailer_batches WHERE id = $1 AND retailer_id = $2`
	row := op.QueryRow(ctx, sql, id, retailerId)
	var retailerBatchBase RetailerBatchBase
	err := row.Scan(
		&retailerBatchBase.Id, &retailerBatchBase.Sku, &retailerBatchBase.Quantity,
		&retailerBatchBase.UnitId, &retailerBatchBase.ExpiresAt, &retailerBatchBase.RetailerId,
	)
	if err != nil {
		common.LoggerFromCtx(ctx).Error("Failed to get retailer batch base", zap.Error(err))
		return RetailerBatchBase{}, common.NewBadRequestFromMessage("Failed to get retailer batch base")
	}
	return retailerBatchBase, nil
}

func (r *RetailerBatchRepository) GetRetailerBatchBase(
	ctx context.Context, retailerId int,
	sku string, expirationDate string,
) (RetailerBatchBase, error) {
	op := common.GetOperator(ctx, r.Pool)
	sql := `
		SELECT id, sku, quantity, unit_id, expires_at, retailer_id
		FROM retailer_batches
		WHERE sku = $1 AND expires_at = $2 and retailer_id = $3
	`
	row := op.QueryRow(ctx, sql, sku, expirationDate, retailerId)
	var retailerBatchBase RetailerBatchBase
	err := row.Scan(
		&retailerBatchBase.Id, &retailerBatchBase.Sku,
		&retailerBatchBase.Quantity, &retailerBatchBase.UnitId,
		&retailerBatchBase.ExpiresAt, &retailerBatchBase.RetailerId,
	)
	if err != nil {
		common.LoggerFromCtx(ctx).Error("Failed to get retailer batch base", zap.Error(err))
		return RetailerBatchBase{}, common.NewBadRequestFromMessage("Failed to get retailer batch base")
	}
	return retailerBatchBase, nil
}

func (r *RetailerBatchRepository) GetRetailerBatches(ctx context.Context, retailerId int, params common.PaginationParams) ([]RetailerBatch, error) {
	op := common.GetOperator(ctx, r.Pool)
	lang := common.GetLanguageParam(ctx)
	rows, err := common.NewPaginationQueryBuilder(
		baseBatchListingSql,
		[]string{"b.expires_at DESC", "b.id DESC"},
	).
		WithOperator(op).
		WithConditions([]string{
			"utx.language_code = $1",
			"AND",
			"b.retailer_id = $2",
		}).
		WithCursorKeys([]string{"b.expires_at", "b.id"}).
		WithParams(params).
		WithCompareSymbols("<", "<=", ">").
		Build().
		Query(ctx, lang, retailerId)
	if err != nil {
		common.LoggerFromCtx(ctx).Error("Failed to get batches", zap.Error(err))
		return nil, common.NewBadRequestFromMessage("Failed to get batches")
	}
	defer rows.Close()
	return r.parseRetailerBatchRows(rows)
}

func (r *RetailerBatchRepository) SearchRetailerBatchesBySku(
	ctx context.Context, retailerId int, sku string,
	params common.PaginationParams,
) ([]RetailerBatch, error) {
	op := common.GetOperator(ctx, r.Pool)
	lang := common.GetLanguageParam(ctx)
	rows, err := common.NewPaginationQueryBuilder(
		baseBatchListingSql,
		[]string{"b.expires_at DESC", "b.id DESC"},
	).
		WithOperator(op).
		WithConditions([]string{
			"utx.language_code = $1",
			"AND",
			"b.retailer_id = $2",
			"AND",
			"b.sku = $3",
		}).
		WithCursorKeys([]string{"b.expires_at", "b.id"}).
		WithParams(params).
		WithCompareSymbols("<", "<=", ">").
		Build().
		Query(ctx, lang, retailerId, sku)
	if err != nil {
		common.LoggerFromCtx(ctx).Error("Failed to get retailer batches", zap.Error(err))
		return nil, common.NewBadRequestFromMessage("Failed to get retailer batches")
	}
	defer rows.Close()
	return r.parseRetailerBatchRows(rows)
}

const baseBatchListingSql = `
select b.id, b.sku, b.quantity, b.expires_at, utx.unit_id, utx.name, utx.symbol,
pvartx.name, pvar.id, pvar.price, pvar.product_id, ptx.name,
rtx.retailer_id, rtx.name
from retailer_batches b
join unit_translations utx on utx.unit_id = b.unit_id
join product_variants pvar on pvar.sku = b.sku
join product_translations ptx on ptx.product_id = pvar.product_id
join product_variant_translations pvartx on pvartx.product_variant_id = pvar.id and utx.language_code = pvartx.language_code
join retailer_translations rtx on rtx.retailer_id = b.retailer_id and utx.language_code = rtx.language_code
`

func (r *RetailerBatchRepository) parseRetailerBatchRows(rows pgx.Rows) ([]RetailerBatch, error) {
	var retailerBatches []RetailerBatch
	for rows.Next() {
		var retailerBatch RetailerBatch
		var productVariantBase product.ProductVariantBase
		var unit unit.Unit
		err := rows.Scan(
			&retailerBatch.Id, &retailerBatch.Sku, &retailerBatch.Quantity, &retailerBatch.ExpiresAt,
			&unit.Id, &unit.Name, &unit.Symbol,
			&productVariantBase.Name, &productVariantBase.Id, &productVariantBase.Price,
			&productVariantBase.ProductId, &retailerBatch.ProductName, &retailerBatch.RetailerId, &retailerBatch.RetailerName,
		)
		if err != nil {
			common.GetLogger().Error("Failed to scan retailer batches", zap.Error(err))
			return []RetailerBatch{}, common.NewBadRequestFromMessage("Failed to scan retailer batches")
		}
		retailerBatch.Unit = unit
		retailerBatch.ProductVariantBase = &productVariantBase
		retailerBatches = append(retailerBatches, retailerBatch)
	}
	return retailerBatches, nil
}

func (r *RetailerBatchRepository) DeleteBatchesOfRetailer(ctx context.Context, retailerId int) error {
	op := common.GetOperator(ctx, r.Pool)
	sql := `DELETE FROM retailer_batches WHERE retailer_id = $1`
	_, err := op.Exec(ctx, sql, retailerId)
	if err != nil {
		common.LoggerFromCtx(ctx).Error("Failed to delete retailer batches", zap.Error(err))
		return common.NewBadRequestFromMessage("Failed to delete retailer batches")
	}
	return nil
}
