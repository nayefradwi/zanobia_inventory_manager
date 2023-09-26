package product

import (
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
	if !isIngredientNameValid(name) {
		return common.ErrorDetails{
			Message: "invalid ingredient name",
			Field:   "name",
		}
	}
	return common.ErrorDetails{}
}

func isIngredientNameValid(name string) bool {
	pattern := "^[A-Za-z0-9\\s]+$"
	re := regexp.MustCompile(pattern)
	return re.MatchString(name)
}

func ValidateBrandName(name string) common.ErrorDetails {
	if !isBrandNameValid(name) {
		return common.ErrorDetails{
			Message: "invalid brand name",
			Field:   "name",
		}
	}
	return common.ErrorDetails{}
}

func isBrandNameValid(name string) bool {
	pattern := "^[A-Za-z0-9\\s.-]+$"
	re := regexp.MustCompile(pattern)
	return re.MatchString(name)
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
