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

func TestGrokCapabilityProbeServiceProbeNowTargetsOnlyUnknownTierCompatibleAccounts(t *testing.T) {
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
			},
		},
	}
	stateSvc := NewGrokAccountStateService(repo)
	upstream := &queuedHTTPUpstream{responses: []*http.Response{newJSONResponse(http.StatusOK, `{"id":"ok"}`)}}
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
	)
	probeNow := time.Date(2026, 4, 19, 10, 0, 0, 0, time.UTC)
	svc.now = func() time.Time { return probeNow }
	stateSvc.now = svc.now

	err := svc.ProbeNow(context.Background())
	require.NoError(t, err)
	require.Len(t, upstream.requests, 1)
	require.Equal(t, "https://api.x.ai/v1/responses", upstream.requests[0].URL.String())
	require.Equal(t, []int64{301}, repo.updatedIDs)
	require.Equal(t, grok.DefaultTestModel, requestBodyModel(t, upstream.requests[0]))

	grokExtra := grokExtraMap(repo.updatedExtra[0])
	require.Equal(t, "2026-04-19T10:00:00Z", getNestedGrokValue(grokExtra, "sync_state", "last_probe_at"))
	require.Equal(t, "2026-04-19T10:00:00Z", getNestedGrokValue(grokExtra, "sync_state", "last_probe_ok_at"))
	require.ElementsMatch(t, []string{"chat"}, grokParseStringSlice(getNestedGrokValue(grokExtra, "capabilities", "operations")))
	require.ElementsMatch(t, []string{grok.DefaultTestModel}, grokParseStringSlice(getNestedGrokValue(grokExtra, "capabilities", "models")))
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
	require.ElementsMatch(t, []string{"grok-2-image", "grok-3", "grok-3-fast"}, grokParseStringSlice(capabilities["models"]))
}
