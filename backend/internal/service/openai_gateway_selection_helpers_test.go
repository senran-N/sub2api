package service

import (
	"context"
	"testing"
	"time"

	"github.com/senran-N/sub2api/internal/pkg/ctxkey"
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

func TestFilterSchedulableOpenAICandidatesForPlatform_Grok(t *testing.T) {
	accounts := []Account{
		{ID: 1, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true},
		{ID: 2, Platform: PlatformGrok, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true},
		{ID: 3, Platform: PlatformGrok, Type: AccountTypeAPIKey, Status: StatusDisabled, Schedulable: true},
	}

	candidates := filterSchedulableOpenAICandidatesForPlatform(accounts, "grok-3", nil, PlatformGrok)
	if len(candidates) != 1 {
		t.Fatalf("expected 1 grok candidate, got %d", len(candidates))
	}
	if candidates[0].ID != 2 {
		t.Fatalf("expected grok account 2, got %d", candidates[0].ID)
	}
}

func TestFilterSchedulableOpenAICandidatesForPlatform_GrokExcludesSessionFromSharedRuntime(t *testing.T) {
	accounts := []Account{
		{
			ID:          10,
			Platform:    PlatformGrok,
			Type:        AccountTypeSession,
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
			ID:          11,
			Platform:    PlatformGrok,
			Type:        AccountTypeAPIKey,
			Status:      StatusActive,
			Schedulable: true,
			Extra: map[string]any{
				"grok": map[string]any{
					"tier": map[string]any{
						"normalized": "heavy",
					},
				},
			},
		},
	}

	candidates := filterSchedulableOpenAICandidatesForPlatform(accounts, "grok-4-fast-reasoning", nil, PlatformGrok)
	if len(candidates) != 1 {
		t.Fatalf("expected 1 shared-runtime grok candidate, got %d", len(candidates))
	}
	if candidates[0].ID != 11 {
		t.Fatalf("expected grok api-key account 11, got %d", candidates[0].ID)
	}
}

func TestFilterSchedulableOpenAICandidatesForPlatform_GrokSessionAllowedForTextRuntime(t *testing.T) {
	accounts := []Account{
		{
			ID:          10,
			Platform:    PlatformGrok,
			Type:        AccountTypeSession,
			Status:      StatusActive,
			Schedulable: true,
			Extra: map[string]any{
				"grok": map[string]any{
					"tier": map[string]any{
						"normalized": "heavy",
					},
				},
			},
		},
	}

	candidates := filterSchedulableOpenAICandidatesForPlatformWithContext(
		WithGrokSessionTextRuntimeAllowed(context.Background()),
		accounts,
		"grok-4-fast-reasoning",
		nil,
		PlatformGrok,
	)
	if len(candidates) != 1 {
		t.Fatalf("expected 1 text-runtime grok session candidate, got %d", len(candidates))
	}
	if candidates[0].ID != 10 {
		t.Fatalf("expected grok session account 10, got %d", candidates[0].ID)
	}
}

func TestFilterSchedulableOpenAICandidatesForPlatform_GrokCapabilityAware(t *testing.T) {
	accounts := []Account{
		{
			ID:          20,
			Platform:    PlatformGrok,
			Type:        AccountTypeAPIKey,
			Status:      StatusActive,
			Schedulable: true,
			Extra: map[string]any{
				"grok": map[string]any{
					"capabilities": map[string]any{
						"operations": []any{"chat"},
					},
				},
			},
		},
		{
			ID:          21,
			Platform:    PlatformGrok,
			Type:        AccountTypeAPIKey,
			Status:      StatusActive,
			Schedulable: true,
			Extra: map[string]any{
				"grok": map[string]any{
					"capabilities": map[string]any{
						"operations": []any{"image"},
					},
				},
			},
		},
	}

	candidates := filterSchedulableOpenAICandidatesForPlatform(accounts, "grok-imagine-image", nil, PlatformGrok)
	if len(candidates) != 1 {
		t.Fatalf("expected 1 image-capable grok candidate, got %d", len(candidates))
	}
	if candidates[0].ID != 21 {
		t.Fatalf("expected image-capable grok account 21, got %d", candidates[0].ID)
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

func TestOpenAIRequestedModelAvailableForPlatform_Grok(t *testing.T) {
	accounts := []Account{
		{ID: 1, Platform: PlatformOpenAI, Type: AccountTypeAPIKey},
		{ID: 2, Platform: PlatformGrok, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true},
	}

	if !openAIRequestedModelAvailableForPlatform(accounts, "grok-3", PlatformGrok) {
		t.Fatal("expected grok model availability check to use grok accounts")
	}
}

func TestOpenAIRequestedModelAvailableForPlatform_GrokSessionOnlyIsUnavailableToSharedRuntime(t *testing.T) {
	accounts := []Account{
		{
			ID:          2,
			Platform:    PlatformGrok,
			Type:        AccountTypeSession,
			Status:      StatusActive,
			Schedulable: true,
			Extra: map[string]any{
				"grok": map[string]any{
					"tier": map[string]any{
						"normalized": "heavy",
					},
				},
			},
		},
	}

	if openAIRequestedModelAvailableForPlatform(accounts, "grok-4-fast-reasoning", PlatformGrok) {
		t.Fatal("expected grok session-only model availability to stay false until shared runtime owns session transport")
	}
}

func TestOpenAIRequestedModelAvailableForPlatform_GrokSessionOnlyIsAvailableForTextRuntime(t *testing.T) {
	accounts := []Account{
		{
			ID:          2,
			Platform:    PlatformGrok,
			Type:        AccountTypeSession,
			Status:      StatusActive,
			Schedulable: true,
			Extra: map[string]any{
				"grok": map[string]any{
					"tier": map[string]any{
						"normalized": "heavy",
					},
				},
			},
		},
	}

	if !openAIRequestedModelAvailableForPlatformWithContext(
		WithGrokSessionTextRuntimeAllowed(context.Background()),
		accounts,
		"grok-4-fast-reasoning",
		PlatformGrok,
	) {
		t.Fatal("expected grok session-only model availability on text runtime routes")
	}
}

func TestOpenAIRequestedModelAvailableForPlatform_GrokAccountCooldownStillCountsAsConfigured(t *testing.T) {
	now := time.Now().UTC()
	accounts := []Account{
		{
			ID:          2,
			Platform:    PlatformGrok,
			Type:        AccountTypeAPIKey,
			Status:      StatusActive,
			Schedulable: true,
			Extra: map[string]any{
				"grok": map[string]any{
					"tier": map[string]any{
						"normalized": "heavy",
					},
					"runtime_state": map[string]any{
						"selection_cooldown_until": now.Add(10 * time.Minute).Format(time.RFC3339),
						"selection_cooldown_scope": "account",
					},
				},
			},
		},
	}

	if !openAIRequestedModelAvailableForPlatform(accounts, "grok-4-fast-reasoning", PlatformGrok) {
		t.Fatal("expected account-scoped grok cooldown to remain model-configured for scheduler errors")
	}

	grokExtra, ok := accounts[0].Extra["grok"].(map[string]any)
	if !ok {
		t.Fatal("expected grok extra state")
	}
	grokExtra["runtime_state"] = map[string]any{
		"selection_cooldown_until": now.Add(10 * time.Minute).Format(time.RFC3339),
		"selection_cooldown_scope": "model",
		"selection_cooldown_model": "grok-4-fast-reasoning",
	}

	if openAIRequestedModelAvailableForPlatform(accounts, "grok-4-fast-reasoning", PlatformGrok) {
		t.Fatal("expected model-scoped grok cooldown to keep model unavailable")
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

func TestResolveOpenAIStickySessionAccount_DeleteOnGrokTierMismatch(t *testing.T) {
	ctx := context.WithValue(context.Background(), ctxkey.ForcePlatform, PlatformGrok)
	cache := &stubGatewayCache{
		sessionBindings: map[string]int64{
			"openai:sticky_model_unsupported": 601,
		},
	}
	svc := &OpenAIGatewayService{
		accountRepo: stubOpenAIAccountRepo{accounts: []Account{
			{
				ID:          601,
				Platform:    PlatformGrok,
				Type:        AccountTypeSession,
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
		}},
		cache: cache,
	}

	account, accountID := svc.resolveOpenAIStickySessionAccount(
		ctx,
		nil,
		"sticky_model_unsupported",
		"grok-4-fast-reasoning",
		nil,
		0,
		openAIStickySessionResolvePolicy{},
	)

	if account != nil {
		t.Fatal("expected nil account on grok tier mismatch")
	}
	if accountID != 0 {
		t.Fatalf("expected accountID=0 on grok tier mismatch, got %d", accountID)
	}
	if cache.deletedSessions["openai:sticky_model_unsupported"] != 1 {
		t.Fatalf("expected grok tier mismatch to delete sticky key, got %d", cache.deletedSessions["openai:sticky_model_unsupported"])
	}
}
