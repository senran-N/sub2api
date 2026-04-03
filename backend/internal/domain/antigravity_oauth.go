package domain

// AntigravityAuthURLResult is the result of generating an authorization URL.
type AntigravityAuthURLResult struct {
	AuthURL   string `json:"auth_url"`
	SessionID string `json:"session_id"`
	State     string `json:"state"`
}

// AntigravityExchangeCodeInput is the input for authorization-code exchange.
type AntigravityExchangeCodeInput struct {
	SessionID string
	State     string
	Code      string
	ProxyID   *int64
}

// AntigravityTokenInfo stores Antigravity OAuth token details.
type AntigravityTokenInfo struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	ExpiresIn        int64  `json:"expires_in"`
	ExpiresAt        int64  `json:"expires_at"`
	TokenType        string `json:"token_type"`
	Email            string `json:"email,omitempty"`
	ProjectID        string `json:"project_id,omitempty"`
	ProjectIDMissing bool   `json:"-"`
	PlanType         string `json:"-"`
}
