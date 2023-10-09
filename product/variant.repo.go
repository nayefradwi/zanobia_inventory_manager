package product

import (
	"context"
	"log"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
	zimutils "github.com/nayefradwi/zanobia_inventory_manager/zim_utils"
)

type IVariantRepository interface {
	CreateVariant(ctx context.Context, variant Variant) error
	AddVariantValues(ctx context.Context, variantId int, values []string) error
}
type VariantRepository struct {
	*pgxpool.Pool
}

func NewVariantRepository(pool *pgxpool.Pool) IVariantRepository {
	return &VariantRepository{pool}
}

func (r *VariantRepository) CreateVariant(ctx context.Context, variant Variant) error {
	err := common.RunWithTransaction(ctx, r.Pool, func(ctx context.Context, tx pgx.Tx) error {
		ctx = common.SetOperator(ctx, tx)
		variantId, err := r.createVariant(ctx, variant)
		if err != nil {
			return err
		}
		translationErr := r.insertVariantTranslation(ctx, variantId, variant.Name, common.DefaultLang)
		if translationErr != nil {
			return translationErr
		}
		addingValuesErr := r.addVariantValues(ctx, variantId, variant.Values)
		return addingValuesErr
	})
	return err
}

func (r *VariantRepository) createVariant(ctx context.Context, variant Variant) (int, error) {
	op := common.GetOperator(ctx, r.Pool)
	sql := `INSERT INTO variants (name) VALUES ($1) RETURNING id`
	var id int
	err := op.QueryRow(ctx, sql, variant.Name).Scan(&id)
	if err != nil {
		log.Printf("failed to create variant: %s", err.Error())
		return 0, common.NewBadRequestError("Failed to create variant", zimutils.GetErrorCodeFromError(err))
	}
	return id, nil
}

func (r *VariantRepository) insertVariantTranslation(ctx context.Context, variantId int, name string, languageCode string) error {
	op := common.GetOperator(context.Background(), r.Pool)
	sql := `INSERT INTO variant_translations (variant_id, name, language_code) VALUES ($1, $2, $3)`
	_, err := op.Exec(ctx, sql, variantId, name, languageCode)
	if err != nil {
		log.Printf("failed to translate variant: %s", err.Error())
		return common.NewBadRequestError("Failed to translate variant", zimutils.GetErrorCodeFromError(err))
	}
	return nil
}

func (r *VariantRepository) AddVariantValues(ctx context.Context, variantId int, values []string) error {
	err := common.RunWithTransaction(ctx, r.Pool, func(ctx context.Context, tx pgx.Tx) error {
		ctx = common.SetOperator(ctx, tx)
		return r.addVariantValues(ctx, variantId, values)
	})
	return err
}

func (r *VariantRepository) addVariantValues(ctx context.Context, variantId int, values []string) error {
	for _, value := range values {
		addValueErr := r.addVariantValue(ctx, variantId, value)
		if addValueErr != nil {
			return addValueErr
		}
	}
	return nil
}

func (r *VariantRepository) addVariantValue(ctx context.Context, variantId int, value string) error {
	id, err := r.createVariantValue(variantId, value)
	if err != nil {
		return err
	}
	translationErr := r.insertVariantValueTranslation(id, value, common.DefaultLang)
	return translationErr
}

func (r *VariantRepository) createVariantValue(variantId int, value string) (int, error) {
	op := common.GetOperator(context.Background(), r.Pool)
	sql := `INSERT INTO variant_values (variant_id, value) VALUES ($1, $2) RETURNING id`
	var id int
	err := op.QueryRow(context.Background(), sql, variantId, value).Scan(&id)
	if err != nil {
		log.Printf("failed to create variant value: %s", err.Error())
		return 0, common.NewBadRequestError("Failed to create variant value", zimutils.GetErrorCodeFromError(err))
	}
	return id, nil
}

func (r *VariantRepository) insertVariantValueTranslation(variantValueId int, value string, languageCode string) error {
	op := common.GetOperator(context.Background(), r.Pool)
	sql := `INSERT INTO variant_value_translations (variant_value_id, value, language_code) VALUES ($1, $2, $3)`
	_, err := op.Exec(context.Background(), sql, variantValueId, value, languageCode)
	if err != nil {
		log.Printf("failed to translate variant value: %s", err.Error())
		return common.NewBadRequestError("Failed to translate variant value", zimutils.GetErrorCodeFromError(err))
	}
	return nil
}
