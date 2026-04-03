package domain

import "time"

// CreateAPIKeyRequest describes an API key creation request.
type CreateAPIKeyRequest struct {
	Name          string   `json:"name"`
	GroupID       *int64   `json:"group_id"`
	CustomKey     *string  `json:"custom_key"`
	IPWhitelist   []string `json:"ip_whitelist"`
	IPBlacklist   []string `json:"ip_blacklist"`
	Quota         float64  `json:"quota"`
	ExpiresInDays *int     `json:"expires_in_days"`
	RateLimit5h   float64  `json:"rate_limit_5h"`
	RateLimit1d   float64  `json:"rate_limit_1d"`
	RateLimit7d   float64  `json:"rate_limit_7d"`
}

// UpdateAPIKeyRequest describes an API key update request.
type UpdateAPIKeyRequest struct {
	Name                *string    `json:"name"`
	GroupID             *int64     `json:"group_id"`
	Status              *string    `json:"status"`
	IPWhitelist         []string   `json:"ip_whitelist"`
	IPBlacklist         []string   `json:"ip_blacklist"`
	Quota               *float64   `json:"quota"`
	ExpiresAt           *time.Time `json:"expires_at"`
	ClearExpiration     bool       `json:"-"`
	ResetQuota          *bool      `json:"reset_quota"`
	RateLimit5h         *float64   `json:"rate_limit_5h"`
	RateLimit1d         *float64   `json:"rate_limit_1d"`
	RateLimit7d         *float64   `json:"rate_limit_7d"`
	ResetRateLimitUsage *bool      `json:"reset_rate_limit_usage"`
}
