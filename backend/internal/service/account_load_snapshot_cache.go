package service

import (
	"context"
	"sync"
	"time"
)

const requestAccountLoadSnapshotTTL = 250 * time.Millisecond

type requestAccountLoadCacheContextKey struct{}

var requestAccountLoadCacheKey = requestAccountLoadCacheContextKey{}

type requestAccountLoadSnapshot struct {
	currentConcurrency int
	waitingCount       int
	expiresAtUnixNano  int64
}

type requestAccountLoadCache struct {
	mu      sync.Mutex
	entries map[int64]requestAccountLoadSnapshot
}

func buildAccountLoadInfo(account AccountWithConcurrency, currentConcurrency, waitingCount int) *AccountLoadInfo {
	loadRate := 0
	if account.MaxConcurrency > 0 {
		loadRate = (currentConcurrency + waitingCount) * 100 / account.MaxConcurrency
	}
	return &AccountLoadInfo{
		AccountID:          account.ID,
		CurrentConcurrency: currentConcurrency,
		WaitingCount:       waitingCount,
		LoadRate:           loadRate,
	}
}

func WithRequestAccountLoadCache(ctx context.Context) context.Context {
	if ctx == nil {
		return nil
	}
	if requestAccountLoadCacheFromContext(ctx) != nil {
		return ctx
	}
	return context.WithValue(ctx, requestAccountLoadCacheKey, &requestAccountLoadCache{
		entries: make(map[int64]requestAccountLoadSnapshot),
	})
}

func requestAccountLoadCacheFromContext(ctx context.Context) *requestAccountLoadCache {
	if ctx == nil {
		return nil
	}
	cache, _ := ctx.Value(requestAccountLoadCacheKey).(*requestAccountLoadCache)
	return cache
}

func buildAccountLoadInfoFromSnapshot(account AccountWithConcurrency, snapshot requestAccountLoadSnapshot) *AccountLoadInfo {
	return buildAccountLoadInfo(account, snapshot.currentConcurrency, snapshot.waitingCount)
}

func (c *requestAccountLoadCache) get(accounts []AccountWithConcurrency, now time.Time) (map[int64]*AccountLoadInfo, []AccountWithConcurrency) {
	if c == nil || len(accounts) == 0 {
		return nil, accounts
	}

	nowUnixNano := now.UnixNano()
	hits := make(map[int64]*AccountLoadInfo, len(accounts))
	missing := make([]AccountWithConcurrency, 0, len(accounts))

	c.mu.Lock()
	defer c.mu.Unlock()

	for _, account := range accounts {
		snapshot, ok := c.entries[account.ID]
		if !ok || snapshot.expiresAtUnixNano <= nowUnixNano {
			if ok {
				delete(c.entries, account.ID)
			}
			missing = append(missing, account)
			continue
		}
		hits[account.ID] = buildAccountLoadInfoFromSnapshot(account, snapshot)
	}

	return hits, missing
}

func (c *requestAccountLoadCache) store(loadMap map[int64]*AccountLoadInfo, now time.Time) {
	if c == nil || len(loadMap) == 0 {
		return
	}

	expiresAtUnixNano := now.Add(requestAccountLoadSnapshotTTL).UnixNano()

	c.mu.Lock()
	defer c.mu.Unlock()

	for accountID, info := range loadMap {
		if info == nil {
			continue
		}
		c.entries[accountID] = requestAccountLoadSnapshot{
			currentConcurrency: info.CurrentConcurrency,
			waitingCount:       info.WaitingCount,
			expiresAtUnixNano:  expiresAtUnixNano,
		}
	}
}
