package domain

// GeminiOAuthCapabilities describes the current Gemini OAuth configuration support.
type GeminiOAuthCapabilities struct {
	AIStudioOAuthEnabled bool     `json:"ai_studio_oauth_enabled"`
	RequiredRedirectURIs []string `json:"required_redirect_uris"`
}
