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
	Reason   string  `json:"reason,omitempty"`
	Cost     float64
}

// what is needed:
// 1. recipe map
// 2. batch bases lookup which includes recipes batch bases and batch bases
// 3. original units which will be used to determine the conversion factor
// 4. batch input map to update will be used with the batch bases lookup to update the batch bases
// 5. batch input map to create will be used to create new batch bases
// 6. recipe batch input map will be used to decrement the recipe batch bases
// 7. sku list will be used to get the original units and batch bases for recipes
// 8. batch ids will be used to get the batch bases
type BulkBatchUpdateRequest struct {
	RecipeMap        map[string]Recipe
	BatchBasesLookup map[string]BatchBase
	OriginalUnitsMap map[string]int
	// need to convert to original units
	BatchInputMapToUpdate map[string]BatchInput
	// need to convert to original units
	BatchInputMapToCreate map[string]BatchInput
	BatchInputLookUp      map[string]BatchInput
	RecipeBatchInputMap   map[string]BatchInput
	SkuList               []string
	BatchIds              []int
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

func ValidateBatchInputsIncrement(inputs []BatchInput) error {
	for _, input := range inputs {
		validationErr := ValidateBatchInputIncrement(input)
		if validationErr != nil {
			return validationErr
		}
	}
	return nil
}

func ValidateBatchInputIncrement(input BatchInput) error {
	validationResults := make([]common.ErrorDetails, 0)
	validationResults = append(validationResults,
		common.ValidateIdPtr(&input.UnitId, "unitId"),
		common.ValidateNotZero(input.Quantity, "quantity"),
		common.ValidateStringLength(input.Sku, "sku", 10, 36),
		common.ValidateAlphanuemericName(input.Reason, "reason"),
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

func ValidateBatchInputsDecrement(inputs []BatchInput) error {
	for _, input := range inputs {
		validationErr := ValidateBatchInputDecrement(input)
		if validationErr != nil {
			return validationErr
		}
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
		common.ValidateAlphanuemericName(input.Reason, "reason"),
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

func (b BatchInput) SetUnitId(unitId int) BatchInput {
	b.UnitId = unitId
	return b
}

func (b BatchInput) SetConvertedQuantity(quantity float64) BatchInput {
	b.Quantity = quantity
	return b
}
