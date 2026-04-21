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
	runtimeState map[string]any
	setErrorID   int64
	setErrorMsg  string
}

func (r *grokAccountStateServiceRepoStub) UpdateExtra(_ context.Context, id int64, updates map[string]any) error {
	r.updatedID = id
	r.updatedExtra = cloneAnyMap(updates)
	return nil
}

func (r *grokAccountStateServiceRepoStub) UpdateGrokRuntimeState(_ context.Context, _ int64, runtimeState map[string]any) error {
	r.runtimeState = cloneAnyMap(runtimeState)
	return nil
}

func (r *grokAccountStateServiceRepoStub) SetError(_ context.Context, id int64, errorMsg string) error {
	r.setErrorID = id
	r.setErrorMsg = errorMsg
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
	require.Equal(t, "error", repo.runtimeState["last_outcome"])
	require.Equal(t, "auth", repo.runtimeState["last_fail_class"])
	require.NotEmpty(t, repo.runtimeState["selection_cooldown_until"])
	require.Zero(t, repo.setErrorID)
}

func TestGrokAccountStateService_PersistProbeResultInvalidCredentials400TriggersAuthCooldown(t *testing.T) {
	repo := &grokAccountStateServiceRepoStub{}
	svc := NewGrokAccountStateService(repo)
	svc.now = func() time.Time {
		return time.Date(2026, 4, 19, 10, 0, 0, 0, time.UTC)
	}

	account := &Account{
		ID:       86,
		Platform: PlatformGrok,
		Type:     AccountTypeSession,
	}

	svc.PersistProbeResult(
		context.Background(),
		account,
		"grok-3",
		&http.Response{StatusCode: http.StatusBadRequest, Status: "400 Bad Request"},
		errors.New(`API returned 400: {"error":"invalid-credentials"}`),
	)

	require.Equal(t, "error", repo.runtimeState["last_outcome"])
	require.Equal(t, "auth", repo.runtimeState["last_fail_class"])
	require.Equal(t, http.StatusBadRequest, grokParseInt(repo.runtimeState["last_fail_status_code"]))
	require.NotEmpty(t, repo.runtimeState["selection_cooldown_until"])
	require.Equal(t, int64(86), repo.setErrorID)
	require.Contains(t, repo.setErrorMsg, "grok invalid credentials")
	require.Contains(t, repo.setErrorMsg, "invalid-credentials")
	require.Equal(t, StatusError, account.Status)
	require.Equal(t, "auth", getNestedGrokValue(account.grokExtraMap(), "runtime_state", "last_fail_class"))
}

func TestGrokAccountStateService_PersistProbeResultSuccessClearsPreviousAuthCooldown(t *testing.T) {
	repo := &grokAccountStateServiceRepoStub{}
	svc := NewGrokAccountStateService(repo)
	svc.now = func() time.Time {
		return time.Date(2026, 4, 19, 11, 0, 0, 0, time.UTC)
	}

	account := &Account{
		ID:       87,
		Platform: PlatformGrok,
		Type:     AccountTypeAPIKey,
		Extra: map[string]any{
			"grok": map[string]any{
				"runtime_state": map[string]any{
					"last_fail_at":             "2026-04-19T10:00:00Z",
					"last_fail_class":          "auth",
					"selection_cooldown_until": "2026-04-19T10:30:00Z",
					"selection_cooldown_scope": "account",
				},
			},
		},
	}

	svc.PersistProbeResult(
		context.Background(),
		account,
		"grok-3-fast",
		&http.Response{StatusCode: http.StatusOK, Status: "200 OK"},
		nil,
	)

	require.Equal(t, "success", repo.runtimeState["last_outcome"])
	require.Equal(t, http.StatusOK, grokParseInt(repo.runtimeState["last_request_status_code"]))
	require.Nil(t, repo.runtimeState["selection_cooldown_until"])
	require.NotEmpty(t, repo.runtimeState["last_use_at"])
	require.Nil(t, getNestedGrokValue(account.grokExtraMap(), "runtime_state", "selection_cooldown_until"))
}

