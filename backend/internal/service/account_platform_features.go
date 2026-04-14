package service

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/senran-N/sub2api/internal/config"
)

func (a *Account) IsPrivacySet() bool {
	switch a.Platform {
	case PlatformOpenAI:
		return a.getExtraString("privacy_mode") == PrivacyModeTrainingOff
	case PlatformAntigravity:
		return a.getExtraString("privacy_mode") == AntigravityPrivacySet
	default:
		return true
	}
}

func (a *Account) IsGemini() bool {
	return a.Platform == PlatformGemini
}

func (a *Account) GeminiOAuthType() string {
	if a.Platform != PlatformGemini || a.Type != AccountTypeOAuth {
		return ""
	}
	oauthType := strings.TrimSpace(a.GetCredential("oauth_type"))
	if oauthType == "" && strings.TrimSpace(a.GetCredential("project_id")) != "" {
		return "code_assist"
	}
	return oauthType
}

func (a *Account) GeminiTierID() string {
	return strings.TrimSpace(a.GetCredential("tier_id"))
}

func (a *Account) IsGeminiCodeAssist() bool {
	if a.Platform != PlatformGemini || a.Type != AccountTypeOAuth {
		return false
	}
	oauthType := a.GeminiOAuthType()
	if oauthType == "" {
		return strings.TrimSpace(a.GetCredential("project_id")) != ""
	}
	return oauthType == "code_assist"
}

func (a *Account) CanGetUsage() bool {
	return a.Type == AccountTypeOAuth
}

func (a *Account) GetBaseURL() string {
	if a.Type != AccountTypeAPIKey && a.Type != AccountTypeUpstream {
		return ""
	}
	baseURL := a.GetCredential("base_url")
	if baseURL == "" {
		return "https://api.anthropic.com"
	}
	if a.Platform == PlatformAntigravity {
		return strings.TrimRight(baseURL, "/") + "/antigravity"
	}
	return baseURL
}

func (a *Account) GetGeminiBaseURL(defaultBaseURL string) string {
	baseURL := strings.TrimSpace(a.GetCredential("base_url"))
	if baseURL == "" {
		return defaultBaseURL
	}
	if a.Platform == PlatformAntigravity && a.Type == AccountTypeAPIKey {
		return strings.TrimRight(baseURL, "/") + "/antigravity"
	}
	return baseURL
}

func (a *Account) GetClaudeUserID() string {
	if value := strings.TrimSpace(a.GetExtraString("claude_user_id")); value != "" {
		return value
	}
	if value := strings.TrimSpace(a.GetExtraString("anthropic_user_id")); value != "" {
		return value
	}
	if value := strings.TrimSpace(a.GetCredential("claude_user_id")); value != "" {
		return value
	}
	if value := strings.TrimSpace(a.GetCredential("anthropic_user_id")); value != "" {
		return value
	}
	return ""
}

func (a *Account) IsCustomErrorCodesEnabled() bool {
	if a.Type != AccountTypeAPIKey || a.Credentials == nil {
		return false
	}
	if value, ok := a.Credentials["custom_error_codes_enabled"]; ok {
		if enabled, ok := value.(bool); ok {
			return enabled
		}
	}
	return false
}

func (a *Account) IsPoolMode() bool {
	if !a.IsAPIKeyOrBedrock() || a.Credentials == nil {
		return false
	}
	if value, ok := a.Credentials["pool_mode"]; ok {
		if enabled, ok := value.(bool); ok {
			return enabled
		}
	}
	return false
}

const (
	defaultPoolModeRetryCount = 3
	maxPoolModeRetryCount     = 10
)

func (a *Account) GetPoolModeRetryCount() int {
	if a == nil || !a.IsPoolMode() || a.Credentials == nil {
		return defaultPoolModeRetryCount
	}
	raw, ok := a.Credentials["pool_mode_retry_count"]
	if !ok || raw == nil {
		return defaultPoolModeRetryCount
	}
	count := parsePoolModeRetryCount(raw)
	if count < 0 {
		return 0
	}
	if count > maxPoolModeRetryCount {
		return maxPoolModeRetryCount
	}
	return count
}

func parsePoolModeRetryCount(value any) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	case json.Number:
		if parsed, err := typed.Int64(); err == nil {
			return int(parsed)
		}
	case string:
		if parsed, err := strconv.Atoi(strings.TrimSpace(typed)); err == nil {
			return parsed
		}
	}
	return defaultPoolModeRetryCount
}

func isPoolModeRetryableStatus(statusCode int) bool {
	switch statusCode {
	case 401, 403, 429:
		return true
	default:
		return false
	}
}

