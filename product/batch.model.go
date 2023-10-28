package product

import (
	"time"

	"github.com/nayefradwi/zanobia_inventory_manager/common"
)

type BatchInput struct {
	Sku       string  `json:"Sku,omitempty"`
	Quantity  float64 `json:"quantity"`
	UnitId    int     `json:"unitId"`
	ExpiresAt string  `json:"expiresAt"`
}

type BatchBase struct {
	Id          *int      `json:"id,omitempty"`
	WarehouseId *int      `json:"warehouseId,omitempty"`
	Sku         string    `json:"Sku"`
	Quantity    float64   `json:"quantity"`
	UnitId      int       `json:"unitId"`
	ExpiresAt   time.Time `json:"expiresAt"`
}

type Batch struct {
	BatchBase
	ProductVariantBase *ProductVariantBase `json:"productVariantBase,omitempty"`
	Unit               Unit                `json:"unit"`
}

func (b BatchBase) SetQuantity(quantity float64) BatchBase {
	b.Quantity = quantity
	return b
}

func (b BatchBase) SetExpiresAt(expiresAt time.Time) BatchBase {
	b.ExpiresAt = expiresAt
	return b
}

func ValidateBatchInput(input BatchInput) error {
	validationResults := make([]common.ErrorDetails, 0)
	validationResults = append(validationResults,
		common.ValidateStringLength(input.ExpiresAt, "expiresAt", 20, 27),
		common.ValidateIdPtr(&input.UnitId, "unitId"),
		common.ValidateNotZero(input.Quantity, "quantity"),
		common.ValidateStringLength(input.Sku, "sku", 10, 36),
	)
	errors := make([]common.ErrorDetails, 0)
	for _, result := range validationResults {
		if len(result.Message) > 0 {
			errors = append(errors, result)
		}
	}
	if len(errors) > 0 {
		return common.NewValidationError("invalid batch input", errors...)
	}
	return nil
}
