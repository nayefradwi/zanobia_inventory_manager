package product

import "github.com/nayefradwi/zanobia_inventory_manager/common"

type VariantInput struct {
	Id     *int     `json:"id"`
	Name   string   `json:"name"`
	Values []string `json:"values"`
}

type VariantValue struct {
	Id    int    `json:"id"`
	Value string `json:"value"`
}

type Variant struct {
	Id     *int           `json:"id"`
	Name   string         `json:"name"`
	Values []VariantValue `json:"values"`
}

func ValidateVariant(variant VariantInput) error {
	validationResults := make([]common.ErrorDetails, 0)
	validationResults = append(validationResults,
		common.ValidateAlphanuemericName(variant.Name, "name"),
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

func ValidateVariantValues(values []string) error {
	validationResult := validateVariantValues(values)
	if len(validationResult.Message) > 0 {
		return common.NewValidationError("invalid variant values", validationResult)
	}
	return nil
}

func validateVariantValues(values []string) common.ErrorDetails {
	if len(values) == 0 || len(values) > 100 {
		return common.ErrorDetails{
			Message: "variant values must be between 1 and 100",
			Field:   "values",
		}
	}
	for _, value := range values {
		result := common.ValidateAlphanuemericName(value, "variant value")
		if len(result.Message) > 0 {
			return result
		}
	}
	return common.ErrorDetails{}
}
