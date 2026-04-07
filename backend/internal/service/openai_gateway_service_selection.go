package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/senran-N/sub2api/internal/config"
)

// SelectAccount selects an OpenAI account with sticky session support.
func (s *OpenAIGatewayService) SelectAccount(ctx context.Context, groupID *int64, sessionHash string) (*Account, error) {
	return s.SelectAccountForModel(ctx, groupID, sessionHash, "")
}

// SelectAccountForModel selects an account supporting the requested model.
func (s *OpenAIGatewayService) SelectAccountForModel(ctx context.Context, groupID *int64, sessionHash string, requestedModel string) (*Account, error) {
	return s.SelectAccountForModelWithExclusions(ctx, groupID, sessionHash, requestedModel, nil)
}

// SelectAccountForModelWithExclusions selects an account supporting the requested model while excluding specified accounts.
func (s *OpenAIGatewayService) SelectAccountForModelWithExclusions(ctx context.Context, groupID *int64, sessionHash string, requestedModel string, excludedIDs map[int64]struct{}) (*Account, error) {
	return s.selectAccountForModelWithExclusions(ctx, groupID, sessionHash, requestedModel, excludedIDs, 0)
}

func (s *OpenAIGatewayService) selectAccountForModelWithExclusions(ctx context.Context, groupID *int64, sessionHash string, requestedModel string, excludedIDs map[int64]struct{}, stickyAccountID int64) (*Account, error) {
	if account := s.tryStickySessionHit(ctx, groupID, sessionHash, requestedModel, excludedIDs, stickyAccountID); account != nil {
		return account, nil
	}

	accounts, err := s.listSchedulableAccounts(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("query accounts failed: %w", err)
	}

	selected := s.selectBestAccount(ctx, accounts, requestedModel, excludedIDs)
	if selected == nil {
		if requestedModel != "" {
			return nil, fmt.Errorf("no available OpenAI accounts supporting model: %s", requestedModel)
		}
		return nil, errors.New("no available OpenAI accounts")
	}

	if sessionHash != "" {
		_ = s.setStickySessionAccountID(ctx, groupID, sessionHash, selected.ID, s.openAIWSSessionStickyTTL())
	}

	return selected, nil
}

// tryStickySessionHit tries to resolve a sticky session to a currently usable account.
func (s *OpenAIGatewayService) tryStickySessionHit(ctx context.Context, groupID *int64, sessionHash, requestedModel string, excludedIDs map[int64]struct{}, stickyAccountID int64) *Account {
	account, _ := s.resolveOpenAIStickySessionAccount(
		ctx,
		groupID,
		sessionHash,
		requestedModel,
		excludedIDs,
		stickyAccountID,
		openAIStickySessionResolvePolicy{
			deleteOnLookupMiss: false,
			refreshTTLOnHit:    true,
			recheckOnResolve:   true,
		},
	)
	return account
}

func (s *OpenAIGatewayService) selectBestAccount(ctx context.Context, accounts []Account, requestedModel string, excludedIDs map[int64]struct{}) *Account {
	candidates := make([]*Account, 0, len(accounts))

	for i := range accounts {
		account := &accounts[i]
		if isOpenAIAccountExcluded(excludedIDs, account.ID) {
			continue
		}

		fresh := s.resolveFreshSchedulableOpenAIAccount(ctx, account, requestedModel)
		if fresh == nil {
			continue
		}
		fresh = s.recheckSelectedOpenAIAccountFromDB(ctx, fresh, requestedModel)
		if fresh == nil {
			continue
		}

		candidates = append(candidates, fresh)
	}

	return selectBestByPriorityAndLastUsed(candidates, nil)
}

