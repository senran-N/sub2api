package service

import "context"

func (s *GatewayService) filterLoadAwareCandidates(
	ctx context.Context,
	accounts []Account,
	requestedModel string,
	platform string,
	useMixed bool,
	excludedIDs map[int64]struct{},
) []*Account {
	candidates := make([]*Account, 0, len(accounts))
	for index := range accounts {
		account := &accounts[index]
		if _, excluded := excludedIDs[account.ID]; excluded {
			continue
		}
		if !s.isAccountSchedulableForSelection(account) {
			continue
		}
		if !s.isAccountAllowedForPlatform(account, platform, useMixed) {
			continue
		}
		if requestedModel != "" && !s.isModelSupportedByAccountWithContext(ctx, account, requestedModel) {
			continue
		}
		if s.isChannelModelRestrictedForSelection(ctx, account, requestedModel) {
			continue
		}
		if !s.isAccountSchedulableForModelSelection(ctx, account, requestedModel) {
			continue
		}
		if !s.isAccountSchedulableForQuota(account) {
			continue
		}
		if !s.isAccountSchedulableForWindowCost(ctx, account, false) {
			continue
		}
		if !s.isAccountSchedulableForRPM(ctx, account, false) {
			continue
		}
		candidates = append(candidates, account)
	}
	return candidates
}
