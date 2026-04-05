package service

import "context"

// RecordUsage 记录使用量并扣费（或更新订阅用量）
func (s *GatewayService) RecordUsage(ctx context.Context, input *RecordUsageInput) error {
	result := input.Result
	apiKey := input.APIKey
	user := input.User
	account := input.Account
	subscription := input.Subscription

	cacheTTLOverridden := prepareGatewayUsageForBilling(result, account, input.ForceCacheBilling)
	multiplier := s.resolveGatewayUsageRateMultiplier(ctx, apiKey, user)
	channelUsage := enrichChannelUsageFields(ctx, s.channelService, apiKey.GroupID, result.Model, result.UpstreamModel, input.ChannelUsageFields)
	billingModel := resolveChannelBillingModel(channelUsage, forwardResultBillingModel(result.Model, result.UpstreamModel))
	cost := s.calculateGatewayUsageCost(ctx, result, apiKey, billingModel, multiplier)

	logBuild := buildGatewayUsageLog(ctx, usageLogBuildInput{
		Result:             result,
		APIKey:             apiKey,
		User:               user,
		Account:            account,
		Subscription:       subscription,
		Cost:               cost,
		Multiplier:         multiplier,
		CacheTTLOverridden: cacheTTLOverridden,
		InboundEndpoint:    input.InboundEndpoint,
		UpstreamEndpoint:   input.UpstreamEndpoint,
		UserAgent:          input.UserAgent,
		IPAddress:          input.IPAddress,
		IncludeMediaType:   true,
		ChannelUsageFields: channelUsage,
	})
	requestID := logBuild.RequestID
	isSubscriptionBilling := logBuild.IsSubscriptionBilling
	accountRateMultiplier := logBuild.AccountRateMultiplier
	usageLog := logBuild.UsageLog
	return s.finalizeGatewayUsageRecord(ctx, gatewayUsageFinalizeInput{
		RequestID:             requestID,
		UsageLog:              usageLog,
		Cost:                  cost,
		User:                  user,
		APIKey:                apiKey,
		Account:               account,
		Subscription:          subscription,
		RequestPayloadHash:    input.RequestPayloadHash,
		IsSubscriptionBilling: isSubscriptionBilling,
		AccountRateMultiplier: accountRateMultiplier,
		APIKeyService:         input.APIKeyService,
	})
}

// RecordUsageWithLongContext 记录使用量并扣费，支持长上下文双倍计费（用于 Gemini）
func (s *GatewayService) RecordUsageWithLongContext(ctx context.Context, input *RecordUsageLongContextInput) error {
	result := input.Result
	apiKey := input.APIKey
	user := input.User
	account := input.Account
	subscription := input.Subscription

	cacheTTLOverridden := prepareGatewayUsageForBilling(result, account, input.ForceCacheBilling)
	multiplier := s.resolveGatewayUsageRateMultiplier(ctx, apiKey, user)
	channelUsage := enrichChannelUsageFields(ctx, s.channelService, apiKey.GroupID, result.Model, result.UpstreamModel, input.ChannelUsageFields)
	billingModel := resolveChannelBillingModel(channelUsage, forwardResultBillingModel(result.Model, result.UpstreamModel))
	cost := s.calculateGatewayLongContextUsageCost(ctx, result, apiKey, billingModel, multiplier, input.LongContextThreshold, input.LongContextMultiplier)

	logBuild := buildGatewayUsageLog(ctx, usageLogBuildInput{
		Result:             result,
		APIKey:             apiKey,
		User:               user,
		Account:            account,
		Subscription:       subscription,
		Cost:               cost,
		Multiplier:         multiplier,
		CacheTTLOverridden: cacheTTLOverridden,
		InboundEndpoint:    input.InboundEndpoint,
		UpstreamEndpoint:   input.UpstreamEndpoint,
		UserAgent:          input.UserAgent,
		IPAddress:          input.IPAddress,
		IncludeMediaType:   false,
		ChannelUsageFields: channelUsage,
	})
	requestID := logBuild.RequestID
	isSubscriptionBilling := logBuild.IsSubscriptionBilling
	accountRateMultiplier := logBuild.AccountRateMultiplier
	usageLog := logBuild.UsageLog
	return s.finalizeGatewayUsageRecord(ctx, gatewayUsageFinalizeInput{
		RequestID:             requestID,
		UsageLog:              usageLog,
		Cost:                  cost,
		User:                  user,
		APIKey:                apiKey,
		Account:               account,
		Subscription:          subscription,
		RequestPayloadHash:    input.RequestPayloadHash,
		IsSubscriptionBilling: isSubscriptionBilling,
		AccountRateMultiplier: accountRateMultiplier,
		APIKeyService:         input.APIKeyService,
	})
}
