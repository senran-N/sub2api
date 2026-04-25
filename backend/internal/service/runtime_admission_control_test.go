package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReserveRuntimeRPMAdmissionDenied(t *testing.T) {
	rpm := &rpmCacheRuntimeLimitsStub{
		reserveOK:    false,
		reserveCount: 9,
	}
	svc := &GatewayService{rpmCache: rpm}
	account := &Account{
		ID:          21,
		Platform:    PlatformAnthropic,
		Type:        AccountTypeOAuth,
		Concurrency: 2,
		Extra: map[string]any{
			"base_rpm":     5,
			"max_sessions": 3,
		},
	}

	result := svc.ReserveRuntimeRPMAdmission(context.Background(), RuntimeRPMAdmissionRequest{
		Account:              account,
		StickyBoundAccountID: account.ID,
	})

	require.Equal(t, RuntimeAdmissionRPMDenied, result.Outcome)
	require.Equal(t, account, result.Account)
	require.True(t, result.StickyBound)
	require.Equal(t, 9, result.RPMCount)
	require.Equal(t, int64(21), rpm.reserveCall.accountID)
	require.Equal(t, 10, rpm.reserveCall.limit)
	require.NoError(t, result.Err)
	require.NoError(t, result.ClearStickyErr)
}

func TestReserveRuntimeRPMAdmissionErrorAllowsRequest(t *testing.T) {
	reserveErr := errors.New("redis unavailable")
	rpm := &rpmCacheRuntimeLimitsStub{reserveErr: reserveErr}
	svc := &GatewayService{rpmCache: rpm}
	account := &Account{
		ID:       22,
		Platform: PlatformAnthropic,
		Type:     AccountTypeOAuth,
		Extra: map[string]any{
			"base_rpm": 5,
		},
	}

	result := svc.ReserveRuntimeRPMAdmission(context.Background(), RuntimeRPMAdmissionRequest{
		Account: account,
	})

	require.Equal(t, RuntimeAdmissionSucceeded, result.Outcome)
	require.ErrorIs(t, result.Err, reserveErr)
	require.False(t, result.StickyBound)
}

func TestReserveRuntimeWindowCostAdmissionDisabled(t *testing.T) {
	svc := &GatewayService{}

	result := svc.ReserveRuntimeWindowCostAdmission(context.Background(), RuntimeWindowCostAdmissionRequest{
		Account: &Account{ID: 23},
	})

	require.Equal(t, RuntimeAdmissionSucceeded, result.Outcome)
	require.Nil(t, result.Reservation)
	require.Zero(t, result.Total)
	require.NoError(t, result.Err)
}
