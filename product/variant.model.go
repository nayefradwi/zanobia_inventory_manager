package product

import "github.com/nayefradwi/zanobia_inventory_manager/common"

type Variant struct {
	Id     *int     `json:"id"`
	Name   string   `json:"name"`
	Values []string `json:"values"`
}

func ValidateVariant(variant Variant) error {
	validationResults := make([]common.ErrorDetails, 0)
	validationResults = append(validationResults,
		validateVariantName(variant.Name),
		validateVariantValues(variant.Values),
	)
	errors := make([]common.ErrorDetails, 0)
	for _, result := range validationResults {
		if len(result.Message) > 0 {
			errors = append(errors, result)
		}
	}
	if len(errors) > 0 {
		return common.NewValidationError("invalid variant input", errors...)
	}
	return nil
}

func validateVariantName(name string) common.ErrorDetails {
	if !isNameValid(name) {
		return common.ErrorDetails{
			Message: "invalid variant name",
			Field:   "name",
		}
	}
	return common.ErrorDetails{}
}

func ValidateVariantValues(values []string) error {
	validationResult := validateVariantValues(values)
	if len(validationResult.Message) > 0 {
		return common.NewValidationError("invalid variant values", validationResult)
	}
	return nil
}

func validateVariantValues(values []string) common.ErrorDetails {
	if len(values) == 0 || len(values) > 10 {
		return common.ErrorDetails{
			Message: "variant values must be between 1 and 10",
			Field:   "values",
		}
	}
	for _, value := range values {
		if !isNameValid(value) {
			return common.ErrorDetails{
				Message: "invalid variant value",
				Field:   "values",
			}
		}
	}
	return common.ErrorDetails{}
}
