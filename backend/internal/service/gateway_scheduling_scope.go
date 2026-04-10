package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/senran-N/sub2api/internal/config"
	"github.com/senran-N/sub2api/internal/pkg/ctxkey"
	"github.com/senran-N/sub2api/internal/pkg/logger"
)

func (s *GatewayService) withGroupContext(ctx context.Context, group *Group) context.Context {
	if !IsGroupContextValid(group) {
		return ctx
	}
	if existing, ok := ctx.Value(ctxkey.Group).(*Group); ok && existing != nil && existing.ID == group.ID && IsGroupContextValid(existing) {
		return ctx
	}
	return context.WithValue(ctx, ctxkey.Group, group)
}

func (s *GatewayService) groupFromContext(ctx context.Context, groupID int64) *Group {
	if group, ok := ctx.Value(ctxkey.Group).(*Group); ok && IsGroupContextValid(group) && group.ID == groupID {
		return group
	}
	return nil
}

func (s *GatewayService) resolveGroupByID(ctx context.Context, groupID int64) (*Group, error) {
	if group := s.groupFromContext(ctx, groupID); group != nil {
		return group, nil
	}
	group, err := s.groupRepo.GetByIDLite(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("get group failed: %w", err)
	}
	return group, nil
}

func (s *GatewayService) ResolveGroupByID(ctx context.Context, groupID int64) (*Group, error) {
	return s.resolveGroupByID(ctx, groupID)
}

func (s *GatewayService) routingAccountIDsForRequest(ctx context.Context, groupID *int64, requestedModel string, platform string) []int64 {
	if groupID == nil || requestedModel == "" || platform != PlatformAnthropic {
		return nil
	}

	group, err := s.resolveGroupByID(ctx, *groupID)
	if err != nil || group == nil {
		if s.debugModelRoutingEnabled() {
			logger.LegacyPrintf(
				"service.gateway",
				"[ModelRoutingDebug] resolve group failed: group_id=%v model=%s platform=%s err=%v",
				derefGroupID(groupID),
				requestedModel,
				platform,
				err,
			)
		}
		return nil
	}

	if group.Platform != PlatformAnthropic {
		if s.debugModelRoutingEnabled() {
			logger.LegacyPrintf(
				"service.gateway",
				"[ModelRoutingDebug] skip: non-anthropic group platform: group_id=%d group_platform=%s model=%s",
				group.ID,
				group.Platform,
				requestedModel,
			)
		}
		return nil
	}

	ids := group.GetRoutingAccountIDs(requestedModel)
	if s.debugModelRoutingEnabled() {
		logger.LegacyPrintf(
			"service.gateway",
			"[ModelRoutingDebug] routing lookup: group_id=%d model=%s enabled=%v rules=%d matched_ids=%v",
			group.ID,
			requestedModel,
			group.ModelRoutingEnabled,
			len(group.ModelRouting),
			ids,
		)
	}
	return ids
}

func (s *GatewayService) resolveGatewayGroup(ctx context.Context, groupID *int64) (*Group, *int64, error) {
	if groupID == nil {
		return nil, nil, nil
	}

	currentID := *groupID
	visited := map[int64]struct{}{}
	for {
		if _, seen := visited[currentID]; seen {
			return nil, nil, fmt.Errorf("fallback group cycle detected")
		}
		visited[currentID] = struct{}{}

		group, err := s.resolveGroupByID(ctx, currentID)
		if err != nil {
			return nil, nil, err
		}

		if !group.ClaudeCodeOnly || IsClaudeCodeClient(ctx) {
			return group, &currentID, nil
		}

		if group.FallbackGroupID == nil {
			return nil, nil, ErrClaudeCodeOnly
		}
		currentID = *group.FallbackGroupID
	}
}

