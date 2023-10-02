package product

import (
	"net/url"
	"regexp"

	"github.com/nayefradwi/zanobia_inventory_manager/common"
)

func ValidateUnitConversion(conversion UnitConversion) error {
	if conversion.ToUnitId == conversion.FromUnitId {
		return common.NewBadRequestFromMessage("Unit and conversion unit cannot be the same")
	}
	if conversion.ConversionFactor <= 0 {
		return common.NewBadRequestFromMessage("Conversion factor must be greater than 0")
	}
	return nil
}

func ValidateUnit(unitInput Unit) error {
	validationResults := make([]common.ErrorDetails, 0)
	validationResults = append(validationResults,
		ValidateUnitName(unitInput.Name),
		ValidateUnitSymbol(unitInput.Symbol),
	)
	errors := make([]common.ErrorDetails, 0)
	for _, result := range validationResults {
		if len(result.Message) > 0 {
			errors = append(errors, result)
		}
	}
	if len(errors) > 0 {
		return common.NewValidationError("invalid unit input", errors...)
	}
	return nil
}

func ValidateUnitName(name string) common.ErrorDetails {
	if len(name) < 3 || len(name) > 50 {
		return common.ErrorDetails{
			Message: "unit name must be between 3 and 50 characters",
			Field:   "name",
		}
	}
	return common.ErrorDetails{}
}

func ValidateUnitSymbol(symbol string) common.ErrorDetails {
	if len(symbol) < 1 || len(symbol) > 10 {
		return common.ErrorDetails{
			Message: "unit symbol must be between 1 and 10 characters",
			Field:   "symbol",
		}
	}
	return common.ErrorDetails{}
}

func ValidateIngredient(ingredient IngredientBase) error {
	validationResults := make([]common.ErrorDetails, 0)
	validationResults = append(validationResults,
		ValidateIngredientName(ingredient.Name),
		ValidateBrandName(ingredient.Brand),
		ValidatePrice(ingredient.Price),
		ValidateExpiresInDays(ingredient.ExpiresInDays),
		ValidateUnitId(ingredient.StandardUnitId),
		ValidateQty(ingredient.StandardQty),
	)
	errors := make([]common.ErrorDetails, 0)
	for _, result := range validationResults {
		if len(result.Message) > 0 {
			errors = append(errors, result)
		}
	}
	if len(errors) > 0 {
		return common.NewValidationError("invalid ingredient input", errors...)
	}
	return nil
}

func ValidateIngredientName(name string) common.ErrorDetails {
	if !isNameValid(name) {
		return common.ErrorDetails{
			Message: "invalid ingredient name",
			Field:   "name",
		}
	}
	return common.ErrorDetails{}
}

func ValidateBrandName(name string) common.ErrorDetails {
	if !isNameValid(name) {
		return common.ErrorDetails{
			Message: "invalid brand name",
			Field:   "name",
		}
	}
	return common.ErrorDetails{}
}

func isNameValid(name string) bool {
	pattern := "^[\\p{L}0-9\\s.-]+$"
	re := regexp.MustCompile(pattern)
	return re.MatchString(name) && len(name) >= 3 && len(name) <= 50
}

func ValidatePrice(price float64) common.ErrorDetails {
	if price <= 0 {
		return common.ErrorDetails{
			Message: "price must be greater than 0",
			Field:   "price",
		}
	}
	return common.ErrorDetails{}
}

func ValidateExpiresInDays(expiresInDays int) common.ErrorDetails {
	if expiresInDays <= 0 {
		return common.ErrorDetails{
			Message: "expires in days must be greater than 0",
			Field:   "expiresInDays",
		}
	}
	return common.ErrorDetails{}
}

func ValidateUnitId(unitId *int) common.ErrorDetails {
	if unitId == nil || *unitId <= 0 {
		return common.ErrorDetails{
			Message: "unit id cannot be empty or less than 0",
			Field:   "unitId",
		}
	}
	return common.ErrorDetails{}
}

func ValidateQty(qty float64) common.ErrorDetails {
	if qty <= 0 {
		return common.ErrorDetails{
			Message: "quantity must be greater than 0",
			Field:   "qty",
		}
	}
	return common.ErrorDetails{}
}

func ValidateInventoryInput(inventoryInput InventoryInput) error {
	validationResults := make([]common.ErrorDetails, 0)
	validationResults = append(validationResults,
		ValidateIngredientId(inventoryInput.IngredientId),
		ValidateUnitId(&inventoryInput.UnitId),
		ValidateQty(inventoryInput.Quantity),
	)
	errors := make([]common.ErrorDetails, 0)
	for _, result := range validationResults {
		if len(result.Message) > 0 {
			errors = append(errors, result)
		}
	}
	if len(errors) > 0 {
		return common.NewValidationError("invalid inventory input", errors...)
	}
	return nil
}

func ValidateIngredientId(ingredientId int) common.ErrorDetails {
	if ingredientId <= 0 {
		return common.ErrorDetails{
			Message: "ingredient id must be a valid id",
			Field:   "ingredientId",
		}
	}
	return common.ErrorDetails{}
}

func ValidateProduct(product ProductBase) error {
	validationResults := make([]common.ErrorDetails, 0)
	validationResults = append(validationResults,
		ValidateProductName(product.Name),
		ValidateProductDescription(product.Description),
		ValidateProductImage(product.Image),
		ValidateProductPrice(product.Price),
		ValidateProductDimension(product.WidthInCm, "widthInCm"),
		ValidateProductDimension(product.HeightInCm, "heightInCm"),
		ValidateProductDimension(product.DepthInCm, "depthInCm"),
		ValidateProductDimension(product.WeightInG, "weightInG"),
		ValidateProductStandardUnitId(product.StandardUnitId),
		ValidateProductCategoryId(product.CategoryId),
		ValidateExpiresInDays(product.ExpiresInDays),
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

func ValidateProductName(name *string) common.ErrorDetails {
	if name == nil || !isNameValid(*name) {
		return common.ErrorDetails{
			Message: "invalid product name",
			Field:   "name",
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

func ValidateProductImage(image *string) common.ErrorDetails {
	if image == nil {
		return common.ErrorDetails{}
	}
	url, err := url.Parse(*image)
	if err != nil || url.Host == "" || url.Scheme == "" {
		return common.ErrorDetails{
			Message: "invalid image url",
			Field:   "image",
		}
	}
	return common.ErrorDetails{}
}

func ValidateProductPrice(price float64) common.ErrorDetails {
	if price <= 0 {
		return common.ErrorDetails{
			Message: "price must be greater than 0",
			Field:   "price",
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
