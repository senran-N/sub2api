package service

import (
	"testing"

	"github.com/senran-N/sub2api/internal/pkg/grok"
	"github.com/stretchr/testify/require"
)

func TestGrokTierStateFallsBackToQuotaWindows(t *testing.T) {
	account := &Account{
		Platform: PlatformGrok,
		Extra: map[string]any{
			"grok": map[string]any{
				"quota_windows": map[string]any{
					grok.QuotaWindowAuto: map[string]any{
						"total": 150,
					},
				},
			},
		},
	}

	require.Equal(t, grok.TierHeavy, account.GrokTierState().Normalized)
}

func TestNormalizeGrokAccountExtraMaterializesCanonicalTierAndQuotaDefaults(t *testing.T) {
	normalized := normalizeGrokAccountExtra(nil, map[string]any{
		"grok": map[string]any{
			"pool":        "super",
			"tier_source": "manual",
		},
	}, AccountTypeSession)

	grokExtra := grokExtraMap(normalized)
	require.Equal(t, AccountTypeSession, grokExtra["auth_mode"])
	require.Equal(t, "super", getNestedGrokValue(grokExtra, "tier", "normalized"))
	require.Equal(t, "manual", getNestedGrokValue(grokExtra, "tier", "source"))

	quotaWindows := grokQuotaWindowsMap(grokExtra["quota_windows"])
	require.Equal(t, 50, grokParseInt(getNestedGrokValue(quotaWindows, grok.QuotaWindowAuto, "total")))
	require.Equal(t, 140, grokParseInt(getNestedGrokValue(quotaWindows, grok.QuotaWindowFast, "total")))
	require.Equal(t, 50, grokParseInt(getNestedGrokValue(quotaWindows, grok.QuotaWindowExpert, "total")))
	_, hasHeavy := quotaWindows[grok.QuotaWindowHeavy]
	require.False(t, hasHeavy)
	_, hasLegacyPool := grokExtra["pool"]
	require.False(t, hasLegacyPool)
}

func TestApplyMutableAccountExtraGrokDeepMergesRuntimeState(t *testing.T) {
	account := &Account{
		Platform: PlatformGrok,
		Type:     AccountTypeSession,
		Extra: map[string]any{
			"grok": map[string]any{
				"auth_mode": "session",
				"tier": map[string]any{
					"normalized": "basic",
				},
				"quota_windows": map[string]any{
					grok.QuotaWindowAuto: map[string]any{
						"remaining": 9,
						"total":     20,
					},
				},
				"sync_state": map[string]any{
					"last_sync_at": "2026-04-19T00:00:00Z",
				},
			},
		},
	}

	err := applyMutableAccountExtra(account, map[string]any{
		"grok": map[string]any{
			"tier": map[string]any{
				"normalized": "heavy",
			},
			"capabilities": map[string]any{
				"operations": []any{"chat", "video"},
			},
		},
	}, false)
	require.NoError(t, err)

	grokExtra := grokExtraMap(account.Extra)
	require.Equal(t, "heavy", getNestedGrokValue(grokExtra, "tier", "normalized"))
	require.Equal(t, "2026-04-19T00:00:00Z", getNestedGrokValue(grokExtra, "sync_state", "last_sync_at"))

	quotaWindows := grokQuotaWindowsMap(grokExtra["quota_windows"])
	require.Equal(t, 9, grokParseInt(getNestedGrokValue(quotaWindows, grok.QuotaWindowAuto, "remaining")))
	require.Equal(t, 150, grokParseInt(getNestedGrokValue(quotaWindows, grok.QuotaWindowAuto, "total")))
	require.Equal(t, 20, grokParseInt(getNestedGrokValue(quotaWindows, grok.QuotaWindowHeavy, "total")))
	require.Equal(t, []any{"chat", "video"}, getNestedGrokValue(grokExtra, "capabilities", "operations"))
}
