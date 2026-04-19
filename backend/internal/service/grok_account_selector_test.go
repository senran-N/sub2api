package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGrokAccountSelectorRequestedModelAvailable(t *testing.T) {
	selector := GrokAccountSelector{}
	accounts := []Account{
		{ID: 1, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true},
		{
			ID:          2,
			Platform:    PlatformGrok,
			Type:        AccountTypeAPIKey,
			Status:      StatusActive,
			Schedulable: true,
			Extra: map[string]any{
				"grok": map[string]any{
					"capabilities": map[string]any{
						"operations": []any{"chat"},
					},
					"tier": map[string]any{
						"normalized": "heavy",
					},
				},
			},
		},
	}

	require.True(t, selector.RequestedModelAvailable(accounts, "grok-4-fast-reasoning"))
	require.False(t, selector.RequestedModelAvailable(accounts, "grok-imagine-image"))
}

func TestGrokAccountSelectorIsRuntimeEligible_ExcludesSessionFromSharedRuntime(t *testing.T) {
	selector := GrokAccountSelector{}
	account := &Account{
		ID:          9,
		Platform:    PlatformGrok,
		Type:        AccountTypeSession,
		Status:      StatusActive,
		Schedulable: true,
		Extra: map[string]any{
			"grok": map[string]any{
				"tier": map[string]any{
					"normalized": "heavy",
				},
			},
		},
	}

	require.False(t, selector.IsRuntimeEligible(account, "grok-4-fast-reasoning"))
	require.True(t, selector.IsRuntimeEligibleWithContext(WithGrokSessionTextRuntimeAllowed(context.Background()), account, "grok-4-fast-reasoning"))
}

func TestGrokAccountSelectorFilterSchedulableCandidates(t *testing.T) {
	selector := GrokAccountSelector{}
	accounts := []Account{
		{
			ID:          10,
			Platform:    PlatformGrok,
			Type:        AccountTypeSession,
			Status:      StatusActive,
			Schedulable: true,
			Extra: map[string]any{
				"grok": map[string]any{
					"tier": map[string]any{
						"normalized": "heavy",
					},
				},
			},
		},
		{
			ID:          11,
			Platform:    PlatformGrok,
			Type:        AccountTypeAPIKey,
			Status:      StatusActive,
			Schedulable: true,
			Extra: map[string]any{
				"grok": map[string]any{
					"tier": map[string]any{
						"normalized": "heavy",
					},
				},
			},
		},
		{ID: 12, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true},
	}

	candidates := selector.FilterSchedulableCandidates(accounts, "grok-4-fast-reasoning", nil)
	require.Len(t, candidates, 1)
	require.Equal(t, int64(11), candidates[0].ID)

	textCandidates := selector.FilterSchedulableCandidatesWithContext(
		WithGrokSessionTextRuntimeAllowed(context.Background()),
		accounts,
		"grok-4-fast-reasoning",
		nil,
	)
	require.Len(t, textCandidates, 2)
	require.Equal(t, int64(10), textCandidates[0].ID)
	require.Equal(t, int64(11), textCandidates[1].ID)
}

func TestGrokAccountSelectorSelectBestCandidate_PrefersExactTierAndFreshQuota(t *testing.T) {
	selector := GrokAccountSelector{}
	now := time.Now().UTC()
	accounts := []*Account{
		{
			ID:          21,
			Platform:    PlatformGrok,
			Type:        AccountTypeAPIKey,
			Status:      StatusActive,
			Schedulable: true,
			Priority:    1,
			Extra: map[string]any{
				"grok": map[string]any{
					"tier": map[string]any{
						"normalized": "heavy",
					},
					"sync_state": map[string]any{
						"last_sync_at": now.Add(-8 * time.Hour).Format(time.RFC3339),
					},
					"quota_windows": map[string]any{
						"auto": map[string]any{
							"remaining": 120,
							"total":     150,
						},
					},
				},
			},
		},
		{
			ID:          22,
			Platform:    PlatformGrok,
			Type:        AccountTypeAPIKey,
			Status:      StatusActive,
			Schedulable: true,
			Priority:    1,
			Extra: map[string]any{
				"grok": map[string]any{
					"tier": map[string]any{
						"normalized": "basic",
					},
					"capabilities": map[string]any{
						"models": []any{"grok-3"},
					},
					"sync_state": map[string]any{
						"last_sync_at": now.Add(-10 * time.Minute).Format(time.RFC3339),
					},
					"quota_windows": map[string]any{
						"auto": map[string]any{
							"remaining": 18,
							"total":     20,
						},
					},
				},
			},
		},
	}

	selected := selector.SelectBestCandidateWithContext(context.Background(), accounts, "grok-3", nil)
	require.NotNil(t, selected)
	require.Equal(t, int64(22), selected.ID)
}

