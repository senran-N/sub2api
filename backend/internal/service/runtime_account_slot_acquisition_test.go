package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAcquireRuntimeAccountSlotUsesPreAcquiredSelection(t *testing.T) {
	released := false
	account := &Account{ID: 10}

	result := AcquireRuntimeAccountSlot(context.Background(), RuntimeAccountSlotRequest{
		Selection: &AccountSelectionResult{
			Account:  account,
			Acquired: true,
			ReleaseFunc: func() {
				released = true
			},
		},
		AcquireOrQueue: func(context.Context, int64, int, int) (*AcquireOrQueueResult, error) {
			t.Fatal("pre-acquired selection must not acquire again")
			return nil, nil
		},
	})

	require.Equal(t, RuntimeAccountSlotSucceeded, result.Outcome)
	require.Equal(t, account, result.Account)
	require.NotNil(t, result.ReleaseFunc)
	result.ReleaseFunc()
	require.True(t, released)
}

func TestAcquireRuntimeAccountSlotReportsNoAvailable(t *testing.T) {
	result := AcquireRuntimeAccountSlot(context.Background(), RuntimeAccountSlotRequest{})
	require.Equal(t, RuntimeAccountSlotNoAvailable, result.Outcome)
	require.Nil(t, result.Account)

	result = AcquireRuntimeAccountSlot(context.Background(), RuntimeAccountSlotRequest{
		Selection: &AccountSelectionResult{Account: &Account{ID: 11}},
	})
	require.Equal(t, RuntimeAccountSlotNoAvailable, result.Outcome)
	require.Equal(t, int64(11), result.Account.ID)
}

func TestAcquireRuntimeAccountSlotQuickAcquireBindsSticky(t *testing.T) {
	var boundAccountID int64
	released := false
	groupID := int64(7)

	result := AcquireRuntimeAccountSlot(context.Background(), RuntimeAccountSlotRequest{
		GroupID:     &groupID,
		SessionHash: "session",
		Selection: &AccountSelectionResult{
			Account: &Account{ID: 12},
			WaitPlan: &AccountWaitPlan{
				MaxConcurrency: 2,
				MaxWaiting:     3,
			},
		},
		AcquireOrQueue: func(_ context.Context, accountID int64, maxConcurrency int, maxWaiting int) (*AcquireOrQueueResult, error) {
			require.Equal(t, int64(12), accountID)
			require.Equal(t, 2, maxConcurrency)
			require.Equal(t, 3, maxWaiting)
			return &AcquireOrQueueResult{
				Acquired: true,
				ReleaseFunc: func() {
					released = true
				},
			}, nil
		},
		BindSticky: func(_ context.Context, groupIDArg *int64, sessionHash string, accountID int64) error {
			require.NotNil(t, groupIDArg)
			require.Equal(t, groupID, *groupIDArg)
			require.Equal(t, "session", sessionHash)
			boundAccountID = accountID
			return nil
		},
	})

	require.Equal(t, RuntimeAccountSlotSucceeded, result.Outcome)
	require.NoError(t, result.BindErr)
	require.Equal(t, int64(12), boundAccountID)
	require.NotNil(t, result.ReleaseFunc)
	result.ReleaseFunc()
	require.True(t, released)
}

func TestAcquireRuntimeAccountSlotQueueFull(t *testing.T) {
	result := AcquireRuntimeAccountSlot(context.Background(), RuntimeAccountSlotRequest{
		Selection: &AccountSelectionResult{
			Account: &Account{ID: 13},
			WaitPlan: &AccountWaitPlan{
				MaxConcurrency: 1,
				MaxWaiting:     2,
			},
		},
		AcquireOrQueue: func(context.Context, int64, int, int) (*AcquireOrQueueResult, error) {
			return &AcquireOrQueueResult{QueueAllowed: false}, nil
		},
	})

	require.Equal(t, RuntimeAccountSlotQueueFull, result.Outcome)
	require.Equal(t, int64(13), result.Account.ID)
}

func TestAcquireRuntimeAccountSlotWaitSuccessReleasesWaitCount(t *testing.T) {
	decremented := 0
	waited := false
	bound := false

	result := AcquireRuntimeAccountSlot(context.Background(), RuntimeAccountSlotRequest{
		SessionHash: "queued",
		Selection: &AccountSelectionResult{
			Account: &Account{ID: 14},
			WaitPlan: &AccountWaitPlan{
				MaxConcurrency: 4,
				MaxWaiting:     5,
				Timeout:        6 * time.Second,
			},
		},
		AcquireOrQueue: func(context.Context, int64, int, int) (*AcquireOrQueueResult, error) {
			return &AcquireOrQueueResult{QueueAllowed: true, WaitCounted: true}, nil
		},
		WaitForSlot: func(_ context.Context, wait RuntimeAccountSlotWaitRequest) (func(), error) {
			require.Equal(t, int64(14), wait.AccountID)
			require.Equal(t, 4, wait.MaxConcurrency)
			require.Equal(t, 6*time.Second, wait.Timeout)
			waited = true
			return func() {}, nil
		},
		DecrementWait: func(context.Context, int64) {
			decremented++
		},
		BindSticky: func(context.Context, *int64, string, int64) error {
			bound = true
			return nil
		},
	})

	require.Equal(t, RuntimeAccountSlotSucceeded, result.Outcome)
	require.True(t, waited)
	require.True(t, bound)
	require.Equal(t, 1, decremented)
	require.NotNil(t, result.ReleaseFunc)
}

func TestAcquireRuntimeAccountSlotWaitErrorReleasesWaitCount(t *testing.T) {
	waitErr := errors.New("wait failed")
	decremented := 0

	result := AcquireRuntimeAccountSlot(context.Background(), RuntimeAccountSlotRequest{
		Selection: &AccountSelectionResult{
			Account:  &Account{ID: 15},
			WaitPlan: &AccountWaitPlan{MaxConcurrency: 1, Timeout: time.Second},
		},
		AcquireOrQueue: func(context.Context, int64, int, int) (*AcquireOrQueueResult, error) {
			return &AcquireOrQueueResult{QueueAllowed: true, WaitCounted: true}, nil
		},
		WaitForSlot: func(context.Context, RuntimeAccountSlotWaitRequest) (func(), error) {
			return nil, waitErr
		},
		DecrementWait: func(context.Context, int64) {
			decremented++
		},
	})

	require.Equal(t, RuntimeAccountSlotWaitAcquireError, result.Outcome)
	require.ErrorIs(t, result.Err, waitErr)
	require.Equal(t, 1, decremented)
	require.Nil(t, result.ReleaseFunc)
}
