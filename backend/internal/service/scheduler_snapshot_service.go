package service

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/senran-N/sub2api/internal/config"
	"github.com/senran-N/sub2api/internal/pkg/logger"
)

var (
	ErrSchedulerCacheNotReady   = errors.New("scheduler cache not ready")
	ErrSchedulerFallbackLimited = errors.New("scheduler db fallback limited")
)

const outboxEventTimeout = 2 * time.Minute

type SchedulerSnapshotService struct {
	cache         SchedulerCache
	outboxRepo    SchedulerOutboxRepository
	accountRepo   AccountRepository
	groupRepo     GroupRepository
	cfg           *config.Config
	stopCh        chan struct{}
	stopOnce      sync.Once
	wg            sync.WaitGroup
	fallbackLimit *fallbackLimiter
	lagMu         sync.Mutex
	lagFailures   int
}

func NewSchedulerSnapshotService(
	cache SchedulerCache,
	outboxRepo SchedulerOutboxRepository,
	accountRepo AccountRepository,
	groupRepo GroupRepository,
	cfg *config.Config,
) *SchedulerSnapshotService {
	maxQPS := 0
	if cfg != nil {
		maxQPS = cfg.Gateway.Scheduling.DbFallbackMaxQPS
	}
	return &SchedulerSnapshotService{
		cache:         cache,
		outboxRepo:    outboxRepo,
		accountRepo:   accountRepo,
		groupRepo:     groupRepo,
		cfg:           cfg,
		stopCh:        make(chan struct{}),
		fallbackLimit: newFallbackLimiter(maxQPS),
	}
}

func (s *SchedulerSnapshotService) ListSchedulableAccounts(ctx context.Context, groupID *int64, platform string, hasForcePlatform bool) ([]Account, bool, error) {
	useMixed := (platform == PlatformAnthropic || platform == PlatformGemini) && !hasForcePlatform
	bucket, useMixed := s.resolveBucket(groupID, platform, hasForcePlatform)

	if s.cache != nil {
		cached, hit, err := s.cache.GetSnapshot(ctx, bucket)
		if err != nil {
			logger.LegacyPrintf("service.scheduler_snapshot", "[Scheduler] cache read failed: bucket=%s err=%v", bucket.String(), err)
		} else if hit {
			return derefAccounts(cached), useMixed, nil
		}
	}

	if err := s.guardFallback(ctx); err != nil {
		return nil, useMixed, err
	}

	fallbackCtx, cancel := s.withFallbackTimeout(ctx)
	defer cancel()

	accounts, err := s.loadAccountsFromDB(fallbackCtx, bucket, useMixed)
	if err != nil {
		return nil, useMixed, err
	}

	if s.cache != nil {
		if err := s.cache.SetSnapshot(fallbackCtx, bucket, accounts); err != nil {
			logger.LegacyPrintf("service.scheduler_snapshot", "[Scheduler] cache write failed: bucket=%s err=%v", bucket.String(), err)
		}
	}

	return accounts, useMixed, nil
}

