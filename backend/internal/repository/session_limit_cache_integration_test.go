//go:build integration

package repository

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type SessionLimitCacheSuite struct {
	IntegrationRedisSuite
	cache *sessionLimitCache
}

func TestSessionLimitCacheSuite(t *testing.T) {
	suite.Run(t, new(SessionLimitCacheSuite))
}

func (s *SessionLimitCacheSuite) SetupTest() {
	s.IntegrationRedisSuite.SetupTest()
	s.cache = NewSessionLimitCache(s.rdb, 5).(*sessionLimitCache)
}

func (s *SessionLimitCacheSuite) TestGetWindowCost_IncludesLiveReservations() {
	accountID := int64(101)
	windowStart := time.Unix(1_700_000_000, 0).UTC()

	require.NoError(s.T(), s.cache.SetWindowCost(s.ctx, accountID, windowStart, 40))

	allowed, total, err := s.cache.ReserveWindowCost(s.ctx, accountID, windowStart, "req-a", 15, 100, time.Minute)
	require.NoError(s.T(), err)
	require.True(s.T(), allowed)
	require.InDelta(s.T(), 55, total, 1e-9)

	cost, hit, err := s.cache.GetWindowCost(s.ctx, accountID, windowStart)
	require.NoError(s.T(), err)
	require.True(s.T(), hit)
	require.InDelta(s.T(), 55, cost, 1e-9)

	batch, err := s.cache.GetWindowCostBatch(s.ctx, map[int64]time.Time{accountID: windowStart})
	require.NoError(s.T(), err)
	require.InDelta(s.T(), 55, batch[accountID], 1e-9)

	require.NoError(s.T(), s.cache.ReleaseWindowCost(s.ctx, accountID, windowStart, "req-a"))

	cost, hit, err = s.cache.GetWindowCost(s.ctx, accountID, windowStart)
	require.NoError(s.T(), err)
	require.True(s.T(), hit)
	require.InDelta(s.T(), 40, cost, 1e-9)
}

func (s *SessionLimitCacheSuite) TestGetWindowCost_IgnoresExpiredActualAndReservations() {
	accountID := int64(102)
	windowStart := time.Unix(1_700_000_600, 0).UTC()
	key := windowCostKey(accountID, windowStart)

	now, err := s.cache.currentUnixSecond(s.ctx)
	require.NoError(s.T(), err)

	require.NoError(s.T(), s.rdb.HSet(s.ctx, key,
		"actual", 40,
		"actual_exp", now-1,
		"r:expired", 9,
		"e:expired", now-1,
		"r:live", 7,
		"e:live", now+60,
	).Err())

	cost, hit, err := s.cache.GetWindowCost(s.ctx, accountID, windowStart)
	require.NoError(s.T(), err)
	require.True(s.T(), hit)
	require.InDelta(s.T(), 7, cost, 1e-9)

	batch, err := s.cache.GetWindowCostBatch(s.ctx, map[int64]time.Time{accountID: windowStart})
	require.NoError(s.T(), err)
	require.InDelta(s.T(), 7, batch[accountID], 1e-9)

	expiredReservationExists, err := s.rdb.HExists(s.ctx, key, "r:expired").Result()
	require.NoError(s.T(), err)
	require.False(s.T(), expiredReservationExists)

	expiredExpiryExists, err := s.rdb.HExists(s.ctx, key, "e:expired").Result()
	require.NoError(s.T(), err)
	require.False(s.T(), expiredExpiryExists)
}

func (s *SessionLimitCacheSuite) TestSetWindowCost_PreservesHigherFreshActual() {
	accountID := int64(103)
	windowStart := time.Unix(1_700_001_200, 0).UTC()

	require.NoError(s.T(), s.cache.SetWindowCost(s.ctx, accountID, windowStart, 12))
	require.NoError(s.T(), s.cache.SetWindowCost(s.ctx, accountID, windowStart, 9))

	cost, hit, err := s.cache.GetWindowCost(s.ctx, accountID, windowStart)
	require.NoError(s.T(), err)
	require.True(s.T(), hit)
	require.InDelta(s.T(), 12, cost, 1e-9)

	require.NoError(s.T(), s.cache.SetWindowCost(s.ctx, accountID, windowStart, 18))

	cost, hit, err = s.cache.GetWindowCost(s.ctx, accountID, windowStart)
	require.NoError(s.T(), err)
	require.True(s.T(), hit)
	require.InDelta(s.T(), 18, cost, 1e-9)
}

func (s *SessionLimitCacheSuite) TestWindowCost_IsolatedByWindowStart() {
	accountID := int64(104)
	oldWindowStart := time.Unix(1_700_001_800, 0).UTC()
	newWindowStart := oldWindowStart.Add(time.Hour)

	require.NoError(s.T(), s.cache.SetWindowCost(s.ctx, accountID, oldWindowStart, 21))
	require.NoError(s.T(), s.cache.SetWindowCost(s.ctx, accountID, newWindowStart, 3))

	oldCost, oldHit, err := s.cache.GetWindowCost(s.ctx, accountID, oldWindowStart)
	require.NoError(s.T(), err)
	require.True(s.T(), oldHit)
	require.InDelta(s.T(), 21, oldCost, 1e-9)

	newCost, newHit, err := s.cache.GetWindowCost(s.ctx, accountID, newWindowStart)
	require.NoError(s.T(), err)
	require.True(s.T(), newHit)
	require.InDelta(s.T(), 3, newCost, 1e-9)
}
