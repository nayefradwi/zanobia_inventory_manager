package product

import (
	"context"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
	"github.com/nayefradwi/zanobia_inventory_manager/unit"
	zimutils "github.com/nayefradwi/zanobia_inventory_manager/zim_utils"
	"go.uber.org/zap"
)

func (r *ProductRepo) GetProductVariantsOfProduct(ctx context.Context, productId int) ([]ProductVariant, error) {
	sql := `
	select pvar.id, pvar.product_id, pvartx.name, pvar.sku, pvar.image, pvar.price, 
	pvar.is_archived, pvar.is_default from product_variants pvar 
	join product_variant_translations pvartx on pvartx.product_variant_id = pvar.id
	where pvar.product_id = $1 and pvartx.language_code = $2;
	`
	op := common.GetOperator(ctx, r.Pool)
	langCode := common.GetLanguageParam(ctx)
	rows, err := op.Query(ctx, sql, productId, langCode)
	if err != nil {
		common.LoggerFromCtx(ctx).Error("failed to get product variants", zap.Error(err))
		return []ProductVariant{}, common.NewBadRequestFromMessage("Failed to get product variants")
	}
	defer rows.Close()
	productVariants := make([]ProductVariant, 0)
	for rows.Next() {
		var productVariant ProductVariant
		err := rows.Scan(
			&productVariant.Id, &productVariant.ProductId, &productVariant.Name, &productVariant.Sku,
			&productVariant.Image, &productVariant.Price, &productVariant.IsArchived, &productVariant.IsDefault,
		)
		if err != nil {
			common.LoggerFromCtx(ctx).Error("failed to scan product variant", zap.Error(err))
			return []ProductVariant{}, common.NewInternalServerError()
		}
		productVariants = append(productVariants, productVariant)
	}

	return productVariants, nil
}

func (r *ProductRepo) GetProductVariant(ctx context.Context, productVariantId int) (ProductVariant, error) {
	sql := `
	select pvar.id, pvar.product_id, pvartx.name, pvar.sku, pvar.image, pvar.price,
	pvar.width_in_cm, pvar.height_in_cm, pvar.depth_in_cm, pvar.weight_in_g,
	pvar.is_archived, pvar.is_default, pvar.expires_in_days, 
	utx.unit_id, utx.name, utx.symbol, ptx.name product_name, p.is_ingredient
	from product_variants pvar 
	join product_variant_translations pvartx on pvar.id = pvartx.product_variant_id
	join unit_translations utx on utx.unit_id = pvar.standard_unit_id
	join product_translations ptx on ptx.product_id = pvar.product_id
	join products p on p.id = pvar.product_id
	where pvar.id = $1 and pvartx.language_code = $2
	`
	op := common.GetOperator(ctx, r.Pool)
	langCode := common.GetLanguageParam(ctx)
	row := op.QueryRow(ctx, sql, productVariantId, langCode)
	var productVariant ProductVariant
	var unit unit.Unit
	err := row.Scan(
		&productVariant.Id, &productVariant.ProductId, &productVariant.Name, &productVariant.Sku,
		&productVariant.Image, &productVariant.Price, &productVariant.WidthInCm, &productVariant.HeightInCm,
		&productVariant.DepthInCm, &productVariant.WeightInG, &productVariant.IsArchived, &productVariant.IsDefault,
		&productVariant.ExpiresInDays, &unit.Id, &unit.Name, &unit.Symbol, &productVariant.ProductName,
		&productVariant.IsIngredient,
	)
	if err != nil {
		common.LoggerFromCtx(ctx).Error("failed to scan product variant", zap.Error(err))
		return ProductVariant{}, common.NewBadRequestError("failed to get product variant", zimutils.GetErrorCodeFromError(err))
	}
	productVariant.StandardUnit = &unit
	return productVariant, nil
}

func (r *ProductRepo) AddProductVariant(ctx context.Context, input ProductVariantInput) error {
	err := common.RunWithTransaction(ctx, r.Pool, func(ctx context.Context, tx pgx.Tx) error {
		id, err := r.addProductVariant(ctx, input.ProductVariant.ProductId, input.ProductVariant)
		if err != nil {
			return err
		}
		input.ProductVariant.Id = &id
		return r.addProductVariantValues(ctx, input.ProductVariant, input.OptionValues)
	})
	return err
}

