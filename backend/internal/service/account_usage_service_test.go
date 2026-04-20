package service

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/senran-N/sub2api/internal/pkg/grok"
	"github.com/stretchr/testify/require"
)

type grokUsageSyncerStub struct {
	syncFn func(ctx context.Context, account *Account) error
}

func (s grokUsageSyncerStub) SyncAccount(ctx context.Context, account *Account) error {
	if s.syncFn == nil {
		return nil
	}
	return s.syncFn(ctx, account)
}

type accountUsageCodexProbeRepo struct {
	stubOpenAIAccountRepo
	updateExtraCh chan map[string]any
	rateLimitCh   chan time.Time
}

func (r *accountUsageCodexProbeRepo) UpdateExtra(_ context.Context, _ int64, updates map[string]any) error {
	if r.updateExtraCh != nil {
		copied := make(map[string]any, len(updates))
		for k, v := range updates {
			copied[k] = v
		}
		r.updateExtraCh <- copied
	}
	return nil
}

func (r *accountUsageCodexProbeRepo) SetRateLimited(_ context.Context, _ int64, resetAt time.Time) error {
	if r.rateLimitCh != nil {
		r.rateLimitCh <- resetAt
	}
	return nil
}

func TestShouldRefreshOpenAICodexSnapshot(t *testing.T) {
	t.Parallel()

	rateLimitedUntil := time.Now().Add(5 * time.Minute)
	now := time.Now()
	usage := &UsageInfo{
		FiveHour: &UsageProgress{Utilization: 0},
		SevenDay: &UsageProgress{Utilization: 0},
	}

	if !shouldRefreshOpenAICodexSnapshot(&Account{RateLimitResetAt: &rateLimitedUntil}, usage, now) {
		t.Fatal("expected rate-limited account to force codex snapshot refresh")
	}

	if shouldRefreshOpenAICodexSnapshot(&Account{}, usage, now) {
		t.Fatal("expected complete non-rate-limited usage to skip codex snapshot refresh")
	}

	if !shouldRefreshOpenAICodexSnapshot(&Account{}, &UsageInfo{FiveHour: nil, SevenDay: &UsageProgress{}}, now) {
		t.Fatal("expected missing 5h snapshot to require refresh")
	}

	staleAt := now.Add(-(openAIProbeCacheTTL + time.Minute)).Format(time.RFC3339)
	if !shouldRefreshOpenAICodexSnapshot(&Account{
		Platform: PlatformOpenAI,
		Type:     AccountTypeOAuth,
		Extra: map[string]any{
			"openai_oauth_responses_websockets_v2_enabled": true,
			"codex_usage_updated_at":                       staleAt,
		},
	}, usage, now) {
		t.Fatal("expected stale ws snapshot to trigger refresh")
	}
}

func TestExtractOpenAICodexProbeUpdatesAccepts429WithCodexHeaders(t *testing.T) {
	t.Parallel()

	headers := make(http.Header)
	headers.Set("x-codex-primary-used-percent", "100")
	headers.Set("x-codex-primary-reset-after-seconds", "604800")
	headers.Set("x-codex-primary-window-minutes", "10080")
	headers.Set("x-codex-secondary-used-percent", "100")
	headers.Set("x-codex-secondary-reset-after-seconds", "18000")
	headers.Set("x-codex-secondary-window-minutes", "300")

	updates, err := extractOpenAICodexProbeUpdates(&http.Response{StatusCode: http.StatusTooManyRequests, Header: headers})
	if err != nil {
		t.Fatalf("extractOpenAICodexProbeUpdates() error = %v", err)
	}
	if len(updates) == 0 {
		t.Fatal("expected codex probe updates from 429 headers")
	}
	if got := updates["codex_5h_used_percent"]; got != 100.0 {
		t.Fatalf("codex_5h_used_percent = %v, want 100", got)
	}
	if got := updates["codex_7d_used_percent"]; got != 100.0 {
		t.Fatalf("codex_7d_used_percent = %v, want 100", got)
	}
}

