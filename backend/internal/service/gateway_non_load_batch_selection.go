package service

import (
	"context"

	"github.com/senran-N/sub2api/internal/config"
)

func cloneExcludedAccountIDs(excludedIDs map[int64]struct{}) map[int64]struct{} {
	cloned := make(map[int64]struct{}, len(excludedIDs))
	for accountID := range excludedIDs {
		cloned[accountID] = struct{}{}
	}
	return cloned
}

func (s *GatewayService) selectAccountWithoutLoadBatch(
	scope *gatewaySelectionScope,
	plan *gatewaySelectionPlan,
	sessionHash string,
	requestedModel string,
	excludedIDs map[int64]struct{},
	cfg config.GatewaySchedulingConfig,
) (*AccountSelectionResult, error) {
	if scope == nil || plan == nil {
		return nil, ErrNoAvailableAccounts
	}

	localExcluded := cloneExcludedAccountIDs(excludedIDs)
	ctx := scope.ctx
	groupID := scope.groupID
	stickyAccountID := scope.stickyAccountID

	for {
		account, err := s.selectLoadlessAccountAttempt(scope, plan, sessionHash, requestedModel, localExcluded)
		if err != nil {
			return nil, err
		}

		result, _, missReason, ok := tryAcquireRuntimeSelectionDetailed(ctx, s.buildGatewayRuntimeAcquireSpec(
			ctx,
			account,
			sessionHash,
			nil,
			s.buildGatewayStickyBindAdapter(ctx, groupID, sessionHash, true),
		))
		if ok {
			return result, nil
		}
		if missReason == runtimeAcquireMissSession {
			s.clearLoadlessRejectedStickyBinding(ctx, groupID, sessionHash)
			localExcluded[account.ID] = struct{}{}
			continue
		}

		if !s.checkAndRegisterSession(ctx, account, sessionHash) {
			s.clearLoadlessRejectedStickyBinding(ctx, groupID, sessionHash)
			localExcluded[account.ID] = struct{}{}
			continue
		}

		if bind := s.buildGatewayStickyBindAdapter(ctx, groupID, sessionHash, true); bind != nil {
			bind(account)
		}
		return buildStickyAwareFallbackWaitPlan(ctx, account, stickyAccountID, cfg, s.concurrencyService), nil
	}
}

func (s *GatewayService) selectLoadlessAccountAttempt(
	scope *gatewaySelectionScope,
	plan *gatewaySelectionPlan,
	sessionHash string,
	requestedModel string,
	excludedIDs map[int64]struct{},
) (*Account, error) {
	return s.selectAccountWithLegacyScheduling(s.buildLegacySelectionInput(scope, sessionHash, requestedModel, excludedIDs, plan, false))
}

func (s *GatewayService) clearLoadlessRejectedStickyBinding(ctx context.Context, groupID *int64, sessionHash string) {
	if sessionHash == "" || s.cache == nil {
		return
	}
	_ = s.cache.DeleteSessionAccountID(ctx, derefGroupID(groupID), sessionHash)
}