func TestGrokAccountStateService_PersistProbeResultSuccessClearsPreviousRateLimitCooldown(t *testing.T) {
	repo := &grokAccountStateServiceRepoStub{}
	svc := NewGrokAccountStateService(repo)
	svc.now = func() time.Time {
		return time.Date(2026, 4, 19, 11, 30, 0, 0, time.UTC)
	}

	account := &Account{
		ID:       88,
		Platform: PlatformGrok,
		Type:     AccountTypeSession,
		Extra: map[string]any{
			"grok": map[string]any{
				"runtime_state": map[string]any{
					"last_fail_at":             "2026-04-19T11:00:00Z",
					"last_fail_class":          "rate_limited",
					"selection_cooldown_until": "2026-04-19T11:40:00Z",
					"selection_cooldown_scope": "account",
				},
			},
		},
	}

	svc.PersistProbeResult(
		context.Background(),
		account,
		"grok-3-fast",
		&http.Response{StatusCode: http.StatusOK, Status: "200 OK"},
		nil,
	)

	require.Equal(t, "success", repo.runtimeState["last_outcome"])
	require.Equal(t, http.StatusOK, grokParseInt(repo.runtimeState["last_request_status_code"]))
	require.Nil(t, repo.runtimeState["selection_cooldown_until"])
	require.NotEmpty(t, repo.runtimeState["last_use_at"])
	require.Nil(t, getNestedGrokValue(account.grokExtraMap(), "runtime_state", "selection_cooldown_until"))
}

func TestGrokAccountStateService_PersistSyncSnapshotSuccessClearsPreviousRateLimitCooldown(t *testing.T) {
	repo := &grokAccountStateServiceRepoStub{}
	svc := NewGrokAccountStateService(repo)
	svc.now = func() time.Time {
		return time.Date(2026, 4, 19, 12, 0, 0, 0, time.UTC)
	}

	account := &Account{
		ID:       89,
		Platform: PlatformGrok,
		Type:     AccountTypeSession,
		Extra: map[string]any{
			"grok": map[string]any{
				"runtime_state": map[string]any{
					"last_fail_at":             "2026-04-19T11:00:00Z",
					"last_fail_class":          "rate_limited",
					"selection_cooldown_until": "2026-04-19T12:10:00Z",
					"selection_cooldown_scope": "account",
				},
			},
		},
	}

	svc.PersistSyncSnapshot(context.Background(), account, grokStateSyncSnapshot{
		AuthMode: AccountTypeSession,
		SyncState: map[string]any{
			"last_sync_at":          "2026-04-19T12:00:00Z",
			"last_sync_ok_at":       "2026-04-19T12:00:00Z",
			"last_sync_status_code": http.StatusOK,
		},
	}, http.StatusOK, nil)

	require.Equal(t, "success", repo.runtimeState["last_outcome"])
	require.Equal(t, http.StatusOK, grokParseInt(repo.runtimeState["last_request_status_code"]))
	require.Nil(t, repo.runtimeState["selection_cooldown_until"])
	require.NotEmpty(t, repo.runtimeState["last_use_at"])
	require.Nil(t, getNestedGrokValue(account.grokExtraMap(), "runtime_state", "selection_cooldown_until"))
}

func TestGrokAccountStateService_PersistProbeResultSuccessKeepsModelUnsupportedCooldown(t *testing.T) {
	repo := &grokAccountStateServiceRepoStub{}
	svc := NewGrokAccountStateService(repo)
	svc.now = func() time.Time {
		return time.Date(2026, 4, 19, 12, 30, 0, 0, time.UTC)
	}

	account := &Account{
		ID:       90,
		Platform: PlatformGrok,
		Type:     AccountTypeAPIKey,
		Extra: map[string]any{
			"grok": map[string]any{
				"runtime_state": map[string]any{
					"last_fail_at":             "2026-04-19T12:00:00Z",
					"last_fail_class":          "model_unsupported",
					"selection_cooldown_until": "2026-04-19T13:00:00Z",
					"selection_cooldown_scope": "model",
					"selection_cooldown_model": "grok-imagine-video",
				},
			},
		},
	}

	svc.PersistProbeResult(
		context.Background(),
		account,
		"grok-3-fast",
		&http.Response{StatusCode: http.StatusOK, Status: "200 OK"},
		nil,
	)

	require.Nil(t, repo.runtimeState)
	require.Equal(t, "2026-04-19T13:00:00Z", getNestedGrokValue(account.grokExtraMap(), "runtime_state", "selection_cooldown_until"))
}
