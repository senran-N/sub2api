package service

import (
	"context"
	"strings"
	"time"

	"github.com/senran-N/sub2api/internal/config"
	"github.com/senran-N/sub2api/internal/pkg/logger"
)

type usageLogBuildInput struct {
	Result             *ForwardResult
	APIKey             *APIKey
	User               *User
	Account            *Account
	Subscription       *UserSubscription
	Cost               *CostBreakdown
	Multiplier         float64
	CacheTTLOverridden bool
	InboundEndpoint    string
	UpstreamEndpoint   string
	UserAgent          string
	IPAddress          string
	IncludeMediaType   bool
}

type usageLogBuildResult struct {
	RequestID             string
	IsSubscriptionBilling bool
	AccountRateMultiplier float64
	UsageLog              *UsageLog
}

type gatewayUsageFinalizeInput struct {
	RequestID             string
	UsageLog              *UsageLog
	Cost                  *CostBreakdown
	User                  *User
	APIKey                *APIKey
	Account               *Account
	Subscription          *UserSubscription
	RequestPayloadHash    string
	IsSubscriptionBilling bool
	AccountRateMultiplier float64
	APIKeyService         APIKeyQuotaUpdater
}

func prepareGatewayUsageForBilling(result *ForwardResult, account *Account, forceCacheBilling bool) bool {
	if forceCacheBilling && result.Usage.InputTokens > 0 {
		logger.LegacyPrintf("service.gateway", "force_cache_billing: %d input_tokens → cache_read_input_tokens (account=%d)",
			result.Usage.InputTokens, account.ID)
		result.Usage.CacheReadInputTokens += result.Usage.InputTokens
		result.Usage.InputTokens = 0
	}

	cacheTTLOverridden := false
	if account.IsCacheTTLOverrideEnabled() {
		applyCacheTTLOverride(&result.Usage, account.GetCacheTTLOverrideTarget())
		cacheTTLOverridden = (result.Usage.CacheCreation5mTokens + result.Usage.CacheCreation1hTokens) > 0
	}

	return cacheTTLOverridden
}

func (s *GatewayService) resolveGatewayUsageRateMultiplier(ctx context.Context, apiKey *APIKey, user *User) float64 {
	multiplier := 1.0
	if s.cfg != nil {
		multiplier = s.cfg.Default.RateMultiplier
	}
	if apiKey != nil && apiKey.GroupID != nil && apiKey.Group != nil && user != nil {
		groupDefault := apiKey.Group.RateMultiplier
		multiplier = s.getUserGroupRateMultiplier(ctx, user.ID, *apiKey.GroupID, groupDefault)
	}
	return multiplier
}

func (s *GatewayService) getUserGroupRateMultiplier(ctx context.Context, userID, groupID int64, groupDefaultMultiplier float64) float64 {
	if s == nil {
		return groupDefaultMultiplier
	}
	resolver := s.userGroupRateResolver
	if resolver == nil {
		resolver = newUserGroupRateResolver(
			s.userGroupRateRepo,
			s.userGroupRateCache,
			resolveUserGroupRateCacheTTL(s.cfg),
			&s.userGroupRateSF,
			"service.gateway",
		)
	}
	return resolver.Resolve(ctx, userID, groupID, groupDefaultMultiplier)
}

