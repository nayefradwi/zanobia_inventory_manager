package product

import "github.com/nayefradwi/zanobia_inventory_manager/common"

func ValidateUnitConversion(conversion UnitConversion) error {
	if conversion.UnitId == conversion.ConversionUnitId {
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
