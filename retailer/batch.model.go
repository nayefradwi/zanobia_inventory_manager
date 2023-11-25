package retailer

import (
	"strconv"
	"time"

	"github.com/nayefradwi/zanobia_inventory_manager/common"
	"github.com/nayefradwi/zanobia_inventory_manager/product"
)

type RetailerBatchInput struct {
	Id         *int    `json:"id,omitempty"`
	Sku        string  `json:"Sku,omitempty"`
	Quantity   float64 `json:"quantity"`
	UnitId     int     `json:"unitId"`
	RetailerId int     `json:"retailerId"`
}

type RetailerBatchBase struct {
	Id         *int      `json:"id,omitempty"`
	RetailerId *int      `json:"retailerId,omitempty"`
	Sku        string    `json:"sku"`
	Quantity   float64   `json:"quantity"`
	UnitId     int       `json:"unitId,omitempty"`
	ExpiresAt  time.Time `json:"expiresAt"`
}

type RetailerBatch struct {
	RetailerBatchBase
	ProductVariantBase *(product.ProductVariantBase) `json:"productVariant,omitempty"`
	Unit               product.Unit                  `json:"unit"`
	ProductName        string                        `json:"productName"`
	RetailerName       string                        `json:"retailerName"`
}

func (b RetailerBatchBase) SetQuantity(quantity float64) RetailerBatchBase {
	b.Quantity = quantity
	return b
}

func (b RetailerBatchBase) SetExpiresAt(expiresAt time.Time) RetailerBatchBase {
	b.ExpiresAt = expiresAt
	return b
}

func ValidateBatchInputIncrement(input RetailerBatchInput) error {
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
		return common.NewValidationError("invalid retailer batch input", errors...)
	}
	return nil
}

func ValidateBatchInputDecrement(input RetailerBatchInput) error {
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
		return common.NewValidationError("invalid retailer batch input", errors...)
	}
	return nil
}

func (b RetailerBatch) GetCursorValue() []string {
	return []string{
		common.GetUtcDateOnlyStringFromTime(b.ExpiresAt),
		strconv.Itoa(*b.Id),
	}
}