// SelectAccountWithLoadAwareness selects an account with load-awareness and wait plan.
func (s *OpenAIGatewayService) SelectAccountWithLoadAwareness(ctx context.Context, groupID *int64, sessionHash string, requestedModel string, excludedIDs map[int64]struct{}) (*AccountSelectionResult, error) {
	cfg := s.schedulingConfig()
	stickyTTL := s.openAIWSSessionStickyTTL()
	stickyAccountID := s.lookupOpenAIStickyAccountID(ctx, groupID, sessionHash)

	if !shouldUseOpenAILoadAwareSelection(s.concurrencyService, cfg) {
		return s.selectOpenAIAccountWithoutLoadBatch(
			ctx,
			groupID,
			sessionHash,
			requestedModel,
			excludedIDs,
			stickyAccountID,
			cfg,
		)
	}

	accounts, err := s.listSchedulableAccounts(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if len(accounts) == 0 {
		return nil, ErrNoAvailableAccounts
	}

	if result, ok := s.trySelectOpenAIStickyLoadAwareAccount(
		ctx,
		groupID,
		sessionHash,
		requestedModel,
		excludedIDs,
		stickyAccountID,
		stickyTTL,
		cfg,
	); ok {
		return result, nil
	}

	candidates, err := s.selectOpenAILoadAwareCandidates(accounts, requestedModel, excludedIDs)
	if err != nil {
		return nil, err
	}

	loadMap, err := s.concurrencyService.GetAccountsLoadBatch(ctx, buildAccountLoadRequests(candidates))
	if err != nil {
		if result, ok := s.trySelectOpenAILegacyFallbackAccount(
			ctx,
			candidates,
			groupID,
			sessionHash,
			stickyTTL,
			requestedModel,
		); ok {
			return result, nil
		}
	} else {
		if result, ok := s.trySelectOpenAILoadAwareAvailableAccount(
			ctx,
			candidates,
			loadMap,
			groupID,
			sessionHash,
			stickyTTL,
			requestedModel,
		); ok {
			return result, nil
		}
	}

	if fresh := s.selectOpenAIWaitPlanCandidate(ctx, candidates, loadMap, requestedModel); fresh != nil {
		return newWaitPlanAccountSelection(fresh, cfg.FallbackWaitTimeout, cfg.FallbackMaxWaiting), nil
	}

	return nil, ErrNoAvailableAccounts
}

func (s *OpenAIGatewayService) listSchedulableAccounts(ctx context.Context, groupID *int64) ([]Account, error) {
	if s.schedulerSnapshot != nil {
		accounts, _, err := s.schedulerSnapshot.ListSchedulableAccounts(ctx, groupID, PlatformOpenAI, false)
		return accounts, err
	}

	var (
		accounts []Account
		err      error
	)
	if s.cfg != nil && s.cfg.RunMode == config.RunModeSimple {
		accounts, err = s.accountRepo.ListSchedulableByPlatform(ctx, PlatformOpenAI)
	} else if groupID != nil {
		accounts, err = s.accountRepo.ListSchedulableByGroupIDAndPlatform(ctx, *groupID, PlatformOpenAI)
	} else {
		accounts, err = s.accountRepo.ListSchedulableUngroupedByPlatform(ctx, PlatformOpenAI)
	}
	if err != nil {
		return nil, fmt.Errorf("query accounts failed: %w", err)
	}
	return accounts, nil
}

func (s *OpenAIGatewayService) tryAcquireAccountSlot(ctx context.Context, accountID int64, maxConcurrency int) (*AcquireResult, error) {
	if s.concurrencyService == nil {
		return &AcquireResult{Acquired: true, ReleaseFunc: func() {}}, nil
	}
	return s.concurrencyService.AcquireAccountSlot(ctx, accountID, maxConcurrency)
}

func (s *OpenAIGatewayService) resolveFreshSchedulableOpenAIAccount(ctx context.Context, account *Account, requestedModel string) *Account {
	if account == nil {
		return nil
	}

	fresh := account
	if s.schedulerSnapshot != nil {
		current, err := s.getSchedulableAccount(ctx, account.ID)
		if err != nil || current == nil {
			return nil
		}
		fresh = current
	}
	if !fresh.IsSchedulable() || !fresh.IsOpenAI() {
		return nil
	}
	if requestedModel != "" && !fresh.IsModelSupported(requestedModel) {
		return nil
	}
	return fresh
}

func (s *OpenAIGatewayService) recheckSelectedOpenAIAccountFromDB(ctx context.Context, account *Account, requestedModel string) *Account {
	if account == nil {
		return nil
	}
	if s.schedulerSnapshot == nil || s.accountRepo == nil {
		return account
	}

	latest, err := s.accountRepo.GetByID(ctx, account.ID)
	if err != nil || latest == nil {
		return nil
	}
	syncOpenAICodexRateLimitFromExtra(ctx, s.accountRepo, latest, time.Now())
	if !latest.IsSchedulable() || !latest.IsOpenAI() {
		return nil
	}
	if requestedModel != "" && !latest.IsModelSupported(requestedModel) {
		return nil
	}
	return latest
}

func (s *OpenAIGatewayService) getSchedulableAccount(ctx context.Context, accountID int64) (*Account, error) {
	var (
		account *Account
		err     error
	)
	if s.schedulerSnapshot != nil {
		account, err = s.schedulerSnapshot.GetAccount(ctx, accountID)
	} else {
		account, err = s.accountRepo.GetByID(ctx, accountID)
	}
	if err != nil || account == nil {
		return account, err
	}
	syncOpenAICodexRateLimitFromExtra(ctx, s.accountRepo, account, time.Now())
	return account, nil
}

func (s *OpenAIGatewayService) schedulingConfig() config.GatewaySchedulingConfig {
	if s.cfg != nil {
		return s.cfg.Gateway.Scheduling
	}
	return config.GatewaySchedulingConfig{
		StickySessionMaxWaiting:  3,
		StickySessionWaitTimeout: 45 * time.Second,
		FallbackWaitTimeout:      30 * time.Second,
		FallbackMaxWaiting:       100,
		LoadBatchEnabled:         true,
		SlotCleanupInterval:      30 * time.Second,
	}
}
