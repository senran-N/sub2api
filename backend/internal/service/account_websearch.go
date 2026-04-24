package service

const (
	WebSearchModeDefault  = "default"
	WebSearchModeEnabled  = "enabled"
	WebSearchModeDisabled = "disabled"
)

// GetWebSearchEmulationMode returns the Anthropic API-key account web-search emulation mode.
func (a *Account) GetWebSearchEmulationMode() string {
	if a == nil || a.Platform != PlatformAnthropic || a.Type != AccountTypeAPIKey || a.Extra == nil {
		return WebSearchModeDefault
	}
	value, _ := a.Extra["web_search_emulation"].(string)
	switch value {
	case WebSearchModeEnabled, WebSearchModeDisabled:
		return value
	default:
		return WebSearchModeDefault
	}
}
