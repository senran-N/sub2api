//go:build unit

package service

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/senran-N/sub2api/internal/pkg/grok"
	"github.com/stretchr/testify/require"
)

type grokQuotaSyncRepoStub struct {
	accounts      []Account
	updatedIDs    []int64
	updatedExtra  []map[string]any
	runtimeStates []map[string]any
	setErrorIDs   []int64
	setErrorMsgs  []string
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

func (r *grokQuotaSyncRepoStub) UpdateGrokRuntimeState(_ context.Context, _ int64, runtimeState map[string]any) error {
	r.runtimeStates = append(r.runtimeStates, cloneAnyMap(runtimeState))
	return nil
}

func (r *grokQuotaSyncRepoStub) SetError(_ context.Context, id int64, errorMsg string) error {
	r.setErrorIDs = append(r.setErrorIDs, id)
	r.setErrorMsgs = append(r.setErrorMsgs, errorMsg)
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
	quotaSvc := NewGrokQuotaSyncService(repo, stateSvc, NewGrokTierService(), nil)
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
	quotaSvc := NewGrokQuotaSyncService(repo, stateSvc, NewGrokTierService(), nil)
	quotaSvc.now = func() time.Time {
		return time.Date(2026, 4, 19, 9, 0, 0, 0, time.UTC)
	}
	stateSvc.now = quotaSvc.now

	err := quotaSvc.SyncNow(context.Background())
	require.NoError(t, err)
	require.Len(t, repo.updatedExtra, 1)

	grokExtra := grokExtraMap(repo.updatedExtra[0])
	require.Equal(t, "basic", getNestedGrokValue(grokExtra, "tier", "normalized"))
	expectedCapabilities := buildGrokCapabilitySyncSnapshot(&repo.accounts[0], grok.TierBasic)
	require.ElementsMatch(t, grokParseStringSlice(expectedCapabilities["operations"]), grokParseStringSlice(getNestedGrokValue(grokExtra, "capabilities", "operations")))
	require.ElementsMatch(t, grokParseStringSlice(expectedCapabilities["models"]), grokParseStringSlice(getNestedGrokValue(grokExtra, "capabilities", "models")))
}

func TestGrokQuotaSyncServiceCurrentIntervalUsesRuntimeSettings(t *testing.T) {
	settingSvc := NewSettingService(&grokRuntimeSettingRepoStub{
		values: map[string]string{
			SettingKeyGrokQuotaSyncIntervalSeconds: "120",
		},
	}, nil)

	svc := NewGrokQuotaSyncService(nil, nil, nil, settingSvc)
	require.Equal(t, 2*time.Minute, svc.currentInterval(context.Background()))
}

func TestGrokQuotaSyncServiceSyncAccountInvalidCredentialsPersistsAuthCooldown(t *testing.T) {
	repo := &grokQuotaSyncRepoStub{}
	stateSvc := NewGrokAccountStateService(repo)
	quotaSvc := NewGrokQuotaSyncService(repo, stateSvc, NewGrokTierService(), nil)
	quotaSvc.httpUpstream = &queuedHTTPUpstream{
		responses: []*http.Response{
			newJSONResponse(http.StatusBadRequest, `{"error":"invalid-credentials"}`),
		},
	}
	now := time.Date(2026, 4, 19, 12, 0, 0, 0, time.UTC)
	quotaSvc.now = func() time.Time { return now }
	stateSvc.now = quotaSvc.now

	account := &Account{
		ID:          93,
		Platform:    PlatformGrok,
		Type:        AccountTypeSession,
		Status:      StatusActive,
		Schedulable: true,
		Credentials: map[string]any{
			"session_token": "session-cookie-93",
		},
	}

	err := quotaSvc.SyncAccount(context.Background(), account)
	require.Error(t, err)
	require.Len(t, repo.updatedExtra, 1)
	require.Len(t, repo.runtimeStates, 1)
	require.Equal(t, "error", repo.runtimeStates[0]["last_outcome"])
	require.Equal(t, "auth", repo.runtimeStates[0]["last_fail_class"])
	require.Equal(t, http.StatusBadRequest, grokParseInt(repo.runtimeStates[0]["last_fail_status_code"]))
	require.NotEmpty(t, repo.runtimeStates[0]["selection_cooldown_until"])
	require.Contains(t, getStringFromMaps(grokNestedMap(grokExtraMap(repo.updatedExtra[0])["sync_state"]), nil, "last_sync_error"), "invalid-credentials")
	require.Equal(t, []int64{93}, repo.setErrorIDs)
	require.Len(t, repo.setErrorMsgs, 1)
	require.Contains(t, repo.setErrorMsgs[0], "grok invalid credentials")
	require.Equal(t, StatusError, account.Status)
}