func (s *SchedulerSnapshotService) ListSchedulableAccountsPage(
	ctx context.Context,
	groupID *int64,
	platform string,
	hasForcePlatform bool,
	offset int,
	limit int,
) ([]Account, bool, bool, error) {
	bucket, useMixed := s.resolveBucket(groupID, platform, hasForcePlatform)

	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = 1
	}

	if s.cache != nil {
		if pager, ok := s.cache.(SchedulerCachePager); ok {
			cached, hit, hasMore, err := pager.GetSnapshotPage(ctx, bucket, offset, limit)
			if err != nil {
				logger.LegacyPrintf("service.scheduler_snapshot", "[Scheduler] cache page read failed: bucket=%s offset=%d limit=%d err=%v", bucket.String(), offset, limit, err)
			} else if hit {
				return derefAccounts(cached), useMixed, hasMore, nil
			}
		} else {
			cached, hit, err := s.cache.GetSnapshot(ctx, bucket)
			if err != nil {
				logger.LegacyPrintf("service.scheduler_snapshot", "[Scheduler] cache read failed: bucket=%s err=%v", bucket.String(), err)
			} else if hit {
				page, hasMore := sliceAccountPage(derefAccounts(cached), offset, limit)
				return page, useMixed, hasMore, nil
			}
		}
	}

	if err := s.guardFallback(ctx); err != nil {
		return nil, useMixed, false, err
	}

	fallbackCtx, cancel := s.withFallbackTimeout(ctx)
	defer cancel()

	accounts, err := s.loadAccountsFromDB(fallbackCtx, bucket, useMixed)
	if err != nil {
		return nil, useMixed, false, err
	}

	if s.cache != nil {
		if err := s.cache.SetSnapshot(fallbackCtx, bucket, accounts); err != nil {
			logger.LegacyPrintf("service.scheduler_snapshot", "[Scheduler] cache write failed: bucket=%s err=%v", bucket.String(), err)
		}
	}

	page, hasMore := sliceAccountPage(accounts, offset, limit)
	return page, useMixed, hasMore, nil
}

func (s *SchedulerSnapshotService) ListSchedulableAccountsByCapabilityPage(
	ctx context.Context,
	groupID *int64,
	platform string,
	hasForcePlatform bool,
	index SchedulerCapabilityIndex,
	offset int,
	limit int,
) ([]Account, bool, bool, error) {
	bucket, useMixed := s.resolveBucket(groupID, platform, hasForcePlatform)
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = 1
	}

	if s.cache != nil {
		if indexed, ok := s.cache.(SchedulerCacheIndexed); ok {
			accounts, hit, hasMore, err := indexed.GetCapabilityIndexPage(ctx, bucket, index, offset, limit)
			if err != nil {
				logger.LegacyPrintf("service.scheduler_snapshot", "[Scheduler] cache index page read failed: bucket=%s index=%s value=%s offset=%d limit=%d err=%v", bucket.String(), index.Kind, index.Value, offset, limit, err)
			} else if hit {
				return derefAccounts(accounts), useMixed, hasMore, nil
			}
		}
	}

	accounts, _, err := s.ListSchedulableAccounts(ctx, groupID, platform, hasForcePlatform)
	if err != nil {
		return nil, useMixed, false, err
	}
	filtered := filterAccountsByCapabilityIndex(accounts, index)
	page, hasMore := sliceAccountPage(filtered, offset, limit)
	return page, useMixed, hasMore, nil
}

func (s *SchedulerSnapshotService) listSchedulableAccountPointersByCapabilityPage(
	ctx context.Context,
	groupID *int64,
	platform string,
	hasForcePlatform bool,
	index SchedulerCapabilityIndex,
	offset int,
	limit int,
) ([]*Account, bool, bool, error) {
	bucket, useMixed := s.resolveBucket(groupID, platform, hasForcePlatform)
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = 1
	}

	if s.cache != nil {
		if indexed, ok := s.cache.(SchedulerCacheIndexed); ok {
			accounts, hit, hasMore, err := indexed.GetCapabilityIndexPage(ctx, bucket, index, offset, limit)
			if err != nil {
				logger.LegacyPrintf("service.scheduler_snapshot", "[Scheduler] cache index page read failed: bucket=%s index=%s value=%s offset=%d limit=%d err=%v", bucket.String(), index.Kind, index.Value, offset, limit, err)
			} else if hit {
				return accounts, useMixed, hasMore, nil
			}
		}
	}

	page, _, hasMore, err := s.ListSchedulableAccountsByCapabilityPage(ctx, groupID, platform, hasForcePlatform, index, offset, limit)
	if err != nil {
		return nil, useMixed, false, err
	}
	return refAccounts(page), useMixed, hasMore, nil
}

