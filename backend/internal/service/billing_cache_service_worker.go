package service

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/senran-N/sub2api/internal/pkg/logger"
)

// Stop 关闭缓存写入工作池
func (s *BillingCacheService) Stop() {
	s.cacheWriteStopOnce.Do(func() {
		s.stopped.Store(true)

		s.cacheWriteMu.Lock()
		channel := s.cacheWriteChan
		if channel != nil {
			close(channel)
		}
		s.cacheWriteMu.Unlock()

		if channel == nil {
			return
		}
		s.cacheWriteWg.Wait()

		s.cacheWriteMu.Lock()
		if s.cacheWriteChan == channel {
			s.cacheWriteChan = nil
		}
		s.cacheWriteMu.Unlock()
	})
}

func (s *BillingCacheService) startCacheWriteWorkers() {
	channel := make(chan cacheWriteTask, cacheWriteBufferSize)
	s.cacheWriteChan = channel
	for i := 0; i < cacheWriteWorkerCount; i++ {
		s.cacheWriteWg.Add(1)
		go s.cacheWriteWorker(channel)
	}
}

// enqueueCacheWrite 尝试将任务入队，队列满时返回 false（并记录告警）。
func (s *BillingCacheService) enqueueCacheWrite(task cacheWriteTask) (enqueued bool) {
	if s.stopped.Load() {
		s.logCacheWriteDrop(task, "closed")
		return false
	}

	s.cacheWriteMu.RLock()
	defer s.cacheWriteMu.RUnlock()

	if s.cacheWriteChan == nil {
		s.logCacheWriteDrop(task, "closed")
		return false
	}

	select {
	case s.cacheWriteChan <- task:
		return true
	default:
		s.logCacheWriteDrop(task, "full")
		return false
	}
}

func (s *BillingCacheService) cacheWriteWorker(channel <-chan cacheWriteTask) {
	defer s.cacheWriteWg.Done()
	for task := range channel {
		ctx, cancel := context.WithTimeout(context.Background(), cacheWriteTimeout)
		switch task.kind {
		case cacheWriteSetBalance:
			s.setBalanceCache(ctx, task.userID, task.balance)
		case cacheWriteSetSubscription:
			s.setSubscriptionCache(ctx, task.userID, task.groupID, task.subscriptionData)
		case cacheWriteUpdateSubscriptionUsage:
			if s.cache != nil {
				if err := s.cache.UpdateSubscriptionUsage(ctx, task.userID, task.groupID, task.amount); err != nil {
					logger.LegacyPrintf("service.billing_cache", "Warning: update subscription cache failed for user %d group %d: %v", task.userID, task.groupID, err)
				}
			}
		case cacheWriteDeductBalance:
			if s.cache != nil {
				if err := s.cache.DeductUserBalance(ctx, task.userID, task.amount); err != nil {
					logger.LegacyPrintf("service.billing_cache", "Warning: deduct balance cache failed for user %d: %v", task.userID, err)
				}
			}
		case cacheWriteUpdateRateLimitUsage:
			if s.cache != nil {
				if err := s.cache.UpdateAPIKeyRateLimitUsage(ctx, task.apiKeyID, task.amount); err != nil {
					logger.LegacyPrintf("service.billing_cache", "Warning: update rate limit usage cache failed for api key %d: %v", task.apiKeyID, err)
				}
			}
		}
		cancel()
	}
}

// cacheWriteKindName 用于日志中的任务类型标识。
func cacheWriteKindName(kind cacheWriteKind) string {
	switch kind {
	case cacheWriteSetBalance:
		return "set_balance"
	case cacheWriteSetSubscription:
		return "set_subscription"
	case cacheWriteUpdateSubscriptionUsage:
		return "update_subscription_usage"
	case cacheWriteDeductBalance:
		return "deduct_balance"
	case cacheWriteUpdateRateLimitUsage:
		return "update_rate_limit_usage"
	default:
		return "unknown"
	}
}

// logCacheWriteDrop 使用节流方式记录丢弃情况，并汇总丢弃数量。
func (s *BillingCacheService) logCacheWriteDrop(task cacheWriteTask, reason string) {
	var (
		countPtr *uint64
		lastPtr  *int64
	)
	switch reason {
	case "full":
		countPtr = &s.cacheWriteDropFullCount
		lastPtr = &s.cacheWriteDropFullLastLog
	case "closed":
		countPtr = &s.cacheWriteDropClosedCount
		lastPtr = &s.cacheWriteDropClosedLastLog
	default:
		return
	}

	atomic.AddUint64(countPtr, 1)
	now := time.Now().UnixNano()
	last := atomic.LoadInt64(lastPtr)
	if now-last < int64(cacheWriteDropLogInterval) {
		return
	}
	if !atomic.CompareAndSwapInt64(lastPtr, last, now) {
		return
	}
	dropped := atomic.SwapUint64(countPtr, 0)
	if dropped == 0 {
		return
	}
	logger.LegacyPrintf("service.billing_cache", "Warning: cache write queue %s, dropped %d tasks in last %s (latest kind=%s user %d group %d)",
		reason,
		dropped,
		cacheWriteDropLogInterval,
		cacheWriteKindName(task.kind),
		task.userID,
		task.groupID,
	)
}