func (a *Account) GetCustomErrorCodes() []int {
	if a.Credentials == nil {
		return nil
	}
	raw, ok := a.Credentials["custom_error_codes"]
	if !ok || raw == nil {
		return nil
	}
	items, ok := raw.([]any)
	if !ok {
		return nil
	}

	result := make([]int, 0, len(items))
	for _, item := range items {
		if value, ok := item.(float64); ok {
			result = append(result, int(value))
		}
	}
	return result
}

func (a *Account) ShouldHandleErrorCode(statusCode int) bool {
	if !a.IsCustomErrorCodesEnabled() {
		return true
	}
	codes := a.GetCustomErrorCodes()
	if len(codes) == 0 {
		return true
	}
	for _, code := range codes {
		if code == statusCode {
			return true
		}
	}
	return false
}

func (a *Account) IsInterceptWarmupEnabled() bool {
	if a.Credentials == nil {
		return false
	}
	if value, ok := a.Credentials["intercept_warmup_requests"]; ok {
		if enabled, ok := value.(bool); ok {
			return enabled
		}
	}
	return false
}

func (a *Account) IsBedrock() bool {
	return a.Platform == PlatformAnthropic && a.Type == AccountTypeBedrock
}

func (a *Account) IsBedrockAPIKey() bool {
	return a.IsBedrock() && a.GetCredential("auth_mode") == "apikey"
}

func (a *Account) IsAPIKeyOrBedrock() bool {
	return a.Type == AccountTypeAPIKey || a.Type == AccountTypeBedrock
}

func (a *Account) IsOpenAI() bool {
	return a.Platform == PlatformOpenAI
}

func (a *Account) IsAnthropic() bool {
	return a.Platform == PlatformAnthropic
}

func (a *Account) IsOpenAIOAuth() bool {
	return a.IsOpenAI() && a.Type == AccountTypeOAuth
}

func (a *Account) IsOpenAIApiKey() bool {
	return a.IsOpenAI() && a.Type == AccountTypeAPIKey
}

func (a *Account) GetOpenAIBaseURL() string {
	if !a.IsOpenAI() {
		return ""
	}
	if a.Type == AccountTypeAPIKey || a.Type == AccountTypeUpstream {
		if baseURL := a.GetCredential("base_url"); baseURL != "" {
			return baseURL
		}
	}
	return "https://api.openai.com"
}

func (a *Account) GetOpenAIAccessToken() string {
	if !a.IsOpenAI() {
		return ""
	}
	return a.GetCredential("access_token")
}

func (a *Account) GetOpenAIRefreshToken() string {
	if !a.IsOpenAIOAuth() {
		return ""
	}
	return a.GetCredential("refresh_token")
}

func (a *Account) GetOpenAIIDToken() string {
	if !a.IsOpenAIOAuth() {
		return ""
	}
	return a.GetCredential("id_token")
}

func (a *Account) GetOpenAIApiKey() string {
	if !a.IsOpenAI() || (a.Type != AccountTypeAPIKey && a.Type != AccountTypeUpstream) {
		return ""
	}
	return a.GetCredential("api_key")
}

func (a *Account) GetOpenAIUserAgent() string {
	if !a.IsOpenAI() {
		return ""
	}
	return a.GetCredential("user_agent")
}

func (a *Account) GetOpenAICodexOriginator() string {
	if !a.IsOpenAI() {
		return ""
	}
	return a.GetCredential("originator")
}

func (a *Account) GetOpenAICodexVersion() string {
	if !a.IsOpenAI() {
		return ""
	}
	return a.GetCredential("version")
}

func (a *Account) GetOpenAICodexBetaFeatures() string {
	if !a.IsOpenAI() {
		return ""
	}
	return a.GetCredential("codex_beta_features")
}

func (a *Account) GetOpenAICodexAcceptLanguage() string {
	if !a.IsOpenAI() {
		return ""
	}
	return a.GetCredential("accept_language")
}

func (a *Account) GetChatGPTAccountID() string {
	if !a.IsOpenAIOAuth() {
		return ""
	}
	return a.GetCredential("chatgpt_account_id")
}

func (a *Account) GetChatGPTUserID() string {
	if !a.IsOpenAIOAuth() {
		return ""
	}
	return a.GetCredential("chatgpt_user_id")
}

func (a *Account) GetOpenAIOrganizationID() string {
	if !a.IsOpenAIOAuth() {
		return ""
	}
	return a.GetCredential("organization_id")
}

func (a *Account) GetOpenAITokenExpiresAt() *time.Time {
	if !a.IsOpenAIOAuth() {
		return nil
	}
	return a.GetCredentialAsTime("expires_at")
}

func (a *Account) IsOpenAITokenExpired() bool {
	expiresAt := a.GetOpenAITokenExpiresAt()
	if expiresAt == nil {
		return false
	}
	return time.Now().Add(60 * time.Second).After(*expiresAt)
}

