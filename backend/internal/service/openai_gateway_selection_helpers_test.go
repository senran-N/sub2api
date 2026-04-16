package service

import (
	"context"
	"testing"
	"time"
)

func TestFilterSchedulableOpenAICandidates(t *testing.T) {
	accounts := []Account{
		{
			ID:          1,
			Platform:    PlatformOpenAI,
			Status:      StatusActive,
			Schedulable: true,
		},
		{
			ID:          2,
			Platform:    PlatformOpenAI,
			Status:      StatusActive,
			Schedulable: false,
		},
		{
			ID:          3,
			Platform:    PlatformOpenAI,
			Status:      StatusActive,
			Schedulable: true,
			Credentials: map[string]any{
				"model_mapping": map[string]any{
					"gpt-3.5": "gpt-3.5",
				},
			},
		},
		{
			ID:          4,
			Platform:    PlatformOpenAI,
			Status:      StatusActive,
			Schedulable: true,
		},
	}

	excluded := map[int64]struct{}{4: {}}
	candidates := filterSchedulableOpenAICandidates(accounts, "gpt-4", excluded)
	if len(candidates) != 1 {
		t.Fatalf("expected 1 candidate, got %d", len(candidates))
	}
	if candidates[0].ID != 1 {
		t.Fatalf("expected account 1, got %d", candidates[0].ID)
	}
}

func TestFilterSchedulableOpenAICandidates_OpenAIReasoningVariantBaseMapping(t *testing.T) {
	accounts := []Account{
		{
			ID:          1,
			Platform:    PlatformOpenAI,
			Status:      StatusActive,
			Schedulable: true,
			Credentials: map[string]any{
				"model_mapping": map[string]any{
					"gpt-5.4": "gpt-5.3-codex-spark",
				},
			},
		},
		{
			ID:          2,
			Platform:    PlatformOpenAI,
			Status:      StatusActive,
			Schedulable: true,
			Credentials: map[string]any{
				"model_mapping": map[string]any{
					"gpt-4.1": "gpt-4.1",
				},
			},
		},
	}

	candidates := filterSchedulableOpenAICandidates(accounts, "gpt-5.4-xhigh", nil)
	if len(candidates) != 1 {
		t.Fatalf("expected 1 candidate, got %d", len(candidates))
	}
	if candidates[0].ID != 1 {
		t.Fatalf("expected account 1, got %d", candidates[0].ID)
	}
}

func TestFilterSchedulableOpenAICandidates_DoesNotPromoteCodexExtraToRateLimit(t *testing.T) {
	resetAt := time.Now().Add(6 * 24 * time.Hour).UTC().Format(time.RFC3339)
	accounts := []Account{
		{
			ID:          1,
			Platform:    PlatformOpenAI,
			Type:        AccountTypeOAuth,
			Status:      StatusActive,
			Schedulable: true,
			Extra: map[string]any{
				"codex_5h_used_percent": 1.0,
				"codex_5h_reset_at":     time.Now().Add(2 * time.Hour).UTC().Format(time.RFC3339),
				"codex_7d_used_percent": 100.0,
				"codex_7d_reset_at":     resetAt,
			},
		},
	}

	candidates := filterSchedulableOpenAICandidates(accounts, "gpt-5.1", nil)
	if len(candidates) != 1 {
		t.Fatalf("expected codex extra to stay informational, got %d candidates", len(candidates))
	}
	if candidates[0].RateLimitResetAt != nil {
		t.Fatalf("unexpected runtime rate limit promoted from codex extra: %v", candidates[0].RateLimitResetAt)
	}
}

func TestOpenAIRequestedModelAvailable_OpenAIReasoningVariantBaseMapping(t *testing.T) {
	accounts := []Account{
		{
			ID:       1,
			Platform: PlatformOpenAI,
			Credentials: map[string]any{
				"model_mapping": map[string]any{
					"gpt-5.4": "gpt-5.3-codex-spark",
				},
			},
		},
	}

	if !openAIRequestedModelAvailable(accounts, "gpt-5.4-xhigh") {
		t.Fatal("expected reasoning variant to be considered available via base-model mapping")
	}
}

func TestSelectBestOpenAIWaitCandidate(t *testing.T) {
	now := time.Now()
	earlier := now.Add(-1 * time.Hour)
	muchEarlier := now.Add(-2 * time.Hour)

	candidates := []*Account{
		{ID: 1, Priority: 1, LastUsedAt: &now},
		{ID: 2, Priority: 1, LastUsedAt: &earlier},
		{ID: 3, Priority: 2, LastUsedAt: &muchEarlier},
	}
	loadMap := map[int64]*AccountLoadInfo{
		1: {AccountID: 1, WaitingCount: 2, LoadRate: 10},
		2: {AccountID: 2, WaitingCount: 2, LoadRate: 5},
		3: {AccountID: 3, WaitingCount: 1, LoadRate: 99},
	}

	best := selectBestOpenAIWaitCandidate(candidates, loadMap)
	if best == nil {
		t.Fatal("expected best candidate, got nil")
	}
	if best.ID != 3 {
		t.Fatalf("expected account 3 by waiting count, got %d", best.ID)
	}
}