func (r *ProductRepo) GetUnitIdOfProductVariantBySku(ctx context.Context, sku string) (int, error) {
	sql := `
		select standard_unit_id from product_variants where sku = $1
	`
	op := common.GetOperator(ctx, r.Pool)
	row := op.QueryRow(ctx, sql, sku)
	var unitId int
	err := row.Scan(&unitId)
	if err != nil {
		common.LoggerFromCtx(ctx).Error("failed to scan unit id", zap.Error(err))
		return 0, common.NewBadRequestError("failed to get unit id", zimutils.GetErrorCodeFromError(err))
	}
	return unitId, nil
}

func (r *ProductRepo) GetProductVariantExpirationDateAndCost(ctx context.Context, sku string) (time.Time, float64, error) {
	sql := `
		select expires_in_days, price from product_variants where sku = $1
	`
	op := common.GetOperator(ctx, r.Pool)
	row := op.QueryRow(ctx, sql, sku)
	var expiresInDays int
	var price float64
	err := row.Scan(&expiresInDays, &price)
	if err != nil {
		common.LoggerFromCtx(ctx).Error("failed to scan expires in days", zap.Error(err))
		return time.Time{}, 0, common.NewBadRequestError("failed to get expires in days", zimutils.GetErrorCodeFromError(err))
	}
	return time.Now().AddDate(0, 0, expiresInDays), price, nil
}

func (r *ProductRepo) GetProductOptions(ctx context.Context, productId int) ([]ProductOption, error) {
	sql := `select id from product_options where product_id = $1 and language_code = $2`
	op := common.GetOperator(ctx, r.Pool)
	langCode := common.GetLanguageParam(ctx)
	rows, err := op.Query(ctx, sql, productId, langCode)
	if err != nil {
		common.LoggerFromCtx(ctx).Error("failed to get product options", zap.Error(err))
		return []ProductOption{}, common.NewBadRequestFromMessage("Failed to get product options")
	}
	defer rows.Close()
	productOptions := make([]ProductOption, 0)
	for rows.Next() {
		var productOption ProductOption
		err := rows.Scan(&productOption.Id)
		if err != nil {
			common.LoggerFromCtx(ctx).Error("failed to scan product option", zap.Error(err))
			return []ProductOption{}, common.NewInternalServerError()
		}
		productOptions = append(productOptions, productOption)
	}
	return productOptions, nil
}

func (r *ProductRepo) GetProductSelectedValues(ctx context.Context, productId int, optionValueIds []int) (map[string]ProductOptionValue, error) {
	sql := `
	select pvl.id, pvl.value, popt.name from product_option_values pvl 
	join product_options popt on pvl.product_option_id = popt.id 
	and popt.language_code = pvl.language_code
	where popt.product_id = $1 and popt.language_code = $2 and pvl.id = any($3);
	`
	op := common.GetOperator(ctx, r.Pool)
	langCode := common.GetLanguageParam(ctx)
	rows, err := op.Query(ctx, sql, productId, langCode, optionValueIds)
	if err != nil {
		common.LoggerFromCtx(ctx).Error("failed to get product selected values", zap.Error(err))
		return map[string]ProductOptionValue{}, common.NewBadRequestFromMessage("Failed to get product selected values")
	}
	defer rows.Close()
	productOptionValues := make(map[string]ProductOptionValue)
	for rows.Next() {
		var productOptionValue ProductOptionValue
		var optionName string
		err := rows.Scan(&productOptionValue.Id, &productOptionValue.Value, &optionName)
		if err != nil {
			common.LoggerFromCtx(ctx).Error("failed to scan product option value", zap.Error(err))
			return map[string]ProductOptionValue{}, common.NewInternalServerError()
		}
		productOptionValues[optionName] = productOptionValue
	}
	return productOptionValues, nil
}

func (r *ProductRepo) UpdateProductVariantDetails(ctx context.Context, update ProductVariantUpdate) error {
	sql := `
	update product_variants set price = $1, width_in_cm = $2, height_in_cm = $3, depth_in_cm = $4,
	weight_in_g = $5, is_archived = $6 where id = $7
	`
	op := common.GetOperator(ctx, r.Pool)
	_, err := op.Exec(ctx, sql, update.Price, update.WidthInCm,
		update.HeightInCm, update.DepthInCm, update.WeightInG,
		update.IsArchived, update.Id,
	)
	if err != nil {
		common.LoggerFromCtx(ctx).Error("failed to update product variant details", zap.Error(err))
		return common.NewInternalServerError()
	}
	return nil
}