func TestGrokAccountSelectorSelectBestCandidate_PenalizesRecentFailures(t *testing.T) {
	selector := GrokAccountSelector{}
	now := time.Now().UTC()
	accounts := []*Account{
		{
			ID:          31,
			Platform:    PlatformGrok,
			Type:        AccountTypeAPIKey,
			Status:      StatusActive,
			Schedulable: true,
			Priority:    1,
			Extra: map[string]any{
				"grok": map[string]any{
					"tier": map[string]any{
						"normalized": "heavy",
					},
					"sync_state": map[string]any{
						"last_sync_at": now.Add(-15 * time.Minute).Format(time.RFC3339),
					},
					"quota_windows": map[string]any{
						"expert": map[string]any{
							"remaining": 150,
							"total":     150,
						},
					},
					"runtime_state": map[string]any{
						"last_fail_at":             now.Add(-3 * time.Minute).Format(time.RFC3339),
						"last_fail_status_code":    403,
						"last_fail_class":          "auth",
						"selection_cooldown_until": now.Add(-2 * time.Minute).Format(time.RFC3339),
						"selection_cooldown_scope": "account",
					},
				},
			},
		},
		{
			ID:          32,
			Platform:    PlatformGrok,
			Type:        AccountTypeAPIKey,
			Status:      StatusActive,
			Schedulable: true,
			Priority:    1,
			Extra: map[string]any{
				"grok": map[string]any{
					"tier": map[string]any{
						"normalized": "heavy",
					},
					"sync_state": map[string]any{
						"last_sync_at": now.Add(-15 * time.Minute).Format(time.RFC3339),
					},
					"quota_windows": map[string]any{
						"expert": map[string]any{
							"remaining": 60,
							"total":     150,
						},
					},
				},
			},
		},
	}

	selected := selector.SelectBestCandidateWithContext(context.Background(), accounts, "grok-4-fast-reasoning", nil)
	require.NotNil(t, selected)
	require.Equal(t, int64(32), selected.ID)
}

func TestGrokAccountSelectorIsRuntimeEligible_RespectsProviderCooldownScope(t *testing.T) {
	selector := GrokAccountSelector{}
	now := time.Now().UTC()

	modelBlocked := &Account{
		ID:          51,
		Platform:    PlatformGrok,
		Type:        AccountTypeAPIKey,
		Status:      StatusActive,
		Schedulable: true,
		Extra: map[string]any{
			"grok": map[string]any{
				"tier": map[string]any{
					"normalized": "heavy",
				},
				"runtime_state": map[string]any{
					"selection_cooldown_until": now.Add(10 * time.Minute).Format(time.RFC3339),
					"selection_cooldown_scope": "model",
					"selection_cooldown_model": "grok-imagine-video",
				},
			},
		},
	}
	require.False(t, selector.IsRuntimeEligible(modelBlocked, "grok-imagine-video"))
	require.True(t, selector.IsRuntimeEligible(modelBlocked, "grok-4-fast-reasoning"))

	accountBlocked := &Account{
		ID:          52,
		Platform:    PlatformGrok,
		Type:        AccountTypeAPIKey,
		Status:      StatusActive,
		Schedulable: true,
		Extra: map[string]any{
			"grok": map[string]any{
				"tier": map[string]any{
					"normalized": "basic",
				},
				"runtime_state": map[string]any{
					"selection_cooldown_until": now.Add(10 * time.Minute).Format(time.RFC3339),
					"selection_cooldown_scope": "account",
				},
			},
		},
	}
	require.False(t, selector.IsRuntimeEligible(accountBlocked, "grok-3-fast"))
}

func TestGrokAccountSelectorSelectBestCandidate_UsesLoadSignals(t *testing.T) {
	selector := GrokAccountSelector{}
	now := time.Now().UTC()
	accounts := []*Account{
		{
			ID:          41,
			Platform:    PlatformGrok,
			Type:        AccountTypeAPIKey,
			Status:      StatusActive,
			Schedulable: true,
			Priority:    1,
			Extra: map[string]any{
				"grok": map[string]any{
					"tier": map[string]any{
						"normalized": "heavy",
					},
					"sync_state": map[string]any{
						"last_sync_at": now.Add(-10 * time.Minute).Format(time.RFC3339),
					},
					"quota_windows": map[string]any{
						"expert": map[string]any{
							"remaining": 120,
							"total":     150,
						},
					},
				},
			},
		},
		{
			ID:          42,
			Platform:    PlatformGrok,
			Type:        AccountTypeAPIKey,
			Status:      StatusActive,
			Schedulable: true,
			Priority:    1,
			Extra: map[string]any{
				"grok": map[string]any{
					"tier": map[string]any{
						"normalized": "heavy",
					},
					"sync_state": map[string]any{
						"last_sync_at": now.Add(-10 * time.Minute).Format(time.RFC3339),
					},
					"quota_windows": map[string]any{
						"expert": map[string]any{
							"remaining": 120,
							"total":     150,
						},
					},
				},
			},
		},
	}

	selected := selector.SelectBestCandidateWithContext(context.Background(), accounts, "grok-4-fast-reasoning", map[int64]*AccountLoadInfo{
		41: {AccountID: 41, LoadRate: 90, WaitingCount: 2},
		42: {AccountID: 42, LoadRate: 10, WaitingCount: 0},
	})
	require.NotNil(t, selected)
	require.Equal(t, int64(42), selected.ID)
}
