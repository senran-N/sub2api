package service

import "context"

// RecordUsageInput 记录使用量的输入参数
type RecordUsageInput struct {
	Result             *ForwardResult
	APIKey             *APIKey
	User               *User
	Account            *Account
	Subscription       *UserSubscription
	InboundEndpoint    string
	UpstreamEndpoint   string
	UserAgent          string
	IPAddress          string
	RequestPayloadHash string
	ForceCacheBilling  bool
	APIKeyService      APIKeyQuotaUpdater
	ChannelUsageFields
}

// APIKeyQuotaUpdater defines the interface for updating API Key quota and rate limit usage.
type APIKeyQuotaUpdater interface {
	UpdateQuotaUsed(ctx context.Context, apiKeyID int64, cost float64) error
	UpdateRateLimitUsage(ctx context.Context, apiKeyID int64, cost float64) error
}

type apiKeyAuthCacheInvalidator interface {
	InvalidateAuthCacheByKey(ctx context.Context, key string)
}

type usageLogBestEffortWriter interface {
	CreateBestEffort(ctx context.Context, log *UsageLog) error
}

// postUsageBillingParams 统一扣费所需的参数
type postUsageBillingParams struct {
	Cost                  *CostBreakdown
	User                  *User
	APIKey                *APIKey
	Account               *Account
	Subscription          *UserSubscription
	RequestPayloadHash    string
	IsSubscriptionBill    bool
	AccountRateMultiplier float64
	APIKeyService         APIKeyQuotaUpdater
}

// RecordUsageLongContextInput 记录使用量的输入参数（支持长上下文双倍计费）
type RecordUsageLongContextInput struct {
	Result                *ForwardResult
	APIKey                *APIKey
	User                  *User
	Account               *Account
	Subscription          *UserSubscription
	InboundEndpoint       string
	UpstreamEndpoint      string
	UserAgent             string
	IPAddress             string
	RequestPayloadHash    string
	LongContextThreshold  int
	LongContextMultiplier float64
	ForceCacheBilling     bool
	APIKeyService         APIKeyQuotaUpdater
	ChannelUsageFields
}
