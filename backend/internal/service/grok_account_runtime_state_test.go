//go:build unit

package service

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/senran-N/sub2api/internal/pkg/grok"
	"github.com/stretchr/testify/require"
)

func TestBuildGrokSyncStateExtraUpdatesDeepMergesProviderSnapshot(t *testing.T) {
	account := &Account{
		Platform: PlatformGrok,
		Type:     AccountTypeSession,
		Extra: map[string]any{
			"grok": map[string]any{
				"auth_mode": "session",
				"tier": map[string]any{
					"normalized": "basic",
					"source":     "manual",
				},
				"quota_windows": map[string]any{
					grok.QuotaWindowAuto: map[string]any{
						"remaining": 9,
						"total":     20,
					},
				},
				"sync_state": map[string]any{
					"last_probe_ok_at": "2026-04-18T00:00:00Z",
				},
			},
		},
	}

	updates := buildGrokSyncStateExtraUpdates(account, grokStateSyncSnapshot{
		Tier: map[string]any{
			"normalized": "heavy",
			"source":     "sync",
		},
		QuotaWindows: map[string]any{
			grok.QuotaWindowAuto: map[string]any{
				"remaining": 17,
			},
			grok.QuotaWindowHeavy: map[string]any{
				"remaining": 3,
			},
		},
		Capabilities: map[string]any{
			"operations": []any{"chat", "video"},
		},
		SyncState: map[string]any{
			"last_sync_at": "2026-04-20T00:00:00Z",
		},
	})

	grokExtra := grokExtraMap(updates)
	require.Equal(t, "session", grokExtra["auth_mode"])
	require.Equal(t, "heavy", getNestedGrokValue(grokExtra, "tier", "normalized"))
	require.Equal(t, "sync", getNestedGrokValue(grokExtra, "tier", "source"))
	require.Equal(t, "2026-04-20T00:00:00Z", getNestedGrokValue(grokExtra, "sync_state", "last_sync_at"))
	require.Equal(t, "2026-04-18T00:00:00Z", getNestedGrokValue(grokExtra, "sync_state", "last_probe_ok_at"))

	quotaWindows := grokQuotaWindowsMap(grokExtra["quota_windows"])
	require.Equal(t, 17, grokParseInt(getNestedGrokValue(quotaWindows, grok.QuotaWindowAuto, "remaining")))
	require.Equal(t, 150, grokParseInt(getNestedGrokValue(quotaWindows, grok.QuotaWindowAuto, "total")))
	require.Equal(t, 3, grokParseInt(getNestedGrokValue(quotaWindows, grok.QuotaWindowHeavy, "remaining")))
	require.Equal(t, 20, grokParseInt(getNestedGrokValue(quotaWindows, grok.QuotaWindowHeavy, "total")))
	require.Equal(t, []any{"chat", "video"}, getNestedGrokValue(grokExtra, "capabilities", "operations"))
}