// checkClaudeCodeRestriction 检查分组的 Claude Code 客户端限制。
func (s *GatewayService) checkClaudeCodeRestriction(ctx context.Context, groupID *int64) (*Group, *int64, error) {
	if groupID == nil {
		return nil, groupID, nil
	}

	if forcePlatform, hasForcePlatform := ctx.Value(ctxkey.ForcePlatform).(string); hasForcePlatform && forcePlatform != "" {
		return nil, groupID, nil
	}

	group, resolvedID, err := s.resolveGatewayGroup(ctx, groupID)
	if err != nil {
		return nil, nil, err
	}

	return group, resolvedID, nil
}

func (s *GatewayService) resolvePlatform(ctx context.Context, groupID *int64, group *Group) (string, bool, error) {
	forcePlatform, hasForcePlatform := ctx.Value(ctxkey.ForcePlatform).(string)
	if hasForcePlatform && forcePlatform != "" {
		return forcePlatform, true, nil
	}
	if group != nil {
		return group.Platform, false, nil
	}
	if groupID != nil {
		group, err := s.resolveGroupByID(ctx, *groupID)
		if err != nil {
			return "", false, err
		}
		return group.Platform, false, nil
	}
	return PlatformAnthropic, false, nil
}

func (s *GatewayService) listSchedulableAccounts(ctx context.Context, groupID *int64, platform string, hasForcePlatform bool) ([]Account, bool, error) {
	if platform == PlatformSora {
		return s.listSoraSchedulableAccounts(ctx, groupID)
	}
	if s.schedulerSnapshot != nil {
		accounts, useMixed, err := s.schedulerSnapshot.ListSchedulableAccounts(ctx, groupID, platform, hasForcePlatform)
		if err == nil {
			logAccountSchedulingSnapshot(groupID, platform, useMixed, accounts)
		}
		return accounts, useMixed, err
	}

	useMixed := (platform == PlatformAnthropic || platform == PlatformGemini) && !hasForcePlatform
	if useMixed {
		accounts, err := s.listMixedSchedulableAccounts(ctx, groupID, platform)
		if err != nil {
			slog.Debug("account_scheduling_list_failed",
				"group_id", derefGroupID(groupID),
				"platform", platform,
				"error", err)
			return nil, useMixed, err
		}
		return accounts, useMixed, nil
	}

	accounts, err := s.listSinglePlatformSchedulableAccounts(ctx, groupID, platform)
	if err != nil {
		slog.Debug("account_scheduling_list_failed",
			"group_id", derefGroupID(groupID),
			"platform", platform,
			"error", err)
		return nil, useMixed, err
	}
	logAccountSchedulingSingle(groupID, platform, accounts)
	return accounts, useMixed, nil
}

func (s *GatewayService) listMixedSchedulableAccounts(ctx context.Context, groupID *int64, platform string) ([]Account, error) {
	platforms := []string{platform, PlatformAntigravity}

	var (
		accounts []Account
		err      error
	)
	if groupID != nil {
		accounts, err = s.accountRepo.ListSchedulableByGroupIDAndPlatforms(ctx, *groupID, platforms)
	} else if s.cfg != nil && s.cfg.RunMode == config.RunModeSimple {
		accounts, err = s.accountRepo.ListSchedulableByPlatforms(ctx, platforms)
	} else {
		accounts, err = s.accountRepo.ListSchedulableUngroupedByPlatforms(ctx, platforms)
	}
	if err != nil {
		return nil, err
	}

	rawCount := len(accounts)
	filtered := make([]Account, 0, rawCount)
	for _, acc := range accounts {
		if acc.Platform == PlatformAntigravity && !acc.IsMixedSchedulingEnabled() {
			continue
		}
		filtered = append(filtered, acc)
	}
	logAccountSchedulingMixedList(groupID, platform, rawCount, filtered)
	return filtered, nil
}

func (s *GatewayService) listSinglePlatformSchedulableAccounts(ctx context.Context, groupID *int64, platform string) ([]Account, error) {
	if s.cfg != nil && s.cfg.RunMode == config.RunModeSimple {
		return s.accountRepo.ListSchedulableByPlatform(ctx, platform)
	}
	if groupID != nil {
		return s.accountRepo.ListSchedulableByGroupIDAndPlatform(ctx, *groupID, platform)
	}
	return s.accountRepo.ListSchedulableUngroupedByPlatform(ctx, platform)
}

