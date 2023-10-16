package product

import (
	"context"
	"log"

	"github.com/jackc/pgx/v4"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
	zimutils "github.com/nayefradwi/zanobia_inventory_manager/zim_utils"
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
		log.Printf("failed to get product variants: %s", err.Error())
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
			log.Printf("failed to scan product variant: %s", err.Error())
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
	utx.unit_id, utx.name, utx.symbol, ptx.name product_name
	from product_variants pvar 
	join product_variant_translations pvartx on pvar.id = pvartx.product_variant_id
	join unit_translations utx on utx.unit_id = pvar.standard_unit_id
	join product_translations ptx on ptx.product_id = pvar.product_id
	where pvar.id = $1 and pvartx.language_code = $2
	`
	op := common.GetOperator(ctx, r.Pool)
	langCode := common.GetLanguageParam(ctx)
	row := op.QueryRow(ctx, sql, productVariantId, langCode)
	var productVariant ProductVariant
	var unit Unit
	err := row.Scan(
		&productVariant.Id, &productVariant.ProductId, &productVariant.Name, &productVariant.Sku,
		&productVariant.Image, &productVariant.Price, &productVariant.WidthInCm, &productVariant.HeightInCm,
		&productVariant.DepthInCm, &productVariant.WeightInG, &productVariant.IsArchived, &productVariant.IsDefault,
		&productVariant.ExpiresInDays, &unit.Id, &unit.Name, &unit.Symbol, &productVariant.ProductName,
	)
	if err != nil {
		log.Printf("failed to scan product variant: %s", err.Error())
		return ProductVariant{}, common.NewBadRequestError("failed to get product variant", zimutils.GetErrorCodeFromError(err))
	}
	productVariant.StandardUnit = &unit
	return productVariant, nil
}

func (r *ProductRepo) AddProductVariant(ctx context.Context, input ProductVariantInput) error {
	err := common.RunWithTransaction(ctx, r.Pool, func(ctx context.Context, tx pgx.Tx) error {
		ctx = common.SetOperator(ctx, tx)
		id, err := r.addProductVariant(ctx, input.ProductVariant.ProductId, input.ProductVariant)
		if err != nil {
			return err
		}
		input.ProductVariant.Id = &id
		return r.addProductVariantValues(ctx, input.ProductVariant, input.VariantValues)
	})
	return err
}
