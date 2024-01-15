package product

import (
	"context"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
	zimutils "github.com/nayefradwi/zanobia_inventory_manager/zim_utils"
	"go.uber.org/zap"
)

type IProductRepo interface {
	CreateProduct(ctx context.Context, product ProductInput) error
	AddProductVariant(ctx context.Context, input ProductVariantInput) error
	TranslateProduct(ctx context.Context, product ProductInput, languageCode string) error
	GetProducts(ctx context.Context, paginationParams common.PaginationParams, isArchive, isIngredient bool) ([]ProductBase, error)
	GetProduct(ctx context.Context, id int) (Product, error)
	GetProductVariantsOfProduct(ctx context.Context, productId int) ([]ProductVariant, error)
	GetProductVariant(ctx context.Context, productVariantId int) (ProductVariant, error)
	GetUnitIdOfProductVariantBySku(ctx context.Context, sku string) (int, error)
	GetProductVariantExpirationDateAndCost(ctx context.Context, sku string) (time.Time, float64, error)
	GetProductOptions(ctx context.Context, productId int) ([]ProductOption, error)
	GetProductSelectedValues(ctx context.Context, productId int, optionValueIds []int) (map[string]ProductOptionValue, error)
	InsertProductOptionValue(ctx context.Context, optionId int, optionValue ProductOptionValue) (int, error)
	UpdateProductVariantDetails(ctx context.Context, update ProductVariantUpdate) error
	GetProductVariantSkuAndIsDefaultFromId(ctx context.Context, id int) (string, bool, error)
	DeleteProductVariant(ctx context.Context, id int, sku string) error
	UpdateProductVariantSku(ctx context.Context, oldSku, newSku string) error
	GetOriginalUnitsBySkuList(ctx context.Context, skuList []string) (map[string]int, error)
	DeleteProduct(ctx context.Context, product Product) error
	UpdateProductArchiveStatus(ctx context.Context, id int, isArchived bool) error
	UpdateProductVariantArchiveStatus(ctx context.Context, id int, isArchived bool) error
	GetProductVariantBySku(ctx context.Context, sku string) (ProductVariant, error)
	SearchProductVariantsByName(ctx context.Context, paginationParams common.PaginationParams, name string) ([]ProductVariant, error)
	AddProductOption(ctx context.Context, input ProductOptionInput) error
}

type ProductRepo struct {
	*pgxpool.Pool
}

func NewProductRepository(dbPool *pgxpool.Pool) IProductRepo {
	return &ProductRepo{dbPool}
}

func (r *ProductRepo) CreateProduct(ctx context.Context, product ProductInput) error {
	err := common.RunWithTransaction(ctx, r.Pool, func(ctx context.Context, tx pgx.Tx) error {
		return r.createProduct(ctx, product)
	})
	return err
}

