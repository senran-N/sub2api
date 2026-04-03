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
	ctx context.Context,
	groupID *int64,
	sessionHash string,
	requestedModel string,
	excludedIDs map[int64]struct{},
	stickyAccountID int64,
	cfg config.GatewaySchedulingConfig,
) (*AccountSelectionResult, error) {
	localExcluded := cloneExcludedAccountIDs(excludedIDs)

	for {
		account, err := s.SelectAccountForModelWithExclusions(ctx, groupID, sessionHash, requestedModel, localExcluded)
		if err != nil {
			return nil, err
		}

		result, err := s.tryAcquireAccountSlot(ctx, account.ID, account.Concurrency)
		if err == nil && result.Acquired {
			if !s.checkAndRegisterSession(ctx, account, sessionHash) {
				result.ReleaseFunc()
				localExcluded[account.ID] = struct{}{}
				continue
			}
			return newAcquiredAccountSelection(account, result.ReleaseFunc), nil
		}

		if !s.checkAndRegisterSession(ctx, account, sessionHash) {
			localExcluded[account.ID] = struct{}{}
			continue
		}

		return s.buildNonLoadBatchWaitPlan(ctx, account, stickyAccountID, cfg), nil
	}
}

func (s *GatewayService) buildNonLoadBatchWaitPlan(
	ctx context.Context,
	account *Account,
	stickyAccountID int64,
	cfg config.GatewaySchedulingConfig,
) *AccountSelectionResult {
	if stickyAccountID > 0 && stickyAccountID == account.ID && s.concurrencyService != nil {
		waitingCount, _ := s.concurrencyService.GetAccountWaitingCount(ctx, account.ID)
		if waitingCount < cfg.StickySessionMaxWaiting {
			return newWaitPlanAccountSelection(account, cfg.StickySessionWaitTimeout, cfg.StickySessionMaxWaiting)
		}
	}

	return newWaitPlanAccountSelection(account, cfg.FallbackWaitTimeout, cfg.FallbackMaxWaiting)
}
