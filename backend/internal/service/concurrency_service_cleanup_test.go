package service

import (
	"context"
	"testing"
	"time"
)

type sweepingConcurrencyCacheForCleanupTest struct {
	requestCacheTestConcurrencyCache
	cleanupAllCalled chan struct{}
}

func (c *sweepingConcurrencyCacheForCleanupTest) CleanupExpiredAccountSlotsAll(_ context.Context) error {
	select {
	case c.cleanupAllCalled <- struct{}{}:
	default:
	}
	return nil
}

func TestStartSlotCleanupWorker_UsesCacheSweeperFastPath(t *testing.T) {
	cache := &sweepingConcurrencyCacheForCleanupTest{
		cleanupAllCalled: make(chan struct{}, 1),
	}
	svc := NewConcurrencyService(cache)

	svc.StartSlotCleanupWorker(nil, 10*time.Millisecond)

	select {
	case <-cache.cleanupAllCalled:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("expected fast-path sweeper to be called")
	}
}
