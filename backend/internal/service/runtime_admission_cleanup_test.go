package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCleanupRuntimeAdmissionDeniedReleasesAndMarksFailed(t *testing.T) {
	calls := []string{}
	failed := map[int64]struct{}{}

	result := CleanupRuntimeAdmissionDenied(RuntimeAdmissionCleanupRequest{
		Account:          &Account{ID: 42},
		FailedAccountIDs: failed,
		QueueRelease: func() {
			calls = append(calls, "queue")
		},
		ClearUpstreamAccepted: func() {
			calls = append(calls, "clear")
		},
		AccountRelease: func() {
			calls = append(calls, "account")
		},
	})

	require.Equal(t, []string{"queue", "clear", "account"}, calls)
	require.Equal(t, int64(42), result.AccountID)
	require.True(t, result.MarkedFailed)
	require.Contains(t, failed, int64(42))
}

func TestCleanupRuntimeAdmissionDeniedNilSafe(t *testing.T) {
	result := CleanupRuntimeAdmissionDenied(RuntimeAdmissionCleanupRequest{})

	require.Zero(t, result.AccountID)
	require.False(t, result.MarkedFailed)
}

func TestCleanupRuntimeAdmissionDeniedWithoutFailedMapStillReleases(t *testing.T) {
	released := false

	result := CleanupRuntimeAdmissionDenied(RuntimeAdmissionCleanupRequest{
		Account: &Account{ID: 43},
		AccountRelease: func() {
			released = true
		},
	})

	require.True(t, released)
	require.Zero(t, result.AccountID)
	require.False(t, result.MarkedFailed)
}