func (a *Account) IsMixedSchedulingEnabled() bool {
	if a.Platform != PlatformAntigravity || a.Extra == nil {
		return false
	}
	if value, ok := a.Extra["mixed_scheduling"]; ok {
		if enabled, ok := value.(bool); ok {
			return enabled
		}
	}
	return false
}

func (a *Account) IsOveragesEnabled() bool {
	if a.Platform != PlatformAntigravity || a.Extra == nil {
		return false
	}
	if value, ok := a.Extra["allow_overages"]; ok {
		if enabled, ok := value.(bool); ok {
			return enabled
		}
	}
	return false
}

func (a *Account) IsOpenAIPassthroughEnabled() bool {
	if a == nil || !a.IsOpenAI() || a.Extra == nil {
		return false
	}
	if enabled, ok := a.Extra["openai_passthrough"].(bool); ok {
		return enabled
	}
	if enabled, ok := a.Extra["openai_oauth_passthrough"].(bool); ok {
		return enabled
	}
	return false
}

func (a *Account) IsOpenAIResponsesWebSocketV2Enabled() bool {
	if a == nil || !a.IsOpenAI() || a.Extra == nil {
		return false
	}
	if a.IsOpenAIOAuth() {
		if enabled, ok := a.Extra["openai_oauth_responses_websockets_v2_enabled"].(bool); ok {
			return enabled
		}
	}
	if a.IsOpenAIApiKey() {
		if enabled, ok := a.Extra["openai_apikey_responses_websockets_v2_enabled"].(bool); ok {
			return enabled
		}
	}
	if enabled, ok := a.Extra["responses_websockets_v2_enabled"].(bool); ok {
		return enabled
	}
	if enabled, ok := a.Extra["openai_ws_enabled"].(bool); ok {
		return enabled
	}
	return false
}

const (
	OpenAIWSIngressModeOff         = "off"
	OpenAIWSIngressModeShared      = "shared"
	OpenAIWSIngressModeDedicated   = "dedicated"
	OpenAIWSIngressModeCtxPool     = "ctx_pool"
	OpenAIWSIngressModePassthrough = "passthrough"
)

func normalizeOpenAIWSIngressMode(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case OpenAIWSIngressModeOff:
		return OpenAIWSIngressModeOff
	case OpenAIWSIngressModeCtxPool:
		return OpenAIWSIngressModeCtxPool
	case OpenAIWSIngressModePassthrough:
		return OpenAIWSIngressModePassthrough
	case OpenAIWSIngressModeShared:
		return OpenAIWSIngressModeShared
	case OpenAIWSIngressModeDedicated:
		return OpenAIWSIngressModeDedicated
	default:
		return ""
	}
}

func normalizeOpenAIWSIngressDefaultMode(mode string) string {
	if normalized := normalizeOpenAIWSIngressMode(mode); normalized != "" {
		if normalized == OpenAIWSIngressModeShared || normalized == OpenAIWSIngressModeDedicated {
			return OpenAIWSIngressModeCtxPool
		}
		return normalized
	}
	return OpenAIWSIngressModeCtxPool
}

func (a *Account) ResolveOpenAIResponsesWebSocketV2Mode(defaultMode string) string {
	resolvedDefault := normalizeOpenAIWSIngressDefaultMode(defaultMode)
	if a == nil || !a.IsOpenAI() {
		return OpenAIWSIngressModeOff
	}
	if a.Extra == nil {
		return resolvedDefault
	}

	resolveModeString := func(key string) (string, bool) {
		raw, ok := a.Extra[key]
		if !ok {
			return "", false
		}
		mode, ok := raw.(string)
		if !ok {
			return "", false
		}
		normalized := normalizeOpenAIWSIngressMode(mode)
		if normalized == "" {
			return "", false
		}
		return normalized, true
	}
	resolveBoolMode := func(key string) (string, bool) {
		raw, ok := a.Extra[key]
		if !ok {
			return "", false
		}
		enabled, ok := raw.(bool)
		if !ok {
			return "", false
		}
		if enabled {
			return OpenAIWSIngressModeCtxPool, true
		}
		return OpenAIWSIngressModeOff, true
	}

	if a.IsOpenAIOAuth() {
		if mode, ok := resolveModeString("openai_oauth_responses_websockets_v2_mode"); ok {
			return mode
		}
		if mode, ok := resolveBoolMode("openai_oauth_responses_websockets_v2_enabled"); ok {
			return mode
		}
	}
	if a.IsOpenAIApiKey() {
		if mode, ok := resolveModeString("openai_apikey_responses_websockets_v2_mode"); ok {
			return mode
		}
		if mode, ok := resolveBoolMode("openai_apikey_responses_websockets_v2_enabled"); ok {
			return mode
		}
	}
	if mode, ok := resolveBoolMode("responses_websockets_v2_enabled"); ok {
		return mode
	}
	if mode, ok := resolveBoolMode("openai_ws_enabled"); ok {
		return mode
	}
	if resolvedDefault == OpenAIWSIngressModeShared || resolvedDefault == OpenAIWSIngressModeDedicated {
		return OpenAIWSIngressModeCtxPool
	}
	return resolvedDefault
}

