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

type IVariantRepository interface {
	CreateVariant(ctx context.Context, variant VariantInput) error
	AddVariantValues(ctx context.Context, variantId int, values []string) error
	UpdateVariantName(ctx context.Context, variantId int, newName string) error
	UpdateVariantValue(ctx context.Context, value VariantValue) error
	GetVariant(ctx context.Context, variantId int) (Variant, error)
	GetVariantsAndValuesFromIds(ctx context.Context, variantIds []int, variantValueIds []int) ([]Variant, error)
	GetProductOptionsIds(ctx context.Context, productId int) ([]Variant, error)
	GetProductSelectedValues(ctx context.Context, productId int) ([]VariantValue, error)
}

type VariantRepository struct {
	*pgxpool.Pool
}

func NewVariantRepository(pool *pgxpool.Pool) IVariantRepository {
	return &VariantRepository{pool}
}

func (r *VariantRepository) CreateVariant(ctx context.Context, variant VariantInput) error {
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

func (r *VariantRepository) createVariant(ctx context.Context, variant VariantInput) (int, error) {
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

func (r *VariantRepository) UpdateVariantName(ctx context.Context, variantId int, newName string) error {
	err := common.RunWithTransaction(ctx, r.Pool, func(ctx context.Context, tx pgx.Tx) error {
		ctx = common.SetOperator(ctx, tx)
		lang := common.GetLanguageParam(ctx)
		err := r.updateVariantTranslation(ctx, variantId, newName, lang)
		if err != nil {
			return err
		}
		return r.updateVariantUpdatedAt(ctx, variantId)
	})
	return err
}

func (r *VariantRepository) updateVariantUpdatedAt(ctx context.Context, variantId int) error {
	updatedAt := time.Now().UTC()
	op := common.GetOperator(ctx, r.Pool)
	sql := `UPDATE variants SET updated_at = $1 where id = $2`
	_, err := op.Exec(ctx, sql, updatedAt, variantId)
	if err != nil {
		log.Printf("failed to update variant updated_at: %s", err.Error())
		return common.NewInternalServerError()
	}
	return nil
}

func (r *VariantRepository) updateVariantTranslation(ctx context.Context, variantId int, newName string, langCode string) error {
	op := common.GetOperator(ctx, r.Pool)
	sql := `UPDATE variant_translations SET name = $1 WHERE language_code = $2 and variant_id = $3`
	_, err := op.Exec(ctx, sql, newName, langCode, variantId)
	if err != nil {
		log.Printf("failed to update variant translation: %s", err.Error())
		return common.NewInternalServerError()
	}
	return nil
}

func (r *VariantRepository) UpdateVariantValue(ctx context.Context, value VariantValue) error {
	op := common.GetOperator(ctx, r.Pool)
	sql := `UPDATE variant_values SET value = $1 WHERE id = $2`
	_, err := op.Exec(ctx, sql, value.Value, value.Id)
	if err != nil {
		log.Printf("failed to update variant value: %s", err.Error())
		return common.NewInternalServerError()
	}
	return nil
}

func (r *VariantRepository) GetVariant(ctx context.Context, variantId int) (Variant, error) {
	op := common.GetOperator(ctx, r.Pool)
	lang := common.GetLanguageParam(ctx)
	sql := `select v.id, vtx.name, variant_values.id, value from variants v
	join variant_translations vtx on vtx.variant_id = v.id
	join variant_values on variant_values.variant_id = v.id
	where v.id = $1 and vtx.language_code = $2;`
	var variant Variant
	var variantValues []VariantValue
	rows, err := op.Query(ctx, sql, variantId, lang)
	if err != nil {
		log.Printf("failed to get variant: %s", err.Error())
		return variant, common.NewInternalServerError()
	}
	defer rows.Close()
	for rows.Next() {
		variantValue := VariantValue{}
		err := rows.Scan(&variant.Id, &variant.Name, &variantValue.Id, &variantValue.Value)
		if err != nil {
			log.Printf("failed to scan variant: %s", err.Error())
			return variant, common.NewInternalServerError()
		}
		variantValues = append(variantValues, variantValue)
	}
	variant.Values = variantValues
	return variant, nil
}

func (r *VariantRepository) GetVariantsAndValuesFromIds(ctx context.Context, variantIds []int, variantValueIds []int) ([]Variant, error) {
	op := common.GetOperator(ctx, r.Pool)
	lang := common.GetLanguageParam(ctx)
	sql := `
		select v.id, vtx.name, variant_values.id, value from variants v
		join variant_translations vtx on vtx.variant_id = v.id
		join variant_values on variant_values.variant_id = v.id
		where v.id = any($1) and vtx.language_code = $2 and variant_values.id = any($3);`
	variants := map[int]Variant{}
	variantValues := map[int][]VariantValue{}
	rows, err := op.Query(ctx, sql, variantIds, lang, variantValueIds)
	if err != nil {
		log.Printf("failed to get variant: %s", err.Error())
		return []Variant{}, common.NewInternalServerError()
	}
	defer rows.Close()
	for rows.Next() {
		variant := Variant{Id: new(int)}
		variantValue := VariantValue{}
		err := rows.Scan(&variant.Id, &variant.Name, &variantValue.Id, &variantValue.Value)
		if err != nil {
			log.Printf("failed to scan variant: %s", err.Error())
			return []Variant{}, common.NewInternalServerError()
		}
		variants[*variant.Id] = variant
		if variantValues[*variant.Id] == nil {
			variantValues[*variant.Id] = []VariantValue{variantValue}
		} else {
			variantValues[*variant.Id] = append(variantValues[*variant.Id], variantValue)
		}
	}
	variantsList := make([]Variant, 0)
	for variantId, variant := range variants {
		variant.Values = variantValues[variantId]
		variantsList = append(variantsList, variant)
	}
	return variantsList, nil
}

func (r *VariantRepository) GetProductOptionsIds(ctx context.Context, productId int) ([]Variant, error) {
	sql := `
	select variants.id from variants join product_options 
	on product_options.variant_id = variants.id where product_options.product_id = $1;
	`
	op := common.GetOperator(ctx, r.Pool)
	rows, err := op.Query(ctx, sql, productId)
	if err != nil {
		log.Printf("failed to get product options: %s", err.Error())
		return []Variant{}, common.NewInternalServerError()
	}
	defer rows.Close()
	variants := make([]Variant, 0)
	for rows.Next() {
		var variant Variant
		err := rows.Scan(&variant.Id)
		if err != nil {
			log.Printf("failed to scan variant: %s", err.Error())
			return []Variant{}, common.NewInternalServerError()
		}
		variants = append(variants, variant)
	}
	return variants, nil
}

func (r *VariantRepository) GetProductSelectedValues(ctx context.Context, productId int) ([]VariantValue, error) {
	sql := `
	select variant_values.id, value from variant_values join product_selected_values 
	on product_selected_values.variant_value_id = variant_values.id where product_selected_values.product_id = $1;
	`
	op := common.GetOperator(ctx, r.Pool)
	rows, err := op.Query(ctx, sql, productId)
	if err != nil {
		log.Printf("failed to get product selected values: %s", err.Error())
		return []VariantValue{}, common.NewInternalServerError()
	}
	defer rows.Close()
	variantValues := make([]VariantValue, 0)
	for rows.Next() {
		var variantValue VariantValue
		err := rows.Scan(&variantValue.Id, &variantValue.Value)
		if err != nil {
			log.Printf("failed to scan variant value: %s", err.Error())
			return []VariantValue{}, common.NewInternalServerError()
		}
		variantValues = append(variantValues, variantValue)
	}
	return variantValues, nil
}
