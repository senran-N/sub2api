package service

import (
	"context"
	"time"

	"github.com/senran-N/sub2api/internal/pkg/logger"
)

func (s *adminServiceImpl) loadGroupAuthCacheKeys(ctx context.Context, groupID int64) []string {
	if s.authCacheInvalidator == nil || s.apiKeyRepo == nil {
		return nil
	}

	keys, err := s.apiKeyRepo.ListKeysByGroupID(ctx, groupID)
	if err != nil {
		return nil
	}
	return keys
}

func (s *adminServiceImpl) invalidateDeletedGroupSubscriptions(affectedUserIDs []int64, groupID int64) {
	if len(affectedUserIDs) == 0 || s.billingCacheService == nil {
		return
	}

	go func() {
		cacheCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		for _, userID := range affectedUserIDs {
			if err := s.billingCacheService.InvalidateSubscription(cacheCtx, userID, groupID); err != nil {
				logger.LegacyPrintf("service.admin", "invalidate subscription cache failed: user_id=%d group_id=%d err=%v", userID, groupID, err)
			}
		}
	}()
}

func (s *adminServiceImpl) invalidateDeletedGroupAuthCache(ctx context.Context, keys []string) {
	if s.authCacheInvalidator == nil {
		return
	}
	for _, key := range keys {
		s.authCacheInvalidator.InvalidateAuthCacheByKey(ctx, key)
	}
}
