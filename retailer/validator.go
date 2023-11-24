package retailer

import "github.com/nayefradwi/zanobia_inventory_manager/common"

func ValidateRetailer(retailer Retailer) error {
	validationResults := make([]common.ErrorDetails, 0)
	validationResults = append(
		validationResults,
		common.ValidateAlphanuemericName(retailer.Name, "Name"),
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
	if err := ValidateRetailerContacts(retailer.Contacts); err != nil {
		return err
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
		common.ValidateStringLength(contact.Email, "Email", 0, 255),
		common.ValidateStringLength(contact.Name, "Name", 1, 50),
		common.ValidateAlphanuemericName(contact.Name, "Name"),
		common.ValidateStringLength(contact.Phone, "Phone", 1, 50),
		common.ValidateAlphanuemericName(contact.Position, "Position"),
		common.ValidateStringLength(contact.Position, "Position", 1, 50),
	)
	if contact.Website != "" {
		validationResults = append(validationResults, common.ValidateUrl(&contact.Website, "Website"))
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
