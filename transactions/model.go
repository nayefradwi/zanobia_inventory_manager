package transactions

import (
	"context"

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
	TransactionReasonTypeAuditIncrease = "audit_increase"
	TransactionReasonTypeAuditDecrease = "audit_decrease"
	TransactionReasonTypeRecipeUse     = "recipeuse"
	TransactionReasonTypeProduced      = "produced"
)

type TransactionReason struct {
	Id          *int   `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"Description,omitempty"`
	IsPositive  bool   `json:"isPositive"`
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
	BatchId    int
	Quantity   float64
	UnitId     int
	Reason     string
	CostPerQty float64
	Comment    string
	Sku        string
}

type CreateRetailerTransactionCommand struct {
	RetailerBatchId int
	RetailerId      int
	Quantity        float64
	UnitId          int
	Reason          string
	CostPerQty      float64
	Comment         string
	Sku             string
}

func ForWarehouseTransactions(ctx context.Context, command CreateWarehouseTransactionCommand) (transactionInput, error) {
	if err := ValidateWarehouseTransactionCommand(command); err != nil {
		return transactionInput{}, err
	}
	userId := user.GetUserFromContext(ctx).Id
	warehouseId := warehouse.GetWarehouseId(ctx)
	amount := command.Quantity * command.CostPerQty
	return transactionInput{
		UserId:      &userId,
		BatchId:     &command.BatchId,
		WarehouseId: &warehouseId,
		Quantity:    command.Quantity,
		UnitId:      &command.UnitId,
		Amount:      amount,
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
	amount := command.Quantity * command.CostPerQty
	return transactionInput{
		UserId:          &userId,
		RetailerBatchId: &command.RetailerBatchId,
		RetailerId:      &command.RetailerId,
		Quantity:        command.Quantity,
		UnitId:          &command.UnitId,
		Amount:          amount,
		Reason:          command.Reason,
		Comment:         command.Comment,
		Sku:             command.Sku,
	}, nil
}
