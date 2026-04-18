package domain

import "time"

// APIKeyAuthSnapshot is the minimal auth cache payload for API key authentication.
type APIKeyAuthSnapshot struct {
	Version     int64                    `json:"version"`
	APIKeyID    int64                    `json:"api_key_id"`
	UserID      int64                    `json:"user_id"`
	GroupID     *int64                   `json:"group_id,omitempty"`
	Status      string                   `json:"status"`
	IPWhitelist []string                 `json:"ip_whitelist,omitempty"`
	IPBlacklist []string                 `json:"ip_blacklist,omitempty"`
	User        APIKeyAuthUserSnapshot   `json:"user"`
	Group       *APIKeyAuthGroupSnapshot `json:"group,omitempty"`

	Quota     float64 `json:"quota"`
	QuotaUsed float64 `json:"quota_used"`

	ExpiresAt *time.Time `json:"expires_at,omitempty"`

	RateLimit5h float64 `json:"rate_limit_5h"`
	RateLimit1d float64 `json:"rate_limit_1d"`
	RateLimit7d float64 `json:"rate_limit_7d"`
}

// APIKeyAuthUserSnapshot is the user subset needed by API key auth.
type APIKeyAuthUserSnapshot struct {
	ID          int64   `json:"id"`
	Status      string  `json:"status"`
	Role        string  `json:"role"`
	Balance     float64 `json:"balance"`
	Concurrency int     `json:"concurrency"`
}

// APIKeyAuthGroupSnapshot is the group subset needed by API key auth.
type APIKeyAuthGroupSnapshot struct {
	ID                              int64                             `json:"id"`
	Name                            string                            `json:"name"`
	Platform                        string                            `json:"platform"`
	Status                          string                            `json:"status"`
	SubscriptionType                string                            `json:"subscription_type"`
	RateMultiplier                  float64                           `json:"rate_multiplier"`
	DailyLimitUSD                   *float64                          `json:"daily_limit_usd,omitempty"`
	WeeklyLimitUSD                  *float64                          `json:"weekly_limit_usd,omitempty"`
	MonthlyLimitUSD                 *float64                          `json:"monthly_limit_usd,omitempty"`
	ImagePrice1K                    *float64                          `json:"image_price_1k,omitempty"`
	ImagePrice2K                    *float64                          `json:"image_price_2k,omitempty"`
	ImagePrice4K                    *float64                          `json:"image_price_4k,omitempty"`
	ClaudeCodeOnly                  bool                              `json:"claude_code_only"`
	FallbackGroupID                 *int64                            `json:"fallback_group_id,omitempty"`
	FallbackGroupIDOnInvalidRequest *int64                            `json:"fallback_group_id_on_invalid_request,omitempty"`
	ModelRouting                    map[string][]int64                `json:"model_routing,omitempty"`
	ModelRoutingEnabled             bool                              `json:"model_routing_enabled"`
	MCPXMLInject                    bool                              `json:"mcp_xml_inject"`
	SupportedModelScopes            []string                          `json:"supported_model_scopes,omitempty"`
	AllowMessagesDispatch           bool                              `json:"allow_messages_dispatch"`
	DefaultMappedModel              string                            `json:"default_mapped_model,omitempty"`
	MessagesDispatchModelConfig     OpenAIMessagesDispatchModelConfig `json:"messages_dispatch_model_config,omitempty"`
}

// APIKeyAuthCacheEntry is a cached auth result with negative-cache support.
type APIKeyAuthCacheEntry struct {
	NotFound bool                `json:"not_found"`
	Snapshot *APIKeyAuthSnapshot `json:"snapshot,omitempty"`
}

const APIKeyAuthSnapshotVersion int64 = 2
