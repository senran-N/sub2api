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

// buildStickySessionWaitPlanIfConcurrencyEnabled preserves legacy sticky behavior:
// once concurrency service exists, always return sticky wait-plan without queue-limit check.
func buildStickySessionWaitPlanIfConcurrencyEnabled(
	account *Account,
	cfg config.GatewaySchedulingConfig,
	concurrencyService *ConcurrencyService,
) (*AccountSelectionResult, bool) {
	if account == nil || concurrencyService == nil {
		return nil, false
	}
	return newWaitPlanAccountSelection(account, cfg.StickySessionWaitTimeout, cfg.StickySessionMaxWaiting), true
}

func tryAcquireOrBuildStickyWaitPlan(
	ctx context.Context,
	account *Account,
	accountID int64,
	cfg config.GatewaySchedulingConfig,
	concurrencyService *ConcurrencyService,
	acquireFn func(context.Context, int64, int) (*AcquireResult, error),
	onAcquired func(*AcquireResult) *AccountSelectionResult,
) (*AccountSelectionResult, bool) {
	if account == nil || acquireFn == nil || onAcquired == nil {
		return nil, false
	}

	result, acquireErr := acquireFn(ctx, accountID, account.Concurrency)
	if acquireErr == nil && result.Acquired {
		return onAcquired(result), true
	}

	return buildStickySessionWaitPlanIfConcurrencyEnabled(account, cfg, concurrencyService)
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
