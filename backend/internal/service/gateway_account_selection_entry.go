package service

import (
	"context"

	"github.com/senran-N/sub2api/internal/pkg/ctxkey"
)

// SelectAccount 选择账号（粘性会话+优先级）
func (s *GatewayService) SelectAccount(ctx context.Context, groupID *int64, sessionHash string) (*Account, error) {
	return s.SelectAccountForModel(ctx, groupID, sessionHash, "")
}

// SelectAccountForModel 选择支持指定模型的账号（粘性会话+优先级+模型映射）
func (s *GatewayService) SelectAccountForModel(ctx context.Context, groupID *int64, sessionHash string, requestedModel string) (*Account, error) {
	return s.SelectAccountForModelWithExclusions(ctx, groupID, sessionHash, requestedModel, nil)
}

// SelectAccountForModelWithExclusions selects an account supporting the requested model while excluding specified accounts.
func (s *GatewayService) SelectAccountForModelWithExclusions(ctx context.Context, groupID *int64, sessionHash string, requestedModel string, excludedIDs map[int64]struct{}) (*Account, error) {
	platform, hasForcePlatform, resolvedGroupID, scopedCtx, err := s.resolveSelectionPlatform(ctx, groupID)
	if err != nil {
		return nil, err
	}
	groupID = resolvedGroupID
	ctx = scopedCtx

	if (platform == PlatformAnthropic || platform == PlatformGemini) && !hasForcePlatform {
		return s.selectAccountWithMixedScheduling(ctx, groupID, sessionHash, requestedModel, excludedIDs, platform)
	}

	return s.selectAccountForModelWithPlatform(ctx, groupID, sessionHash, requestedModel, excludedIDs, platform)
}

func (s *GatewayService) resolveSelectionPlatform(ctx context.Context, groupID *int64) (platform string, hasForcePlatform bool, resolvedGroupID *int64, scopedCtx context.Context, err error) {
	scopedCtx = ctx

	forcePlatform, ok := ctx.Value(ctxkey.ForcePlatform).(string)
	if ok && forcePlatform != "" {
		return forcePlatform, true, groupID, scopedCtx, nil
	}

	if groupID != nil {
		group, resolvedID, resolveErr := s.resolveGatewayGroup(ctx, groupID)
		if resolveErr != nil {
			return "", false, nil, scopedCtx, resolveErr
		}
		return group.Platform, false, resolvedID, s.withGroupContext(ctx, group), nil
	}

	return PlatformAnthropic, false, groupID, scopedCtx, nil
}

// selectAccountForModelWithPlatform 选择单平台账户（完全隔离）
func (s *GatewayService) selectAccountForModelWithPlatform(ctx context.Context, groupID *int64, sessionHash string, requestedModel string, excludedIDs map[int64]struct{}, platform string) (*Account, error) {
	preferOAuth := platform == PlatformGemini
	singlePlatformCheck := func(a *Account) bool { return a.Platform == platform }
	singleOAuthTieBreaker := func(a, b *Account) bool {
		return preferOAuth && a.Type != b.Type && a.Type == AccountTypeOAuth
	}

	return s.selectAccountWithLegacyScheduling(&legacyAccountSelectionInput{
		ctx:                 ctx,
		groupID:             groupID,
		sessionHash:         sessionHash,
		requestedModel:      requestedModel,
		excludedIDs:         excludedIDs,
		platform:            platform,
		allowForcePlatform:  true,
		debugBeginLabel:     "[ModelRoutingDebug] legacy routed begin: group_id=%v model=%s platform=%s session=%s routed_ids=%v",
		debugStickyHitLabel: "[ModelRoutingDebug] legacy routed sticky hit: group_id=%v model=%s session=%s account=%d",
		debugSelectLabel:    "[ModelRoutingDebug] legacy routed select: group_id=%v model=%s session=%s account=%d",
		stickyPlatformCheck: singlePlatformCheck,
		oauthTieBreaker:     singleOAuthTieBreaker,
		failureUseMixed:     false,
	})
}

// selectAccountWithMixedScheduling 选择账户（支持混合调度）
// 查询原生平台账户 + 启用 mixed_scheduling 的 antigravity 账户
func (s *GatewayService) selectAccountWithMixedScheduling(ctx context.Context, groupID *int64, sessionHash string, requestedModel string, excludedIDs map[int64]struct{}, nativePlatform string) (*Account, error) {
	preferOAuth := nativePlatform == PlatformGemini
	mixedPlatformCheck := func(a *Account) bool {
		return a.Platform == nativePlatform || (a.Platform == PlatformAntigravity && a.IsMixedSchedulingEnabled())
	}
	mixedAntigravityFilter := func(a *Account) bool {
		return a.Platform != PlatformAntigravity || a.IsMixedSchedulingEnabled()
	}
	mixedOAuthTieBreaker := func(a, b *Account) bool {
		return preferOAuth && a.Platform == PlatformGemini && b.Platform == PlatformGemini && a.Type != b.Type && a.Type == AccountTypeOAuth
	}

	return s.selectAccountWithLegacyScheduling(&legacyAccountSelectionInput{
		ctx:                 ctx,
		groupID:             groupID,
		sessionHash:         sessionHash,
		requestedModel:      requestedModel,
		excludedIDs:         excludedIDs,
		platform:            nativePlatform,
		allowForcePlatform:  false,
		debugBeginLabel:     "[ModelRoutingDebug] legacy mixed routed begin: group_id=%v model=%s platform=%s session=%s routed_ids=%v",
		debugStickyHitLabel: "[ModelRoutingDebug] legacy mixed routed sticky hit: group_id=%v model=%s session=%s account=%d",
		debugSelectLabel:    "[ModelRoutingDebug] legacy mixed routed select: group_id=%v model=%s session=%s account=%d",
		stickyPlatformCheck: mixedPlatformCheck,
		platformFilter:      mixedAntigravityFilter,
		oauthTieBreaker:     mixedOAuthTieBreaker,
		failureUseMixed:     true,
	})
}
