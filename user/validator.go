package user

import "github.com/nayefradwi/zanobia_inventory_manager/common"

func ValidateUser(userInput UserInput) error {
	errors := make([]common.ErrorDetails, 0)
	errors = append(errors,
		ValidateEmail(userInput.Email),
		ValidatePassword(userInput.Password),
		ValidateFirstName(userInput.FirstName),
		ValidateLastName(userInput.LastName),
	)
	for _, err := range errors {
		if len(err.Message) > 0 {
			return common.NewValidationError("invalid user input", errors...)
		}
	}
	return nil
}

func ValidateEmail(email string) common.ErrorDetails {
	return common.ErrorDetails{}
}

func ValidatePassword(password string) common.ErrorDetails {
	return common.ErrorDetails{}
}

func ValidateFirstName(firstName string) common.ErrorDetails {
	return common.ErrorDetails{}
}

func ValidateLastName(lastName string) common.ErrorDetails {
	return common.ErrorDetails{}
}

func ValidatePermission(permission Permission) error {
	errors := make([]common.ErrorDetails, 0)
	errors = append(errors,
		ValidatePermissionName(permission.Name),
		ValidatePermissionDescription(permission.Description),
	)
	for _, err := range errors {
		if len(err.Message) > 0 {
			return common.NewValidationError("invalid permission input", errors...)
		}
	}
	return nil
}

func ValidatePermissionName(name string) common.ErrorDetails {
	if len(name) < 3 || len(name) > 50 {
		return common.ErrorDetails{
			Message: "permission name must be between 3 and 50 characters",
			Field:   "name",
		}
	}
	return common.ErrorDetails{}
}

func ValidatePermissionDescription(description string) common.ErrorDetails {
	if len(description) > 255 {
		return common.ErrorDetails{
			Message: "permission description must be less than 255 characters",
			Field:   "description",
		}
	}
	return common.ErrorDetails{}
}
