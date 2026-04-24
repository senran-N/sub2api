package service

import (
	"context"
	"strings"
	"time"

	"github.com/senran-N/sub2api/internal/config"
	"github.com/senran-N/sub2api/internal/pkg/logger"
)

// OpenAIRecordUsageInput input for recording usage
type OpenAIRecordUsageInput struct {
	Result             *OpenAIForwardResult
	APIKey             *APIKey
	User               *User
	Account            *Account
	Subscription       *UserSubscription
	InboundEndpoint    string
	UpstreamEndpoint   string
	UserAgent          string
	IPAddress          string
	RequestPayloadHash string
	APIKeyService      APIKeyQuotaUpdater
	ChannelUsageFields
}

// RecordUsage records usage and deducts balance
func (s *OpenAIGatewayService) RecordUsage(ctx context.Context, input *OpenAIRecordUsageInput) error {
	result := input.Result
	if result.Usage.InputTokens == 0 && result.Usage.OutputTokens == 0 &&
		result.Usage.CacheCreationInputTokens == 0 && result.Usage.CacheReadInputTokens == 0 &&
		result.Usage.ImageOutputTokens == 0 {
		return nil
	}

	apiKey := input.APIKey
	user := input.User
	account := input.Account
	subscription := input.Subscription

	actualInputTokens := result.Usage.InputTokens - result.Usage.CacheReadInputTokens
	if actualInputTokens < 0 {
		actualInputTokens = 0
	}

	tokens := UsageTokens{
		InputTokens:         actualInputTokens,
		OutputTokens:        result.Usage.OutputTokens,
		CacheCreationTokens: result.Usage.CacheCreationInputTokens,
		CacheReadTokens:     result.Usage.CacheReadInputTokens,
		ImageOutputTokens:   result.Usage.ImageOutputTokens,
	}

	multiplier := s.cfg.Default.RateMultiplier
	if apiKey.GroupID != nil && apiKey.Group != nil {
		resolver := s.userGroupRateResolver
		if resolver == nil {
			resolver = newUserGroupRateResolver(nil, nil, resolveUserGroupRateCacheTTL(s.cfg), nil, "service.openai_gateway")
		}
		multiplier = resolver.Resolve(ctx, user.ID, *apiKey.GroupID, apiKey.Group.RateMultiplier)
	}

	channelUsage := enrichChannelUsageFields(ctx, s.channelService, apiKey.GroupID, result.Model, result.UpstreamModel, input.ChannelUsageFields)
	billingModel := forwardResultBillingModel(result.Model, result.UpstreamModel)
	serviceTier := ""
	if result.ServiceTier != nil {
		serviceTier = strings.TrimSpace(*result.ServiceTier)
	}
	if result.BillingModel != "" {
		billingModel = strings.TrimSpace(result.BillingModel)
	}
	billingModel = resolveChannelBillingModel(channelUsage, billingModel)
	var cost *CostBreakdown
	var err error
	if resolved := resolveChannelPricing(ctx, s.resolver, apiKey, billingModel); resolved != nil {
		groupID := apiKey.Group.ID
		cost, err = s.billingService.CalculateCostUnified(CostInput{
			Ctx:            ctx,
			Model:          billingModel,
			GroupID:        &groupID,
			Tokens:         tokens,
			RequestCount:   1,
			RateMultiplier: multiplier,
			ServiceTier:    serviceTier,
			Resolver:       s.resolver,
			Resolved:       resolved,
		})
	} else {
		cost, err = s.billingService.CalculateCostWithServiceTier(billingModel, tokens, multiplier, serviceTier)
	}
	if err != nil {
		cost = &CostBreakdown{ActualCost: 0}
	}

	isSubscriptionBilling := subscription != nil && apiKey.Group != nil && apiKey.Group.IsSubscriptionType()
	billingType := BillingTypeBalance
	if isSubscriptionBilling {
		billingType = BillingTypeSubscription
	}

	durationMs := int(result.Duration.Milliseconds())
	accountRateMultiplier := account.BillingRateMultiplier()
	requestID := resolveUsageBillingRequestID(ctx, result.RequestID)
	usageLog := &UsageLog{
		UserID:                user.ID,
		APIKeyID:              apiKey.ID,
		AccountID:             account.ID,
		RequestID:             requestID,
		Model:                 result.Model,
		RequestedModel:        channelUsage.OriginalModel,
		UpstreamModel:         optionalNonEqualStringPtr(result.UpstreamModel, result.Model),
		ChannelID:             optionalInt64Ptr(channelUsage.ChannelID),
		ModelMappingChain:     optionalTrimmedStringPtr(channelUsage.ModelMappingChain),
		ServiceTier:           result.ServiceTier,
		ReasoningEffort:       result.ReasoningEffort,
		InboundEndpoint:       optionalTrimmedStringPtr(input.InboundEndpoint),
		UpstreamEndpoint:      optionalTrimmedStringPtr(input.UpstreamEndpoint),
		InputTokens:           actualInputTokens,
		OutputTokens:          result.Usage.OutputTokens,
		CacheCreationTokens:   result.Usage.CacheCreationInputTokens,
		CacheReadTokens:       result.Usage.CacheReadInputTokens,
		ImageOutputTokens:     result.Usage.ImageOutputTokens,
		InputCost:             cost.InputCost,
		ImageOutputCost:       cost.ImageOutputCost,
		OutputCost:            cost.OutputCost,
		CacheCreationCost:     cost.CacheCreationCost,
		CacheReadCost:         cost.CacheReadCost,
		TotalCost:             cost.TotalCost,
		ActualCost:            cost.ActualCost,
		RateMultiplier:        multiplier,
		AccountRateMultiplier: &accountRateMultiplier,
		BillingType:           billingType,
		BillingMode:           resolveOpenAIBillingMode(result, cost),
		Stream:                result.Stream,
		OpenAIWSMode:          result.OpenAIWSMode,
		DurationMs:            &durationMs,
		FirstTokenMs:          result.FirstTokenMs,
		CreatedAt:             time.Now(),
	}
	if usageLog.RequestedModel == "" {
		usageLog.RequestedModel = result.Model
	}
	if input.UserAgent != "" {
		usageLog.UserAgent = &input.UserAgent
	}
	if input.IPAddress != "" {
		usageLog.IPAddress = &input.IPAddress
	}
	if apiKey.GroupID != nil {
		usageLog.GroupID = apiKey.GroupID
	}
	if subscription != nil {
		usageLog.SubscriptionID = &subscription.ID
	}

	if s.cfg != nil && s.cfg.RunMode == config.RunModeSimple {
		writeUsageLogBestEffort(ctx, s.usageLogRepo, usageLog, "service.openai_gateway")
		logger.LegacyPrintf("service.openai_gateway", "[SIMPLE MODE] Usage recorded (not billed): user=%d, tokens=%d", usageLog.UserID, usageLog.TotalTokens())
		s.deferredService.ScheduleLastUsedUpdate(account.ID)
		return nil
	}

	_, billingErr := applyUsageBilling(ctx, requestID, usageLog, &postUsageBillingParams{
		Cost:                  cost,
		User:                  user,
		APIKey:                apiKey,
		Account:               account,
		Subscription:          subscription,
		RequestPayloadHash:    resolveUsageBillingPayloadFingerprint(ctx, input.RequestPayloadHash),
		IsSubscriptionBill:    isSubscriptionBilling,
		AccountRateMultiplier: accountRateMultiplier,
		APIKeyService:         input.APIKeyService,
	}, s.billingDeps(), s.usageBillingRepo)
	if billingErr != nil {
		return billingErr
	}

	writeUsageLogBestEffort(ctx, s.usageLogRepo, usageLog, "service.openai_gateway")
	return nil
}

func resolveOpenAIBillingMode(result *OpenAIForwardResult, cost *CostBreakdown) *string {
	if result == nil {
		return nil
	}
	mode := string(BillingModeToken)
	if cost != nil && cost.BillingMode != "" {
		mode = cost.BillingMode
	}
	return &mode
}