func TestAccountUsageService_PersistOpenAICodexProbeSnapshotOnlyUpdatesExtra(t *testing.T) {
	t.Parallel()

	repo := &accountUsageCodexProbeRepo{
		updateExtraCh: make(chan map[string]any, 1),
		rateLimitCh:   make(chan time.Time, 1),
	}
	svc := &AccountUsageService{accountRepo: repo}

	svc.persistOpenAICodexProbeSnapshot(321, map[string]any{
		"codex_7d_used_percent": 100.0,
		"codex_7d_reset_at":     time.Now().Add(2 * time.Hour).UTC().Truncate(time.Second).Format(time.RFC3339),
	})

	select {
	case updates := <-repo.updateExtraCh:
		if got := updates["codex_7d_used_percent"]; got != 100.0 {
			t.Fatalf("codex_7d_used_percent = %v, want 100", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("waiting for codex probe extra persistence timed out")
	}

	select {
	case got := <-repo.rateLimitCh:
		t.Fatalf("unexpected rate limit resetAt persisted from probe snapshot: %v", got)
	case <-time.After(200 * time.Millisecond):
	}
}

func TestAccountUsageService_GetOpenAIUsage_DoesNotPromoteCodexExtraToRateLimit(t *testing.T) {
	t.Parallel()

	resetAt := time.Now().Add(6 * 24 * time.Hour).UTC().Truncate(time.Second)
	repo := &accountUsageCodexProbeRepo{
		rateLimitCh: make(chan time.Time, 1),
	}
	svc := &AccountUsageService{accountRepo: repo}
	account := &Account{
		Platform: PlatformOpenAI,
		Type:     AccountTypeOAuth,
		Extra: map[string]any{
			"codex_5h_used_percent": 1.0,
			"codex_5h_reset_at":     time.Now().Add(2 * time.Hour).UTC().Truncate(time.Second).Format(time.RFC3339),
			"codex_7d_used_percent": 100.0,
			"codex_7d_reset_at":     resetAt.Format(time.RFC3339),
		},
	}

	usage, err := svc.getOpenAIUsage(context.Background(), account)
	if err != nil {
		t.Fatalf("getOpenAIUsage() error = %v", err)
	}
	if usage.SevenDay == nil || usage.SevenDay.Utilization != 100.0 {
		t.Fatalf("expected visible 7d usage snapshot, got %#v", usage.SevenDay)
	}
	if account.RateLimitResetAt != nil {
		t.Fatalf("unexpected runtime rate limit promoted from codex extra: %v", account.RateLimitResetAt)
	}
	select {
	case got := <-repo.rateLimitCh:
		t.Fatalf("unexpected rate limit persistence from codex extra: %v", got)
	case <-time.After(200 * time.Millisecond):
	}
}

func TestBuildCodexUsageProgressFromExtra_ZerosExpiredWindow(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 3, 16, 12, 0, 0, 0, time.UTC)

	t.Run("expired 5h window zeroes utilization", func(t *testing.T) {
		extra := map[string]any{
			"codex_5h_used_percent": 42.0,
			"codex_5h_reset_at":     "2026-03-16T10:00:00Z", // 2h ago
		}
		progress := buildCodexUsageProgressFromExtra(extra, "5h", now)
		if progress == nil {
			t.Fatal("expected non-nil progress")
		}
		if progress.Utilization != 0 {
			t.Fatalf("expected Utilization=0 for expired window, got %v", progress.Utilization)
		}
		if progress.RemainingSeconds != 0 {
			t.Fatalf("expected RemainingSeconds=0, got %v", progress.RemainingSeconds)
		}
	})

	t.Run("active 5h window keeps utilization", func(t *testing.T) {
		resetAt := now.Add(2 * time.Hour).Format(time.RFC3339)
		extra := map[string]any{
			"codex_5h_used_percent": 42.0,
			"codex_5h_reset_at":     resetAt,
		}
		progress := buildCodexUsageProgressFromExtra(extra, "5h", now)
		if progress == nil {
			t.Fatal("expected non-nil progress")
		}
		if progress.Utilization != 42.0 {
			t.Fatalf("expected Utilization=42, got %v", progress.Utilization)
		}
	})

	t.Run("expired 7d window zeroes utilization", func(t *testing.T) {
		extra := map[string]any{
			"codex_7d_used_percent": 88.0,
			"codex_7d_reset_at":     "2026-03-15T00:00:00Z", // yesterday
		}
		progress := buildCodexUsageProgressFromExtra(extra, "7d", now)
		if progress == nil {
			t.Fatal("expected non-nil progress")
		}
		if progress.Utilization != 0 {
			t.Fatalf("expected Utilization=0 for expired 7d window, got %v", progress.Utilization)
		}
	})
}

func TestAccountUsageService_BuildOpenAICodexProbeRequestUsesStablePersona(t *testing.T) {
	t.Parallel()

	svc := &AccountUsageService{
		identityCache: &identityCacheStub{
			fingerprint: &Fingerprint{UserAgent: "leaked-local-client/9.9"},
		},
	}
	account := &Account{
		ID:       42,
		Platform: PlatformOpenAI,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"chatgpt_account_id": "chatgpt-acc",
		},
	}

	req, err := svc.buildOpenAICodexProbeRequest(context.Background(), account, "access-token", []byte(`{"model":"gpt-5"}`))
	require.NoError(t, err)
	require.Equal(t, "chatgpt.com", req.Host)
	require.Equal(t, "Bearer access-token", req.Header.Get("Authorization"))
	require.Equal(t, "text/event-stream", req.Header.Get("Accept"))
	require.Equal(t, openAICodexHTTPBetaValue, req.Header.Get("OpenAI-Beta"))
	require.Equal(t, "codex_cli_rs", req.Header.Get("Originator"))
	require.Empty(t, req.Header.Get("Version"))
	require.Equal(t, codexCLIUserAgent, req.Header.Get("User-Agent"))
	require.Equal(t, "chatgpt-acc", req.Header.Get("chatgpt-account-id"))
	require.Empty(t, req.Header.Get("Accept-Language"))
	require.Empty(t, req.Header.Get("x-codex-beta-features"))
}

