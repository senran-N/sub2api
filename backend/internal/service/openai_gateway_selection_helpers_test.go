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

func TestResolveOpenAIStickySessionAccount_DeleteOnLookupMiss(t *testing.T) {
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
	if cache.deletedSessions["openai:sticky_missing"] != 1 {
		t.Fatalf("expected sticky key to be deleted once, got %d", cache.deletedSessions["openai:sticky_missing"])
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
