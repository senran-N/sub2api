package domain

// GenerateAuthURLResult contains the authorization URL and session info.
type GenerateAuthURLResult struct {
	AuthURL   string `json:"auth_url"`
	SessionID string `json:"session_id"`
}

// ExchangeCodeInput represents the input for code exchange.
type ExchangeCodeInput struct {
	SessionID string
	Code      string
	ProxyID   *int64
}

// TokenInfo represents the token information stored in credentials.
type TokenInfo struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	ExpiresAt    int64  `json:"expires_at"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
	OrgUUID      string `json:"org_uuid,omitempty"`
	AccountUUID  string `json:"account_uuid,omitempty"`
	EmailAddress string `json:"email_address,omitempty"`
}

// CookieAuthInput represents the input for cookie-based authentication.
type CookieAuthInput struct {
	SessionKey string
	ProxyID    *int64
	Scope      string // "full" or "inference"
}
