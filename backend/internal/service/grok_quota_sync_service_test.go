//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/senran-N/sub2api/internal/pkg/grok"
	"github.com/stretchr/testify/require"
)

type grokQuotaSyncRepoStub struct {
	accounts     []Account
	updatedIDs   []int64
	updatedExtra []map[string]any
}

func (r *grokQuotaSyncRepoStub) ListByPlatform(_ context.Context, platform string) ([]Account, error) {
	if platform != PlatformGrok {
		return nil, nil
	}
	result := make([]Account, len(r.accounts))
	copy(result, r.accounts)
	return result, nil
}

func (r *grokQuotaSyncRepoStub) UpdateExtra(_ context.Context, id int64, updates map[string]any) error {
	r.updatedIDs = append(r.updatedIDs, id)
	r.updatedExtra = append(r.updatedExtra, cloneAnyMap(updates))
	return nil
}

func TestGrokTierServiceBuildSnapshotInfersQuotaWindowSource(t *testing.T) {
	svc := NewGrokTierService()
	account := &Account{
		Platform: PlatformGrok,
		Type:     AccountTypeAPIKey,
		Extra: map[string]any{
			"grok": map[string]any{
				"quota_windows": map[string]any{
					grok.QuotaWindowAuto: map[string]any{
						"total": 50,
					},
				},
			},
		},
	}

	snapshot := svc.BuildSnapshot(account)
	require.Equal(t, "super", snapshot["normalized"])
	require.Equal(t, grokTierSourceQuotaWindows, snapshot["source"])
}

func TestGrokQuotaSyncServiceSyncNowPersistsNormalizedSchedulerState(t *testing.T) {
	repo := &grokQuotaSyncRepoStub{
		accounts: []Account{
			{
				ID:       91,
				Platform: PlatformGrok,
				Type:     AccountTypeAPIKey,
				Extra: map[string]any{
					"grok": map[string]any{
						"quota_windows": map[string]any{
							grok.QuotaWindowAuto: map[string]any{
								"remaining": 0,
								"total":     50,
								"reset_at":  "2026-04-19T07:00:00Z",
							},
						},
						"capabilities": map[string]any{
							"models": []any{"grok-3"},
						},
						"sync_state": map[string]any{
							"last_probe_ok_at": "2026-04-19T06:00:00Z",
						},
					},
				},
			},
		},
	}
	stateSvc := NewGrokAccountStateService(repo)
	quotaSvc := NewGrokQuotaSyncService(repo, stateSvc, NewGrokTierService())
	quotaSvc.now = func() time.Time {
		return time.Date(2026, 4, 19, 8, 0, 0, 0, time.UTC)
	}
	stateSvc.now = quotaSvc.now

	err := quotaSvc.SyncNow(context.Background())
	require.NoError(t, err)
	require.Equal(t, []int64{91}, repo.updatedIDs)
	require.Len(t, repo.updatedExtra, 1)

	grokExtra := grokExtraMap(repo.updatedExtra[0])
	require.Equal(t, AccountTypeAPIKey, grokExtra["auth_mode"])
	require.Equal(t, "super", getNestedGrokValue(grokExtra, "tier", "normalized"))
	require.Equal(t, grokTierSourceQuotaWindows, getNestedGrokValue(grokExtra, "tier", "source"))
	require.Equal(t, "2026-04-19T08:00:00Z", getNestedGrokValue(grokExtra, "sync_state", "last_sync_at"))
	require.Equal(t, "2026-04-19T06:00:00Z", getNestedGrokValue(grokExtra, "sync_state", "last_probe_ok_at"))
	require.ElementsMatch(t, []string{"grok-3"}, grokParseStringSlice(getNestedGrokValue(grokExtra, "capabilities", "models")))

	autoWindow := grokNestedMap(grokQuotaWindowsMap(grokExtra["quota_windows"])[grok.QuotaWindowAuto])
	require.Equal(t, 50, grokParseInt(autoWindow["remaining"]))
	require.Equal(t, 50, grokParseInt(autoWindow["total"]))
	_, hasResetAt := autoWindow["reset_at"]
	require.False(t, hasResetAt)
}

func TestGrokQuotaSyncServiceSyncNowWidensSimpleChatProbeCapabilitiesOnceTierIsKnown(t *testing.T) {
	repo := &grokQuotaSyncRepoStub{
		accounts: []Account{
			{
				ID:       92,
				Platform: PlatformGrok,
				Type:     AccountTypeAPIKey,
				Extra: map[string]any{
					"grok": map[string]any{
						"quota_windows": map[string]any{
							grok.QuotaWindowAuto: map[string]any{
								"remaining": 20,
								"total":     20,
							},
						},
						"capabilities": map[string]any{
							"models":     []any{"grok-3"},
							"operations": []any{"chat"},
						},
					},
				},
			},
		},
	}
	stateSvc := NewGrokAccountStateService(repo)
	quotaSvc := NewGrokQuotaSyncService(repo, stateSvc, NewGrokTierService())
	quotaSvc.now = func() time.Time {
		return time.Date(2026, 4, 19, 9, 0, 0, 0, time.UTC)
	}
	stateSvc.now = quotaSvc.now

	err := quotaSvc.SyncNow(context.Background())
	require.NoError(t, err)
	require.Len(t, repo.updatedExtra, 1)

	grokExtra := grokExtraMap(repo.updatedExtra[0])
	require.Equal(t, "basic", getNestedGrokValue(grokExtra, "tier", "normalized"))
	require.ElementsMatch(t, []string{"chat", "image"}, grokParseStringSlice(getNestedGrokValue(grokExtra, "capabilities", "operations")))
	require.ElementsMatch(t, []string{"grok-2-image", "grok-3", "grok-3-fast"}, grokParseStringSlice(getNestedGrokValue(grokExtra, "capabilities", "models")))
}
