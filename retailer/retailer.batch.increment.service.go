package retailer

import "context"

func (s *RetailerBatchService) IncrementBatch(ctx context.Context, batchInput RetailerBatchInput) error {
	return s.BulkIncrementBatch(ctx, []RetailerBatchInput{batchInput})
}

func (s *RetailerBatchService) BulkIncrementBatch(ctx context.Context, inputs []RetailerBatchInput) error {
	if err := ValidateBatchInputsIncrement(inputs); err != nil {
		return err
	}
}