func (s *GatewayService) listSoraSchedulableAccounts(ctx context.Context, groupID *int64) ([]Account, bool, error) {
	const useMixed = false

	var (
		accounts []Account
		err      error
	)
	if s.cfg != nil && s.cfg.RunMode == config.RunModeSimple {
		accounts, err = s.accountRepo.ListByPlatform(ctx, PlatformSora)
	} else if groupID != nil {
		accounts, err = s.accountRepo.ListByGroup(ctx, *groupID)
	} else {
		accounts, err = s.accountRepo.ListByPlatform(ctx, PlatformSora)
	}
	if err != nil {
		slog.Debug("account_scheduling_list_failed",
			"group_id", derefGroupID(groupID),
			"platform", PlatformSora,
			"error", err)
		return nil, useMixed, err
	}

	filtered := make([]Account, 0, len(accounts))
	for _, acc := range accounts {
		if acc.Platform != PlatformSora || !s.isSoraAccountSchedulable(&acc) {
			continue
		}
		filtered = append(filtered, acc)
	}
	slog.Debug("account_scheduling_list_sora",
		"group_id", derefGroupID(groupID),
		"platform", PlatformSora,
		"raw_count", len(accounts),
		"filtered_count", len(filtered))
	logAccountSchedulingDetails(filtered)
	return filtered, useMixed, nil
}

func logAccountSchedulingSnapshot(groupID *int64, platform string, useMixed bool, accounts []Account) {
	slog.Debug("account_scheduling_list_snapshot",
		"group_id", derefGroupID(groupID),
		"platform", platform,
		"use_mixed", useMixed,
		"count", len(accounts))
	logAccountSchedulingDetails(accounts)
}

func logAccountSchedulingSingle(groupID *int64, platform string, accounts []Account) {
	slog.Debug("account_scheduling_list_single",
		"group_id", derefGroupID(groupID),
		"platform", platform,
		"count", len(accounts))
	logAccountSchedulingDetails(accounts)
}

func logAccountSchedulingMixedList(groupID *int64, platform string, rawCount int, accounts []Account) {
	slog.Debug("account_scheduling_list_mixed",
		"group_id", derefGroupID(groupID),
		"platform", platform,
		"raw_count", rawCount,
		"filtered_count", len(accounts))
	logAccountSchedulingDetails(accounts)
}

func logAccountSchedulingDetails(accounts []Account) {
	for _, acc := range accounts {
		slog.Debug("account_scheduling_account_detail",
			"account_id", acc.ID,
			"name", acc.Name,
			"platform", acc.Platform,
			"type", acc.Type,
			"status", acc.Status,
			"tls_fingerprint", acc.IsTLSFingerprintEnabled())
	}
}

// IsSingleAntigravityAccountGroup 检查指定分组是否只有一个 antigravity 平台的可调度账号。
func (s *GatewayService) IsSingleAntigravityAccountGroup(ctx context.Context, groupID *int64) bool {
	accounts, _, err := s.listSchedulableAccounts(ctx, groupID, PlatformAntigravity, true)
	if err != nil {
		return false
	}
	return len(accounts) == 1
}

func (s *GatewayService) isAccountAllowedForPlatform(account *Account, platform string, useMixed bool) bool {
	if account == nil {
		return false
	}
	if useMixed {
		if account.Platform == platform {
			return true
		}
		return account.Platform == PlatformAntigravity && account.IsMixedSchedulingEnabled()
	}
	return account.Platform == platform
}

func (s *GatewayService) isSoraAccountSchedulable(account *Account) bool {
	return s.soraUnschedulableReason(account) == ""
}

