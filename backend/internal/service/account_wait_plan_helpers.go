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
	if concurrencyService == nil {
		return nil, false
	}
	result, _, ok := tryBuildRuntimeWaitPlan(ctx, runtimeWaitPlanSpec{
		account: account,
		allowWait: func(account *Account) (string, bool) {
			waitingCount, _ := concurrencyService.GetAccountWaitingCount(ctx, account.ID)
			if waitingCount >= cfg.StickySessionMaxWaiting {
				return "wait_queue_full", false
			}
			return "", true
		},
		timeout:    cfg.StickySessionWaitTimeout,
		maxWaiting: cfg.StickySessionMaxWaiting,
	})
	return result, ok
}

// buildStickySessionWaitPlanIfConcurrencyEnabled preserves legacy sticky behavior:
// once concurrency service exists, always return sticky wait-plan without queue-limit check.
func buildStickySessionWaitPlanIfConcurrencyEnabled(
	account *Account,
	cfg config.GatewaySchedulingConfig,
	concurrencyService *ConcurrencyService,
) (*AccountSelectionResult, bool) {
	if concurrencyService == nil {
		return nil, false
	}
	result, _, ok := tryBuildRuntimeWaitPlan(context.Background(), runtimeWaitPlanSpec{
		account:    account,
		timeout:    cfg.StickySessionWaitTimeout,
		maxWaiting: cfg.StickySessionMaxWaiting,
	})
	return result, ok
}

func tryAcquireOrBuildStickyWaitPlan(
	ctx context.Context,
	account *Account,
	accountID int64,
	cfg config.GatewaySchedulingConfig,
	concurrencyService *ConcurrencyService,
	acquireFn func(context.Context, int64, int) (*AcquireResult, error),
	onAcquired func(account *Account, acquired *AcquireResult) *AccountSelectionResult,
) (*AccountSelectionResult, bool) {
	if account == nil || acquireFn == nil || onAcquired == nil {
		return nil, false
	}

	result, _, ok := trySelectStickyRuntimeSelection(stickyRuntimeSelectionSpec{
		tryAcquire: func() (*AccountSelectionResult, string, bool) {
			result, acquireErr, missReason, ok := tryAcquireRuntimeSelectionDetailed(ctx, runtimeAcquireSelectionSpec{
				account: account,
				acquire: func(account *Account) (*AcquireResult, error) {
					return acquireFn(ctx, accountID, account.Concurrency)
				},
				onAcquired: onAcquired,
			})
			if acquireErr != nil {
				return nil, missReason, false
			}
			return result, missReason, ok
		},
		buildWaitPlan: func() (*AccountSelectionResult, string, bool) {
			result, ok := buildStickySessionWaitPlanIfConcurrencyEnabled(account, cfg, concurrencyService)
			return result, "", ok
		},
	})
	return result, ok
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