func (s *SchedulerSnapshotService) MatchSchedulableAccountsCapability(
	ctx context.Context,
	groupID *int64,
	platform string,
	hasForcePlatform bool,
	index SchedulerCapabilityIndex,
	accountIDs []int64,
) (map[int64]bool, bool, error) {
	bucket, _ := s.resolveBucket(groupID, platform, hasForcePlatform)
	if len(accountIDs) == 0 {
		return map[int64]bool{}, true, nil
	}

	if s.cache != nil {
		if indexed, ok := s.cache.(SchedulerCacheIndexed); ok {
			matches, hit, err := indexed.HasCapabilityIndexMembers(ctx, bucket, index, accountIDs)
			if err != nil {
				logger.LegacyPrintf("service.scheduler_snapshot", "[Scheduler] cache index membership failed: bucket=%s index=%s value=%s err=%v", bucket.String(), index.Kind, index.Value, err)
			} else if hit {
				return matches, true, nil
			}
		}
	}

	accounts, _, err := s.ListSchedulableAccounts(ctx, groupID, platform, hasForcePlatform)
	if err != nil {
		return nil, false, err
	}
	allowed := filterAccountsByCapabilityIndex(accounts, index)
	matches := make(map[int64]bool, len(accountIDs))
	allowedSet := make(map[int64]struct{}, len(allowed))
	for i := range allowed {
		allowedSet[allowed[i].ID] = struct{}{}
	}
	for _, accountID := range accountIDs {
		if _, ok := allowedSet[accountID]; ok {
			matches[accountID] = true
		}
	}
	return matches, true, nil
}

func (s *SchedulerSnapshotService) ListSchedulableCapabilityIndexValues(
	ctx context.Context,
	groupID *int64,
	platform string,
	hasForcePlatform bool,
	kind SchedulerCapabilityIndexKind,
) ([]string, bool, error) {
	bucket, useMixed := s.resolveBucket(groupID, platform, hasForcePlatform)
	_ = useMixed

	if s.cache != nil {
		if indexed, ok := s.cache.(SchedulerCacheIndexed); ok {
			values, hit, err := indexed.ListCapabilityIndexValues(ctx, bucket, kind)
			if err != nil {
				logger.LegacyPrintf("service.scheduler_snapshot", "[Scheduler] cache index values failed: bucket=%s kind=%s err=%v", bucket.String(), kind, err)
			} else if hit {
				return values, true, nil
			}
		}
	}

	accounts, _, err := s.ListSchedulableAccounts(ctx, groupID, platform, hasForcePlatform)
	if err != nil {
		return nil, false, err
	}
	return collectCapabilityIndexValues(accounts, kind), true, nil
}

func (s *SchedulerSnapshotService) GetAccount(ctx context.Context, accountID int64) (*Account, error) {
	if accountID <= 0 {
		return nil, nil
	}
	if s.cache != nil {
		account, err := s.cache.GetAccount(ctx, accountID)
		if err != nil {
			logger.LegacyPrintf("service.scheduler_snapshot", "[Scheduler] account cache read failed: id=%d err=%v", accountID, err)
		} else if account != nil {
			return account, nil
		}
	}

	if err := s.guardFallback(ctx); err != nil {
		return nil, err
	}
	fallbackCtx, cancel := s.withFallbackTimeout(ctx)
	defer cancel()
	return s.accountRepo.GetByID(fallbackCtx, accountID)
}

// GetGroupByID 获取分组信息（供调度器使用）
func (s *SchedulerSnapshotService) GetGroupByID(ctx context.Context, groupID int64) (*Group, error) {
	if s.groupRepo == nil {
		return nil, nil
	}
	return s.groupRepo.GetByID(ctx, groupID)
}

// UpdateAccountInCache 立即更新 Redis 中单个账号的数据（用于模型限流后立即生效）
func (s *SchedulerSnapshotService) UpdateAccountInCache(ctx context.Context, account *Account) error {
	if s.cache == nil || account == nil {
		return nil
	}
	return s.cache.SetAccount(ctx, account)
}
