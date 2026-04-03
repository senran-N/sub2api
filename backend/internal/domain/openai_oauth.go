package domain

// OpenAIAuthURLResult contains the authorization URL and session info.
type OpenAIAuthURLResult struct {
	AuthURL   string `json:"auth_url"`
	SessionID string `json:"session_id"`
}

// OpenAIExchangeCodeInput represents the input for code exchange.
type OpenAIExchangeCodeInput struct {
	SessionID   string
	Code        string
	State       string
	RedirectURI string
	ProxyID     *int64
}

// OpenAITokenInfo represents the token information for OpenAI.
type OpenAITokenInfo struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	IDToken          string `json:"id_token,omitempty"`
	ExpiresIn        int64  `json:"expires_in"`
	ExpiresAt        int64  `json:"expires_at"`
	ClientID         string `json:"client_id,omitempty"`
	Email            string `json:"email,omitempty"`
	ChatGPTAccountID string `json:"chatgpt_account_id,omitempty"`
	ChatGPTUserID    string `json:"chatgpt_user_id,omitempty"`
	OrganizationID   string `json:"organization_id,omitempty"`
	PlanType         string `json:"plan_type,omitempty"`
	PrivacyMode      string `json:"privacy_mode,omitempty"`
}
