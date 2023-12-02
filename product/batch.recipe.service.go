package product

import (
	"context"
	"log"

	"github.com/nayefradwi/zanobia_inventory_manager/common"
	"github.com/nayefradwi/zanobia_inventory_manager/transactions"
)

func (s *BatchService) createRecipeBatchInputs(
	ctx context.Context,
	recipeBatches []BatchBase,
	recipeLookUp map[string]Recipe,
	batchInputLookUp map[string]BatchInput,
) (map[string]BatchInput, []int, error) {
	recipeBatchMap := make(map[string]BatchInput, len(recipeBatches))
	recipeBatchBaseIds := make([]int, len(recipeBatches))
	for _, recipeBatchBase := range recipeBatches {
		if recipeBatchBase.Id == nil && recipeBatchBase.Sku == "" {
			return nil, nil, common.NewBadRequestFromMessage("invalid recipe batch")
		}
		recipe, recipeFound := recipeLookUp[recipeBatchBase.Sku]
		if !recipeFound {
			return nil, nil, common.NewBadRequestFromMessage("invalid recipe batch")
		}
		resultBatch, resultFound := batchInputLookUp[recipe.ResultVariantSku]
		if !resultFound {
			return nil, nil, common.NewBadRequestFromMessage("invalid recipe batch")
		}
		recipeBatchInput, err := s.createRecipeBatchInput(ctx, recipe, recipeBatchBase, resultBatch)
		if err != nil {
			return nil, nil, err
		}
		// this will add the quantities of the same recipe batch inputs
		recipeBatchInput = s.mergeRecipBatchInput(recipeBatchInput, recipeBatchMap)
		recipeBatchMap[recipe.RecipeVariantSku] = recipeBatchInput
		recipeBatchBaseIds = append(recipeBatchBaseIds, *recipeBatchBase.Id)
	}
	return recipeBatchMap, recipeBatchBaseIds, nil
}

func (s *BatchService) createRecipeBatchInput(
	ctx context.Context,
	recipe Recipe, recipeBatchBase BatchBase,
	resultBatch BatchInput,
) (BatchInput, error) {
	recipeConvertedQuantityToRecipeStandardUnit, err := s.unitService.ConvertUnit(
		ctx,
		ConvertUnitInput{
			ToUnitId:   recipe.IngredientStandardUnit.Id,
			FromUnitId: recipe.Unit.Id,
			Quantity:   recipe.Quantity,
		},
	)
	if err != nil {
		log.Printf("failed to convert unit: %s", err.Error())
		return BatchInput{}, err
	}
	recipeCostPerQty := recipe.IngredientCost *
		recipeConvertedQuantityToRecipeStandardUnit.Quantity *
		resultBatch.Quantity
	recipeBatchInput := BatchInput{
		Id:       recipeBatchBase.Id,
		Quantity: resultBatch.Quantity * recipe.Quantity,
		UnitId:   recipeBatchBase.UnitId,
		Sku:      recipe.RecipeVariantSku,
		Reason:   transactions.TransactionReasonTypeRecipeUse,
		// TODO: check if cost is correct
		Cost: recipeCostPerQty,
	}
	return recipeBatchInput, nil
}

func (s *BatchService) mergeRecipBatchInput(
	recipeInputToMerge BatchInput,
	recipeBatchInputMap map[string]BatchInput,
) BatchInput {
	recipeBatchInput, ok := recipeBatchInputMap[recipeInputToMerge.Sku]
	if !ok {
		return recipeInputToMerge
	}
	recipeBatchInput.Quantity += recipeInputToMerge.Quantity
	return recipeBatchInput
}

func (s *BatchService) getRecipeBatchBases(ctx context.Context, recipeSkus []string) ([]BatchBase, error) {
	useMostExpired := common.GetBoolFromContext(ctx, UseMostExpiredKey{})
	if useMostExpired {
		return s.batchRepo.getMostExpiredBatchBasesFromSkus(ctx, recipeSkus)
	}
	return s.batchRepo.getLeastExpiredBatchBasesFromSkus(ctx, recipeSkus)
}