func (s *GatewayService) soraUnschedulableReason(account *Account) string {
	if account == nil {
		return "account_nil"
	}
	if account.Status != StatusActive {
		return fmt.Sprintf("status=%s", account.Status)
	}
	if !account.Schedulable {
		return "schedulable=false"
	}
	if account.TempUnschedulableUntil != nil && time.Now().Before(*account.TempUnschedulableUntil) {
		return fmt.Sprintf("temp_unschedulable_until=%s", account.TempUnschedulableUntil.UTC().Format(time.RFC3339))
	}
	return ""
}

func (s *GatewayService) isAccountSchedulableForSelection(account *Account) bool {
	if account == nil {
		return false
	}
	if account.Platform == PlatformSora {
		return s.isSoraAccountSchedulable(account)
	}
	return account.IsSchedulable()
}

func (s *GatewayService) isAccountSchedulableForModelSelection(ctx context.Context, account *Account, requestedModel string) bool {
	if account == nil {
		return false
	}
	if account.Platform == PlatformSora {
		if !s.isSoraAccountSchedulable(account) {
			return false
		}
		return account.GetRateLimitRemainingTimeWithContext(ctx, requestedModel) <= 0
	}
	return account.IsSchedulableForModelWithContext(ctx, requestedModel)
}

func (s *GatewayService) isAccountInGroup(account *Account, groupID *int64) bool {
	if account == nil {
		return false
	}
	if groupID == nil {
		return len(account.AccountGroups) == 0
	}
	for _, accountGroup := range account.AccountGroups {
		if accountGroup.GroupID == *groupID {
			return true
		}
	}
	return false
}

func (s *GatewayService) isStickyAccountFullySchedulable(
	ctx context.Context,
	account *Account,
	groupID *int64,
	requestedModel string,
	isStickyPath bool,
	platformCheck func(*Account) bool,
) bool {
	if !s.isAccountInGroup(account, groupID) {
		return false
	}
	if platformCheck != nil && !platformCheck(account) {
		return false
	}
	if requestedModel != "" && !s.isModelSupportedByAccountWithContext(ctx, account, requestedModel) {
		return false
	}
	if s.isChannelModelRestrictedForSelectionWithGroup(ctx, groupID, account, requestedModel) {
		return false
	}
	if !s.isAccountSchedulableForModelSelection(ctx, account, requestedModel) {
		return false
	}
	if !s.isAccountSchedulableForQuota(account) {
		return false
	}
	if !s.isAccountSchedulableForWindowCost(ctx, account, isStickyPath) {
		return false
	}
	if !s.isAccountSchedulableForRPM(ctx, account, isStickyPath) {
		return false
	}
	return true
}

type candidateFilterParams struct {
	ctx            context.Context
	requestedModel string
	excludedIDs    map[int64]struct{}
	routingSet     map[int64]struct{}
	schedGroup     *Group
	platformFilter func(*Account) bool
}

func (s *GatewayService) filterCandidates(accounts []Account, params *candidateFilterParams) []*Account {
	var candidates []*Account
	for index := range accounts {
		account := &accounts[index]
		if params.routingSet != nil {
			if _, ok := params.routingSet[account.ID]; !ok {
				continue
			}
		}
		if _, excluded := params.excludedIDs[account.ID]; excluded {
			continue
		}
		if !s.isAccountSchedulableForSelection(account) {
			continue
		}
		if params.schedGroup != nil && params.schedGroup.RequirePrivacySet && !account.IsPrivacySet() {
			continue
		}
		if params.platformFilter != nil && !params.platformFilter(account) {
			continue
		}
		if params.requestedModel != "" && !s.isModelSupportedByAccountWithContext(params.ctx, account, params.requestedModel) {
			continue
		}
		if s.isChannelModelRestrictedForSelection(params.ctx, account, params.requestedModel) {
			continue
		}
		if !s.isAccountSchedulableForModelSelection(params.ctx, account, params.requestedModel) {
			continue
		}
		if !s.isAccountSchedulableForQuota(account) {
			continue
		}
		if !s.isAccountSchedulableForWindowCost(params.ctx, account, false) {
			continue
		}
		if !s.isAccountSchedulableForRPM(params.ctx, account, false) {
			continue
		}
		candidates = append(candidates, account)
	}
	return candidates
}
