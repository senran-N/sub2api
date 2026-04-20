//go:build unit

package service

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/senran-N/sub2api/internal/pkg/grok"
	"github.com/stretchr/testify/require"
)

type grokAccountStateServiceRepoStub struct {
	updatedID    int64
	updatedExtra map[string]any
}

func (r *grokAccountStateServiceRepoStub) UpdateExtra(_ context.Context, id int64, updates map[string]any) error {
	r.updatedID = id
	r.updatedExtra = cloneAnyMap(updates)
	return nil
}

func TestGrokAccountStateService_PersistProbeResultSuccessWritesNormalizedProviderState(t *testing.T) {
	repo := &grokAccountStateServiceRepoStub{}
	svc := NewGrokAccountStateService(repo)
	svc.now = func() time.Time {
		return time.Date(2026, 4, 19, 8, 0, 0, 0, time.UTC)
	}

	account := &Account{
		ID:       84,
		Platform: PlatformGrok,
		Type:     AccountTypeSession,
		Extra: map[string]any{
			"grok": map[string]any{
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
					"last_sync_at": "2026-04-19T00:00:00Z",
				},
			},
		},
	}

	svc.PersistProbeResult(context.Background(), account, "grok-3-fast", &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
	}, nil)

	require.Equal(t, int64(84), repo.updatedID)
	grokExtra := grokExtraMap(repo.updatedExtra)
	require.Equal(t, "basic", getNestedGrokValue(grokExtra, "tier", "normalized"))
	require.Equal(t, "manual", getNestedGrokValue(grokExtra, "tier", "source"))
	require.Equal(t, "2026-04-19T00:00:00Z", getNestedGrokValue(grokExtra, "sync_state", "last_sync_at"))
	require.Equal(t, "2026-04-19T08:00:00Z", getNestedGrokValue(grokExtra, "sync_state", "last_probe_at"))
	require.Equal(t, "2026-04-19T08:00:00Z", getNestedGrokValue(grokExtra, "sync_state", "last_probe_ok_at"))
	require.Equal(t, 200, grokParseInt(getNestedGrokValue(grokExtra, "sync_state", "last_probe_status_code")))
	expectedCapabilities := buildGrokCapabilitySyncSnapshot(account, grok.TierBasic)
	require.ElementsMatch(t, grokParseStringSlice(expectedCapabilities["operations"]), grokParseStringSlice(getNestedGrokValue(grokExtra, "capabilities", "operations")))
	require.ElementsMatch(t, grokParseStringSlice(expectedCapabilities["models"]), grokParseStringSlice(getNestedGrokValue(grokExtra, "capabilities", "models")))
	require.Equal(t, "2026-04-19T08:00:00Z", getNestedGrokValue(account.grokExtraMap(), "sync_state", "last_probe_at"))
}

func TestGrokAccountStateService_PersistProbeResultFailurePreservesTierAndQuotaSignals(t *testing.T) {
	repo := &grokAccountStateServiceRepoStub{}
	svc := NewGrokAccountStateService(repo)
	svc.now = func() time.Time {
		return time.Date(2026, 4, 19, 9, 0, 0, 0, time.UTC)
	}

	account := &Account{
		ID:       85,
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
				"sync_state": map[string]any{
					"last_sync_at":     "2026-04-19T00:00:00Z",
					"last_probe_ok_at": "2026-04-18T00:00:00Z",
				},
			},
		},
	}

	svc.PersistProbeResult(context.Background(), account, "grok-3-fast", &http.Response{
		StatusCode: http.StatusUnauthorized,
		Status:     "401 Unauthorized",
	}, errors.New("API returned 401: unauthorized"))

	grokExtra := grokExtraMap(repo.updatedExtra)
	require.Equal(t, "super", getNestedGrokValue(grokExtra, "tier", "normalized"))
	require.Equal(t, "sync", getNestedGrokValue(grokExtra, "tier", "source"))
	require.Equal(t, "2026-04-19T00:00:00Z", getNestedGrokValue(grokExtra, "sync_state", "last_sync_at"))
	require.Equal(t, "2026-04-18T00:00:00Z", getNestedGrokValue(grokExtra, "sync_state", "last_probe_ok_at"))
	require.Equal(t, "2026-04-19T09:00:00Z", getNestedGrokValue(grokExtra, "sync_state", "last_probe_at"))
	require.Equal(t, "2026-04-19T09:00:00Z", getNestedGrokValue(grokExtra, "sync_state", "last_probe_error_at"))
	require.Equal(t, 401, grokParseInt(getNestedGrokValue(grokExtra, "sync_state", "last_probe_status_code")))
	require.Contains(t, getStringFromMaps(grokNestedMap(grokExtra["sync_state"]), nil, "last_probe_error"), "API returned 401")
	require.Equal(t, 7, grokParseInt(getNestedGrokValue(grokQuotaWindowsMap(grokExtra["quota_windows"]), grok.QuotaWindowAuto, "remaining")))
}
