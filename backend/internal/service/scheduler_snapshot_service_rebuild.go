package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/senran-N/sub2api/internal/pkg/logger"
)

func (s *SchedulerSnapshotService) rebuildByAccount(ctx context.Context, account *Account, groupIDs []int64, reason string) error {
	if account == nil {
		return nil
	}
	groupIDs = s.normalizeGroupIDs(groupIDs)
	if len(groupIDs) == 0 {
		return nil
	}

	var firstErr error
	if err := s.rebuildBucketsForPlatform(ctx, account.Platform, groupIDs, reason); err != nil && firstErr == nil {
		firstErr = err
	}
	if account.Platform == PlatformAntigravity && account.IsMixedSchedulingEnabled() {
		if err := s.rebuildBucketsForPlatform(ctx, PlatformAnthropic, groupIDs, reason); err != nil && firstErr == nil {
			firstErr = err
		}
		if err := s.rebuildBucketsForPlatform(ctx, PlatformGemini, groupIDs, reason); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (s *SchedulerSnapshotService) rebuildByGroupIDs(ctx context.Context, groupIDs []int64, reason string) error {
	groupIDs = s.normalizeGroupIDs(groupIDs)
	if len(groupIDs) == 0 {
		return nil
	}
	platforms := []string{PlatformAnthropic, PlatformGemini, PlatformOpenAI, PlatformAntigravity}
	var firstErr error
	for _, platform := range platforms {
		if err := s.rebuildBucketsForPlatform(ctx, platform, groupIDs, reason); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (s *SchedulerSnapshotService) rebuildBucketsForPlatform(ctx context.Context, platform string, groupIDs []int64, reason string) error {
	if platform == "" {
		return nil
	}
	var firstErr error
	for _, gid := range groupIDs {
		if err := s.rebuildBucket(ctx, SchedulerBucket{GroupID: gid, Platform: platform, Mode: SchedulerModeSingle}, reason); err != nil && firstErr == nil {
			firstErr = err
		}
		if err := s.rebuildBucket(ctx, SchedulerBucket{GroupID: gid, Platform: platform, Mode: SchedulerModeForced}, reason); err != nil && firstErr == nil {
			firstErr = err
		}
		if platform == PlatformAnthropic || platform == PlatformGemini {
			if err := s.rebuildBucket(ctx, SchedulerBucket{GroupID: gid, Platform: platform, Mode: SchedulerModeMixed}, reason); err != nil && firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

func (s *SchedulerSnapshotService) rebuildBuckets(ctx context.Context, buckets []SchedulerBucket, reason string) error {
	var firstErr error
	for _, bucket := range buckets {
		if err := s.rebuildBucket(ctx, bucket, reason); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (s *SchedulerSnapshotService) rebuildBucket(ctx context.Context, bucket SchedulerBucket, reason string) error {
	if s.cache == nil {
		return ErrSchedulerCacheNotReady
	}
	ok, err := s.cache.TryLockBucket(ctx, bucket, 30*time.Second)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	rebuildCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	accounts, err := s.loadAccountsFromDB(rebuildCtx, bucket, bucket.Mode == SchedulerModeMixed)
	if err != nil {
		logger.LegacyPrintf("service.scheduler_snapshot", "[Scheduler] rebuild failed: bucket=%s reason=%s err=%v", bucket.String(), reason, err)
		return err
	}
	if err := s.cache.SetSnapshot(rebuildCtx, bucket, accounts); err != nil {
		logger.LegacyPrintf("service.scheduler_snapshot", "[Scheduler] rebuild cache failed: bucket=%s reason=%s err=%v", bucket.String(), reason, err)
		return err
	}
	slog.Debug("[Scheduler] rebuild ok", "bucket", bucket.String(), "reason", reason, "size", len(accounts))
	return nil
}

func (s *SchedulerSnapshotService) triggerFullRebuild(reason string) error {
	if s.cache == nil {
		return ErrSchedulerCacheNotReady
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	buckets, err := s.cache.ListBuckets(ctx)
	if err != nil {
		logger.LegacyPrintf("service.scheduler_snapshot", "[Scheduler] list buckets failed: %v", err)
		return err
	}
	if len(buckets) == 0 {
		buckets, err = s.defaultBuckets(ctx)
		if err != nil {
			logger.LegacyPrintf("service.scheduler_snapshot", "[Scheduler] default buckets failed: %v", err)
			return err
		}
	}
	return s.rebuildBuckets(ctx, buckets, reason)
}

func (s *SchedulerSnapshotService) defaultBuckets(ctx context.Context) ([]SchedulerBucket, error) {
	buckets := make([]SchedulerBucket, 0)
	platforms := []string{PlatformAnthropic, PlatformGemini, PlatformOpenAI, PlatformAntigravity}
	for _, platform := range platforms {
		buckets = append(buckets, SchedulerBucket{GroupID: 0, Platform: platform, Mode: SchedulerModeSingle})
		buckets = append(buckets, SchedulerBucket{GroupID: 0, Platform: platform, Mode: SchedulerModeForced})
		if platform == PlatformAnthropic || platform == PlatformGemini {
			buckets = append(buckets, SchedulerBucket{GroupID: 0, Platform: platform, Mode: SchedulerModeMixed})
		}
	}

	if s.isRunModeSimple() || s.groupRepo == nil {
		return dedupeBuckets(buckets), nil
	}

	groups, err := s.groupRepo.ListActive(ctx)
	if err != nil {
		return dedupeBuckets(buckets), nil
	}
	for _, group := range groups {
		if group.Platform == "" {
			continue
		}
		buckets = append(buckets, SchedulerBucket{GroupID: group.ID, Platform: group.Platform, Mode: SchedulerModeSingle})
		buckets = append(buckets, SchedulerBucket{GroupID: group.ID, Platform: group.Platform, Mode: SchedulerModeForced})
		if group.Platform == PlatformAnthropic || group.Platform == PlatformGemini {
			buckets = append(buckets, SchedulerBucket{GroupID: group.ID, Platform: group.Platform, Mode: SchedulerModeMixed})
		}
	}
	return dedupeBuckets(buckets), nil
}

func dedupeBuckets(in []SchedulerBucket) []SchedulerBucket {
	seen := make(map[string]struct{}, len(in))
	out := make([]SchedulerBucket, 0, len(in))
	for _, bucket := range in {
		key := bucket.String()
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, bucket)
	}
	return out
}
