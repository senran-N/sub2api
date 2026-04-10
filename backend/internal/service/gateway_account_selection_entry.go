package service

import (
	"context"
)

func (s *GatewayService) buildLegacySelectionInput(
	scope *gatewaySelectionScope,
	sessionHash string,
	requestedModel string,
	excludedIDs map[int64]struct{},
	plan *gatewaySelectionPlan,
	bindStickyOnSelection bool,
) *legacyAccountSelectionInput {
	if scope == nil {
		return nil
	}

	if plan == nil {
		plan = s.buildGatewaySelectionPlan(scope, requestedModel)
	}
	beginLabel, stickyHitLabel, selectLabel := plan.legacyDebugLabels()
	return &legacyAccountSelectionInput{
		ctx:                   scope.ctx,
		groupID:               scope.groupID,
		sessionHash:           sessionHash,
		requestedModel:        requestedModel,
		excludedIDs:           excludedIDs,
		plan:                  plan,
		debugBeginLabel:       beginLabel,
		debugStickyHitLabel:   stickyHitLabel,
		debugSelectLabel:      selectLabel,
		bindStickyOnSelection: bindStickyOnSelection,
	}
}

func buildGatewayDirectSelectionScope(
	ctx context.Context,
	groupID *int64,
	platform string,
	hasForcePlatform bool,
	useMixed bool,
) *gatewaySelectionScope {
	return &gatewaySelectionScope{
		ctx:              ctx,
		groupID:          groupID,
		platform:         platform,
		hasForcePlatform: hasForcePlatform,
		useMixed:         useMixed,
		preferOAuth:      platform == PlatformGemini,
	}
}

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
	scope, err := s.prepareGatewaySelectionScope(ctx, groupID, sessionHash)
	if err != nil {
		return nil, err
	}
	return s.selectAccountWithLegacyScheduling(s.buildLegacySelectionInput(scope, sessionHash, requestedModel, excludedIDs, nil, true))
}

// selectAccountForModelWithPlatform 选择单平台账户（完全隔离）
func (s *GatewayService) selectAccountForModelWithPlatform(ctx context.Context, groupID *int64, sessionHash string, requestedModel string, excludedIDs map[int64]struct{}, platform string) (*Account, error) {
	scope := buildGatewayDirectSelectionScope(ctx, groupID, platform, true, false)
	return s.selectAccountWithLegacyScheduling(s.buildLegacySelectionInput(scope, sessionHash, requestedModel, excludedIDs, nil, true))
}

// selectAccountWithMixedScheduling 选择账户（支持混合调度）
// 查询原生平台账户 + 启用 mixed_scheduling 的 antigravity 账户
func (s *GatewayService) selectAccountWithMixedScheduling(ctx context.Context, groupID *int64, sessionHash string, requestedModel string, excludedIDs map[int64]struct{}, nativePlatform string) (*Account, error) {
	scope := buildGatewayDirectSelectionScope(ctx, groupID, nativePlatform, false, true)
	return s.selectAccountWithLegacyScheduling(s.buildLegacySelectionInput(scope, sessionHash, requestedModel, excludedIDs, nil, true))
}
