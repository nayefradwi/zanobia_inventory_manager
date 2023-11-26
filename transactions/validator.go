package transactions

import "github.com/nayefradwi/zanobia_inventory_manager/common"

func ValidateTransactionReason(reason TransactionReason) error {
	validationResults := make([]common.ErrorDetails, 0)
	validationResults = append(validationResults,
		common.ValidateStringLength(reason.Name, "name", 3, 50),
		common.ValidateStringLength(reason.Description, "description", 0, 255),
	)
	errors := make([]common.ErrorDetails, 0)
	for _, result := range validationResults {
		if len(result.Message) > 0 {
			errors = append(errors, result)
		}
	}
	if len(errors) > 0 {
		return common.NewValidationError("invalid transaction reason input", errors...)
	}
	return nil
}

func ValidateWarehouseTransactionCommand(command CreateWarehouseTransactionCommand) error {
	validationResults := make([]common.ErrorDetails, 0)
	validationResults = append(validationResults,
		common.ValidateId(command.BatchId, "batchId"),
		common.ValidateAmountPositive(command.Quantity, "quantity"),
		common.ValidateId(command.UnitId, "unitId"),
		common.ValidateStringLength(command.Reason, "reason", 3, 50),
		common.ValidateAmountPositive(command.CostPerQty, "costPerQty"),
		common.ValidateStringLength(command.Comment, "comment", 0, 255),
		common.ValidateStringLength(command.Sku, "sku", 10, 36),
	)
	errors := make([]common.ErrorDetails, 0)
	for _, result := range validationResults {
		if len(result.Message) > 0 {
			errors = append(errors, result)
		}
	}
	if len(errors) > 0 {
		return common.NewValidationError("invalid transaction input", errors...)
	}
	return nil
}

func ValidateRetailerTransactionCommand(command CreateRetailerTransactionCommand) error {
	validationResults := make([]common.ErrorDetails, 0)
	validationResults = append(validationResults,
		common.ValidateId(command.RetailerBatchId, "retailerBatchId"),
		common.ValidateId(command.RetailerId, "retailerId"),
		common.ValidateAmountPositive(command.Quantity, "quantity"),
		common.ValidateId(command.UnitId, "unitId"),
		common.ValidateStringLength(command.Reason, "reason", 3, 50),
		common.ValidateAmountPositive(command.CostPerQty, "costPerQty"),
		common.ValidateStringLength(command.Comment, "comment", 0, 255),
		common.ValidateStringLength(command.Sku, "sku", 10, 36),
	)
	errors := make([]common.ErrorDetails, 0)
	for _, result := range validationResults {
		if len(result.Message) > 0 {
			errors = append(errors, result)
		}
	}
	if len(errors) > 0 {
		return common.NewValidationError("invalid transaction input", errors...)
	}
	return nil
}
