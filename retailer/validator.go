package retailer

import "github.com/nayefradwi/zanobia_inventory_manager/common"

func ValidateRetailer(retailer Retailer) error {
	validationResults := make([]common.ErrorDetails, 0)
	if err := ValidateRetailerContacts(retailer.Contacts); err != nil {
		return err
	}
	validationResults = append(
		validationResults,
		common.ValidateAlphanuemericName("Name", retailer.Name),
		common.ValidateStringLength("Name", retailer.Name, 1, 50),
	)
	errors := make([]common.ErrorDetails, 0)
	for _, result := range validationResults {
		if len(result.Message) > 0 {
			errors = append(errors, result)
		}
	}
	if len(errors) > 0 {
		return common.NewValidationError("invalid retailer input", errors...)
	}
	return nil
}

func ValidateRetailerContacts(contacts []RetailerContact) error {
	for _, contact := range contacts {
		if err := ValidateRetailerContact(contact); err != nil {
			return err
		}
	}
	return nil
}

func ValidateRetailerContact(contact RetailerContact) error {
	validationResults := make([]common.ErrorDetails, 0)
	validationResults = append(
		validationResults,
		common.ValidateStringLength("Email", contact.Email, 0, 255),
		common.ValidateStringLength("Name", contact.Name, 1, 50),
		common.ValidateAlphanuemericName("Name", contact.Name),
		common.ValidateStringLength("Phone", contact.Phone, 1, 50),
		common.ValidateAlphanuemericName("Position", contact.Position),
		common.ValidateStringLength("Position", contact.Position, 1, 50),
	)
	if contact.Website != "" {
		validationResults = append(validationResults, common.ValidateUrl("Website", &contact.Website))
	}
	errors := make([]common.ErrorDetails, 0)
	for _, result := range validationResults {
		if len(result.Message) > 0 {
			errors = append(errors, result)
		}
	}
	if len(errors) > 0 {
		return common.NewValidationError("invalid retailer contact input", errors...)
	}
	return nil
}
