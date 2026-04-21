//go:build unit

package service

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/senran-N/sub2api/internal/config"
	"github.com/senran-N/sub2api/internal/pkg/grok"
	"github.com/stretchr/testify/require"
)

type grokCapabilityProbeRepoStub struct {
	accounts     []Account
	updatedIDs   []int64
	updatedExtra []map[string]any
}

func (r *grokCapabilityProbeRepoStub) ListByPlatform(_ context.Context, platform string) ([]Account, error) {
	if platform != PlatformGrok {
		return nil, nil
	}
	result := make([]Account, len(r.accounts))
	copy(result, r.accounts)
	return result, nil
}

func (r *grokCapabilityProbeRepoStub) UpdateExtra(_ context.Context, id int64, updates map[string]any) error {
	r.updatedIDs = append(r.updatedIDs, id)
	r.updatedExtra = append(r.updatedExtra, cloneAnyMap(updates))
	return nil
}

func TestGrokCapabilityProbeServiceProbeNowTargetsUnknownTierAccountsAcrossTransports(t *testing.T) {
	repo := &grokCapabilityProbeRepoStub{
		accounts: []Account{
			{
				ID:          301,
				Platform:    PlatformGrok,
				Type:        AccountTypeAPIKey,
				Status:      StatusActive,
				Schedulable: true,
				Concurrency: 1,
				Credentials: map[string]any{
					"api_key": "xai-test-key",
				},
			},
			{
				ID:          302,
				Platform:    PlatformGrok,
				Type:        AccountTypeAPIKey,
				Status:      StatusActive,
				Schedulable: true,
				Extra: map[string]any{
					"grok": map[string]any{
						"tier": map[string]any{
							"normalized": "basic",
						},
					},
				},
			},
			{
				ID:          303,
				Platform:    PlatformGrok,
				Type:        AccountTypeSession,
				Status:      StatusActive,
				Schedulable: true,
				Credentials: map[string]any{
					"session_token": "session-cookie-303",
				},
			},
		},
	}
	stateSvc := NewGrokAccountStateService(repo)
	upstream := &queuedHTTPUpstream{
		responses: []*http.Response{
			newJSONResponse(http.StatusOK, `{"id":"ok"}`),
			newJSONResponse(http.StatusOK, `{"conversationId":"conv_303"}`),
		},
	}
	svc := NewGrokCapabilityProbeService(
		repo,
		stateSvc,
		upstream,
		&config.Config{
			Security: config.SecurityConfig{
				URLAllowlist: config.URLAllowlistConfig{Enabled: false},
			},
		},
		nil,
		nil,
	)
	probeNow := time.Date(2026, 4, 19, 10, 0, 0, 0, time.UTC)
	svc.now = func() time.Time { return probeNow }
	stateSvc.now = svc.now

	err := svc.ProbeNow(context.Background())
	require.NoError(t, err)
	require.Len(t, upstream.requests, 2)
	urls := []string{upstream.requests[0].URL.String(), upstream.requests[1].URL.String()}
	require.ElementsMatch(t, []string{
		"https://api.x.ai/v1/responses",
		"https://grok.com/rest/app-chat/conversations/new",
	}, urls)
	require.ElementsMatch(t, []int64{301, 303}, repo.updatedIDs)

	var apiRequest *http.Request
	var sessionRequest *http.Request
	for _, req := range upstream.requests {
		switch req.URL.String() {
		case "https://api.x.ai/v1/responses":
			apiRequest = req
		case "https://grok.com/rest/app-chat/conversations/new":
			sessionRequest = req
		}
	}
	require.NotNil(t, apiRequest)
	require.NotNil(t, sessionRequest)
	require.Equal(t, grokCapabilityTierBootstrapModelID, requestBodyModel(t, apiRequest))
	require.Equal(t, requireGrokSessionCookieHeader(t, "session-cookie-303"), sessionRequest.Header.Get("Cookie"))
	require.Equal(t, grokSessionProbeUserAgent, sessionRequest.Header.Get("User-Agent"))
	require.Equal(t, grokSessionModeExpert, requestBodyModeID(t, sessionRequest))

	extraByID := map[int64]map[string]any{}
	for i, id := range repo.updatedIDs {
		extraByID[id] = grokExtraMap(repo.updatedExtra[i])
	}

	grokExtra := extraByID[301]
	require.Equal(t, "2026-04-19T10:00:00Z", getNestedGrokValue(grokExtra, "sync_state", "last_probe_at"))
	require.Equal(t, "2026-04-19T10:00:00Z", getNestedGrokValue(grokExtra, "sync_state", "last_probe_ok_at"))
	expectedAPIKeyCapabilities := buildGrokProbeCapabilities(&repo.accounts[0], grokCapabilityTierBootstrapModelID)
	require.ElementsMatch(t, grokParseStringSlice(expectedAPIKeyCapabilities["operations"]), grokParseStringSlice(getNestedGrokValue(grokExtra, "capabilities", "operations")))
	require.ElementsMatch(t, grokParseStringSlice(expectedAPIKeyCapabilities["models"]), grokParseStringSlice(getNestedGrokValue(grokExtra, "capabilities", "models")))

	sessionExtra := extraByID[303]
	require.Equal(t, "2026-04-19T10:00:00Z", getNestedGrokValue(sessionExtra, "sync_state", "last_probe_at"))
	require.Equal(t, "2026-04-19T10:00:00Z", getNestedGrokValue(sessionExtra, "sync_state", "last_probe_ok_at"))
	expectedSessionCapabilities := buildGrokProbeCapabilities(&repo.accounts[2], grokCapabilityTierBootstrapModelID)
	require.ElementsMatch(t, grokParseStringSlice(expectedSessionCapabilities["operations"]), grokParseStringSlice(getNestedGrokValue(sessionExtra, "capabilities", "operations")))
	require.ElementsMatch(t, grokParseStringSlice(expectedSessionCapabilities["models"]), grokParseStringSlice(getNestedGrokValue(sessionExtra, "capabilities", "models")))
}

