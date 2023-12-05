package product

// func (s *BatchService) unlockBatchUpdateRequest(ctx context.Context, batchUpdateRequest BulkBatchUpdateInfo) {
// 	s.unlockBatchesByIds(ctx, batchUpdateRequest.BatchIds)
// 	s.unlockBatchesBySkus(ctx, batchUpdateRequest.SkuList)

// }
// func (s *BatchService) lockBatchUpdateRequest(ctx context.Context, batchUpdateRequest BulkBatchUpdateInfo) error {
// 	_, lockErrForIds := s.lockBatchesByIds(ctx, batchUpdateRequest.BatchIds)
// 	if lockErrForIds != nil {
// 		return lockErrForIds
// 	}
// 	_, lockErrForSkus := s.lockBatchesBySkus(ctx, batchUpdateRequest.SkuList)
// 	if lockErrForSkus != nil {
// 		return lockErrForSkus
// 	}
// 	return nil
// }

// func (s *BatchService) lockBatchesBySkus(ctx context.Context, skus []string) ([]common.Lock, error) {
// 	locks := make([]common.Lock, len(skus))
// 	for i, sku := range skus {
// 		lockKey := "batch:" + sku + ":lock"
// 		lock, err := s.lockingService.Acquire(ctx, lockKey)
// 		if err != nil {
// 			return nil, err
// 		}
// 		locks[i] = lock
// 	}
// 	return locks, nil
// }

// func (s *BatchService) lockBatchesByIds(ctx context.Context, ids []int) ([]common.Lock, error) {
// 	locks := make([]common.Lock, len(ids))
// 	for i, id := range ids {
// 		lockKey := "batch:" + strconv.Itoa(id) + ":lock"
// 		lock, err := s.lockingService.Acquire(ctx, lockKey)
// 		if err != nil {
// 			return nil, err
// 		}
// 		locks[i] = lock
// 	}
// 	return locks, nil
// }

// func (s *BatchService) unlockBatchesByIds(ctx context.Context, ids []int) {
// 	for _, id := range ids {
// 		lockKey := "batch:" + strconv.Itoa(id) + ":lock"
// 		s.lockingService.Release(ctx, common.Lock{Name: lockKey})
// 	}
// }

// func (s *BatchService) unlockBatchesBySkus(ctx context.Context, skus []string) {
// 	for _, sku := range skus {
// 		lockKey := "batch:" + sku + ":lock"
// 		s.lockingService.Release(ctx, common.Lock{Name: lockKey})
// 	}
// }
