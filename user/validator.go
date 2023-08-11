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
