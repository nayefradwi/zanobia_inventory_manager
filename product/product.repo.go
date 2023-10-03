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
	CreateProduct(ctx context.Context, product ProductBase) error
	TranslateProduct(ctx context.Context, product ProductBase, languageCode string) error
	GetProducts(ctx context.Context, pageSize int, endCursor string, isArchive bool) ([]Product, error)
	GetProduct(ctx context.Context, id int) (Product, error)
}

type ProductRepo struct {
	*pgxpool.Pool
}

func NewProductRepository(dbPool *pgxpool.Pool) IProductRepo {
	return &ProductRepo{dbPool}
}

func (r *ProductRepo) CreateProduct(ctx context.Context, product ProductBase) error {
	err := common.RunWithTransaction(ctx, r.Pool, func(ctx context.Context, tx pgx.Tx) error {
		ctx = common.SetOperator(ctx, tx)
		id, addErr := r.addProduct(ctx, product)
		if addErr != nil {
			return addErr
		}
		product.Id = &id
		translationErr := r.insertTranslation(ctx, product, common.DefaultLang)
		if translationErr != nil {
			return translationErr
		}
		return nil
	})

	return err
}

func (r *ProductRepo) addProduct(ctx context.Context, product ProductBase) (int, error) {
	sql := `INSERT INTO products 
			(image, price, width_in_cm, height_in_cm, depth_in_cm, 
			weight_in_g, standard_unit_id, category_id, is_archived,
			expires_in_days) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) 
			RETURNING id`
	var id int
	op := common.GetOperator(ctx, r.Pool)
	err := op.QueryRow(ctx, sql,
		product.Image, product.Price, product.WidthInCm, product.HeightInCm, product.DepthInCm,
		product.WeightInG, product.StandardUnitId, product.CategoryId, product.IsArchived,
		product.ExpiresInDays,
	).Scan(&id)
	if err != nil {
		log.Printf("failed to create ingredient: %s", err.Error())
		return 0, common.NewBadRequestError("Failed to create ingredient", zimutils.GetErrorCodeFromError(err))
	}
	return id, nil
}

func (r *ProductRepo) insertTranslation(ctx context.Context, product ProductBase, languageCode string) error {
	sql := `INSERT INTO product_translations (product_id, name, description, standard_unit_id, language_code) VALUES ($1, $2, $3, $4, $5)`
	op := common.GetOperator(ctx, r.Pool)
	_, err := op.Exec(ctx, sql, product.Id, product.Name, product.Description, product.StandardUnitId, languageCode)
	if err != nil {
		log.Printf("failed to insert translation: %s", err.Error())
		return common.NewBadRequestError("Failed to insert translation", zimutils.GetErrorCodeFromError(err))
	}
	return nil
}

func (r *ProductRepo) TranslateProduct(ctx context.Context, product ProductBase, languageCode string) error {
	return r.insertTranslation(ctx, product, languageCode)
}

func (r *ProductRepo) GetProducts(ctx context.Context, pageSize int, endCursor string, isArchive bool) ([]Product, error) {
	sql := `select p.id, ptx.name,  p.image, p.is_archived, p.expires_in_days, utx.name, utx.symbol, utx.unit_id from products p
	join unit_translations utx on utx.unit_id = p.standard_unit_id
	join product_translations ptx on p.id = ptx.product_id
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
	products := make([]Product, 0)
	for rows.Next() {
		var product Product
		var unit Unit
		err := rows.Scan(&product.Id, &product.Name, &product.Image, &product.IsArchived, &product.ExpiresInDays, &unit.Name, &unit.Symbol, &unit.Id)
		if err != nil {
			log.Printf("failed to scan product: %s", err.Error())
			return nil, common.NewInternalServerError()
		}
		product.StandardUnit = &unit
		products = append(products, product)
	}
	return products, nil
}

func (r *ProductRepo) GetProduct(ctx context.Context, id int) (Product, error) {
	sql := `select p.id, p.image, p.price, p.width_in_cm, p.height_in_cm, 
	p.depth_in_cm, p.weight_in_g, p.standard_unit_id, p.category_id,
	p.is_archived, p.expires_in_days, ptx.name, ptx.description, utx.name, utx.symbol, utx.unit_id,
	ctx.name, ctx.category_id
	from products p
	join product_translations ptx on p.id = ptx.product_id
	join unit_translations utx on utx.unit_id = p.standard_unit_id
	left join category_translations ctx on ctx.category_id = p.category_id
	where p.id = $1 and ptx.language_code = $2;`
	op := common.GetOperator(ctx, r.Pool)
	languageCode := common.GetLanguageParam(ctx)
	var product Product
	var unit Unit
	var category Category
	err := op.QueryRow(ctx, sql, id, languageCode).Scan(
		&product.Id, &product.Image, &product.Price, &product.WidthInCm, &product.HeightInCm,
		&product.DepthInCm, &product.WeightInG, &product.StandardUnitId, &product.CategoryId,
		&product.IsArchived, &product.ExpiresInDays, &product.Name, &product.Description,
		&unit.Name, &unit.Symbol, &unit.Id, &category.Name, &category.Id,
	)
	if err != nil {
		log.Printf("failed to get product: %s", err.Error())
		return Product{}, common.NewBadRequestError("Failed to get product", zimutils.GetErrorCodeFromError(err))
	}
	product.StandardUnit = &unit
	if category.Id != nil {
		product.Category = &category
	}
	return product, nil

}
