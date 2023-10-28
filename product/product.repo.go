package product

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
	zimutils "github.com/nayefradwi/zanobia_inventory_manager/zim_utils"
)

type IProductRepo interface {
	CreateProduct(ctx context.Context, product ProductInput) error
	AddProductVariant(ctx context.Context, input ProductVariantInput) error
	TranslateProduct(ctx context.Context, product ProductInput, languageCode string) error
	GetProducts(ctx context.Context, pageSize int, endCursor string, isArchive bool) ([]ProductBase, error)
	GetProduct(ctx context.Context, id int) (Product, error)
	GetProductVariantsOfProduct(ctx context.Context, productId int) ([]ProductVariant, error)
	GetProductVariant(ctx context.Context, productVariantId int) (ProductVariant, error)
	GetUnitIdOfProductVariantBySku(ctx context.Context, sku string) (int, error)
	GetProductVariantExpirationDate(ctx context.Context, sku string) (time.Time, error)
}

type ProductRepo struct {
	*pgxpool.Pool
}

func NewProductRepository(dbPool *pgxpool.Pool) IProductRepo {
	return &ProductRepo{dbPool}
}

func (r *ProductRepo) CreateProduct(ctx context.Context, product ProductInput) error {
	err := common.RunWithTransaction(ctx, r.Pool, func(ctx context.Context, tx pgx.Tx) error {
		ctx = common.SetOperator(ctx, tx)
		return r.createProduct(ctx, product)
	})
	return err
}

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
	defaultProductVariantId, addVariantErr := r.addProductVariant(ctx, product.Id, product.DefaultProductVariant)
	if addVariantErr != nil {
		return addVariantErr
	}
	product.DefaultProductVariant.Id = &defaultProductVariantId
	if err := r.addProductVariantValues(ctx, product.DefaultProductVariant, product.DefaultOptionValues); err != nil {
		return err
	}
	if err := r.addProductOptions(ctx, product.Id, product.Options); err != nil {
		return err
	}
	return nil
}

func (r *ProductRepo) addProduct(ctx context.Context, product ProductInput) (int, error) {
	sql := `INSERT INTO products (image, category_id, is_archived) 
			VALUES ($1, $2, $3) RETURNING id`
	var id int
	op := common.GetOperator(ctx, r.Pool)
	err := op.QueryRow(ctx, sql, product.Image, product.CategoryId, product.IsArchived).Scan(&id)
	if err != nil {
		log.Printf("failed to create ingredient: %s", err.Error())
		return 0, common.NewBadRequestError("Failed to create ingredient", zimutils.GetErrorCodeFromError(err))
	}
	return id, nil
}

