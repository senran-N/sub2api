package domain

// GeminiAuthURLResult is the result of generating a Gemini authorization URL.
type GeminiAuthURLResult struct {
	AuthURL   string `json:"auth_url"`
	SessionID string `json:"session_id"`
	State     string `json:"state"`
}

// GeminiExchangeCodeInput is the input for Gemini authorization-code exchange.
type GeminiExchangeCodeInput struct {
	SessionID string
	State     string
	Code      string
	ProxyID   *int64
	OAuthType string
	TierID    string
}

// GeminiTokenInfo stores Gemini OAuth token details.
type GeminiTokenInfo struct {
	AccessToken  string         `json:"access_token"`
	RefreshToken string         `json:"refresh_token"`
	ExpiresIn    int64          `json:"expires_in"`
	ExpiresAt    int64          `json:"expires_at"`
	TokenType    string         `json:"token_type"`
	Scope        string         `json:"scope,omitempty"`
	ProjectID    string         `json:"project_id,omitempty"`
	OAuthType    string         `json:"oauth_type,omitempty"`
	TierID       string         `json:"tier_id,omitempty"`
	Extra        map[string]any `json:"extra,omitempty"`
}
