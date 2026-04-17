package service

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/senran-N/sub2api/internal/config"
)

func (s *SchedulerSnapshotService) loadAccountsFromDB(ctx context.Context, bucket SchedulerBucket, useMixed bool) ([]Account, error) {
	if s.accountRepo == nil {
		return nil, ErrSchedulerCacheNotReady
	}
	groupID := bucket.GroupID
	if s.isRunModeSimple() {
		groupID = 0
	}

	if useMixed {
		platforms := []string{bucket.Platform, PlatformAntigravity}
		var (
			accounts []Account
			err      error
		)
		if groupID > 0 {
			accounts, err = s.accountRepo.ListSchedulableByGroupIDAndPlatforms(ctx, groupID, platforms)
		} else if s.isRunModeSimple() {
			accounts, err = s.accountRepo.ListSchedulableByPlatforms(ctx, platforms)
		} else {
			accounts, err = s.accountRepo.ListSchedulableUngroupedByPlatforms(ctx, platforms)
		}
		if err != nil {
			return nil, err
		}
		filtered := make([]Account, 0, len(accounts))
		for _, acc := range accounts {
			if acc.Platform == PlatformAntigravity && !acc.IsMixedSchedulingEnabled() {
				continue
			}
			filtered = append(filtered, acc)
		}
		return filtered, nil
	}

	if groupID > 0 {
		return s.accountRepo.ListSchedulableByGroupIDAndPlatform(ctx, groupID, bucket.Platform)
	}
	if s.isRunModeSimple() {
		return s.accountRepo.ListSchedulableByPlatform(ctx, bucket.Platform)
	}
	return s.accountRepo.ListSchedulableUngroupedByPlatform(ctx, bucket.Platform)
}

func (s *SchedulerSnapshotService) bucketFor(groupID *int64, platform string, mode string) SchedulerBucket {
	return SchedulerBucket{
		GroupID:  s.normalizeGroupID(groupID),
		Platform: platform,
		Mode:     mode,
	}
}

func (s *SchedulerSnapshotService) normalizeGroupID(groupID *int64) int64 {
	if s.isRunModeSimple() {
		return 0
	}
	if groupID == nil || *groupID <= 0 {
		return 0
	}
	return *groupID
}

func (s *SchedulerSnapshotService) normalizeGroupIDs(groupIDs []int64) []int64 {
	if s.isRunModeSimple() {
		return []int64{0}
	}
	if len(groupIDs) == 0 {
		return []int64{0}
	}
	seen := make(map[int64]struct{}, len(groupIDs))
	out := make([]int64, 0, len(groupIDs))
	for _, id := range groupIDs {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	if len(out) == 0 {
		return []int64{0}
	}
	return out
}

func (s *SchedulerSnapshotService) resolveMode(platform string, hasForcePlatform bool) string {
	if hasForcePlatform {
		return SchedulerModeForced
	}
	if platform == PlatformAnthropic || platform == PlatformGemini {
		return SchedulerModeMixed
	}
	return SchedulerModeSingle
}

func (s *SchedulerSnapshotService) guardFallback(ctx context.Context) error {
	if s.cfg == nil || s.cfg.Gateway.Scheduling.DbFallbackEnabled {
		if s.fallbackLimit == nil || s.fallbackLimit.Allow() {
			return nil
		}
		return ErrSchedulerFallbackLimited
	}
	return ErrSchedulerCacheNotReady
}

func (s *SchedulerSnapshotService) withFallbackTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if s.cfg == nil || s.cfg.Gateway.Scheduling.DbFallbackTimeoutSeconds <= 0 {
		return context.WithCancel(ctx)
	}
	timeout := time.Duration(s.cfg.Gateway.Scheduling.DbFallbackTimeoutSeconds) * time.Second
	if deadline, ok := ctx.Deadline(); ok {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return context.WithCancel(ctx)
		}
		if remaining < timeout {
			timeout = remaining
		}
	}
	return context.WithTimeout(ctx, timeout)
}

func (s *SchedulerSnapshotService) isRunModeSimple() bool {
	return s.cfg != nil && s.cfg.RunMode == config.RunModeSimple
}

func (s *SchedulerSnapshotService) outboxPollInterval() time.Duration {
	if s.cfg == nil {
		return time.Second
	}
	sec := s.cfg.Gateway.Scheduling.OutboxPollIntervalSeconds
	if sec <= 0 {
		return time.Second
	}
	return time.Duration(sec) * time.Second
}

func (s *SchedulerSnapshotService) fullRebuildInterval() time.Duration {
	if s.cfg == nil {
		return 0
	}
	sec := s.cfg.Gateway.Scheduling.FullRebuildIntervalSeconds
	if sec <= 0 {
		return 0
	}
	return time.Duration(sec) * time.Second
}

