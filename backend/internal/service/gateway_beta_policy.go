package service

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/senran-N/sub2api/internal/pkg/claude"
	"github.com/tidwall/gjson"
)

func (s *GatewayService) getBetaHeader(modelID string, clientBetaHeader string) string {
	if clientBetaHeader != "" {
		if strings.Contains(clientBetaHeader, claude.OAuthBetaToken()) {
			return clientBetaHeader
		}

		parts := strings.Split(clientBetaHeader, ",")
		for index, part := range parts {
			parts[index] = strings.TrimSpace(part)
		}

		claudeCodeIndex := -1
		for index, part := range parts {
			if part == claude.ClaudeCodeBetaToken() {
				claudeCodeIndex = index
				break
			}
		}

		if claudeCodeIndex >= 0 {
			withOAuth := make([]string, 0, len(parts)+1)
			withOAuth = append(withOAuth, parts[:claudeCodeIndex+1]...)
			withOAuth = append(withOAuth, claude.OAuthBetaToken())
			withOAuth = append(withOAuth, parts[claudeCodeIndex+1:]...)
			return strings.Join(withOAuth, ",")
		}

		return claude.OAuthBetaToken() + "," + clientBetaHeader
	}

	if strings.Contains(strings.ToLower(modelID), "haiku") {
		return claude.HaikuAnthropicBetaHeader()
	}

	return claude.DefaultAnthropicBetaHeader()
}

func requestNeedsBetaFeatures(body []byte) bool {
	tools := gjson.GetBytes(body, "tools")
	if tools.Exists() && tools.IsArray() && len(tools.Array()) > 0 {
		return true
	}

	thinkingType := gjson.GetBytes(body, "thinking.type").String()
	return strings.EqualFold(thinkingType, "enabled") || strings.EqualFold(thinkingType, "adaptive")
}

func defaultAPIKeyBetaHeader(body []byte) string {
	modelID := gjson.GetBytes(body, "model").String()
	if strings.Contains(strings.ToLower(modelID), "haiku") {
		return claude.APIKeyHaikuAnthropicBetaHeader()
	}
	return claude.APIKeyAnthropicBetaHeader()
}

func applyClaudeOAuthHeaderDefaults(req *http.Request) {
	if req == nil {
		return
	}
	if getHeaderRaw(req.Header, "Accept") == "" {
		setHeaderRaw(req.Header, "Accept", "application/json")
	}
	for key, value := range claude.StableHeaders() {
		if value == "" {
			continue
		}
		if getHeaderRaw(req.Header, key) == "" {
			setHeaderRaw(req.Header, resolveWireCasing(key), value)
		}
	}
}

func mergeAnthropicBeta(required []string, incoming string) string {
	seen := make(map[string]struct{}, len(required)+8)
	merged := make([]string, 0, len(required)+8)

	add := func(value string) {
		value = strings.TrimSpace(value)
		if value == "" {
			return
		}
		if _, exists := seen[value]; exists {
			return
		}
		seen[value] = struct{}{}
		merged = append(merged, value)
	}

	for _, value := range required {
		add(value)
	}
	for _, value := range strings.Split(incoming, ",") {
		add(value)
	}

	return strings.Join(merged, ",")
}

func mergeAnthropicBetaDropping(required []string, incoming string, drop map[string]struct{}) string {
	merged := mergeAnthropicBeta(required, incoming)
	if merged == "" || len(drop) == 0 {
		return merged
	}

	filtered := make([]string, 0, 8)
	for _, value := range strings.Split(merged, ",") {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, exists := drop[value]; exists {
			continue
		}
		filtered = append(filtered, value)
	}
	return strings.Join(filtered, ",")
}

func stripBetaTokens(header string, tokens []string) string {
	if header == "" || len(tokens) == 0 {
		return header
	}
	return stripBetaTokensWithSet(header, buildBetaTokenSet(tokens))
}

func stripBetaTokensWithSet(header string, drop map[string]struct{}) string {
	if header == "" || len(drop) == 0 {
		return header
	}

	parts := strings.Split(header, ",")
	filtered := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if _, exists := drop[part]; exists {
			continue
		}
		filtered = append(filtered, part)
	}
	if len(filtered) == len(parts) {
		return header
	}
	return strings.Join(filtered, ",")
}

type BetaBlockedError struct {
	Message string
}

func (e *BetaBlockedError) Error() string { return e.Message }

type betaPolicyResult struct {
	blockErr  *BetaBlockedError
	filterSet map[string]struct{}
}

