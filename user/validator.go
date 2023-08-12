package user

import (
	"regexp"

	"github.com/nayefradwi/zanobia_inventory_manager/common"
)

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
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(pattern)
	if !re.MatchString(email) {
		return common.ErrorDetails{
			Message: "invalid email address",
			Field:   "email",
		}
	}
	return common.ErrorDetails{}
}

func ValidatePassword(password string) common.ErrorDetails {
	isAtLeast8 := len(password) >= 8
	hasDigit, _ := regexp.MatchString(`\d`, password)
	hasSymbol, _ := regexp.MatchString(`[!@#$%^&*()_+{}[\]:;<>,.?~\\/-]`, password)
	if !hasDigit || !hasSymbol || !isAtLeast8 {
		return common.ErrorDetails{
			Message: "password must be at least 8 characters long and contain at least one number and one special character",
			Field:   "password",
		}
	}
	return common.ErrorDetails{}
}

func ValidateFirstName(firstName string) common.ErrorDetails {

	if !isNameValid(firstName) {
		return common.ErrorDetails{
			Message: "invalid first name",
			Field:   "firstName",
		}
	}
	return common.ErrorDetails{}
}

func ValidateLastName(lastName string) common.ErrorDetails {
	if !isNameValid(lastName) {
		return common.ErrorDetails{
			Message: "invalid last name",
			Field:   "lastName",
		}
	}
	return common.ErrorDetails{}
}

func isNameValid(name string) bool {
	pattern := `^[A-Za-z]+([-'][A-Za-z]+)*$`
	re := regexp.MustCompile(pattern)
	return re.MatchString(name)
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