func (a *Account) IsOpenAIWSForceHTTPEnabled() bool {
	if a == nil || !a.IsOpenAI() || a.Extra == nil {
		return false
	}
	enabled, ok := a.Extra["openai_ws_force_http"].(bool)
	return ok && enabled
}

func (a *Account) IsOpenAIWSAllowStoreRecoveryEnabled() bool {
	if a == nil || !a.IsOpenAI() || a.Extra == nil {
		return false
	}
	enabled, ok := a.Extra["openai_ws_allow_store_recovery"].(bool)
	return ok && enabled
}

func (a *Account) IsOpenAIOAuthPassthroughEnabled() bool {
	return a != nil && a.IsOpenAIOAuth() && a.IsOpenAIPassthroughEnabled()
}

func (a *Account) IsAnthropicAPIKeyPassthroughEnabled() bool {
	if a == nil || a.Platform != PlatformAnthropic || a.Type != AccountTypeAPIKey || a.Extra == nil {
		return false
	}
	enabled, ok := a.Extra["anthropic_passthrough"].(bool)
	return ok && enabled
}

func (a *Account) IsCodexCLIOnlyEnabled() bool {
	if a == nil || !a.IsOpenAIOAuth() || a.Extra == nil {
		return false
	}
	enabled, ok := a.Extra["codex_cli_only"].(bool)
	return ok && enabled
}

type WindowCostSchedulability int

const (
	WindowCostSchedulable WindowCostSchedulability = iota
	WindowCostStickyOnly
	WindowCostNotSchedulable
)

func (a *Account) IsAnthropicOAuthOrSetupToken() bool {
	return a.Platform == PlatformAnthropic && (a.Type == AccountTypeOAuth || a.Type == AccountTypeSetupToken)
}

func (a *Account) IsTLSFingerprintEnabled() bool {
	if !a.IsAnthropicOAuthOrSetupToken() || a.Extra == nil {
		return false
	}
	if value, ok := a.Extra["enable_tls_fingerprint"]; ok {
		if enabled, ok := value.(bool); ok {
			return enabled
		}
	}
	return false
}

func (a *Account) GetTLSFingerprintProfileID() int64 {
	if a.Extra == nil {
		return 0
	}
	value, ok := a.Extra["tls_fingerprint_profile_id"]
	if !ok {
		return 0
	}
	return a.parseInt64ExtraValue(value)
}

func (a *Account) parseInt64ExtraValue(value any) int64 {
	switch typed := value.(type) {
	case float64:
		return int64(typed)
	case int64:
		return typed
	case int:
		return int64(typed)
	default:
		return int64(ParseExtraInt(value))
	}
}

func (a *Account) GetUserMsgQueueMode() string {
	if a.Extra == nil {
		return ""
	}
	if mode, ok := a.Extra["user_msg_queue_mode"].(string); ok && mode != "" {
		if mode == config.UMQModeSerialize || mode == config.UMQModeThrottle {
			return mode
		}
		return ""
	}
	if enabled, ok := a.Extra["user_msg_queue_enabled"].(bool); ok && enabled {
		return config.UMQModeSerialize
	}
	return ""
}

func (a *Account) IsSessionIDMaskingEnabled() bool {
	if !a.IsAnthropicOAuthOrSetupToken() || a.Extra == nil {
		return false
	}
	if value, ok := a.Extra["session_id_masking_enabled"]; ok {
		if enabled, ok := value.(bool); ok {
			return enabled
		}
	}
	return false
}

func (a *Account) IsCustomBaseURLEnabled() bool {
	if !a.IsAnthropicOAuthOrSetupToken() || a.Extra == nil {
		return false
	}
	if value, ok := a.Extra["custom_base_url_enabled"]; ok {
		if enabled, ok := value.(bool); ok {
			return enabled
		}
	}
	return false
}

func (a *Account) GetCustomBaseURL() string {
	return a.GetExtraString("custom_base_url")
}

func (a *Account) IsCacheTTLOverrideEnabled() bool {
	if !a.IsAnthropicOAuthOrSetupToken() || a.Extra == nil {
		return false
	}
	if value, ok := a.Extra["cache_ttl_override_enabled"]; ok {
		if enabled, ok := value.(bool); ok {
			return enabled
		}
	}
	return false
}

func (a *Account) GetCacheTTLOverrideTarget() string {
	if a.Extra == nil {
		return "5m"
	}
	if value, ok := a.Extra["cache_ttl_override_target"]; ok {
		if target, ok := value.(string); ok && (target == "5m" || target == "1h") {
			return target
		}
	}
	return "5m"
}