func TestSelectBestOpenAIWaitCandidate_PreferNeverUsedOnFullTie(t *testing.T) {
	now := time.Now()

	candidates := []*Account{
		{ID: 1, Priority: 1, LastUsedAt: &now},
		{ID: 2, Priority: 1, LastUsedAt: nil},
	}
	loadMap := map[int64]*AccountLoadInfo{
		1: {AccountID: 1, WaitingCount: 0, LoadRate: 10},
		2: {AccountID: 2, WaitingCount: 0, LoadRate: 10},
	}

	best := selectBestOpenAIWaitCandidate(candidates, loadMap)
	if best == nil {
		t.Fatal("expected best candidate, got nil")
	}
	if best.ID != 2 {
		t.Fatalf("expected account 2 (never used), got %d", best.ID)
	}
}

func TestNormalizeOpenAIWaitLoadMapAndAccountLoadFallback(t *testing.T) {
	normalized := normalizeOpenAIWaitLoadMap(nil)
	if normalized == nil {
		t.Fatal("expected non-nil normalized map")
	}
	info := accountLoadInfoOrDefault(normalized, 11)
	if info.AccountID != 11 {
		t.Fatalf("expected fallback account id 11, got %d", info.AccountID)
	}
	if info.WaitingCount != 0 || info.LoadRate != 0 {
		t.Fatalf("expected zero-value fallback load info, got %+v", *info)
	}
}

func TestResolveOpenAIStickySessionAccount_KeepBindingOnLookupMiss(t *testing.T) {
	cache := &stubGatewayCache{
		sessionBindings: map[string]int64{
			"openai:sticky_missing": 404,
		},
	}
	svc := &OpenAIGatewayService{
		accountRepo: stubOpenAIAccountRepo{accounts: nil},
		cache:       cache,
	}

	account, accountID := svc.resolveOpenAIStickySessionAccount(
		context.Background(),
		nil,
		"sticky_missing",
		"",
		nil,
		0,
		openAIStickySessionResolvePolicy{
			deleteOnLookupMiss: true,
			refreshTTLOnHit:    false,
		},
	)

	if account != nil {
		t.Fatal("expected nil account on lookup miss")
	}
	if accountID != 0 {
		t.Fatalf("expected accountID=0 on lookup miss, got %d", accountID)
	}
	if len(cache.deletedSessions) != 0 {
		t.Fatalf("expected lookup miss to keep sticky binding, got %+v", cache.deletedSessions)
	}
}

func TestResolveOpenAIStickySessionAccount_KeepOnLookupMissWhenDisabled(t *testing.T) {
	cache := &stubGatewayCache{
		sessionBindings: map[string]int64{
			"openai:sticky_missing_no_delete": 405,
		},
	}
	svc := &OpenAIGatewayService{
		accountRepo: stubOpenAIAccountRepo{accounts: nil},
		cache:       cache,
	}

	account, accountID := svc.resolveOpenAIStickySessionAccount(
		context.Background(),
		nil,
		"sticky_missing_no_delete",
		"",
		nil,
		0,
		openAIStickySessionResolvePolicy{
			deleteOnLookupMiss: false,
			refreshTTLOnHit:    false,
		},
	)

	if account != nil {
		t.Fatal("expected nil account on lookup miss")
	}
	if accountID != 0 {
		t.Fatalf("expected accountID=0 on lookup miss, got %d", accountID)
	}
	if len(cache.deletedSessions) != 0 {
		t.Fatalf("expected no delete side-effect, got %+v", cache.deletedSessions)
	}
}

func TestResolveOpenAIStickySessionAccount_DeleteOnModelUnsupported(t *testing.T) {
	cache := &stubGatewayCache{
		sessionBindings: map[string]int64{
			"openai:sticky_model_unsupported": 501,
		},
	}
	svc := &OpenAIGatewayService{
		accountRepo: stubOpenAIAccountRepo{accounts: []Account{
			{
				ID:          501,
				Platform:    PlatformOpenAI,
				Status:      StatusActive,
				Schedulable: true,
				Credentials: map[string]any{
					"model_mapping": map[string]any{
						"gpt-4.1": "gpt-4.1",
					},
				},
			},
		}},
		cache: cache,
	}

	account, accountID := svc.resolveOpenAIStickySessionAccount(
		context.Background(),
		nil,
		"sticky_model_unsupported",
		"gpt-5.1",
		nil,
		0,
		openAIStickySessionResolvePolicy{},
	)

	if account != nil {
		t.Fatal("expected nil account on model mismatch")
	}
	if accountID != 0 {
		t.Fatalf("expected accountID=0 on model mismatch, got %d", accountID)
	}
	if cache.deletedSessions["openai:sticky_model_unsupported"] != 1 {
		t.Fatalf("expected model mismatch to delete sticky key, got %d", cache.deletedSessions["openai:sticky_model_unsupported"])
	}
}