func (r *ProductRepo) insertTranslation(ctx context.Context, product ProductInput, languageCode string) error {
	sql := `INSERT INTO product_translations (product_id, name, description, language_code) VALUES ($1, $2, $3, $4)`
	op := common.GetOperator(ctx, r.Pool)
	_, err := op.Exec(ctx, sql, product.Id, product.Name, product.Description, languageCode)
	if err != nil {
		log.Printf("failed to insert translation: %s", err.Error())
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

func (r *ProductRepo) addProductVariant(ctx context.Context, productId *int, productVariant ProductVariant) (int, error) {
	// TODO change sql
	id, err := r.insertProductVariant(ctx, productId, productVariant)
	if err != nil {
		return 0, err
	}
	productVariant.Id = &id
	// TODO change sql
	if err := r.insertProductVariantTranslation(ctx, productVariant, common.DefaultLang); err != nil {
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
		log.Printf("failed to insert product variant: %s", err.Error())
		return 0, common.NewBadRequestError("Failed to insert product variant", zimutils.GetErrorCodeFromError(err))
	}
	return id, nil
}

func (r *ProductRepo) insertProductVariantTranslation(ctx context.Context, productVariant ProductVariant, languageCode string) error {
	// TODO change sql
	op := common.GetOperator(ctx, r.Pool)
	sql := `INSERT INTO product_variant_translations (product_variant_id, name, language_code) VALUES ($1, $2, $3)`
	_, err := op.Exec(ctx, sql, productVariant.Id, productVariant.Name, languageCode)
	if err != nil {
		log.Printf("failed to insert product variant translation: %s", err.Error())
		return common.NewBadRequestError("Failed to insert product variant translation", zimutils.GetErrorCodeFromError(err))
	}
	return nil
}

func (r *ProductRepo) addProductVariantValues(ctx context.Context, productVariant ProductVariant, variantValues []ProductOptionValue) error {
	for _, variantValue := range variantValues {
		// TODO change sql
		if err := r.addProductVariantSelectedValue(ctx, productVariant.Id, variantValue); err != nil {
			return err
		}
	}
	return nil
}

func (r *ProductRepo) addProductVariantSelectedValue(ctx context.Context, productVariantId *int, value ProductOptionValue) error {
	// TODO change sql
	op := common.GetOperator(ctx, r.Pool)
	sql := `INSERT INTO product_variant_selected_values (product_variant_id, variant_value_id) VALUES ($1, $2)`
	_, err := op.Exec(ctx, sql, productVariantId, value.Id)
	if err != nil {
		log.Printf("failed to insert product variant selected value: %s", err.Error())
		return common.NewBadRequestError("Failed to insert product variant selected value", zimutils.GetErrorCodeFromError(err))
	}
	return nil
}

func (r *ProductRepo) mapProductToVariantValue(ctx context.Context, productId *int, variantValue ProductOptionValue) error {
	// TODO change sql
	op := common.GetOperator(ctx, r.Pool)
	sql := `INSERT INTO product_selected_values (product_id, variant_value_id) VALUES ($1, $2)`
	_, err := op.Exec(ctx, sql, productId, variantValue.Id)
	if err != nil {
		log.Printf("failed to map product to variant value: %s", err.Error())
		return common.NewBadRequestError("Failed to map product to variant value", zimutils.GetErrorCodeFromError(err))
	}
	return nil
}

func (r *ProductRepo) addProductOptions(ctx context.Context, productId *int, variants []ProductOption) error {
	for _, variant := range variants {
		// TODO change sql
		if err := r.addProductOption(ctx, productId, variant); err != nil {
			return err
		}
		for _, value := range variant.Values {
			// TODO change sql
			if err := r.mapProductToVariantValue(ctx, productId, value); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *ProductRepo) addProductOption(ctx context.Context, productId *int, variant ProductOption) error {
	op := common.GetOperator(ctx, r.Pool)
	sql := `INSERT INTO product_options (product_id, variant_id) VALUES ($1, $2)`
	_, err := op.Exec(ctx, sql, productId, variant.Id)
	if err != nil {
		log.Printf("failed to add product option: %s", err.Error())
		return common.NewBadRequestError("Failed to add product option", zimutils.GetErrorCodeFromError(err))
	}
	return nil
}

func (r *ProductRepo) GetProducts(ctx context.Context, pageSize int, endCursor string, isArchive bool) ([]ProductBase, error) {
	sql := `select p.id, ptx.name, ptx.description, p.image, p.is_archived, p.category_id
	from products p join product_translations ptx on p.id = ptx.product_id
	where is_archived = $2 and ptx.language_code = $3
	and (
		p.id < $1 or $1 = 0
	)
	order by created_at desc limit $4;`
	op := common.GetOperator(ctx, r.Pool)
	languageCode := common.GetLanguageParam(ctx)
	rows, err := op.Query(ctx, sql, endCursor, isArchive, languageCode, pageSize)
	if err != nil {
		log.Printf("failed to get products: %s", err.Error())
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
			log.Printf("failed to scan product: %s", err.Error())
			return nil, common.NewInternalServerError()
		}
		products = append(products, product)
	}
	return products, nil
}

func (r *ProductRepo) GetProduct(ctx context.Context, id int) (Product, error) {
	sql := `
	select p.id, ptx.name, ptx.description, p.image, p.is_archived, p.category_id, vartx.variant_id,
	vartx.name, varvl.id, varvl.value from products p
	join product_options popt on popt.product_id = p.id 
	join variant_translations vartx on vartx.variant_id = popt.variant_id
	join product_selected_values pvl on pvl.product_id = p.id
	join variant_values varvl on pvl.variant_value_id = varvl.id and varvl.variant_id = popt.variant_id
	join product_translations ptx on ptx.product_id = p.id
	where p.id = $1 and ptx.language_code = $2;
	`
	op := common.GetOperator(ctx, r.Pool)
	langCode := common.GetLanguageParam(ctx)
	rows, err := op.Query(ctx, sql, id, langCode)
	if err != nil {
		log.Printf("failed to get product: %s", err.Error())
		return Product{}, common.NewBadRequestFromMessage("Failed to get product")
	}
	defer rows.Close()
	var product Product
	options := map[string]ProductOption{}
	for rows.Next() {
		var option ProductOption
		var optionValue ProductOptionValue
		err := rows.Scan(
			&product.Id, &product.Name, &product.Description, &product.Image, &product.IsArchived,
			&product.CategoryId, &option.Id, &option.Name, &optionValue.Id, &optionValue.Value,
		)
		if err != nil {
			log.Printf("failed to scan product: %s", err.Error())
			return Product{}, common.NewInternalServerError()
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
