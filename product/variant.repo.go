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
	sql := `INSERT INTO variants DEFAULT VALUES RETURNING id`
	var id int
	err := op.QueryRow(ctx, sql).Scan(&id)
	if err != nil {
		log.Printf("failed to create variant: %s", err.Error())
		return 0, common.NewBadRequestError("Failed to create variant", zimutils.GetErrorCodeFromError(err))
	}
	return id, nil
}

func (r *VariantRepository) insertVariantTranslation(ctx context.Context, variantId int, name string, languageCode string) error {
	op := common.GetOperator(ctx, r.Pool)
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
		addValueErr := r.addVariantValue(ctx, variantId, value, common.DefaultLang)
		if addValueErr != nil {
			return addValueErr
		}
	}
	return nil
}

func (r *VariantRepository) addVariantValue(ctx context.Context, variantId int, value string, langCode string) error {
	return r.createVariantValue(ctx, variantId, value, langCode)
}

func (r *VariantRepository) createVariantValue(ctx context.Context, variantId int, value string, langCode string) error {
	op := common.GetOperator(ctx, r.Pool)
	sql := `INSERT INTO variant_values (variant_id, value, language_code) VALUES ($1, $2, $3) RETURNING id`
	_, err := op.Exec(ctx, sql, variantId, value, langCode)
	if err != nil {
		log.Printf("failed to create variant value: %s", err.Error())
		return common.NewBadRequestError("Failed to create variant value", zimutils.GetErrorCodeFromError(err))
	}
	return nil
}
