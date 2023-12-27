package retailer

import (
	"context"
	"strconv"

	"github.com/nayefradwi/zanobia_inventory_manager/common"
)

func (s *RetailerBatchService) unlockBatchUpdateRequest(ctx context.Context, batchUpdateRequest BulkRetailerBatchUpdateInfo) {
	for _, lock := range batchUpdateRequest.locks {
		s.lockingService.Release(ctx, lock)
	}
}

func (s *RetailerBatchService) lockBatchUpdateRequest(
	ctx context.Context,
	batchUpdateRequest BulkRetailerBatchUpdateInfo,
) (BulkRetailerBatchUpdateInfo, error) {
	idLocks, lockErrForIds := s.lockBatchesByIds(ctx, batchUpdateRequest.Ids)
	batchUpdateRequest.locks = idLocks
	if lockErrForIds != nil {
		return batchUpdateRequest, lockErrForIds
	}
	skuLocks, lockErrForSkus := s.lockBatchesBySkus(ctx, batchUpdateRequest.SkuList)
	batchUpdateRequest.locks = append(batchUpdateRequest.locks, skuLocks...)
	if lockErrForSkus != nil {
		return batchUpdateRequest, lockErrForSkus
	}
	return batchUpdateRequest, nil
}

func (s *RetailerBatchService) lockBatchesBySkus(ctx context.Context, skus []string) ([]common.Lock, error) {
	locks := make([]common.Lock, len(skus))
	for i, sku := range skus {
		lockKey := s.createBatchLockKey(sku)
		lock, err := s.lockingService.Acquire(ctx, lockKey)
		if err != nil {
			return locks, common.NewBadRequestFromMessage("Failed to acquire lock for sku: " + sku)
		}
		locks[i] = lock
	}
	return locks, nil
}

func (s *RetailerBatchService) lockBatchesByIds(ctx context.Context, ids []int) ([]common.Lock, error) {
	locks := make([]common.Lock, len(ids))
	for i, id := range ids {
		lockKey := s.createBatchLockKey(strconv.Itoa(id))
		lock, err := s.lockingService.Acquire(ctx, lockKey)
		if err != nil {
			return locks, common.NewBadRequestFromMessage("Failed to acquire lock for id: " + strconv.Itoa(id))
		}
		locks[i] = lock
	}
	return locks, nil
}

func (s *RetailerBatchService) createBatchLockKey(idOrSku string) string {
	return "retailer-batch:" + idOrSku + ":lock"
}
