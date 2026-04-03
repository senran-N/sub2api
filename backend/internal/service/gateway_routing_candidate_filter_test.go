//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFilterRoutedCandidates_CollectsFilteringStats(t *testing.T) {
	resetAt := time.Now().Add(time.Minute).UTC().Format(time.RFC3339)
	svc := &GatewayService{
		sessionLimitCache: &sessionLimitRuntimeStub{
			windowHit:  true,
			windowCost: 20,
		},
		cfg: testConfig(),
	}

	accountByID := map[int64]*Account{
		1: {
			ID:          1,
			Platform:    PlatformAnthropic,
			Type:        AccountTypeOAuth,
			Status:      StatusActive,
			Schedulable: true,
		},
		3: {
			ID:          3,
			Platform:    PlatformAnthropic,
			Type:        AccountTypeOAuth,
			Status:      StatusDisabled,
			Schedulable: true,
		},
		4: {
			ID:          4,
			Platform:    PlatformGemini,
			Type:        AccountTypeOAuth,
			Status:      StatusActive,
			Schedulable: true,
		},
		5: {
			ID:          5,
			Platform:    PlatformAnthropic,
			Type:        AccountTypeOAuth,
			Status:      StatusActive,
			Schedulable: true,
			Credentials: map[string]any{
				"model_mapping": map[string]any{
					"other-model": "other-model",
				},
			},
		},
		6: {
			ID:          6,
			Platform:    PlatformAnthropic,
			Type:        AccountTypeOAuth,
			Status:      StatusActive,
			Schedulable: true,
			Extra: map[string]any{
				modelRateLimitsKey: map[string]any{
					"claude-test": map[string]any{
						"rate_limit_reset_at": resetAt,
					},
				},
			},
		},
		7: {
			ID:          7,
			Platform:    PlatformAnthropic,
			Type:        AccountTypeOAuth,
			Status:      StatusActive,
			Schedulable: true,
			Extra: map[string]any{
				"window_cost_limit": 10.0,
			},
		},
		8: {
			ID:          8,
			Platform:    PlatformAnthropic,
			Type:        AccountTypeOAuth,
			Status:      StatusActive,
			Schedulable: true,
		},
	}

	result := svc.filterRoutedCandidates(
		context.Background(),
		accountByID,
		[]int64{1, 2, 3, 4, 5, 6, 7, 8},
		"claude-test",
		PlatformAnthropic,
		false,
		map[int64]struct{}{1: {}},
	)

	require.Len(t, result.Candidates, 1)
	require.Equal(t, int64(8), result.Candidates[0].ID)
	require.Equal(t, 1, result.Stats.FilteredExcluded)
	require.Equal(t, 1, result.Stats.FilteredMissing)
	require.Equal(t, 1, result.Stats.FilteredUnsched)
	require.Equal(t, 1, result.Stats.FilteredPlatform)
	require.Equal(t, 1, result.Stats.FilteredModelMap)
	require.Equal(t, 1, result.Stats.FilteredModelScope)
	require.Equal(t, []int64{6}, result.Stats.ModelScopeSkippedID)
	require.Equal(t, 1, result.Stats.FilteredWindowCost)
}
