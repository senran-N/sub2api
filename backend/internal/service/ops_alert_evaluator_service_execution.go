package service

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/senran-N/sub2api/internal/pkg/logger"
)

func (s *OpsAlertEvaluatorService) evaluateOnce(interval time.Duration) {
	if s == nil || s.opsRepo == nil {
		return
	}
	if s.cfg != nil && !s.cfg.Ops.Enabled {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), opsAlertEvaluatorTimeout)
	defer cancel()

	if s.opsService != nil && !s.opsService.IsMonitoringEnabled(ctx) {
		return
	}

	runtimeCfg := defaultOpsAlertRuntimeSettings()
	if s.opsService != nil {
		if loaded, err := s.opsService.GetOpsAlertRuntimeSettings(ctx); err == nil && loaded != nil {
			runtimeCfg = loaded
		}
	}

	release, ok := s.tryAcquireLeaderLock(ctx, runtimeCfg.DistributedLock)
	if !ok {
		return
	}
	if release != nil {
		defer release()
	}

	startedAt := time.Now().UTC()
	runAt := startedAt

	rules, err := s.opsRepo.ListAlertRules(ctx)
	if err != nil {
		s.recordHeartbeatError(runAt, time.Since(startedAt), err)
		logger.LegacyPrintf("service.ops_alert_evaluator", "[OpsAlertEvaluator] list rules failed: %v", err)
		return
	}

	rulesTotal := len(rules)
	rulesEnabled := 0
	rulesEvaluated := 0
	eventsCreated := 0
	eventsResolved := 0
	emailsSent := 0

	now := time.Now().UTC()
	safeEnd := now.Truncate(time.Minute)
	if safeEnd.IsZero() {
		safeEnd = now
	}

	systemMetrics, _ := s.opsRepo.GetLatestSystemMetrics(ctx, 1)
	s.pruneRuleStates(rules)

	for _, rule := range rules {
		if rule == nil || !rule.Enabled || rule.ID <= 0 {
			continue
		}
		rulesEnabled++

		scopePlatform, scopeGroupID, scopeRegion := parseOpsAlertRuleScope(rule.Filters)
		windowMinutes := rule.WindowMinutes
		if windowMinutes <= 0 {
			windowMinutes = 1
		}
		windowStart := safeEnd.Add(-time.Duration(windowMinutes) * time.Minute)
		windowEnd := safeEnd

		metricValue, ok := s.computeRuleMetric(ctx, rule, systemMetrics, windowStart, windowEnd, scopePlatform, scopeGroupID)
		if !ok {
			s.resetRuleState(rule.ID, now)
			continue
		}
		rulesEvaluated++

		breachedNow := compareMetric(metricValue, rule.Operator, rule.Threshold)
		required := requiredSustainedBreaches(rule.SustainedMinutes, interval)
		consecutive := s.updateRuleBreaches(rule.ID, now, interval, breachedNow)

		activeEvent, err := s.opsRepo.GetActiveAlertEvent(ctx, rule.ID)
		if err != nil {
			logger.LegacyPrintf("service.ops_alert_evaluator", "[OpsAlertEvaluator] get active event failed (rule=%d): %v", rule.ID, err)
			continue
		}

		if breachedNow && consecutive >= required {
			if activeEvent != nil {
				continue
			}

			if s.opsService != nil {
				platform := strings.TrimSpace(scopePlatform)
				region := scopeRegion
				if platform != "" {
					if ok, err := s.opsService.IsAlertSilenced(ctx, rule.ID, platform, scopeGroupID, region, now); err == nil && ok {
						continue
					}
				}
			}

			latestEvent, err := s.opsRepo.GetLatestAlertEvent(ctx, rule.ID)
			if err != nil {
				logger.LegacyPrintf("service.ops_alert_evaluator", "[OpsAlertEvaluator] get latest event failed (rule=%d): %v", rule.ID, err)
				continue
			}
			if latestEvent != nil && rule.CooldownMinutes > 0 {
				cooldown := time.Duration(rule.CooldownMinutes) * time.Minute
				if now.Sub(latestEvent.FiredAt) < cooldown {
					continue
				}
			}

			firedEvent := &OpsAlertEvent{
				RuleID:         rule.ID,
				Severity:       strings.TrimSpace(rule.Severity),
				Status:         OpsAlertStatusFiring,
				Title:          fmt.Sprintf("%s: %s", strings.TrimSpace(rule.Severity), strings.TrimSpace(rule.Name)),
				Description:    buildOpsAlertDescription(rule, metricValue, windowMinutes, scopePlatform, scopeGroupID),
				MetricValue:    float64Ptr(metricValue),
				ThresholdValue: float64Ptr(rule.Threshold),
				Dimensions:     buildOpsAlertDimensions(scopePlatform, scopeGroupID),
				FiredAt:        now,
				CreatedAt:      now,
			}

			created, err := s.opsRepo.CreateAlertEvent(ctx, firedEvent)
			if err != nil {
				logger.LegacyPrintf("service.ops_alert_evaluator", "[OpsAlertEvaluator] create event failed (rule=%d): %v", rule.ID, err)
				continue
			}

			eventsCreated++
			if created != nil && created.ID > 0 && s.maybeSendAlertEmail(ctx, runtimeCfg, rule, created) {
				emailsSent++
			}
			continue
		}

		if activeEvent != nil {
			resolvedAt := now
			if err := s.opsRepo.UpdateAlertEventStatus(ctx, activeEvent.ID, OpsAlertStatusResolved, &resolvedAt); err != nil {
				logger.LegacyPrintf("service.ops_alert_evaluator", "[OpsAlertEvaluator] resolve event failed (event=%d): %v", activeEvent.ID, err)
			} else {
				eventsResolved++
			}
		}
	}

	result := truncateString(fmt.Sprintf("rules=%d enabled=%d evaluated=%d created=%d resolved=%d emails_sent=%d", rulesTotal, rulesEnabled, rulesEvaluated, eventsCreated, eventsResolved, emailsSent), 2048)
	s.recordHeartbeatSuccess(runAt, time.Since(startedAt), result)
}

