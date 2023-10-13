package product

import (
	"context"
	"log"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
	zimutils "github.com/nayefradwi/zanobia_inventory_manager/zim_utils"
)

type IProductRepo interface {
	CreateProduct(ctx context.Context, product ProductInput) error
	TranslateProduct(ctx context.Context, product ProductInput, languageCode string) error
	GetProducts(ctx context.Context, pageSize int, endCursor string, isArchive bool) ([]Product, error)
	GetProduct(ctx context.Context, id int) (Product, error)
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
		id, addErr := r.addProduct(ctx, product)
		if addErr != nil {
			return addErr
		}
		product.Id = &id
		if translationErr := r.insertTranslation(ctx, product, common.DefaultLang); translationErr != nil {
			return translationErr
		}
		product = product.GenerateProductDetails()
		defaultProductVariantId, addVariantErr := r.addProductVariant(ctx, product.DefaultProductVariant)
		if addVariantErr != nil {
			return addVariantErr
		}
		product.DefaultProductVariant.Id = &defaultProductVariantId
		if err := r.addDefaultProductVariantValues(ctx, product.DefaultProductVariant, product.DefaultValues); err != nil {
			return err
		}
		if err := r.mapProductToVariantValue(ctx, product.Id, product.DefaultValues[0]); err != nil {
			return err
		}
		if err := r.addProductOptions(ctx, product.Id, product.Variants); err != nil {
			return err
		}
		return nil
	})

	return err
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
	sql := `INSERT INTO product_translations (product_id, name, description, language_code) VALUES ($1, $2, $3, $4, $5)`
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

func (r *ProductRepo) addProductVariants(ctx context.Context, productVariants []ProductVariant) error {
	for _, productVariant := range productVariants {
		if _, err := r.addProductVariant(ctx, productVariant); err != nil {
			return err
		}
	}
	return nil
}

func (r *ProductRepo) addProductVariant(ctx context.Context, productVariant ProductVariant) (int, error) {
	id, err := r.insertProductVariant(ctx, productVariant)
	if err != nil {
		return 0, err
	}
	productVariant.Id = &id
	if err := r.insertProductVariantTranslation(ctx, productVariant, common.DefaultLang); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *ProductRepo) insertProductVariant(ctx context.Context, productVariant ProductVariant) (int, error) {
	// TODO insert product variant
	return 0, nil
}

func (r *ProductRepo) insertProductVariantTranslation(ctx context.Context, productVariant ProductVariant, languageCode string) error {
	// TODO insert product variant translation
	return nil
}

func (r *ProductRepo) addDefaultProductVariantValues(ctx context.Context, productVariant ProductVariant, variantValues []VariantValue) error {
	for _, variantValue := range variantValues {
		if err := r.addProductVariantSelectedValue(ctx, productVariant.Id, variantValue); err != nil {
			return err
		}
	}
	return nil
}

func (r *ProductRepo) addProductVariantSelectedValue(ctx context.Context, productVariantId *int, value VariantValue) error {
	// TODO add product_variant_selected_value
	return nil
}

func (r *ProductRepo) mapProductToVariantValue(ctx context.Context, productId *int, variantValue VariantValue) error {
	// TODO map product to variant values
	return nil
}

func (r *ProductRepo) addProductOptions(ctx context.Context, productId *int, variants []Variant) error {
	for _, variant := range variants {
		if err := r.addProductOption(ctx, productId, variant); err != nil {
			return err
		}
	}
	return nil
}

func (r *ProductRepo) addProductOption(ctx context.Context, productId *int, variant Variant) error {
	// TODO add product option
	return nil
}

func (r *ProductRepo) GetProducts(ctx context.Context, pageSize int, endCursor string, isArchive bool) ([]Product, error) {
	// TODO refactor
	// sql := `select p.id, ptx.name,  p.image, p.is_archived, p.expires_in_days, utx.name, utx.symbol, utx.unit_id from products p
	// join unit_translations utx on utx.unit_id = p.standard_unit_id
	// join product_translations ptx on p.id = ptx.product_id
	// where is_archived = $2 and ptx.language_code = $3
	// and (
	// 	p.id < $1 or $1 = 0
	// )
	// order by created_at desc limit $4;`
	// op := common.GetOperator(ctx, r.Pool)
	// languageCode := common.GetLanguageParam(ctx)
	// rows, err := op.Query(ctx, sql, endCursor, isArchive, languageCode, pageSize)
	// if err != nil {
	// 	log.Printf("failed to get products: %s", err.Error())
	// 	return nil, common.NewBadRequestError("Failed to get products", zimutils.GetErrorCodeFromError(err))
	// }
	// defer rows.Close()
	// products := make([]Product, 0)
	// for rows.Next() {
	// 	var product Product
	// 	var unit Unit
	// 	err := rows.Scan(&product.Id, &product.Name, &product.Image, &product.IsArchived, &product.ExpiresInDays, &unit.Name, &unit.Symbol, &unit.Id)
	// 	if err != nil {
	// 		log.Printf("failed to scan product: %s", err.Error())
	// 		return nil, common.NewInternalServerError()
	// 	}
	// 	product.StandardUnit = &unit
	// 	products = append(products, product)
	// }
	// return products, nil
	return []Product{}, nil
}

func (r *ProductRepo) GetProduct(ctx context.Context, id int) (Product, error) {
	// TODO refactor
	// sql := `select p.id, p.image, p.price, p.width_in_cm, p.height_in_cm,
	// p.depth_in_cm, p.weight_in_g, p.standard_unit_id, p.category_id,
	// p.is_archived, p.expires_in_days, ptx.name, ptx.description, utx.name, utx.symbol, utx.unit_id,
	// ctx.name, ctx.category_id
	// from products p
	// join product_translations ptx on p.id = ptx.product_id
	// join unit_translations utx on utx.unit_id = p.standard_unit_id
	// left join category_translations ctx on ctx.category_id = p.category_id
	// where p.id = $1 and ptx.language_code = $2;`
	// op := common.GetOperator(ctx, r.Pool)
	// languageCode := common.GetLanguageParam(ctx)
	// var product Product
	// var unit Unit
	// var category Category
	// err := op.QueryRow(ctx, sql, id, languageCode).Scan(
	// 	&product.Id, &product.Image, &product.Price, &product.WidthInCm, &product.HeightInCm,
	// 	&product.DepthInCm, &product.WeightInG, &product.StandardUnitId, &product.CategoryId,
	// 	&product.IsArchived, &product.ExpiresInDays, &product.Name, &product.Description,
	// 	&unit.Name, &unit.Symbol, &unit.Id, &category.Name, &category.Id,
	// )
	// if err != nil {
	// 	log.Printf("failed to get product: %s", err.Error())
	// 	return Product{}, common.NewBadRequestError("Failed to get product", zimutils.GetErrorCodeFromError(err))
	// }
	// product.StandardUnit = &unit
	// if category.Id != nil {
	// 	product.Category = &category
	// }
	// return product, nil
	return Product{}, nil
}
