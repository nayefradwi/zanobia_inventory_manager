package unit

import "github.com/nayefradwi/zanobia_inventory_manager/common"

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
		common.ValidateStringLength(unitInput.Name, "name", 3, 50),
		common.ValidateStringLength(unitInput.Symbol, "symbol", 1, 10),
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