func (r *ProductRepo) GetProductVariantSkuAndIsDefaultFromId(ctx context.Context, id int) (string, bool, error) {
	sql := `
	select sku, is_default from product_variants where id = $1
	`
	op := common.GetOperator(ctx, r.Pool)
	row := op.QueryRow(ctx, sql, id)
	var sku string
	var isDefault bool
	err := row.Scan(&sku, &isDefault)
	if err != nil {
		common.LoggerFromCtx(ctx).Error("failed to scan product variant sku", zap.Error(err))
		return "", false, common.NewInternalServerError()
	}
	return sku, isDefault, nil
}

func (r *ProductRepo) DeleteProductVariant(ctx context.Context, id int, sku string) error {
	batch := r.createDeleteProductVariantBatch(id, sku)
	op := common.GetOperator(ctx, r.Pool)
	results := op.SendBatch(ctx, batch)
	defer results.Close()
	_, deleteVariantTransationsErr := results.Exec()
	if deleteVariantTransationsErr != nil {
		common.LoggerFromCtx(ctx).Error("failed to delete product variant translations", zap.Error(deleteVariantTransationsErr))
		return common.NewBadRequestFromMessage("Failed to delete product variant translations")
	}
	_, deleteVariantValuesErr := results.Exec()
	if deleteVariantValuesErr != nil {
		common.LoggerFromCtx(ctx).Error("failed to delete product variant values", zap.Error(deleteVariantValuesErr))
		return common.NewBadRequestFromMessage("Failed to delete product variant values")
	}
	_, deleteRecipesErr := results.Exec()
	if deleteRecipesErr != nil {
		common.LoggerFromCtx(ctx).Error("failed to delete recipes", zap.Error(deleteRecipesErr))
		return common.NewBadRequestFromMessage("Failed to delete recipes")
	}
	_, deleteBatchesErr := results.Exec()
	if deleteBatchesErr != nil {
		common.LoggerFromCtx(ctx).Error("failed to delete batches", zap.Error(deleteBatchesErr))
		return common.NewBadRequestFromMessage("Failed to delete batches")
	}
	_, deleteRetailerBatchesErr := results.Exec()
	if deleteRetailerBatchesErr != nil {
		common.LoggerFromCtx(ctx).Error("failed to delete retailer batches", zap.Error(deleteRetailerBatchesErr))
		return common.NewBadRequestFromMessage("Failed to delete retailer batches")
	}
	_, deleteErr := results.Exec()
	if deleteErr != nil {
		common.LoggerFromCtx(ctx).Error("failed to delete product variant", zap.Error(deleteErr))
		return common.NewBadRequestFromMessage("Failed to delete product variant")
	}
	return nil
}

func (r *ProductRepo) createDeleteProductVariantBatch(id int, sku string) *pgx.Batch {
	batch := &pgx.Batch{}
	r.addToDeleteVariantQueriesToBatch(batch, id, sku)
	return batch
}

func (r *ProductRepo) addToDeleteVariantQueriesToBatch(batch *pgx.Batch, id int, sku string) {
	batch.Queue("delete from product_variant_translations where product_variant_id = $1", id)
	batch.Queue("delete from product_variant_values where product_variant_id = $1", id)
	batch.Queue("delete from recipes where recipe_variant_sku = $1 OR result_variant_sku = $1", sku)
	batch.Queue("delete from batches where sku = $1", sku)
	batch.Queue("delete from retailer_batches where sku = $1", sku)
	batch.Queue("delete from product_variants where id = $1 RETURNING is_default", id)
}

func (r *ProductRepo) UpdateProductVariantSku(ctx context.Context, oldSku, newSku string) error {
	sql := `
	update product_variants set sku = $1 where sku = $2
	`
	op := common.GetOperator(ctx, r.Pool)
	_, err := op.Exec(ctx, sql, newSku, oldSku)
	if err != nil {
		common.LoggerFromCtx(ctx).Error("failed to update product variant sku", zap.Error(err))
		return common.NewBadRequestFromMessage("Failed to update product variant sku")
	}
	return nil
}

func (r *ProductRepo) UpdateProductVariantArchiveStatus(ctx context.Context, id int, isArchived bool) error {
	sql := `update product_variants set is_archived = $1 where id = $2`
	op := common.GetOperator(ctx, r.Pool)
	_, err := op.Exec(ctx, sql, isArchived, id)
	if err != nil {
		common.LoggerFromCtx(ctx).Error("failed to archive product variant", zap.Error(err))
		return common.NewBadRequestFromMessage("Failed to archive product variant")
	}
	return nil
}