// TODO: this will probably need to be optimized
func (r *ProductRepo) createProduct(ctx context.Context, product ProductInput) error {
	id, addErr := r.addProduct(ctx, product)
	if addErr != nil {
		return addErr
	}
	product.Id = &id
	if translationErr := r.insertTranslation(ctx, product, common.DefaultLang); translationErr != nil {
		return translationErr
	}
	product = product.GenerateProductDetails()
	optionValueIdMap, err := r.addProductOptions(ctx, product.Id, product.Options)
	if err != nil {
		return err
	}

	variantIdsLookup, addVariantsErr := r.addProductVariants(ctx, product.Id, product.ProductVariants)
	if addVariantsErr != nil {
		return addVariantsErr
	}
	for sku, variantId := range variantIdsLookup {
		values := product.skuOptionValuesLookup[sku]
		for _, value := range values {
			valueId := optionValueIdMap[value.Value]
			if err := r.addProductVariantSelectedValue(ctx, &variantId, valueId); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *ProductRepo) addProduct(ctx context.Context, product ProductInput) (int, error) {
	sql := `INSERT INTO products (image, category_id, is_archived, is_ingredient)
			VALUES ($1, $2, $3, $4) RETURNING id`
	var id int
	op := common.GetOperator(ctx, r.Pool)
	err := op.QueryRow(ctx, sql, product.Image, product.CategoryId, product.IsArchived, product.IsIngredient).Scan(&id)
	if err != nil {
		common.LoggerFromCtx(ctx).Error("failed to create ingredient", zap.Error(err))
		return 0, common.NewBadRequestError("Failed to create ingredient", zimutils.GetErrorCodeFromError(err))
	}
	return id, nil
}

func (r *ProductRepo) insertTranslation(ctx context.Context, product ProductInput, languageCode string) error {
	sql := `INSERT INTO product_translations (product_id, name, description, language_code) VALUES ($1, $2, $3, $4)`
	op := common.GetOperator(ctx, r.Pool)
	_, err := op.Exec(ctx, sql, product.Id, product.Name, product.Description, languageCode)
	if err != nil {
		common.LoggerFromCtx(ctx).Error("failed to insert translation", zap.Error(err))
		return common.NewBadRequestError("Failed to insert translation", zimutils.GetErrorCodeFromError(err))
	}
	return nil
}

func (r *ProductRepo) TranslateProduct(ctx context.Context, product ProductInput, languageCode string) error {
	return r.insertTranslation(ctx, product, languageCode)
}

func (r *ProductRepo) AddProductVariants(ctx context.Context, productId *int, productVariants []ProductVariant) error {
	for _, productVariant := range productVariants {
		if _, err := r.addProductVariant(ctx, productId, productVariant); err != nil {
			return err
		}
	}
	return nil
}

func (r *ProductRepo) addProductVariants(ctx context.Context, productId *int, variants []ProductVariant) (map[string]int, error) {
	ids := make(map[string]int, len(variants))
	for _, variant := range variants {
		id, err := r.addProductVariant(ctx, productId, variant)
		if err != nil {
			return map[string]int{}, err
		}
		ids[variant.Sku] = id
	}
	return ids, nil
}

func (r *ProductRepo) addProductVariant(ctx context.Context, productId *int, productVariant ProductVariant) (int, error) {
	id, err := r.insertProductVariant(ctx, productId, productVariant)
	if err != nil {
		return 0, err
	}
	productVariant.Id = &id
	if err := r.insertProductVariantTranslation(ctx, productId, productVariant, common.DefaultLang); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *ProductRepo) insertProductVariant(ctx context.Context, productId *int, productVariant ProductVariant) (int, error) {
	op := common.GetOperator(ctx, r.Pool)
	sql := `INSERT INTO product_variants
	(
		product_id, price, sku, is_archived, is_default, image,
		standard_unit_id, expires_in_days
	) values (
		$1, $2, $3, $4, $5, $6, $7, $8
	) RETURNING id`
	var id int
	err := op.QueryRow(
		ctx, sql, productId, productVariant.Price, productVariant.Sku, productVariant.IsArchived,
		productVariant.IsDefault, productVariant.Image, productVariant.StandardUnitId,
		productVariant.ExpiresInDays,
	).Scan(&id)
	if err != nil {
		common.LoggerFromCtx(ctx).Error("failed to insert product variant", zap.Error(err))
		return 0, common.NewBadRequestError("Failed to insert product variant", zimutils.GetErrorCodeFromError(err))
	}
	return id, nil
}

func (r *ProductRepo) insertProductVariantTranslation(ctx context.Context, productId *int, productVariant ProductVariant, languageCode string) error {
	op := common.GetOperator(ctx, r.Pool)
	sql := `INSERT INTO product_variant_translations (product_id, product_variant_id, name, language_code) VALUES ($1, $2, $3, $4)`
	_, err := op.Exec(ctx, sql, productId, productVariant.Id, productVariant.Name, languageCode)
	if err != nil {
		common.LoggerFromCtx(ctx).Error("failed to insert product variant translation", zap.Error(err))
		return common.NewBadRequestError("Failed to insert product variant translation", zimutils.GetErrorCodeFromError(err))
	}
	return nil
}

func (r *ProductRepo) addProductVariantValues(
	ctx context.Context,
	productVariant ProductVariant,
	values []ProductOptionValue,
) error {
	for _, optionValue := range values {
		if err := r.addProductVariantSelectedValue(ctx, productVariant.Id, *optionValue.Id); err != nil {
			return err
		}
	}
	return nil
}

func (r *ProductRepo) addProductVariantSelectedValue(ctx context.Context, productVariantId *int, valueId int) error {
	op := common.GetOperator(ctx, r.Pool)
	sql := `INSERT INTO product_variant_values (product_variant_id, product_option_value_id) VALUES ($1, $2)`
	_, err := op.Exec(ctx, sql, productVariantId, valueId)
	if err != nil {
		common.LoggerFromCtx(ctx).Error("failed to insert product variant selected value", zap.Error(err))
		return common.NewBadRequestError("Failed to insert product variant selected value", zimutils.GetErrorCodeFromError(err))
	}
	return nil
}

func (r *ProductRepo) addProductOptions(ctx context.Context, productId *int, variants []ProductOption) (map[string]int, error) {
	optionValueIdMap := make(map[string]int)
	for _, variant := range variants {
		optionId, insertErr := r.InsertProductOption(ctx, productId, variant.Name)
		if insertErr != nil {
			return map[string]int{}, insertErr
		}
		for _, value := range variant.Values {
			valueId, insertValueErr := r.InsertProductOptionValue(ctx, optionId, value)
			if insertValueErr != nil {
				return map[string]int{}, insertValueErr
			}
			optionValueIdMap[value.Value] = valueId
		}
	}
	return optionValueIdMap, nil
}
func (r *ProductRepo) InsertProductOption(ctx context.Context, productId *int, option string) (int, error) {
	id, addErr := r.addProductOption(ctx, productId, option, common.DefaultLang)
	if addErr != nil {
		return 0, addErr
	}
	return id, nil
}

func (r *ProductRepo) addProductOption(ctx context.Context, productId *int, option string, languageCode string) (int, error) {
	op := common.GetOperator(ctx, r.Pool)
	sql := `INSERT INTO product_options (product_id, name, language_code) VALUES ($1, $2, $3) returning id`
	var id int
	err := op.QueryRow(ctx, sql, productId, option, languageCode).Scan(&id)
	if err != nil {
		common.LoggerFromCtx(ctx).Error("failed to translate product option", zap.Error(err))
		return 0, common.NewBadRequestError("Failed to translate product option", zimutils.GetErrorCodeFromError(err))
	}
	return id, nil
}

func (r *ProductRepo) InsertProductOptionValue(ctx context.Context, optionId int, optionValue ProductOptionValue) (int, error) {
	id, err := r.addProductOptionValue(ctx, optionId, optionValue, common.DefaultLang)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *ProductRepo) addProductOptionValue(ctx context.Context, optionId int, productOptionValue ProductOptionValue, languageCode string) (int, error) {
	op := common.GetOperator(ctx, r.Pool)
	sql := `INSERT INTO product_option_values(product_option_id, value, language_code) VALUES ($1, $2, $3) returning id`
	var id int
	err := op.QueryRow(ctx, sql, optionId, productOptionValue.Value, languageCode).Scan(&id)
	if err != nil {
		common.LoggerFromCtx(ctx).Error("failed to translate product option value", zap.Error(err))
		return 0, common.NewBadRequestError("Failed to translate product option value", zimutils.GetErrorCodeFromError(err))
	}
	return id, nil
}

func (r *ProductRepo) GetProducts(
	ctx context.Context,
	paginationParams common.PaginationParams,
	isArchive bool,
	isIngredient bool,
) ([]ProductBase, error) {
	op := common.GetOperator(ctx, r.Pool)
	languageCode := common.GetLanguageParam(ctx)
	sqlBuilder := common.NewPaginationQueryBuilder(
		`
		select p.id, ptx.name, ptx.description, p.image, p.is_archived, p.category_id
		from products p join product_translations ptx on p.id = ptx.product_id
		`,
		[]string{"p.id ASC"},
	)
	rows, err := sqlBuilder.
		WithOperator(op).
		WithConditions([]string{
			"is_archived = $1",
			"and",
			"ptx.language_code = $2",
			"and",
			"is_ingredient = $3",
		}).
		WithCursorKeys([]string{"p.id"}).
		WithParams(paginationParams).
		Build().
		Query(ctx, isArchive, languageCode, isIngredient)

	if err != nil {
		common.LoggerFromCtx(ctx).Error("failed to get products", zap.Error(err))
		return nil, common.NewBadRequestError("Failed to get products", zimutils.GetErrorCodeFromError(err))
	}
	defer rows.Close()
	products := make([]ProductBase, 0)
	for rows.Next() {
		var product ProductBase
		err := rows.Scan(&product.Id, &product.Name,
			&product.Description, &product.Image, &product.IsArchived, &product.CategoryId,
		)
		if err != nil {
			common.LoggerFromCtx(ctx).Error("failed to scan product", zap.Error(err))
			return nil, common.NewInternalServerError()
		}
		products = append(products, product)
	}
	return products, nil
}

func (r *ProductRepo) GetProduct(ctx context.Context, id int) (Product, error) {
	sql := `
	select p.id, ptx.name, ptx.description, p.image, p.is_archived, p.category_id, popt.id,
	popt.name, pvl.id, pvl.value from products p
	left join product_options popt on popt.product_id = p.id
	left join product_option_values pvl on pvl.product_option_id = popt.id
	join product_translations ptx on ptx.product_id = p.id
	where p.id = $1 and ptx.language_code = $2;
	`
	op := common.GetOperator(ctx, r.Pool)
	langCode := common.GetLanguageParam(ctx)
	rows, err := op.Query(ctx, sql, id, langCode)
	if err != nil {
		common.LoggerFromCtx(ctx).Error("failed to get product", zap.Error(err))
		return Product{}, common.NewBadRequestFromMessage("Failed to get product")
	}

	defer rows.Close()
	var product Product
	options := map[string]ProductOption{}
	for rows.Next() {
		var optionId, optionValueId *int
		var optionName, optionValueName *string
		err := rows.Scan(
			&product.Id, &product.Name, &product.Description, &product.Image, &product.IsArchived,
			&product.CategoryId, &optionId, &optionName, &optionValueId, &optionValueName,
		)
		if err != nil {
			common.LoggerFromCtx(ctx).Error("failed to scan product", zap.Error(err))
			return Product{}, common.NewInternalServerError()
		}
		if optionId == nil || optionValueId == nil {
			continue
		}
		option := ProductOption{
			Id:   optionId,
			Name: *optionName,
		}
		optionValue := ProductOptionValue{
			Id:    optionValueId,
			Value: *optionValueName,
		}
		if _variant, ok := options[option.Name]; ok {
			option.Values = append(_variant.Values, optionValue)
		} else {
			option.Values = []ProductOptionValue{optionValue}
		}
		options[option.Name] = option
	}
	product.Options = options
	return product, nil
}

func (r *ProductRepo) GetOriginalUnitsBySkuList(ctx context.Context, skuList []string) (map[string]int, error) {
	sql := `select standard_unit_id, sku from product_variants where sku = ANY($1)`
	op := common.GetOperator(ctx, r.Pool)
	rows, err := op.Query(ctx, sql, skuList)
	if err != nil {
		common.LoggerFromCtx(ctx).Error("failed to get original units", zap.Error(err))
		return nil, common.NewBadRequestFromMessage("Failed to get original units")
	}
	defer rows.Close()
	units := make(map[string]int)
	for rows.Next() {
		var unitId int
		var sku string
		err := rows.Scan(&unitId, &sku)
		if err != nil {
			common.LoggerFromCtx(ctx).Error("failed to scan unit", zap.Error(err))
			return nil, common.NewInternalServerError()
		}
		units[sku] = unitId
	}
	return units, nil
}

func (r *ProductRepo) DeleteProduct(ctx context.Context, product Product) error {
	batch := r.createDeleteProductBatch(ctx, product)
	op := common.GetOperator(ctx, r.Pool)
	results := op.SendBatch(ctx, batch)
	defer results.Close()
	for i := 0; i < batch.Len(); i++ {
		if _, err := results.Exec(); err != nil {
			common.LoggerFromCtx(ctx).Error("failed to delete product", zap.Error(err))
			return common.NewBadRequestFromMessage("Failed to delete product")
		}
	}
	return nil
}

func (r *ProductRepo) createDeleteProductBatch(ctx context.Context, product Product) *pgx.Batch {
	batch := &pgx.Batch{}
	batch.Queue("DELETE FROM product_translations WHERE product_id = $1", product.Id)
	for _, variant := range product.ProductVariants {
		r.addToDeleteVariantQueriesToBatch(batch, *variant.Id, variant.Sku)
	}
	batch.Queue(`
		DELETE FROM product_option_values
		USING product_options
		WHERE product_option_values.product_option_id = product_options.id 
		AND product_options.product_id = $1;`,
		product.Id,
	)
	batch.Queue("DELETE FROM product_options WHERE product_id = $1", product.Id)
	batch.Queue("DELETE FROM products WHERE id = $1", product.Id)
	return batch
}

func (r *ProductRepo) UpdateProductArchiveStatus(ctx context.Context, id int, isArchive bool) error {
	batch := r.createArchiveProductBatch(ctx, id, isArchive)
	op := common.GetOperator(ctx, r.Pool)
	results := op.SendBatch(ctx, batch)
	defer results.Close()
	for i := 0; i < batch.Len(); i++ {
		if _, err := results.Exec(); err != nil {
			common.LoggerFromCtx(ctx).Error("failed to archive product", zap.Error(err))
			return common.NewBadRequestFromMessage("Failed to archive product")
		}
	}
	return nil
}

func (r *ProductRepo) createArchiveProductBatch(ctx context.Context, id int, isArchive bool) *pgx.Batch {
	batch := &pgx.Batch{}
	batch.Queue("UPDATE products SET is_archived = $1 WHERE id = $2", isArchive, id)
	batch.Queue("UPDATE product_variants SET is_archived = $1 WHERE product_id = $2", isArchive, id)
	return batch
}
