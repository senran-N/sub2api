package service

import (
	"context"
	"errors"
	"fmt"

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
	if account := s.tryStickySessionHit(ctx, groupID, sessionHash, requestedModel, excludedIDs, stickyAccountID, OpenAIUpstreamTransportAny); account != nil {
		return account, nil
	}

	if s.schedulerSnapshot != nil {
		selected, supported, err := s.selectBestAccountFromIndexedSnapshot(ctx, groupID, requestedModel, excludedIDs)
		if err != nil {
			return nil, fmt.Errorf("query accounts failed: %w", err)
		}
		if selected == nil {
			if requestedModel != "" {
				if !supported {
					return nil, newOpenAIRequestedModelUnavailableError(requestedModel)
				}
				return nil, ErrNoAvailableAccounts
			}
			return nil, errors.New("no available OpenAI accounts")
		}

		_ = s.BindStickySessionIfUnbound(ctx, groupID, sessionHash, selected.ID)
		return selected, nil
	}

	accounts, err := s.listSchedulableAccounts(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("query accounts failed: %w", err)
	}

	selected := s.selectBestAccount(ctx, accounts, requestedModel, excludedIDs)
	if selected == nil {
		if requestedModel != "" {
			if !openAIRequestedModelAvailable(accounts, requestedModel) {
				return nil, newOpenAIRequestedModelUnavailableError(requestedModel)
			}
			return nil, ErrNoAvailableAccounts
		}
		return nil, errors.New("no available OpenAI accounts")
	}

	_ = s.BindStickySessionIfUnbound(ctx, groupID, sessionHash, selected.ID)

	return selected, nil
}

// tryStickySessionHit tries to resolve a sticky session to a currently usable account.
func (s *OpenAIGatewayService) tryStickySessionHit(ctx context.Context, groupID *int64, sessionHash, requestedModel string, excludedIDs map[int64]struct{}, stickyAccountID int64, requiredTransport OpenAIUpstreamTransport) *Account {
	if stickyAccountID <= 0 {
		stickyAccountID = s.resolveCodexChainState(ctx, codexChainStateInput{
			GroupID:     derefGroupID(groupID),
			SessionHash: sessionHash,
			Transport:   requiredTransport,
		}).SessionStickyAccount
	}
	if s.isOpenAITransportFallbackCooling(stickyAccountID, requiredTransport) {
		recordOpenAIStickyBindingDisposition(ctx, stickyBindingKindSession, newStickyBindingSoftMiss("transport_cooling"), stickyAccountID, sessionHash, "")
		return nil
	}
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
	cfg := gatewaySchedulingConfigOrDefault(s.cfg)
	stickyTTL := s.openAIWSSessionStickyTTL()
	stickyAccountID := s.resolveCodexChainState(ctx, codexChainStateInput{
		GroupID:     derefGroupID(groupID),
		SessionHash: sessionHash,
		Transport:   OpenAIUpstreamTransportAny,
	}).SessionStickyAccount

	if s.concurrencyService == nil || !cfg.LoadBatchEnabled {
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

	if s.schedulerSnapshot != nil {
		result, supported, err := s.selectOpenAIAccountWithLoadAwarenessFromIndexedSnapshot(
			ctx,
			groupID,
			sessionHash,
			requestedModel,
			excludedIDs,
			cfg,
		)
		if err != nil {
			return nil, err
		}
		if result != nil {
			return result, nil
		}
		if requestedModel != "" && !supported {
			return nil, newOpenAIRequestedModelUnavailableError(requestedModel)
		}
		return nil, ErrNoAvailableAccounts
	}

	accounts, err := s.listSchedulableAccounts(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if len(accounts) == 0 {
		return nil, ErrNoAvailableAccounts
	}

	candidates := filterSchedulableOpenAICandidates(accounts, requestedModel, excludedIDs)
	if len(candidates) == 0 {
		if requestedModel != "" && !openAIRequestedModelAvailable(accounts, requestedModel) {
			return nil, newOpenAIRequestedModelUnavailableError(requestedModel)
		}
		return nil, ErrNoAvailableAccounts
	}

	loadMap, err := s.concurrencyService.GetAccountsLoadBatch(ctx, buildAccountLoadRequests(candidates))
	if err != nil {
		if result, ok := s.trySelectOpenAILegacyFallbackAccount(
			ctx,
			candidates,
			groupID,
			sessionHash,
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
			requestedModel,
		); ok {
			return result, nil
		}
	}

	if waitCandidate := selectBestOpenAIWaitCandidate(candidates, loadMap); waitCandidate != nil {
		if waitPlan, ok := s.tryBuildOpenAIWaitPlanSelection(ctx, waitCandidate, requestedModel, OpenAIUpstreamTransportAny, cfg); ok {
			return waitPlan, nil
		}
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
	if !isOpenAIAccountRuntimeEligible(fresh, requestedModel) {
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
	if !isOpenAIAccountRuntimeEligible(latest, requestedModel) {
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
	return account, nil
}
