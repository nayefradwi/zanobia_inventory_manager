package warehouse

import (
	"regexp"

	"github.com/nayefradwi/zanobia_inventory_manager/common"
)

func ValidateWarehouse(warehouse Warehouse) error {
	validationResults := make([]common.ErrorDetails, 0)
	validationResults = append(validationResults,
		ValidateName(warehouse.Name),
		ValidateLatLng(warehouse.Lat, warehouse.Lng),
	)
	errors := make([]common.ErrorDetails, 0)
	for _, result := range validationResults {
		if len(result.Message) > 0 {
			errors = append(errors, result)
		}
	}
	if len(errors) > 0 {
		return common.NewValidationError("invalid warehouse input", errors...)
	}
	return nil
}

func ValidateName(firstName string) common.ErrorDetails {
	if !isNameValid(firstName) {
		return common.ErrorDetails{
			Message: "invalid first name",
			Field:   "firstName",
		}
	}
	return common.ErrorDetails{}
}

func isNameValid(name string) bool {
	pattern := `^[A-Za-z]+([-'][A-Za-z]+)*$`
	re := regexp.MustCompile(pattern)
	return re.MatchString(name)
}

func ValidateLatLng(lat *float64, lng *float64) common.ErrorDetails {
	if lat == nil || lng == nil {
		return common.ErrorDetails{
			Message: "invalid lat lng",
			Field:   "latlng",
		}
	}
	return common.ErrorDetails{}
}