func derefAccounts(accounts []*Account) []Account {
	if len(accounts) == 0 {
		return []Account{}
	}
	out := make([]Account, 0, len(accounts))
	for _, account := range accounts {
		if account == nil {
			continue
		}
		out = append(out, *account)
	}
	return out
}

func refAccounts(accounts []Account) []*Account {
	if len(accounts) == 0 {
		return []*Account{}
	}
	out := make([]*Account, 0, len(accounts))
	for i := range accounts {
		out = append(out, &accounts[i])
	}
	return out
}

func (s *SchedulerSnapshotService) resolveBucket(groupID *int64, platform string, hasForcePlatform bool) (SchedulerBucket, bool) {
	useMixed := (platform == PlatformAnthropic || platform == PlatformGemini) && !hasForcePlatform
	mode := s.resolveMode(platform, hasForcePlatform)
	return s.bucketFor(groupID, platform, mode), useMixed
}

func sliceAccountPage(accounts []Account, offset int, limit int) ([]Account, bool) {
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = 1
	}
	if offset >= len(accounts) {
		return []Account{}, false
	}

	end := offset + limit
	if end > len(accounts) {
		end = len(accounts)
	}

	page := make([]Account, end-offset)
	copy(page, accounts[offset:end])
	return page, end < len(accounts)
}

func filterAccountsByCapabilityIndex(accounts []Account, index SchedulerCapabilityIndex) []Account {
	filtered := make([]Account, 0, len(accounts))
	for i := range accounts {
		account := accounts[i]
		if !matchesCapabilityIndex(&account, index) {
			continue
		}
		filtered = append(filtered, account)
	}
	return filtered
}

func matchesCapabilityIndex(account *Account, index SchedulerCapabilityIndex) bool {
	if account == nil {
		return false
	}
	switch index.Kind {
	case SchedulerCapabilityIndexAll:
		return true
	case SchedulerCapabilityIndexPrivacySet:
		return account.IsPrivacySet()
	case SchedulerCapabilityIndexOpenAIWS:
		return isPotentialOpenAIWSCandidate(account)
	case SchedulerCapabilityIndexModelAny:
		return len(account.GetModelMapping()) == 0
	case SchedulerCapabilityIndexModelExact:
		model := strings.TrimSpace(index.Value)
		if model == "" {
			return false
		}
		mapping := account.GetModelMapping()
		if len(mapping) == 0 {
			return false
		}
		_, ok := mapping[model]
		return ok
	case SchedulerCapabilityIndexModelPattern:
		pattern := strings.TrimSpace(index.Value)
		if pattern == "" || !strings.Contains(pattern, "*") {
			return false
		}
		mapping := account.GetModelMapping()
		if len(mapping) == 0 {
			return false
		}
		_, ok := mapping[pattern]
		return ok
	default:
		return false
	}
}

func collectCapabilityIndexValues(accounts []Account, kind SchedulerCapabilityIndexKind) []string {
	valueSet := make(map[string]struct{})
	for i := range accounts {
		account := &accounts[i]
		switch kind {
		case SchedulerCapabilityIndexModelPattern:
			for pattern := range account.GetModelMapping() {
				pattern = strings.TrimSpace(pattern)
				if strings.Contains(pattern, "*") {
					valueSet[pattern] = struct{}{}
				}
			}
		}
	}
	values := make([]string, 0, len(valueSet))
	for value := range valueSet {
		values = append(values, value)
	}
	sort.Strings(values)
	return values
}

func isPotentialOpenAIWSCandidate(account *Account) bool {
	if account == nil || !account.IsOpenAI() {
		return false
	}
	if account.IsOpenAIWSForceHTTPEnabled() {
		return false
	}
	if !account.IsOpenAIOAuth() && !account.IsOpenAIApiKey() {
		return false
	}
	if account.Concurrency <= 0 {
		return false
	}
	return account.ResolveOpenAIResponsesWebSocketV2Mode(OpenAIWSIngressModeCtxPool) != OpenAIWSIngressModeOff
}

type fallbackLimiter struct {
	maxQPS int
	mu     sync.Mutex
	window time.Time
	count  int
}

func newFallbackLimiter(maxQPS int) *fallbackLimiter {
	if maxQPS <= 0 {
		return nil
	}
	return &fallbackLimiter{
		maxQPS: maxQPS,
		window: time.Now(),
	}
}

func (l *fallbackLimiter) Allow() bool {
	if l == nil || l.maxQPS <= 0 {
		return true
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	if now.Sub(l.window) >= time.Second {
		l.window = now
		l.count = 0
	}
	if l.count >= l.maxQPS {
		return false
	}
	l.count++
	return true
}
