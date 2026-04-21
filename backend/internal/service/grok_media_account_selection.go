package service

import (
	"context"
	"errors"
	"fmt"
)

type grokAccountSelectionFilter func(*Account) bool

func selectSchedulableGrokMediaAccount(
	ctx context.Context,
	gatewayService *GatewayService,
	groupID *int64,
	requestedModel string,
	excludedIDs map[int64]struct{},
	filter grokAccountSelectionFilter,
	noAccountsErr string,
) (*Account, error) {
	if gatewayService == nil {
		return nil, errors.New("grok media service is not configured")
	}
	ctx = WithGrokSessionMediaRuntimeAllowed(ctx)

	accounts, _, err := gatewayService.listSchedulableAccounts(ctx, groupID, PlatformGrok, true)
	if err != nil {
		return nil, err
	}

	candidates := defaultGrokAccountSelector.FilterSchedulableCandidatesWithContext(ctx, accounts, requestedModel, excludedIDs)
	if len(candidates) > 0 && filter != nil {
		filtered := make([]*Account, 0, len(candidates))
		for i := range candidates {
			if !filter(candidates[i]) {
				continue
			}
			filtered = append(filtered, candidates[i])
		}
		candidates = filtered
	}
	if len(candidates) == 0 {
		if !defaultGrokAccountSelector.RequestedModelAvailableWithContext(ctx, accounts, requestedModel) {
			return nil, fmt.Errorf("requested model unavailable:%s", requestedModel)
		}
		return nil, errors.New(firstNonEmpty(noAccountsErr, "no compatible grok media accounts"))
	}

	var loadMap map[int64]*AccountLoadInfo
	if gatewayService.concurrencyService != nil {
		if snapshot, loadErr := gatewayService.concurrencyService.GetAccountsLoadBatch(ctx, buildAccountLoadRequests(candidates)); loadErr == nil {
			loadMap = snapshot
		}
	}

	selected := defaultGrokAccountSelector.SelectBestCandidateWithContext(ctx, candidates, requestedModel, loadMap)
	if selected == nil {
		return nil, errors.New(firstNonEmpty(noAccountsErr, "no compatible grok media accounts"))
	}

	hydrated, err := gatewayService.hydrateSelectedAccount(ctx, selected)
	if err != nil {
		return nil, err
	}
	if hydrated == nil {
		return nil, errors.New(firstNonEmpty(noAccountsErr, "no compatible grok media accounts"))
	}
	if filter != nil && !filter(hydrated) {
		return nil, errors.New(firstNonEmpty(noAccountsErr, "no compatible grok media accounts"))
	}
	return hydrated, nil
}