func TestBuildGrokProbeStateExtraUpdatesPreservesExistingSignalsOnFailure(t *testing.T) {
	now := time.Date(2026, 4, 19, 8, 0, 0, 0, time.UTC)
	account := &Account{
		Platform: PlatformGrok,
		Type:     AccountTypeAPIKey,
		Extra: map[string]any{
			"grok": map[string]any{
				"tier": map[string]any{
					"normalized": "super",
					"source":     "sync",
				},
				"quota_windows": map[string]any{
					grok.QuotaWindowAuto: map[string]any{
						"remaining": 7,
						"total":     50,
					},
				},
				"capabilities": map[string]any{
					"models": []any{"grok-imagine-video"},
				},
				"sync_state": map[string]any{
					"last_sync_at":     "2026-04-19T00:00:00Z",
					"last_probe_ok_at": "2026-04-18T00:00:00Z",
				},
			},
		},
	}

	updates := buildGrokProbeStateExtraUpdates(
		account,
		"grok-3-fast",
		&http.Response{StatusCode: http.StatusUnauthorized, Status: "401 Unauthorized"},
		errors.New("API returned 401: unauthorized"),
		now,
	)

	grokExtra := grokExtraMap(updates)
	require.Equal(t, "super", getNestedGrokValue(grokExtra, "tier", "normalized"))
	require.Equal(t, "sync", getNestedGrokValue(grokExtra, "tier", "source"))
	require.Equal(t, "2026-04-19T00:00:00Z", getNestedGrokValue(grokExtra, "sync_state", "last_sync_at"))
	require.Equal(t, "2026-04-18T00:00:00Z", getNestedGrokValue(grokExtra, "sync_state", "last_probe_ok_at"))
	require.Equal(t, "2026-04-19T08:00:00Z", getNestedGrokValue(grokExtra, "sync_state", "last_probe_at"))
	require.Equal(t, "2026-04-19T08:00:00Z", getNestedGrokValue(grokExtra, "sync_state", "last_probe_error_at"))
	require.Equal(t, 401, grokParseInt(getNestedGrokValue(grokExtra, "sync_state", "last_probe_status_code")))
	require.Contains(t, getStringFromMaps(grokNestedMap(grokExtra["sync_state"]), nil, "last_probe_error"), "API returned 401")

	quotaWindows := grokQuotaWindowsMap(grokExtra["quota_windows"])
	require.Equal(t, 7, grokParseInt(getNestedGrokValue(quotaWindows, grok.QuotaWindowAuto, "remaining")))
	require.Equal(t, 50, grokParseInt(getNestedGrokValue(quotaWindows, grok.QuotaWindowAuto, "total")))
	require.ElementsMatch(t, []string{"grok-imagine-video"}, grokParseStringSlice(getNestedGrokValue(grokExtra, "capabilities", "models")))
}

func TestBuildGrokRuntimeCapabilityExtraUpdatesPersistsObservedSuccessIntoPatchLayer(t *testing.T) {
	account := &Account{
		Platform: PlatformGrok,
		Type:     AccountTypeSession,
		Extra: map[string]any{
			"grok": map[string]any{
				"tier": map[string]any{
					"normalized": "basic",
					"source":     "sync",
				},
				"sync_state": map[string]any{
					"last_probe_ok_at": "2026-04-18T00:00:00Z",
				},
			},
		},
	}

	updates := buildGrokRuntimeCapabilityExtraUpdates(
		account,
		GrokRuntimeFeedbackInput{
			Account:        account,
			RequestedModel: "grok-3-fast",
			Result: &OpenAIForwardResult{
				Model:         "grok-3-fast",
				UpstreamModel: "grok-4.20-fast",
			},
			StatusCode:     http.StatusOK,
			ProtocolFamily: grok.ProtocolFamilyChatCompletions,
		},
		"grok-3-fast",
		grok.CapabilityChat,
	)

	grokExtra := updates
	require.Nil(t, getNestedGrokValue(grokExtra, "sync_state", "last_probe_ok_at"))
	require.Nil(t, getNestedGrokValue(grokExtra, "tier", "normalized"))
	require.Nil(t, getNestedGrokValue(grokExtra, "runtime_state", "last_outcome"))
	require.ElementsMatch(t, []string{"grok-4.20-fast"}, grokParseStringSlice(getNestedGrokValue(grokExtra, "capabilities", "models")))
	require.ElementsMatch(t, []string{"chat"}, grokParseStringSlice(getNestedGrokValue(grokExtra, "capabilities", "operations")))
}

func TestBuildGrokRuntimeCapabilityExtraUpdatesPrunesUnsupportedModelsWithoutDroppingExactSignal(t *testing.T) {
	account := &Account{
		Platform: PlatformGrok,
		Type:     AccountTypeAPIKey,
		Extra: map[string]any{
			"grok": map[string]any{
				"capabilities": map[string]any{
					"models":     []any{"grok-3-fast"},
					"operations": []any{"chat"},
				},
			},
		},
	}

	updates := buildGrokRuntimeCapabilityExtraUpdates(
		account,
		GrokRuntimeFeedbackInput{
			Account:        account,
			RequestedModel: "grok-3-fast",
			StatusCode:     http.StatusBadRequest,
			ProtocolFamily: grok.ProtocolFamilyChatCompletions,
			Err: &UpstreamFailoverError{
				StatusCode:   http.StatusBadRequest,
				ResponseBody: []byte(`{"error":{"code":"model_not_found","message":"model not found: grok-3-fast"}}`),
			},
		},
		"grok-3-fast",
		grok.CapabilityChat,
	)

	grokExtra := updates
	capabilities := grokNestedMap(grokExtra["capabilities"])
	modelsRaw, hasModels := capabilities["models"]

	require.True(t, hasModels)
	require.Empty(t, grokParseStringSlice(modelsRaw))
	require.Nil(t, getNestedGrokValue(grokExtra, "runtime_state", "last_outcome"))
	require.Nil(t, getNestedGrokValue(grokExtra, "sync_state", "last_runtime_error_at"))
}

