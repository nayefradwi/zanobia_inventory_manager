package batchlocking

import (
	"context"
	"strconv"

	"github.com/nayefradwi/zanobia_inventory_manager/common"
)

func UnlockBatchUpdateRequest(ctx context.Context, lockingService common.IDistributedLockingService, locks []common.Lock) {
	for _, lock := range locks {
		lockingService.Release(ctx, lock)
	}
}

func LockBatchUpdateRequest(
	ctx context.Context,
	lockingService common.IDistributedLockingService,
	ids []int,
	skuList []string,
	getKey func(string) string,
) ([]common.Lock, error) {
	idLocks, lockErrForIds := lockBatchesByIds(ctx, lockingService, ids, getKey)
	if lockErrForIds != nil {
		return idLocks, lockErrForIds
	}
	skuLocks, lockErrForSkus := lockBatchesBySkus(ctx, lockingService, skuList, getKey)
	idLocks = append(idLocks, skuLocks...)
	return idLocks, lockErrForSkus
}

func lockBatchesBySkus(
	ctx context.Context,
	lockingService common.IDistributedLockingService,
	skus []string,
	getKey func(string) string,
) ([]common.Lock, error) {
	locks := make([]common.Lock, len(skus))
	for i, sku := range skus {
		lockKey := getKey(sku)
		lock, err := lockingService.Acquire(ctx, lockKey)
		if err != nil {
			return locks, common.NewBadRequestFromMessage("Failed to acquire lock for sku: " + sku)
		}
		locks[i] = lock
	}
	return locks, nil
}

func lockBatchesByIds(
	ctx context.Context,
	lockingService common.IDistributedLockingService,
	ids []int,
	getKey func(string) string,
) ([]common.Lock, error) {
	locks := make([]common.Lock, len(ids))
	for i, id := range ids {
		lockKey := getKey(strconv.Itoa(id))
		lock, err := lockingService.Acquire(ctx, lockKey)
		if err != nil {
			return locks, common.NewBadRequestFromMessage("Failed to acquire lock for id: " + strconv.Itoa(id))
		}
		locks[i] = lock
	}
	return locks, nil
}
