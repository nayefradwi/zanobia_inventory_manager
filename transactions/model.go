package transactions

import (
	"context"
	"time"

	"github.com/nayefradwi/zanobia_inventory_manager/unit"
	"github.com/nayefradwi/zanobia_inventory_manager/user"
	"github.com/nayefradwi/zanobia_inventory_manager/warehouse"
)

const (
	TransactionReasonTypeSold          = "sold"
	TransactionReasonTypeBought        = "bought"
	TransactionReasonTypeExpired       = "expired"
	TransactionReasonTypeDamaged       = "damaged"
	TransactionReasonTypeLost          = "lost"
	TransactionReasonTypeFound         = "found"
	TransactionReasonTypeReturn        = "return"
	TransactionReasonTypeAuditIncrease = "auditIncrease"
	TransactionReasonTypeAuditDecrease = "auditDecrease"
	TransactionReasonTypeRecipeUse     = "recipeUse"
	TransactionReasonTypeProduced      = "produced"
	TransactionReasonTypeTransferIn    = "transferIn"
	TransactionReasonTypeTransferOut   = "transferOut"
)

type TransactionReason struct {
	Id          *int   `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"Description,omitempty"`
	IsPositive  bool   `json:"isPositive"`
}

type Transaction struct {
	Id              *int              `json:"id,omitempty"`
	UserId          *int              `json:"userId,omitempty"`
	BatchId         *int              `json:"batchId,omitempty"`
	RetailerBatchId *int              `json:"retailerBatchId,omitempty"`
	WarehouseId     *int              `json:"warehouseId,omitempty"`
	RetailerId      *int              `json:"retailerId,omitempty"`
	Quantity        float64           `json:"quantity"`
	Unit            *unit.Unit        `json:"unit,omitempty"`
	Amount          float64           `json:"amount,omitempty"`
	Reason          TransactionReason `json:"reason,omitempty"`
	Comment         string            `json:"comment,omitempty"`
	Sku             string            `json:"sku,omitempty"`
	CreatedAt       time.Time         `json:"createdAt,omitempty"`
}

type transactionInput struct {
	UserId          *int    `json:"userId,omitempty"`
	BatchId         *int    `json:"batchId,omitempty"`
	RetailerBatchId *int    `json:"retailerBatchId,omitempty"`
	WarehouseId     *int    `json:"warehouseId,omitempty"`
	RetailerId      *int    `json:"retailerId,omitempty"`
	Quantity        float64 `json:"quantity,omitempty"`
	UnitId          *int    `json:"unitId,omitempty"`
	Amount          float64 `json:"amount,omitempty"`
	Reason          string  `json:"reason,omitempty"`
	Comment         string  `json:"comment,omitempty"`
	Sku             string  `json:"sku,omitempty"`
}

type CreateWarehouseTransactionCommand struct {
	BatchId  int
	Quantity float64
	UnitId   int
	Reason   string
	Cost     float64
	Comment  string
	Sku      string
}

type CreateRetailerTransactionCommand struct {
	RetailerBatchId int
	RetailerId      int
	Quantity        float64
	UnitId          int
	Reason          string
	Cost            float64
	Comment         string
	Sku             string
}

func ForWarehouseTransactions(ctx context.Context, command CreateWarehouseTransactionCommand) (transactionInput, error) {
	if err := ValidateWarehouseTransactionCommand(command); err != nil {
		return transactionInput{}, err
	}
	userId := user.GetUserFromContext(ctx).Id
	warehouseId := warehouse.GetWarehouseId(ctx)
	return transactionInput{
		UserId:      &userId,
		BatchId:     &command.BatchId,
		WarehouseId: &warehouseId,
		Quantity:    command.Quantity,
		UnitId:      &command.UnitId,
		Amount:      command.Cost,
		Reason:      command.Reason,
		Comment:     command.Comment,
		Sku:         command.Sku,
	}, nil
}

func ForRetailerTransactions(ctx context.Context, command CreateRetailerTransactionCommand) (transactionInput, error) {
	if err := ValidateRetailerTransactionCommand(command); err != nil {
		return transactionInput{}, err
	}
	userId := user.GetUserFromContext(ctx).Id
	warehouseId := warehouse.GetWarehouseId(ctx)
	return transactionInput{
		UserId:          &userId,
		RetailerBatchId: &command.RetailerBatchId,
		RetailerId:      &command.RetailerId,
		Quantity:        command.Quantity,
		UnitId:          &command.UnitId,
		Amount:          command.Cost,
		Reason:          command.Reason,
		Comment:         command.Comment,
		Sku:             command.Sku,
		WarehouseId:     &warehouseId,
	}, nil
}

var initalTransactionReasons = []TransactionReason{
	{
		Name:        TransactionReasonTypeSold,
		Description: "Sold to customer",
		IsPositive:  false,
	},
	{
		Name:        TransactionReasonTypeBought,
		Description: "Bought from supplier",
		IsPositive:  true,
	},
	{
		Name:        TransactionReasonTypeExpired,
		Description: "Expired",
		IsPositive:  false,
	},
	{
		Name:        TransactionReasonTypeDamaged,
		Description: "Damaged",
		IsPositive:  false,
	},
	{
		Name:        TransactionReasonTypeLost,
		Description: "Lost",
		IsPositive:  false,
	},
	{
		Name:        TransactionReasonTypeFound,
		Description: "Found",
		IsPositive:  true,
	},
	{
		Name:        TransactionReasonTypeReturn,
		Description: "Returned to supplier",
		IsPositive:  false,
	},
	{
		Name:        TransactionReasonTypeAuditIncrease,
		Description: "Audit increase",
		IsPositive:  true,
	},
	{
		Name:        TransactionReasonTypeAuditDecrease,
		Description: "Audit decrease",
		IsPositive:  false,
	},
	{
		Name:        TransactionReasonTypeRecipeUse,
		Description: "Recipe use",
		IsPositive:  false,
	},
	{
		Name:        TransactionReasonTypeProduced,
		Description: "Produced",
		IsPositive:  true,
	},
}
