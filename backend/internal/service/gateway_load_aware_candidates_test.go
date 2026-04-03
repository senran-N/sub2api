//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFilterLoadAwareCandidates_AppliesEligibilityRules(t *testing.T) {
	svc := &GatewayService{
		sessionLimitCache: &sessionLimitRuntimeStub{
			windowHit:  true,
			windowCost: 20,
		},
		cfg: testConfig(),
	}

	accounts := []Account{
		{
			ID:          1,
			Platform:    PlatformAnthropic,
			Type:        AccountTypeOAuth,
			Status:      StatusActive,
			Schedulable: true,
		},
		{
			ID:          2,
			Platform:    PlatformGemini,
			Type:        AccountTypeOAuth,
			Status:      StatusActive,
			Schedulable: true,
		},
		{
			ID:          3,
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
		{
			ID:          4,
			Platform:    PlatformAnthropic,
			Type:        AccountTypeOAuth,
			Status:      StatusActive,
			Schedulable: true,
			Extra: map[string]any{
				"window_cost_limit": 10.0,
			},
		},
		{
			ID:          5,
			Platform:    PlatformAnthropic,
			Type:        AccountTypeOAuth,
			Status:      StatusActive,
			Schedulable: true,
		},
	}

	candidates := svc.filterLoadAwareCandidates(
		context.Background(),
		accounts,
		"claude-test",
		PlatformAnthropic,
		false,
		map[int64]struct{}{1: {}},
	)

	require.Len(t, candidates, 1)
	require.Equal(t, int64(5), candidates[0].ID)
}
