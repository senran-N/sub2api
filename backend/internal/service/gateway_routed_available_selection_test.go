//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSelectAvailableRoutedAccount_SelectsAndBindsSticky(t *testing.T) {
	cache := &mockGatewayCacheForPlatform{}
	concurrencyCache := &mockConcurrencyCache{}
	svc := &GatewayService{
		cache:              cache,
		cfg:                testConfig(),
		concurrencyService: NewConcurrencyService(concurrencyCache),
	}

	available := []accountWithLoad{
		{account: &Account{ID: 1, Priority: 1, Concurrency: 5}, loadInfo: &AccountLoadInfo{AccountID: 1, LoadRate: 80}},
		{account: &Account{ID: 2, Priority: 1, Concurrency: 5}, loadInfo: &AccountLoadInfo{AccountID: 2, LoadRate: 20}},
	}

	result, ok := svc.selectAvailableRoutedAccount(context.Background(), nil, "route", "claude-3-5-sonnet-20241022", available, time.Second, 1)

	require.True(t, ok)
	require.NotNil(t, result)
	require.NotNil(t, result.Account)
	require.Equal(t, int64(2), result.Account.ID)
	require.Equal(t, int64(2), cache.sessionBindings["route"])
}

func TestSelectAvailableRoutedAccount_ReturnsWaitPlanWhenAllAcquireFail(t *testing.T) {
	concurrencyCache := &mockConcurrencyCache{
		acquireResults: map[int64]bool{1: false, 2: false},
		waitCounts:     map[int64]int{1: 0, 2: 0},
	}
	svc := &GatewayService{
		cfg:                testConfig(),
		concurrencyService: NewConcurrencyService(concurrencyCache),
	}

	available := []accountWithLoad{
		{account: &Account{ID: 1, Priority: 1, Concurrency: 5}, loadInfo: &AccountLoadInfo{AccountID: 1, LoadRate: 10}},
		{account: &Account{ID: 2, Priority: 1, Concurrency: 5}, loadInfo: &AccountLoadInfo{AccountID: 2, LoadRate: 20}},
	}

	result, ok := svc.selectAvailableRoutedAccount(context.Background(), nil, "route-full", "claude-3-5-sonnet-20241022", available, time.Second, 1)

	require.True(t, ok)
	require.NotNil(t, result)
	require.NotNil(t, result.WaitPlan)
	require.Equal(t, int64(1), result.Account.ID)
}
