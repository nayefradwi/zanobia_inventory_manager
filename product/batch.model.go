package product

import (
	"strconv"
	"time"

	"github.com/nayefradwi/zanobia_inventory_manager/common"
)

type BatchInput struct {
	Id       *int    `json:"id,omitempty"`
	Sku      string  `json:"Sku,omitempty"`
	Quantity float64 `json:"quantity"`
	UnitId   int     `json:"unitId"`
}

type BatchBase struct {
	Id          *int      `json:"id,omitempty"`
	WarehouseId *int      `json:"warehouseId,omitempty"`
	Sku         string    `json:"sku"`
	Quantity    float64   `json:"quantity"`
	UnitId      int       `json:"unitId"`
	ExpiresAt   time.Time `json:"expiresAt"`
}

type Batch struct {
	BatchBase
	ProductVariantBase *ProductVariantBase `json:"productVariant,omitempty"`
	Unit               Unit                `json:"unit"`
	ProductName        string              `json:"productName"`
}

func (b BatchBase) SetQuantity(quantity float64) BatchBase {
	b.Quantity = quantity
	return b
}

func (b BatchBase) SetExpiresAt(expiresAt time.Time) BatchBase {
	b.ExpiresAt = expiresAt
	return b
}

func ValidateBatchInputIncrement(input BatchInput) error {
	validationResults := make([]common.ErrorDetails, 0)
	validationResults = append(validationResults,
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

func ValidateBatchInputDecrement(input BatchInput) error {
	validationResults := make([]common.ErrorDetails, 0)
	validationResults = append(validationResults,
		common.ValidateIdPtr(&input.UnitId, "unitId"),
		common.ValidateNotZero(input.Quantity, "quantity"),
		common.ValidateStringLength(input.Sku, "sku", 10, 36),
		common.ValidateIdPtr(input.Id, "id"),
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

func (b Batch) GetCursorValue() []string {
	return []string{
		common.GetUtcDateOnlyStringFromTime(b.ExpiresAt),
		strconv.Itoa(*b.Id),
	}
}
