package service

import (
	"context"
	"strings"
)

func resolveCompatibleSelectionPlatform(ctx context.Context, fallback string) string {
	return ResolveCompatibleGatewayPlatform(ctx, fallback)
}

func isCompatibleSelectionPlatformAccount(account *Account, platform string) bool {
	if account == nil {
		return false
	}
	return NormalizeCompatibleGatewayPlatform(account.Platform) == ResolveCompatibleGatewayPlatform(context.TODO(), platform)
}

func isCompatibleAccountBaseEligibleForPlatform(account *Account, platform string) bool {
	return isCompatibleSelectionPlatformAccount(account, platform) && account.IsSchedulable()
}

func isCompatibleAccountModelEligible(account *Account, requestedModel string, platform string) bool {
	switch ResolveCompatibleGatewayPlatform(context.TODO(), platform) {
	case PlatformGrok:
		return defaultGrokAccountSelector.IsModelEligible(account, requestedModel)
	default:
		return isOpenAIAccountModelEligible(account, requestedModel)
	}
}

func isCompatibleAccountRuntimeEligibleForPlatformWithContext(
	ctx context.Context,
	account *Account,
	requestedModel string,
	platform string,
) bool {
	switch ResolveCompatibleGatewayPlatform(context.TODO(), platform) {
	case PlatformGrok:
		return defaultGrokAccountSelector.IsRuntimeEligibleWithContext(ctx, account, requestedModel)
	}
	if account == nil || !isCompatibleSelectionPlatformAccount(account, platform) {
		return false
	}
	if !account.IsSchedulable() {
		return false
	}
	if oauthSelectionCredentialIssue(account) != "" {
		return false
	}
	if requestedModel != "" && !isCompatibleAccountModelEligible(account, requestedModel, platform) {
		return false
	}
	return true
}

func filterSchedulableCompatibleCandidatesForPlatformWithContext(
	ctx context.Context,
	accounts []Account,
	requestedModel string,
	excludedIDs map[int64]struct{},
	platform string,
) []*Account {
	if ResolveCompatibleGatewayPlatform(context.TODO(), platform) == PlatformGrok {
		return defaultGrokAccountSelector.FilterSchedulableCandidatesWithContext(ctx, accounts, requestedModel, excludedIDs)
	}

	candidates := make([]*Account, 0, len(accounts))
	for i := range accounts {
		account := &accounts[i]
		if isOpenAIAccountExcluded(excludedIDs, account.ID) {
			continue
		}
		if !isCompatibleAccountRuntimeEligibleForPlatformWithContext(ctx, account, requestedModel, platform) {
			continue
		}
		candidates = append(candidates, account)
	}
	return candidates
}

func filterSchedulableCompatibleAccountPointersForPlatformWithContext(
	ctx context.Context,
	accounts []*Account,
	requestedModel string,
	excludedIDs map[int64]struct{},
	platform string,
) []*Account {
	if ResolveCompatibleGatewayPlatform(context.TODO(), platform) == PlatformGrok {
		return defaultGrokAccountSelector.FilterSchedulableAccountPointersWithContext(ctx, accounts, requestedModel, excludedIDs)
	}

	candidates := make([]*Account, 0, len(accounts))
	for _, account := range accounts {
		if account == nil {
			continue
		}
		if isOpenAIAccountExcluded(excludedIDs, account.ID) {
			continue
		}
		if !isCompatibleAccountRuntimeEligibleForPlatformWithContext(ctx, account, requestedModel, platform) {
			continue
		}
		candidates = append(candidates, account)
	}
	return candidates
}

func compatibleRequestedModelAvailableForScheduling(
	ctx context.Context,
	account *Account,
	requestedModel string,
	platform string,
) bool {
	if strings.TrimSpace(requestedModel) == "" {
		return true
	}
	if ResolveCompatibleGatewayPlatform(context.TODO(), platform) == PlatformGrok {
		return defaultGrokAccountSelector.IsModelAvailableWithContext(ctx, account, requestedModel)
	}
	return isCompatibleAccountModelEligible(account, requestedModel, platform)
}

func compatibleRequestedModelAvailableForPlatformWithContext(
	ctx context.Context,
	accounts []Account,
	requestedModel string,
	platform string,
) bool {
	model := strings.TrimSpace(requestedModel)
	if model == "" {
		return true
	}
	if ResolveCompatibleGatewayPlatform(context.TODO(), platform) == PlatformGrok {
		return defaultGrokAccountSelector.RequestedModelAvailableWithContext(ctx, accounts, model)
	}
	for i := range accounts {
		account := &accounts[i]
		if !isCompatibleSelectionPlatformAccount(account, platform) {
			continue
		}
		if isCompatibleAccountModelEligible(account, model, platform) {
			return true
		}
	}
	return false
}