func TestGrokCapabilityProbeServiceProbeNowFallsBackToDefaultProbeWhenBootstrapModelFails(t *testing.T) {
	repo := &grokCapabilityProbeRepoStub{
		accounts: []Account{
			{
				ID:          303,
				Platform:    PlatformGrok,
				Type:        AccountTypeSession,
				Status:      StatusActive,
				Schedulable: true,
				Credentials: map[string]any{
					"session_token": "session-cookie-303",
				},
			},
		},
	}
	stateSvc := NewGrokAccountStateService(repo)
	upstream := &queuedHTTPUpstream{
		responses: []*http.Response{
			newJSONResponse(http.StatusForbidden, `{"error":"tier required"}`),
			newJSONResponse(http.StatusOK, `{"conversationId":"conv_303"}`),
		},
	}
	svc := NewGrokCapabilityProbeService(
		repo,
		stateSvc,
		upstream,
		&config.Config{
			Security: config.SecurityConfig{
				URLAllowlist: config.URLAllowlistConfig{Enabled: false},
			},
		},
		nil,
		nil,
	)

	err := svc.ProbeNow(context.Background())
	require.NoError(t, err)
	require.Len(t, upstream.requests, 2)
	require.Equal(t, grokSessionModeExpert, requestBodyModeID(t, upstream.requests[0]))
	require.Equal(t, grokSessionModeAuto, requestBodyModeID(t, upstream.requests[1]))
	require.Len(t, repo.updatedExtra, 1)
	require.ElementsMatch(t, []string{"chat"}, grokParseStringSlice(getNestedGrokValue(grokExtraMap(repo.updatedExtra[0]), "capabilities", "operations")))
}

func TestBuildGrokCapabilitySyncSnapshotWidensSimpleChatProbeWhenTierKnown(t *testing.T) {
	account := &Account{
		Platform: PlatformGrok,
		Type:     AccountTypeAPIKey,
		Extra: map[string]any{
			"grok": map[string]any{
				"capabilities": map[string]any{
					"models":     []any{"grok-3"},
					"operations": []any{"chat"},
				},
			},
		},
	}

	capabilities := buildGrokCapabilitySyncSnapshot(account, grok.TierBasic)
	require.ElementsMatch(t, []string{"chat", "image"}, grokParseStringSlice(capabilities["operations"]))
	models := grokParseStringSlice(capabilities["models"])
	require.Contains(t, models, "grok-4.20-fast")
	require.Contains(t, models, "grok-4.20-auto")
	require.Contains(t, models, "grok-imagine-image-lite")
	require.NotContains(t, models, "grok-4.20-heavy")
}

