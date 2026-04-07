package service

import (
	"context"

	"github.com/senran-N/sub2api/internal/config"
)

func tryBuildStickySessionWaitPlan(
	ctx context.Context,
	account *Account,
	cfg config.GatewaySchedulingConfig,
	concurrencyService *ConcurrencyService,
) (*AccountSelectionResult, bool) {
	if account == nil || concurrencyService == nil {
		return nil, false
	}

	waitingCount, _ := concurrencyService.GetAccountWaitingCount(ctx, account.ID)
	if waitingCount >= cfg.StickySessionMaxWaiting {
		return nil, false
	}

	return newWaitPlanAccountSelection(account, cfg.StickySessionWaitTimeout, cfg.StickySessionMaxWaiting), true
}

func buildStickyAwareFallbackWaitPlan(
	ctx context.Context,
	account *Account,
	stickyAccountID int64,
	cfg config.GatewaySchedulingConfig,
	concurrencyService *ConcurrencyService,
) *AccountSelectionResult {
	if account != nil && stickyAccountID > 0 && stickyAccountID == account.ID {
		if result, ok := tryBuildStickySessionWaitPlan(ctx, account, cfg, concurrencyService); ok {
			return result
		}
	}

	return newWaitPlanAccountSelection(account, cfg.FallbackWaitTimeout, cfg.FallbackMaxWaiting)
}
