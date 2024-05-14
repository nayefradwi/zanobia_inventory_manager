package retailer

import (
	"context"

	batchlocking "github.com/nayefradwi/zanobia_inventory_manager/batch_locking"
)

func (s *RetailerBatchService) unlockBatchUpdateRequest(ctx context.Context, batchUpdateRequest BulkRetailerBatchUpdateInfo) {
	batchlocking.UnlockBatchUpdateRequest(ctx, s.lockingService, batchUpdateRequest.locks)
}

func (s *RetailerBatchService) lockBatchUpdateRequest(
	ctx context.Context,
	batchUpdateRequest BulkRetailerBatchUpdateInfo,
) (BulkRetailerBatchUpdateInfo, error) {
	locks, err := batchlocking.LockBatchUpdateRequest(
		ctx,
		s.lockingService,
		batchUpdateRequest.Ids,
		batchUpdateRequest.SkuList,
		s.createBatchLockKey,
	)
	batchUpdateRequest.locks = locks
	return batchUpdateRequest, err
}

func (s *RetailerBatchService) createBatchLockKey(idOrSku string) string {
	return "retailer-batch:" + idOrSku + ":lock"
}
