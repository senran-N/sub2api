package service

import (
	"context"
	"strings"
)

type schedulerIndexedAccountSource struct {
	index   SchedulerCapabilityIndex
	offset  int
	hasMore bool
}

type schedulerIndexedAccountPager struct {
	snapshot         *SchedulerSnapshotService
	groupID          *int64
	platform         string
	hasForcePlatform bool
	sources          []schedulerIndexedAccountSource
	seenAccountIDs   map[int64]struct{}
}

func newSchedulerIndexedAccountPager(
	snapshot *SchedulerSnapshotService,
	groupID *int64,
	platform string,
	hasForcePlatform bool,
	sources []SchedulerCapabilityIndex,
) *schedulerIndexedAccountPager {
	if snapshot == nil || len(sources) == 0 {
		return nil
	}
	pagerSources := make([]schedulerIndexedAccountSource, 0, len(sources))
	seenSources := make(map[string]struct{}, len(sources))
	for _, source := range sources {
		key := string(source.Kind) + "\x00" + strings.TrimSpace(source.Value)
		if _, exists := seenSources[key]; exists {
			continue
		}
		seenSources[key] = struct{}{}
		pagerSources = append(pagerSources, schedulerIndexedAccountSource{
			index:   source,
			hasMore: true,
		})
	}
	if len(pagerSources) == 0 {
		return nil
	}
	return &schedulerIndexedAccountPager{
		snapshot:         snapshot,
		groupID:          groupID,
		platform:         platform,
		hasForcePlatform: hasForcePlatform,
		sources:          pagerSources,
		seenAccountIDs:   make(map[int64]struct{}),
	}
}

func (p *schedulerIndexedAccountPager) Next(ctx context.Context, limit int) ([]Account, bool, error) {
	if p == nil || p.snapshot == nil || len(p.sources) == 0 {
		return nil, false, nil
	}
	if limit <= 0 {
		limit = 1
	}

	for {
		batch := make([]Account, 0, limit)
		remaining := false

		for i := range p.sources {
			source := &p.sources[i]
			if !source.hasMore {
				continue
			}

			page, _, hasMore, err := p.snapshot.ListSchedulableAccountsByCapabilityPage(
				ctx,
				p.groupID,
				p.platform,
				p.hasForcePlatform,
				source.index,
				source.offset,
				limit,
			)
			if err != nil {
				return nil, false, err
			}
			defaultSchedulingRuntimeKernelStats.indexPageFetches.Add(1)
			defaultSchedulingRuntimeKernelStats.indexFetchedAccounts.Add(int64(len(page)))

			source.offset += limit
			source.hasMore = hasMore
			if hasMore {
				remaining = true
			}

			for _, account := range page {
				if _, exists := p.seenAccountIDs[account.ID]; exists {
					continue
				}
				p.seenAccountIDs[account.ID] = struct{}{}
				batch = append(batch, account)
			}
		}

		if len(batch) > 0 {
			defaultSchedulingRuntimeKernelStats.indexReturnedBatches.Add(1)
			defaultSchedulingRuntimeKernelStats.indexReturnedAccounts.Add(int64(len(batch)))
			for i := range p.sources {
				if p.sources[i].hasMore {
					remaining = true
					break
				}
			}
			return batch, remaining, nil
		}

		active := false
		for i := range p.sources {
			if p.sources[i].hasMore {
				active = true
				break
			}
		}
		if !active {
			return nil, false, nil
		}
	}
}

func buildRequestedModelCapabilitySources(
	ctx context.Context,
	snapshot *SchedulerSnapshotService,
	groupID *int64,
	platform string,
	hasForcePlatform bool,
	requestedModel string,
) ([]SchedulerCapabilityIndex, error) {
	model := strings.TrimSpace(requestedModel)
	if snapshot == nil || model == "" {
		return nil, nil
	}

	sources := []SchedulerCapabilityIndex{
		{Kind: SchedulerCapabilityIndexModelAny},
		{Kind: SchedulerCapabilityIndexModelExact, Value: model},
	}
	patterns, _, err := snapshot.ListSchedulableCapabilityIndexValues(ctx, groupID, platform, hasForcePlatform, SchedulerCapabilityIndexModelPattern)
	if err != nil {
		return nil, err
	}
	for _, pattern := range patterns {
		if !matchWildcard(pattern, model) {
			continue
		}
		sources = append(sources, SchedulerCapabilityIndex{
			Kind:  SchedulerCapabilityIndexModelPattern,
			Value: pattern,
		})
	}
	return sources, nil
}