func TestBuildGrokRuntimeCapabilityExtraUpdatesPrunesUnsupportedCapabilityFamily(t *testing.T) {
	account := &Account{
		Platform: PlatformGrok,
		Type:     AccountTypeAPIKey,
		Extra: map[string]any{
			"grok": map[string]any{
				"capabilities": map[string]any{
					"models":     []any{"grok-3", "grok-imagine-image", "grok-imagine-image-pro"},
					"operations": []any{"chat", "image"},
				},
			},
		},
	}

	updates := buildGrokRuntimeCapabilityExtraUpdates(
		account,
		GrokRuntimeFeedbackInput{
			Account:        account,
			RequestedModel: "grok-imagine-image",
			StatusCode:     http.StatusForbidden,
			ProtocolFamily: grok.ProtocolFamilyResponses,
			Err: &UpstreamFailoverError{
				StatusCode:   http.StatusForbidden,
				ResponseBody: []byte(`{"error":{"code":"tier_required","message":"image tier required"}}`),
			},
		},
		"grok-imagine-image",
		grok.CapabilityImage,
	)

	grokExtra := updates
	capabilities := grokNestedMap(grokExtra["capabilities"])
	require.ElementsMatch(t, []string{"grok-4.20-auto"}, grokParseStringSlice(capabilities["models"]))
	require.ElementsMatch(t, []string{"chat"}, grokParseStringSlice(capabilities["operations"]))
	require.Equal(t, false, capabilities["image"])
}

func TestBuildGrokRuntimeCapabilityExtraUpdatesPrunesHigherTierChatVariantsOnly(t *testing.T) {
	account := &Account{
		Platform: PlatformGrok,
		Type:     AccountTypeAPIKey,
		Extra: map[string]any{
			"grok": map[string]any{
				"capabilities": map[string]any{
					"models":     []any{"grok-3", "grok-4.20-0309-super", "grok-4.20-heavy"},
					"operations": []any{"chat"},
				},
			},
		},
	}

	updates := buildGrokRuntimeCapabilityExtraUpdates(
		account,
		GrokRuntimeFeedbackInput{
			Account:        account,
			RequestedModel: "grok-4.20-0309-super",
			StatusCode:     http.StatusForbidden,
			ProtocolFamily: grok.ProtocolFamilyResponses,
			Err: &UpstreamFailoverError{
				StatusCode:   http.StatusForbidden,
				ResponseBody: []byte(`{"error":{"code":"tier_required","message":"requires super"}}`),
			},
		},
		"grok-4.20-0309-super",
		grok.CapabilityChat,
	)

	grokExtra := updates
	capabilities := grokNestedMap(grokExtra["capabilities"])
	require.ElementsMatch(t, []string{"grok-4.20-auto"}, grokParseStringSlice(capabilities["models"]))
	require.ElementsMatch(t, []string{"chat"}, grokParseStringSlice(capabilities["operations"]))
	_, hasChatNegative := capabilities["chat"]
	require.False(t, hasChatNegative)
}

func TestGrokCapabilitiesTreatsEmptyModelsAsExactSignal(t *testing.T) {
	account := &Account{
		Platform: PlatformGrok,
		Extra: map[string]any{
			"grok": map[string]any{
				"capabilities": map[string]any{
					"models":     []any{},
					"operations": []any{"chat"},
				},
			},
		},
	}

	state := account.grokCapabilities()
	require.True(t, state.hasModelSignal)
	require.Empty(t, state.models)
	require.True(t, state.operations[grok.CapabilityChat])
}
