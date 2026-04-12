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
	result := defaultGatewaySchedulingConfig()
	if cfg == nil {
		return result
	}

	if cfg.Gateway.Scheduling.StickySessionMaxWaiting > 0 {
		result.StickySessionMaxWaiting = cfg.Gateway.Scheduling.StickySessionMaxWaiting
	}
	if cfg.Gateway.Scheduling.StickySessionWaitTimeout > 0 {
		result.StickySessionWaitTimeout = cfg.Gateway.Scheduling.StickySessionWaitTimeout
	}
	if cfg.Gateway.Scheduling.FallbackWaitTimeout > 0 {
		result.FallbackWaitTimeout = cfg.Gateway.Scheduling.FallbackWaitTimeout
	}
	if cfg.Gateway.Scheduling.FallbackMaxWaiting > 0 {
		result.FallbackMaxWaiting = cfg.Gateway.Scheduling.FallbackMaxWaiting
	}
	if cfg.Gateway.Scheduling.SlotCleanupInterval > 0 {
		result.SlotCleanupInterval = cfg.Gateway.Scheduling.SlotCleanupInterval
	}
	result.LoadBatchEnabled = cfg.Gateway.Scheduling.LoadBatchEnabled
	if cfg.Gateway.Scheduling.FallbackSelectionMode != "" {
		result.FallbackSelectionMode = cfg.Gateway.Scheduling.FallbackSelectionMode
	}
	if cfg.Gateway.Scheduling.SnapshotPageSize > 0 {
		result.SnapshotPageSize = cfg.Gateway.Scheduling.SnapshotPageSize
	}
	return result
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