func TestAccountUsageService_GetGrokUsage_UsesSyncedQuotaWindows(t *testing.T) {
	t.Parallel()

	account := &Account{
		ID:       501,
		Platform: PlatformGrok,
		Type:     AccountTypeSession,
		Extra: map[string]any{
			"grok": map[string]any{
				"quota_windows": map[string]any{
					grok.QuotaWindowAuto: map[string]any{
						"remaining":      20,
						"total":          20,
						"window_seconds": 72000,
						"source":         "default",
					},
				},
			},
		},
	}

	svc := &AccountUsageService{
		grokQuotaSyncer: grokUsageSyncerStub{
			syncFn: func(_ context.Context, account *Account) error {
				account.Extra = map[string]any{
					"grok": map[string]any{
						"quota_windows": map[string]any{
							grok.QuotaWindowAuto: map[string]any{
								"remaining":      7,
								"total":          20,
								"window_seconds": 72000,
								"source":         "live",
								"reset_at":       "2099-03-16T14:00:00Z",
							},
							grok.QuotaWindowFast: map[string]any{
								"remaining":      40,
								"total":          60,
								"window_seconds": 72000,
								"source":         "live",
								"reset_at":       "2099-03-16T14:00:00Z",
							},
						},
						"sync_state": map[string]any{
							"last_probe_at": "2099-03-16T12:00:00Z",
						},
					},
				}
				return nil
			},
		},
	}

	usage, err := svc.getGrokUsage(context.Background(), account)
	require.NoError(t, err)
	require.NotNil(t, usage)
	require.NotNil(t, usage.GrokQuotaWindows)
	require.Len(t, usage.GrokQuotaWindows, 2)
	require.Equal(t, int64(13), usage.GrokQuotaWindows[grok.QuotaWindowAuto].UsedRequests)
	require.Equal(t, int64(20), usage.GrokQuotaWindows[grok.QuotaWindowAuto].LimitRequests)
	require.Equal(t, 65.0, usage.GrokQuotaWindows[grok.QuotaWindowAuto].Utilization)
	require.Equal(t, int64(20), usage.GrokQuotaWindows[grok.QuotaWindowFast].UsedRequests)
	require.InDelta(t, 33.33333333333333, usage.GrokQuotaWindows[grok.QuotaWindowFast].Utilization, 1e-9)
	require.NotNil(t, usage.UpdatedAt)
	require.Equal(t, "2099-03-16T12:00:00Z", usage.UpdatedAt.UTC().Format(time.RFC3339))
}

func TestNewOpenAICodexOAuthResponsesRequestHonorsAccountOverrides(t *testing.T) {
	t.Parallel()

	account := &Account{
		ID:       7,
		Platform: PlatformOpenAI,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"user_agent":          "Codex Desktop/0.200.0 (Windows 10.0.19045; x86_64)",
			"originator":          "codex_vscode",
			"version":             "v0.200.0",
			"accept_language":     "en-US",
			"codex_beta_features": "feature_a,feature_b",
		},
	}

	req, err := newOpenAICodexOAuthResponsesRequest(context.Background(), chatgptCodexURL, "token", []byte(`{}`), "application/json", account)
	require.NoError(t, err)
	require.Equal(t, "application/json", req.Header.Get("Accept"))
	require.Equal(t, "codex_cli_rs/0.200.0", req.Header.Get("User-Agent"))
	require.Equal(t, "codex_cli_rs", req.Header.Get("Originator"))
	require.Empty(t, req.Header.Get("Version"))
	require.Equal(t, "en-US", req.Header.Get("Accept-Language"))
	require.Equal(t, "feature_a,feature_b", req.Header.Get("x-codex-beta-features"))
	require.Equal(t, openAICodexHTTPBetaValue, req.Header.Get("OpenAI-Beta"))
}

func TestNewOpenAICodexOAuthResponsesRequestCompactSetsVersion(t *testing.T) {
	t.Parallel()

	account := &Account{
		ID:       8,
		Platform: PlatformOpenAI,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"version": "0.210.0",
		},
	}

	req, err := newOpenAICodexOAuthResponsesRequest(context.Background(), chatgptCodexURL+"/compact", "token", []byte(`{}`), "application/json", account)
	require.NoError(t, err)
	require.Equal(t, "0.210.0", req.Header.Get("Version"))
}

func TestNewOpenAICodexOAuthResponsesRequestIgnoresNonOfficialOverrideFingerprint(t *testing.T) {
	t.Parallel()

	account := &Account{
		ID:       9,
		Platform: PlatformOpenAI,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"user_agent": "Mozilla/5.0 CustomApp/9.9",
			"originator": "custom_desktop",
			"version":    "not-a-version",
		},
	}

	req, err := newOpenAICodexOAuthResponsesRequest(context.Background(), chatgptCodexURL+"/compact", "token", []byte(`{}`), "application/json", account)
	require.NoError(t, err)
	require.Equal(t, codexCLIUserAgent, req.Header.Get("User-Agent"))
	require.Equal(t, "codex_cli_rs", req.Header.Get("Originator"))
	require.Equal(t, codexCLIVersion, req.Header.Get("Version"))
}