func TestGrokCapabilityProbeServiceShouldProbeAccountUsesRuntimeSettings(t *testing.T) {
	settingSvc := NewSettingService(&grokRuntimeSettingRepoStub{
		values: map[string]string{
			SettingKeyGrokCapabilityProbeIntervalSeconds: "60",
		},
	}, nil)

	svc := NewGrokCapabilityProbeService(
		nil,
		nil,
		nil,
		&config.Config{},
		nil,
		settingSvc,
	)
	now := time.Date(2026, 4, 19, 10, 0, 0, 0, time.UTC)
	lastProbeAt := now.Add(-2 * time.Minute)
	account := &Account{
		Platform:    PlatformGrok,
		Type:        AccountTypeAPIKey,
		Status:      StatusActive,
		Schedulable: true,
		Credentials: map[string]any{
			"api_key": "xai-test-key",
		},
		Extra: map[string]any{
			"grok": map[string]any{
				"sync_state": map[string]any{
					"last_probe_at": lastProbeAt.Format(time.RFC3339),
				},
				"capabilities": map[string]any{
					"models": []any{"grok-3"},
				},
			},
		},
	}

	require.True(t, svc.shouldProbeAccount(account, now, svc.currentInterval(context.Background())))
}

func TestGrokCapabilityProbeServiceShouldProbeKnownTierAccountsWhenProbeRefreshIsDue(t *testing.T) {
	svc := NewGrokCapabilityProbeService(nil, nil, nil, &config.Config{}, nil, nil)
	now := time.Date(2026, 4, 19, 10, 0, 0, 0, time.UTC)
	lastProbeAt := now.Add(-8 * time.Hour)
	account := &Account{
		Platform:    PlatformGrok,
		Type:        AccountTypeSession,
		Status:      StatusActive,
		Schedulable: true,
		Credentials: map[string]any{
			"session_token": "grok-session-token",
		},
		Extra: map[string]any{
			"grok": map[string]any{
				"tier": map[string]any{
					"normalized": "super",
				},
				"sync_state": map[string]any{
					"last_probe_at": lastProbeAt.Format(time.RFC3339),
				},
				"capabilities": map[string]any{
					"operations": []any{"chat", "image", "video"},
					"models":     []any{"grok-3", "grok-imagine-video"},
				},
			},
		},
	}

	require.True(t, svc.shouldProbeAccount(account, now, DefaultGrokRuntimeSettings().CapabilityProbeInterval()))
	require.Equal(t, []string{grokCapabilityTierBootstrapModelID, grok.DefaultTestModel}, grokCapabilityProbeModelCandidates(account, ""))
}

func TestGrokCapabilityProbeServiceShouldProbeSessionAccounts(t *testing.T) {
	svc := NewGrokCapabilityProbeService(nil, nil, nil, &config.Config{}, nil, nil)
	now := time.Date(2026, 4, 19, 10, 0, 0, 0, time.UTC)
	account := &Account{
		Platform:    PlatformGrok,
		Type:        AccountTypeSession,
		Status:      StatusActive,
		Schedulable: true,
		Credentials: map[string]any{
			"session_token": "grok-session-token",
		},
	}

	require.True(t, svc.shouldProbeAccount(account, now, DefaultGrokRuntimeSettings().CapabilityProbeInterval()))
}

func TestBuildGrokProbeCapabilitiesWidensKnownTierBaseline(t *testing.T) {
	account := &Account{
		Platform: PlatformGrok,
		Type:     AccountTypeSession,
		Extra: map[string]any{
			"grok": map[string]any{
				"tier": map[string]any{
					"normalized": "basic",
				},
			},
		},
	}

	capabilities := buildGrokProbeCapabilities(account, "grok-3-fast")
	require.ElementsMatch(t, []string{"chat", "image"}, grokParseStringSlice(capabilities["operations"]))
	models := grokParseStringSlice(capabilities["models"])
	require.Contains(t, models, "grok-4.20-fast")
	require.Contains(t, models, "grok-4.20-auto")
	require.Contains(t, models, "grok-imagine-image-lite")
	require.NotContains(t, models, "grok-4.20-heavy")
}

func TestGrokTierStateInfersHighTierFromCapabilityModelsWithoutDowngradingBasicProbe(t *testing.T) {
	heavyAccount := &Account{
		Platform: PlatformGrok,
		Extra: map[string]any{
			"grok": map[string]any{
				"capabilities": map[string]any{
					"models": []any{"grok-4.20-heavy"},
				},
			},
		},
	}
	require.Equal(t, grok.TierHeavy, heavyAccount.GrokTierState().Normalized)
	require.Equal(t, "capability_models", heavyAccount.GrokTierState().Source)

	basicOnlyAccount := &Account{
		Platform: PlatformGrok,
		Extra: map[string]any{
			"grok": map[string]any{
				"capabilities": map[string]any{
					"models": []any{"grok-3"},
				},
			},
		},
	}
	require.Equal(t, grok.TierUnknown, basicOnlyAccount.GrokTierState().Normalized)
}