func buildGatewayUsageLog(ctx context.Context, input usageLogBuildInput) usageLogBuildResult {
	isSubscriptionBilling := input.Subscription != nil && input.APIKey.Group != nil && input.APIKey.Group.IsSubscriptionType()
	billingType := BillingTypeBalance
	if isSubscriptionBilling {
		billingType = BillingTypeSubscription
	}

	durationMs := int(input.Result.Duration.Milliseconds())
	var imageSize *string
	if input.Result.ImageSize != "" {
		imageSize = &input.Result.ImageSize
	}
	var mediaType *string
	if input.IncludeMediaType && strings.TrimSpace(input.Result.MediaType) != "" {
		mediaType = &input.Result.MediaType
	}

	accountRateMultiplier := input.Account.BillingRateMultiplier()
	requestID := resolveUsageBillingRequestID(ctx, input.Result.RequestID)
	usageLog := &UsageLog{
		UserID:                input.User.ID,
		APIKeyID:              input.APIKey.ID,
		AccountID:             input.Account.ID,
		RequestID:             requestID,
		Model:                 input.Result.Model,
		RequestedModel:        input.Result.Model,
		UpstreamModel:         optionalNonEqualStringPtr(input.Result.UpstreamModel, input.Result.Model),
		ReasoningEffort:       input.Result.ReasoningEffort,
		InboundEndpoint:       optionalTrimmedStringPtr(input.InboundEndpoint),
		UpstreamEndpoint:      optionalTrimmedStringPtr(input.UpstreamEndpoint),
		InputTokens:           input.Result.Usage.InputTokens,
		OutputTokens:          input.Result.Usage.OutputTokens,
		CacheCreationTokens:   input.Result.Usage.CacheCreationInputTokens,
		CacheReadTokens:       input.Result.Usage.CacheReadInputTokens,
		CacheCreation5mTokens: input.Result.Usage.CacheCreation5mTokens,
		CacheCreation1hTokens: input.Result.Usage.CacheCreation1hTokens,
		InputCost:             input.Cost.InputCost,
		OutputCost:            input.Cost.OutputCost,
		CacheCreationCost:     input.Cost.CacheCreationCost,
		CacheReadCost:         input.Cost.CacheReadCost,
		TotalCost:             input.Cost.TotalCost,
		ActualCost:            input.Cost.ActualCost,
		RateMultiplier:        input.Multiplier,
		AccountRateMultiplier: &accountRateMultiplier,
		BillingType:           billingType,
		Stream:                input.Result.Stream,
		DurationMs:            &durationMs,
		FirstTokenMs:          input.Result.FirstTokenMs,
		ImageCount:            input.Result.ImageCount,
		ImageSize:             imageSize,
		MediaType:             mediaType,
		CacheTTLOverridden:    input.CacheTTLOverridden,
		CreatedAt:             time.Now(),
	}

	if input.UserAgent != "" {
		usageLog.UserAgent = &input.UserAgent
	}
	if input.IPAddress != "" {
		usageLog.IPAddress = &input.IPAddress
	}
	if input.APIKey.GroupID != nil {
		usageLog.GroupID = input.APIKey.GroupID
	}
	if input.Subscription != nil {
		usageLog.SubscriptionID = &input.Subscription.ID
	}

	return usageLogBuildResult{
		RequestID:             requestID,
		IsSubscriptionBilling: isSubscriptionBilling,
		AccountRateMultiplier: accountRateMultiplier,
		UsageLog:              usageLog,
	}
}

func (s *GatewayService) finalizeGatewayUsageRecord(ctx context.Context, input gatewayUsageFinalizeInput) error {
	if s.cfg != nil && s.cfg.RunMode == config.RunModeSimple {
		writeUsageLogBestEffort(ctx, s.usageLogRepo, input.UsageLog, "service.gateway")
		logger.LegacyPrintf("service.gateway", "[SIMPLE MODE] Usage recorded (not billed): user=%d, tokens=%d", input.UsageLog.UserID, input.UsageLog.TotalTokens())
		s.deferredService.ScheduleLastUsedUpdate(input.Account.ID)
		return nil
	}

	_, err := applyUsageBilling(ctx, input.RequestID, input.UsageLog, &postUsageBillingParams{
		Cost:                  input.Cost,
		User:                  input.User,
		APIKey:                input.APIKey,
		Account:               input.Account,
		Subscription:          input.Subscription,
		RequestPayloadHash:    resolveUsageBillingPayloadFingerprint(ctx, input.RequestPayloadHash),
		IsSubscriptionBill:    input.IsSubscriptionBilling,
		AccountRateMultiplier: input.AccountRateMultiplier,
		APIKeyService:         input.APIKeyService,
	}, s.billingDeps(), s.usageBillingRepo)
	if err != nil {
		return err
	}

	writeUsageLogBestEffort(ctx, s.usageLogRepo, input.UsageLog, "service.gateway")
	return nil
}
