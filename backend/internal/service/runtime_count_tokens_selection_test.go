package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSelectRuntimeCountTokensAccountSkipsRPMDeniedAccount(t *testing.T) {
	accounts := []*Account{{ID: 1}, {ID: 2}}
	var excludedSnapshots []map[int64]struct{}

	result := selectRuntimeCountTokensAccount(context.Background(), RuntimeCountTokensSelectionRequest{
		Model: "claude-sonnet",
	}, runtimeCountTokensSelectionHooks{
		selectAccount: func(_ context.Context, _ *int64, _ string, _ string, excluded map[int64]struct{}) (*Account, error) {
			snapshot := make(map[int64]struct{}, len(excluded))
			for id := range excluded {
				snapshot[id] = struct{}{}
			}
			excludedSnapshots = append(excludedSnapshots, snapshot)
			for _, account := range accounts {
				if _, ok := excluded[account.ID]; !ok {
					return account, nil
				}
			}
			return nil, errors.New("no accounts")
		},
		reserveRPM: func(_ context.Context, req RuntimeRPMAdmissionRequest) RuntimeRPMAdmissionResult {
			if req.Account.ID == 1 {
				return RuntimeRPMAdmissionResult{
					Outcome:  RuntimeAdmissionRPMDenied,
					Account:  req.Account,
					RPMCount: 10,
				}
			}
			return RuntimeRPMAdmissionResult{Outcome: RuntimeAdmissionSucceeded, Account: req.Account}
		},
	})

	require.NoError(t, result.Err)
	require.Equal(t, int64(2), result.Account.ID)
	require.Contains(t, result.FailedAccountIDs, int64(1))
	require.Len(t, result.AdmissionEvents, 1)
	require.Equal(t, int64(1), result.AdmissionEvents[0].Account.ID)
	require.Len(t, excludedSnapshots, 2)
	require.Empty(t, excludedSnapshots[0])
	require.Contains(t, excludedSnapshots[1], int64(1))
}

func TestSelectRuntimeCountTokensAccountAllowsRPMReservationError(t *testing.T) {
	reserveErr := errors.New("redis unavailable")

	result := selectRuntimeCountTokensAccount(context.Background(), RuntimeCountTokensSelectionRequest{}, runtimeCountTokensSelectionHooks{
		selectAccount: func(context.Context, *int64, string, string, map[int64]struct{}) (*Account, error) {
			return &Account{ID: 11}, nil
		},
		reserveRPM: func(_ context.Context, req RuntimeRPMAdmissionRequest) RuntimeRPMAdmissionResult {
			return RuntimeRPMAdmissionResult{
				Outcome: RuntimeAdmissionSucceeded,
				Account: req.Account,
				Err:     reserveErr,
			}
		},
	})

	require.NoError(t, result.Err)
	require.Equal(t, int64(11), result.Account.ID)
	require.Empty(t, result.FailedAccountIDs)
	require.Len(t, result.AdmissionEvents, 1)
	require.ErrorIs(t, result.AdmissionEvents[0].Admission.Err, reserveErr)
}

func TestSelectRuntimeCountTokensAccountReturnsSelectionError(t *testing.T) {
	selectionErr := errors.New("no available accounts")

	result := selectRuntimeCountTokensAccount(context.Background(), RuntimeCountTokensSelectionRequest{}, runtimeCountTokensSelectionHooks{
		selectAccount: func(context.Context, *int64, string, string, map[int64]struct{}) (*Account, error) {
			return nil, selectionErr
		},
		reserveRPM: func(context.Context, RuntimeRPMAdmissionRequest) RuntimeRPMAdmissionResult {
			t.Fatal("rpm admission should not run after selection failure")
			return RuntimeRPMAdmissionResult{}
		},
	})

	require.ErrorIs(t, result.Err, selectionErr)
	require.Nil(t, result.Account)
	require.Empty(t, result.FailedAccountIDs)
	require.Empty(t, result.AdmissionEvents)
}
