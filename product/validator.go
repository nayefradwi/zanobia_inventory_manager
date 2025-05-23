package product

import (
	"github.com/nayefradwi/zanobia_inventory_manager/common"
)

func ValidateProduct(product ProductInput) error {
	validationResults := make([]common.ErrorDetails, 0)
	validationResults = append(validationResults,
		common.ValidateAlphaNuemericPtr(product.Name, "name"),
		common.ValidateStringLength(product.Description, "description", 0, 255),
		common.ValidateNotZero(product.Price, "price"),
		common.ValidateIdPtr(product.StandardUnitId, "standardUnitId"),
		common.ValidateNotZero(product.ExpiresInDays, "expiresInDays"),
		validateProductOptions(product.Options),
	)
	errors := make([]common.ErrorDetails, 0)
	for _, result := range validationResults {
		if len(result.Message) > 0 {
			errors = append(errors, result)
		}
	}
	if len(errors) > 0 {
		return common.NewValidationError("invalid product input", errors...)
	}
	return nil
}
func validateProductOptions(options []ProductOption) common.ErrorDetails {
	if len(options) == 0 {
		return common.ErrorDetails{}
	}
	for _, option := range options {
		if len(option.Name) < 1 || len(option.Name) > 50 {
			return common.ErrorDetails{
				Message: "option name must be between 1 and 50 characters",
				Field:   "options",
			}
		}
		for _, value := range option.Values {
			if len(value.Value) < 1 || len(value.Value) > 50 {
				return common.ErrorDetails{
					Message: "option value must be between 1 and 50 characters",
					Field:   "options",
				}
			}
		}
	}
	return common.ErrorDetails{}
}

func ValidateProductDescription(description string) common.ErrorDetails {
	if len(description) > 255 {
		return common.ErrorDetails{
			Message: "description cannot be more than 255 characters",
			Field:   "description",
		}
	}
	return common.ErrorDetails{}
}

func ValidateProductDimension(value *float64, field string) common.ErrorDetails {
	if value == nil {
		return common.ErrorDetails{}
	}
	if *value <= 0 {
		return common.ErrorDetails{
			Message: field + " must be greater than 0",
			Field:   field,
		}
	}
	return common.ErrorDetails{}
}

func ValidateProductStandardUnitId(unitId *int) common.ErrorDetails {
	if unitId == nil {
		return common.ErrorDetails{
			Message: "unit id cannot be empty",
			Field:   "standardUnitId",
		}
	}
	return common.ErrorDetails{}
}

func ValidateProductCategoryId(categoryId *int) common.ErrorDetails {
	if categoryId == nil {
		return common.ErrorDetails{}
	}
	if *categoryId <= 0 {
		return common.ErrorDetails{
			Message: "category id must be greater than 0",
			Field:   "categoryId",
		}
	}
	return common.ErrorDetails{}
}

func ValidateRecipe(recipe RecipeBase) error {
	qtyValidation := common.ValidateNotZero(recipe.Quantity, "quantity")
	if len(qtyValidation.Message) > 0 {
		return common.NewValidationError("invalid recipe input", qtyValidation)
	}
	return nil
}

func ValidateRecipes(recipes []RecipeBase) error {
	for _, recipe := range recipes {
		err := ValidateRecipe(recipe)
		if err != nil {
			return err
		}
	}
	return nil
}

func ValidateProductVariant(input ProductVariantInput, min, max int) error {
	validationResults := make([]common.ErrorDetails, 0)
	validationResults = append(validationResults,
		common.ValidateNotZero(input.ProductVariant.Price, "price"),
		common.ValidateIdPtr(input.ProductVariant.StandardUnitId, "standardUnitId"),
		common.ValidateIdPtr(input.ProductVariant.ProductId, "productId"),
		common.ValidateNotZero(input.ProductVariant.ExpiresInDays, "expiresInDays"),
		ValidateProductVariantSelectedValues(input.OptionValueIds, min, max),
	)
	errors := make([]common.ErrorDetails, 0)
	for _, result := range validationResults {
		if len(result.Message) > 0 {
			errors = append(errors, result)
		}
	}
	if len(errors) > 0 {
		return common.NewValidationError("invalid product input", errors...)
	}
	return nil
}

func ValidateProductVariantSelectedValues(valueIds []int, min, max int) common.ErrorDetails {
	if details := common.ValidateSliceSize[int](valueIds, "variantValues", min, max); details.Message != "" {
		return details
	}
	for _, id := range valueIds {
		if id <= 0 {
			return common.ErrorDetails{
				Message: "invalid variant id",
				Field:   "variantValues",
			}
		}
	}
	return common.ErrorDetails{}
}

func ValidateProductOptionInput(input ProductOptionInput) error {
	details := make([]common.ErrorDetails, 0)
	details = append(details,
		common.ValidateId(input.ProductId, "productId"),
		common.ValidateStringLength(input.Name, "name", 1, 50),
	)
	if len(input.Values) == 0 {
		details = append(details, common.ErrorDetails{
			Message: "values cannot be empty",
			Field:   "values",
		})
	}
	for _, value := range input.Values {
		if len(value.Value) < 1 || len(value.Value) > 50 {
			details = append(details, common.ErrorDetails{Message: "value must be between 1 and 50 characters", Field: "values"})
			continue
		}
	}
	errors := make([]common.ErrorDetails, 0)
	for _, detail := range details {
		if len(detail.Message) > 0 {
			errors = append(errors, detail)
		}
	}
	if len(errors) > 0 {
		return common.NewValidationError("invalid product option input", errors...)
	}
	return nil
}
