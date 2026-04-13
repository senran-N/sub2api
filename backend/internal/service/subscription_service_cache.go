package service

import (
	"context"
	"log"
	"math/rand/v2"
	"strconv"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/senran-N/sub2api/internal/config"
)

const subscriptionCacheInvalidateTimeout = 5 * time.Second

func (s *SubscriptionService) initSubCache(cfg *config.Config) {
	if cfg == nil {
		return
	}
	sc := cfg.SubscriptionCache
	if sc.L1Size <= 0 || sc.L1TTLSeconds <= 0 {
		return
	}

	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: int64(sc.L1Size) * 10,
		MaxCost:     int64(sc.L1Size),
		BufferItems: 64,
	})
	if err != nil {
		log.Printf("Warning: failed to init subscription L1 cache: %v", err)
		return
	}

	s.subCacheL1 = cache
	s.subCacheTTL = time.Duration(sc.L1TTLSeconds) * time.Second
	s.subCacheJitter = sc.JitterPercent
}

func subCacheKey(userID, groupID int64) string {
	return "sub:" + strconv.FormatInt(userID, 10) + ":" + strconv.FormatInt(groupID, 10)
}

func (s *SubscriptionService) jitteredTTL(ttl time.Duration) time.Duration {
	if ttl <= 0 || s.subCacheJitter <= 0 {
		return ttl
	}
	pct := s.subCacheJitter
	if pct > 100 {
		pct = 100
	}
	delta := float64(pct) / 100
	factor := 1 - delta + rand.Float64()*(2*delta)
	if factor <= 0 {
		return ttl
	}
	return time.Duration(float64(ttl) * factor)
}

func (s *SubscriptionService) InvalidateSubCache(userID, groupID int64) {
	if s.subCacheL1 == nil {
		return
	}
	s.subCacheL1.Del(subCacheKey(userID, groupID))
}

func (s *SubscriptionService) invalidateSubscriptionCaches(ctx context.Context, userID, groupID int64) {
	s.InvalidateSubCache(userID, groupID)
	if s.billingCacheService == nil {
		return
	}

	go func(parent context.Context) {
		cacheCtx, cancel := newDetachedTimeoutContext(parent, subscriptionCacheInvalidateTimeout)
		defer cancel()
		_ = s.billingCacheService.InvalidateSubscription(cacheCtx, userID, groupID)
	}(ctx)
}

func (s *SubscriptionService) GetActiveSubscription(ctx context.Context, userID, groupID int64) (*UserSubscription, error) {
	key := subCacheKey(userID, groupID)

	if s.subCacheL1 != nil {
		if value, ok := s.subCacheL1.Get(key); ok {
			if sub, ok := value.(*UserSubscription); ok {
				cp := *sub
				return &cp, nil
			}
		}
	}

	value, err, _ := s.subCacheGroup.Do(key, func() (any, error) {
		sub, getErr := s.userSubRepo.GetActiveByUserIDAndGroupID(ctx, userID, groupID)
		if getErr != nil {
			return nil, getErr
		}
		if s.subCacheL1 != nil {
			_ = s.subCacheL1.SetWithTTL(key, sub, 1, s.jitteredTTL(s.subCacheTTL))
		}
		return sub, nil
	})
	if err != nil {
		return nil, err
	}

	sub, ok := value.(*UserSubscription)
	if !ok || sub == nil {
		return nil, ErrSubscriptionNotFound
	}
	cp := *sub
	return &cp, nil
}