func (s *OpsAlertEvaluatorService) pruneRuleStates(rules []*OpsAlertRule) {
	s.mu.Lock()
	defer s.mu.Unlock()

	live := map[int64]struct{}{}
	for _, rule := range rules {
		if rule != nil && rule.ID > 0 {
			live[rule.ID] = struct{}{}
		}
	}
	for id := range s.ruleStates {
		if _, ok := live[id]; !ok {
			delete(s.ruleStates, id)
		}
	}
}

func (s *OpsAlertEvaluatorService) resetRuleState(ruleID int64, now time.Time) {
	if ruleID <= 0 {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	state, ok := s.ruleStates[ruleID]
	if !ok {
		state = &opsAlertRuleState{}
		s.ruleStates[ruleID] = state
	}
	state.LastEvaluatedAt = now
	state.ConsecutiveBreaches = 0
}

func (s *OpsAlertEvaluatorService) updateRuleBreaches(ruleID int64, now time.Time, interval time.Duration, breached bool) int {
	if ruleID <= 0 {
		return 0
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	state, ok := s.ruleStates[ruleID]
	if !ok {
		state = &opsAlertRuleState{}
		s.ruleStates[ruleID] = state
	}

	if !state.LastEvaluatedAt.IsZero() && interval > 0 && now.Sub(state.LastEvaluatedAt) > interval*2 {
		state.ConsecutiveBreaches = 0
	}

	state.LastEvaluatedAt = now
	if breached {
		state.ConsecutiveBreaches++
	} else {
		state.ConsecutiveBreaches = 0
	}
	return state.ConsecutiveBreaches
}

func requiredSustainedBreaches(sustainedMinutes int, interval time.Duration) int {
	if sustainedMinutes <= 0 {
		return 1
	}
	if interval <= 0 {
		return sustainedMinutes
	}
	required := int(math.Ceil(float64(sustainedMinutes*60) / interval.Seconds()))
	if required < 1 {
		return 1
	}
	return required
}

func parseOpsAlertRuleScope(filters map[string]any) (platform string, groupID *int64, region *string) {
	if filters == nil {
		return "", nil, nil
	}
	if value, ok := filters["platform"]; ok {
		if text, ok := value.(string); ok {
			platform = strings.TrimSpace(text)
		}
	}
	if value, ok := filters["group_id"]; ok {
		switch typed := value.(type) {
		case float64:
			if typed > 0 {
				id := int64(typed)
				groupID = &id
			}
		case int64:
			if typed > 0 {
				id := typed
				groupID = &id
			}
		case int:
			if typed > 0 {
				id := int64(typed)
				groupID = &id
			}
		case string:
			n, err := strconv.ParseInt(strings.TrimSpace(typed), 10, 64)
			if err == nil && n > 0 {
				groupID = &n
			}
		}
	}
	if value, ok := filters["region"]; ok {
		if text, ok := value.(string); ok {
			trimmed := strings.TrimSpace(text)
			if trimmed != "" {
				region = &trimmed
			}
		}
	}
	return platform, groupID, region
}

func (s *OpsAlertEvaluatorService) computeRuleMetric(
	ctx context.Context,
	rule *OpsAlertRule,
	systemMetrics *OpsSystemMetricsSnapshot,
	start time.Time,
	end time.Time,
	platform string,
	groupID *int64,
) (float64, bool) {
	if rule == nil {
		return 0, false
	}

	switch strings.TrimSpace(rule.MetricType) {
	case "cpu_usage_percent":
		if systemMetrics != nil && systemMetrics.CPUUsagePercent != nil {
			return *systemMetrics.CPUUsagePercent, true
		}
		return 0, false
	case "memory_usage_percent":
		if systemMetrics != nil && systemMetrics.MemoryUsagePercent != nil {
			return *systemMetrics.MemoryUsagePercent, true
		}
		return 0, false
	case "concurrency_queue_depth":
		if systemMetrics != nil && systemMetrics.ConcurrencyQueueDepth != nil {
			return float64(*systemMetrics.ConcurrencyQueueDepth), true
		}
		return 0, false
	case "group_available_accounts":
		if groupID == nil || *groupID <= 0 || s == nil || s.opsService == nil {
			return 0, false
		}
		availability, err := s.opsService.GetAccountAvailability(ctx, platform, groupID)
		if err != nil || availability == nil {
			return 0, false
		}
		if availability.Group == nil {
			return 0, true
		}
		return float64(availability.Group.AvailableCount), true
	case "group_available_ratio":
		if groupID == nil || *groupID <= 0 || s == nil || s.opsService == nil {
			return 0, false
		}
		availability, err := s.opsService.GetAccountAvailability(ctx, platform, groupID)
		if err != nil || availability == nil {
			return 0, false
		}
		return computeGroupAvailableRatio(availability.Group), true
	case "account_rate_limited_count":
		if s == nil || s.opsService == nil {
			return 0, false
		}
		availability, err := s.opsService.GetAccountAvailability(ctx, platform, groupID)
		if err != nil || availability == nil {
			return 0, false
		}
		return float64(countAccountsByCondition(availability.Accounts, func(acc *AccountAvailability) bool {
			return acc.IsRateLimited
		})), true
	case "account_error_count":
		if s == nil || s.opsService == nil {
			return 0, false
		}
		availability, err := s.opsService.GetAccountAvailability(ctx, platform, groupID)
		if err != nil || availability == nil {
			return 0, false
		}
		return float64(countAccountsByCondition(availability.Accounts, func(acc *AccountAvailability) bool {
			return acc.HasError && acc.TempUnschedulableUntil == nil
		})), true
	case "group_rate_limit_ratio":
		if groupID == nil || *groupID <= 0 || s == nil || s.opsService == nil {
			return 0, false
		}
		availability, err := s.opsService.GetAccountAvailability(ctx, platform, groupID)
		if err != nil || availability == nil {
			return 0, false
		}
		if availability.Group == nil || availability.Group.TotalAccounts <= 0 {
			return 0, true
		}
		return (float64(availability.Group.RateLimitCount) / float64(availability.Group.TotalAccounts)) * 100, true
	case "account_error_ratio":
		if s == nil || s.opsService == nil {
			return 0, false
		}
		availability, err := s.opsService.GetAccountAvailability(ctx, platform, groupID)
		if err != nil || availability == nil {
			return 0, false
		}
		total := int64(len(availability.Accounts))
		if total <= 0 {
			return 0, true
		}
		errorCount := countAccountsByCondition(availability.Accounts, func(acc *AccountAvailability) bool {
			return acc.HasError && acc.TempUnschedulableUntil == nil
		})
		return (float64(errorCount) / float64(total)) * 100, true
	case "overload_account_count":
		if s == nil || s.opsService == nil {
			return 0, false
		}
		availability, err := s.opsService.GetAccountAvailability(ctx, platform, groupID)
		if err != nil || availability == nil {
			return 0, false
		}
		return float64(countAccountsByCondition(availability.Accounts, func(acc *AccountAvailability) bool {
			return acc.IsOverloaded
		})), true
	}

	overview, err := s.opsRepo.GetDashboardOverview(ctx, &OpsDashboardFilter{
		StartTime: start,
		EndTime:   end,
		Platform:  platform,
		GroupID:   groupID,
		QueryMode: OpsQueryModeRaw,
	})
	if err != nil || overview == nil {
		return 0, false
	}

	switch strings.TrimSpace(rule.MetricType) {
	case "success_rate":
		if overview.RequestCountSLA <= 0 {
			return 0, false
		}
		return overview.SLA * 100, true
	case "error_rate":
		if overview.RequestCountSLA <= 0 {
			return 0, false
		}
		return overview.ErrorRate * 100, true
	case "upstream_error_rate":
		if overview.RequestCountSLA <= 0 {
			return 0, false
		}
		return overview.UpstreamErrorRate * 100, true
	default:
		return 0, false
	}
}

func compareMetric(value float64, operator string, threshold float64) bool {
	switch strings.TrimSpace(operator) {
	case ">":
		return value > threshold
	case ">=":
		return value >= threshold
	case "<":
		return value < threshold
	case "<=":
		return value <= threshold
	case "==":
		return value == threshold
	case "!=":
		return value != threshold
	default:
		return false
	}
}

func buildOpsAlertDimensions(platform string, groupID *int64) map[string]any {
	dimensions := map[string]any{}
	if strings.TrimSpace(platform) != "" {
		dimensions["platform"] = strings.TrimSpace(platform)
	}
	if groupID != nil && *groupID > 0 {
		dimensions["group_id"] = *groupID
	}
	if len(dimensions) == 0 {
		return nil
	}
	return dimensions
}

func buildOpsAlertDescription(rule *OpsAlertRule, value float64, windowMinutes int, platform string, groupID *int64) string {
	if rule == nil {
		return ""
	}
	scope := "overall"
	if strings.TrimSpace(platform) != "" {
		scope = fmt.Sprintf("platform=%s", strings.TrimSpace(platform))
	}
	if groupID != nil && *groupID > 0 {
		scope = fmt.Sprintf("%s group_id=%d", scope, *groupID)
	}
	if windowMinutes <= 0 {
		windowMinutes = 1
	}
	return fmt.Sprintf("%s %s %.2f (current %.2f) over last %dm (%s)",
		strings.TrimSpace(rule.MetricType),
		strings.TrimSpace(rule.Operator),
		rule.Threshold,
		value,
		windowMinutes,
		strings.TrimSpace(scope),
	)
}

// computeGroupAvailableRatio returns the available percentage for a group.
// Formula: (AvailableCount / TotalAccounts) * 100.
// Returns 0 when TotalAccounts is 0.
func computeGroupAvailableRatio(group *GroupAvailability) float64 {
	if group == nil || group.TotalAccounts <= 0 {
		return 0
	}
	return (float64(group.AvailableCount) / float64(group.TotalAccounts)) * 100
}

// countAccountsByCondition counts accounts that satisfy the given condition.
func countAccountsByCondition(accounts map[int64]*AccountAvailability, condition func(*AccountAvailability) bool) int64 {
	if len(accounts) == 0 || condition == nil {
		return 0
	}
	var count int64
	for _, account := range accounts {
		if account != nil && condition(account) {
			count++
		}
	}
	return count
}
