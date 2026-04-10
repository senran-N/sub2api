package service

import (
	"context"
	"time"

	"github.com/senran-N/sub2api/internal/config"
)

func defaultGatewaySchedulingConfig() config.GatewaySchedulingConfig {
	return config.GatewaySchedulingConfig{
		StickySessionMaxWaiting:  3,
		StickySessionWaitTimeout: 45 * time.Second,
		FallbackWaitTimeout:      30 * time.Second,
		FallbackMaxWaiting:       100,
		LoadBatchEnabled:         true,
		SlotCleanupInterval:      30 * time.Second,
	}
}

func gatewaySchedulingConfigOrDefault(cfg *config.Config) config.GatewaySchedulingConfig {
	if cfg != nil {
		return cfg.Gateway.Scheduling
	}
	return defaultGatewaySchedulingConfig()
}

func snapshotPageSizeOrDefault(cfg *config.Config) int {
	if cfg != nil && cfg.Gateway.Scheduling.SnapshotPageSize > 0 {
		return cfg.Gateway.Scheduling.SnapshotPageSize
	}
	return 128
}

func acquireAccountSlotWithConcurrencyService(
	ctx context.Context,
	concurrencyService *ConcurrencyService,
	accountID int64,
	maxConcurrency int,
) (*AcquireResult, error) {
	if concurrencyService == nil {
		return &AcquireResult{Acquired: true, ReleaseFunc: func() {}}, nil
	}
	return concurrencyService.AcquireAccountSlot(ctx, accountID, maxConcurrency)
}