func (s *GatewayService) evaluateBetaPolicy(ctx context.Context, betaHeader string, account *Account) betaPolicyResult {
	if s.settingService == nil {
		return betaPolicyResult{}
	}

	settings, err := s.settingService.GetBetaPolicySettings(ctx)
	if err != nil || settings == nil {
		return betaPolicyResult{}
	}

	isOAuth := account.IsOAuth()
	isBedrock := account.IsBedrock()
	var result betaPolicyResult

	for _, rule := range settings.Rules {
		if !betaPolicyScopeMatches(rule.Scope, isOAuth, isBedrock) {
			continue
		}

		switch rule.Action {
		case BetaPolicyActionBlock:
			if result.blockErr == nil && betaHeader != "" && containsBetaToken(betaHeader, rule.BetaToken) {
				message := rule.ErrorMessage
				if message == "" {
					message = "beta feature " + rule.BetaToken + " is not allowed"
				}
				result.blockErr = &BetaBlockedError{Message: message}
			}
		case BetaPolicyActionFilter:
			if result.filterSet == nil {
				result.filterSet = make(map[string]struct{})
			}
			result.filterSet[rule.BetaToken] = struct{}{}
		}
	}

	return result
}

func mergeDropSets(policySet map[string]struct{}, extra ...string) map[string]struct{} {
	if len(policySet) == 0 && len(extra) == 0 {
		return defaultDroppedBetasSet
	}

	merged := make(map[string]struct{}, len(defaultDroppedBetasSet)+len(policySet)+len(extra))
	for token := range defaultDroppedBetasSet {
		merged[token] = struct{}{}
	}
	for token := range policySet {
		merged[token] = struct{}{}
	}
	for _, token := range extra {
		merged[token] = struct{}{}
	}

	return merged
}

const betaPolicyFilterSetKey = "betaPolicyFilterSet"

func (s *GatewayService) getBetaPolicyFilterSet(ctx context.Context, c *gin.Context, account *Account) map[string]struct{} {
	if c != nil {
		if value, ok := c.Get(betaPolicyFilterSetKey); ok {
			if filterSet, ok := value.(map[string]struct{}); ok {
				return filterSet
			}
		}
	}
	return s.evaluateBetaPolicy(ctx, "", account).filterSet
}

func betaPolicyScopeMatches(scope string, isOAuth bool, isBedrock bool) bool {
	switch scope {
	case BetaPolicyScopeAll:
		return true
	case BetaPolicyScopeOAuth:
		return isOAuth
	case BetaPolicyScopeAPIKey:
		return !isOAuth && !isBedrock
	case BetaPolicyScopeBedrock:
		return isBedrock
	default:
		return true
	}
}

func droppedBetaSet(extra ...string) map[string]struct{} {
	merged := make(map[string]struct{}, len(defaultDroppedBetasSet)+len(extra))
	for token := range defaultDroppedBetasSet {
		merged[token] = struct{}{}
	}
	for _, token := range extra {
		merged[token] = struct{}{}
	}
	return merged
}

func containsBetaToken(header, token string) bool {
	if header == "" || token == "" {
		return false
	}

	for _, part := range strings.Split(header, ",") {
		if strings.TrimSpace(part) == token {
			return true
		}
	}
	return false
}

func filterBetaTokens(tokens []string, filterSet map[string]struct{}) []string {
	if len(tokens) == 0 || len(filterSet) == 0 {
		return tokens
	}

	filtered := make([]string, 0, len(tokens))
	for _, token := range tokens {
		if _, blocked := filterSet[token]; blocked {
			continue
		}
		filtered = append(filtered, token)
	}
	return filtered
}

func (s *GatewayService) resolveBedrockBetaTokensForRequest(
	ctx context.Context,
	account *Account,
	betaHeader string,
	body []byte,
	modelID string,
) ([]string, error) {
	policy := s.evaluateBetaPolicy(ctx, betaHeader, account)
	if policy.blockErr != nil {
		return nil, policy.blockErr
	}

	betaTokens := ResolveBedrockBetaTokens(betaHeader, body, modelID)
	if blockErr := s.checkBetaPolicyBlockForTokens(ctx, betaTokens, account); blockErr != nil {
		return nil, blockErr
	}

	return filterBetaTokens(betaTokens, policy.filterSet), nil
}

func (s *GatewayService) checkBetaPolicyBlockForTokens(ctx context.Context, tokens []string, account *Account) *BetaBlockedError {
	if s.settingService == nil || len(tokens) == 0 {
		return nil
	}

	settings, err := s.settingService.GetBetaPolicySettings(ctx)
	if err != nil || settings == nil {
		return nil
	}

	isOAuth := account.IsOAuth()
	isBedrock := account.IsBedrock()
	tokenSet := buildBetaTokenSet(tokens)

	for _, rule := range settings.Rules {
		if rule.Action != BetaPolicyActionBlock {
			continue
		}
		if !betaPolicyScopeMatches(rule.Scope, isOAuth, isBedrock) {
			continue
		}
		if _, present := tokenSet[rule.BetaToken]; !present {
			continue
		}

		message := rule.ErrorMessage
		if message == "" {
			message = "beta feature " + rule.BetaToken + " is not allowed"
		}
		return &BetaBlockedError{Message: message}
	}

	return nil
}

func buildBetaTokenSet(tokens []string) map[string]struct{} {
	set := make(map[string]struct{}, len(tokens))
	for _, token := range tokens {
		if token == "" {
			continue
		}
		set[token] = struct{}{}
	}
	return set
}

var defaultDroppedBetasSet = buildBetaTokenSet(claude.DroppedBetas)
