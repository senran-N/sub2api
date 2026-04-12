package service

import (
	"context"
	"log/slog"

	"github.com/senran-N/sub2api/internal/pkg/ctxkey"
)

func gatewayGroupIDFromContext(ctx context.Context) *int64 {
	group, ok := ctx.Value(ctxkey.Group).(*Group)
	if !ok || !IsGroupContextValid(group) {
		return nil
	}
	groupID := group.ID
	return &groupID
}

func (s *GatewayService) ResolveChannelMapping(ctx context.Context, groupID int64, model string) ChannelMappingResult {
	if s == nil || s.channelService == nil {
		return ChannelMappingResult{MappedModel: model}
	}
	return s.channelService.ResolveChannelMapping(ctx, groupID, model)
}

// ReplaceModelInBody 替换请求体中的模型名，供 handler 侧在转发前覆写 model 字段。
func (s *GatewayService) ReplaceModelInBody(body []byte, newModel string) []byte {
	return ReplaceModelInBody(body, newModel)
}

func (s *GatewayService) IsModelRestricted(ctx context.Context, groupID int64, model string) bool {
	if s == nil || s.channelService == nil {
		return false
	}
	return s.channelService.IsModelRestricted(ctx, groupID, model)
}

// ResolveChannelMappingAndRestrict 解析渠道映射。
// 模型限制检查已移至调度阶段，restricted 始终返回 false。
func (s *GatewayService) ResolveChannelMappingAndRestrict(ctx context.Context, groupID *int64, model string) (ChannelMappingResult, bool) {
	if s == nil || s.channelService == nil {
		return ChannelMappingResult{MappedModel: model}, false
	}
	return s.channelService.ResolveChannelMappingAndRestrict(ctx, groupID, model)
}

// checkChannelPricingRestriction 根据渠道计费基准检查模型是否受定价列表限制。
// requested/channel_mapped 模式可在调度前直接判定；upstream 模式需逐账号判断。
func (s *GatewayService) checkChannelPricingRestriction(ctx context.Context, groupID *int64, requestedModel string) bool {
	if groupID == nil || s == nil || s.channelService == nil || requestedModel == "" {
		return false
	}
	mapping := s.channelService.ResolveChannelMapping(ctx, *groupID, requestedModel)
	billingModel := billingModelForRestriction(mapping.BillingModelSource, requestedModel, mapping.MappedModel)
	if billingModel == "" {
		return false
	}
	return s.channelService.IsModelRestricted(ctx, *groupID, billingModel)
}

func billingModelForRestriction(source, requestedModel, channelMappedModel string) string {
	switch source {
	case BillingModelSourceRequested:
		return requestedModel
	case BillingModelSourceUpstream:
		return ""
	case BillingModelSourceChannelMapped:
		return channelMappedModel
	default:
		return channelMappedModel
	}
}

// isUpstreamModelRestrictedByChannel 检查账号映射后的上游模型是否受渠道定价限制。
func (s *GatewayService) isUpstreamModelRestrictedByChannel(ctx context.Context, groupID int64, account *Account, requestedModel string) bool {
	if s == nil || s.channelService == nil || account == nil {
		return false
	}
	upstreamModel := resolveAccountUpstreamModel(account, requestedModel)
	if upstreamModel == "" {
		return false
	}
	return s.channelService.IsModelRestricted(ctx, groupID, upstreamModel)
}

func resolveAccountUpstreamModel(account *Account, requestedModel string) string {
	if account == nil {
		return ""
	}
	if account.Platform == PlatformAntigravity {
		return mapAntigravityModel(account, requestedModel)
	}
	if account.Platform == PlatformOpenAI {
		return resolveOpenAIForwardModel(account, requestedModel, "")
	}
	if mappedModel, matched := resolveMappedModelWithOpenAIReasoningFallback(account, requestedModel); matched {
		return mappedModel
	}
	return account.GetMappedModel(requestedModel)
}

// needsUpstreamChannelRestrictionCheck 判断是否需要在调度循环中逐账号检查上游模型的渠道限制。
func (s *GatewayService) needsUpstreamChannelRestrictionCheck(ctx context.Context, groupID *int64) bool {
	if groupID == nil || s == nil || s.channelService == nil {
		return false
	}
	channel, err := s.channelService.GetChannelForGroup(ctx, *groupID)
	if err != nil {
		slog.Warn("failed to check channel upstream restriction", "group_id", *groupID, "error", err)
		return false
	}
	if channel == nil || !channel.RestrictModels {
		return false
	}
	return channel.BillingModelSource == BillingModelSourceUpstream
}

func (s *GatewayService) isChannelModelRestrictedForSelection(ctx context.Context, account *Account, requestedModel string) bool {
	return s.isChannelModelRestrictedForSelectionWithGroup(ctx, gatewayGroupIDFromContext(ctx), account, requestedModel)
}

func (s *GatewayService) isChannelModelRestrictedForSelectionWithGroup(ctx context.Context, groupID *int64, account *Account, requestedModel string) bool {
	if requestedModel == "" {
		return false
	}
	if s.checkChannelPricingRestriction(ctx, groupID, requestedModel) {
		return true
	}
	if groupID == nil || account == nil || !s.needsUpstreamChannelRestrictionCheck(ctx, groupID) {
		return false
	}
	return s.isUpstreamModelRestrictedByChannel(ctx, *groupID, account, requestedModel)
}
