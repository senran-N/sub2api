package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/senran-N/sub2api/internal/pkg/tlsfingerprint"
	"github.com/senran-N/sub2api/internal/util/soraerror"
)

const (
	soraMeAPIURL            = "https://sora.chatgpt.com/backend/me"
	soraBillingAPIURL       = "https://sora.chatgpt.com/backend/billing/subscriptions"
	soraInviteMineURL       = "https://sora.chatgpt.com/backend/project_y/invite/mine"
	soraBootstrapURL        = "https://sora.chatgpt.com/backend/m/bootstrap"
	soraRemainingURL        = "https://sora.chatgpt.com/backend/nf/check"
	defaultSoraTestCooldown = 10 * time.Second
)

type soraProbeStep struct {
	Name       string `json:"name"`
	Status     string `json:"status"`
	HTTPStatus int    `json:"http_status,omitempty"`
	ErrorCode  string `json:"error_code,omitempty"`
	Message    string `json:"message,omitempty"`
}

type soraProbeSummary struct {
	Status string          `json:"status"`
	Steps  []soraProbeStep `json:"steps"`
}

type soraProbeRecorder struct {
	steps []soraProbeStep
}

func (r *soraProbeRecorder) addStep(name, status string, httpStatus int, errorCode, message string) {
	r.steps = append(r.steps, soraProbeStep{
		Name:       name,
		Status:     status,
		HTTPStatus: httpStatus,
		ErrorCode:  strings.TrimSpace(errorCode),
		Message:    strings.TrimSpace(message),
	})
}

func (r *soraProbeRecorder) finalize() soraProbeSummary {
	meSuccess := false
	partial := false
	for _, step := range r.steps {
		if step.Name == "me" {
			meSuccess = strings.EqualFold(step.Status, "success")
			continue
		}
		if strings.EqualFold(step.Status, "failed") {
			partial = true
		}
	}

	status := "success"
	if !meSuccess {
		status = "failed"
	} else if partial {
		status = "partial_success"
	}

	return soraProbeSummary{
		Status: status,
		Steps:  append([]soraProbeStep(nil), r.steps...),
	}
}

func (s *AccountTestService) emitSoraProbeSummary(c *gin.Context, rec *soraProbeRecorder) {
	if rec == nil {
		return
	}
	summary := rec.finalize()
	code := ""
	for _, step := range summary.Steps {
		if strings.EqualFold(step.Status, "failed") && strings.TrimSpace(step.ErrorCode) != "" {
			code = step.ErrorCode
			break
		}
	}
	s.sendEvent(c, TestEvent{
		Type:   "sora_test_result",
		Status: summary.Status,
		Code:   code,
		Data:   summary,
	})
}

func (s *AccountTestService) acquireSoraTestPermit(accountID int64) (time.Duration, bool) {
	if accountID <= 0 {
		return 0, true
	}
	s.soraTestGuardMu.Lock()
	defer s.soraTestGuardMu.Unlock()

	if s.soraTestLastRun == nil {
		s.soraTestLastRun = make(map[int64]time.Time)
	}
	cooldown := s.soraTestCooldown
	if cooldown <= 0 {
		cooldown = defaultSoraTestCooldown
	}

	now := time.Now()
	if lastRun, ok := s.soraTestLastRun[accountID]; ok {
		elapsed := now.Sub(lastRun)
		if elapsed < cooldown {
			return cooldown - elapsed, false
		}
	}
	s.soraTestLastRun[accountID] = now
	return 0, true
}

func ceilSeconds(d time.Duration) int {
	if d <= 0 {
		return 1
	}
	sec := int(d / time.Second)
	if d%time.Second != 0 {
		sec++
	}
	if sec < 1 {
		sec = 1
	}
	return sec
}

// testSoraAPIKeyAccountConnection 测试 Sora apikey 类型账号的连通性。
// 向上游 base_url 发送轻量级 prompt-enhance 请求验证连通性和 API Key 有效性。
func (s *AccountTestService) testSoraAPIKeyAccountConnection(c *gin.Context, account *Account) error {
	ctx := c.Request.Context()

	apiKey := account.GetCredential("api_key")
	if apiKey == "" {
		return s.sendErrorAndEnd(c, "Sora apikey 账号缺少 api_key 凭证")
	}

	baseURL := account.GetBaseURL()
	if baseURL == "" {
		return s.sendErrorAndEnd(c, "Sora apikey 账号缺少 base_url")
	}

	normalizedBaseURL, err := s.validateUpstreamBaseURL(baseURL)
	if err != nil {
		return s.sendErrorAndEnd(c, fmt.Sprintf("base_url 无效: %s", err.Error()))
	}
	upstreamURL := strings.TrimSuffix(normalizedBaseURL, "/") + "/sora/v1/chat/completions"

	s.prepareTestStream(c)
	if wait, ok := s.acquireSoraTestPermit(account.ID); !ok {
		return s.sendErrorAndEnd(c, fmt.Sprintf("Sora 账号测试过于频繁，请 %d 秒后重试", ceilSeconds(wait)))
	}

	s.sendEvent(c, TestEvent{Type: "test_start", Model: "sora-upstream"})

	testPayload := map[string]any{
		"model":    "prompt-enhance-short-10s",
		"messages": []map[string]string{{"role": "user", "content": "test"}},
		"stream":   false,
	}
	payloadBytes, _ := json.Marshal(testPayload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, upstreamURL, bytes.NewReader(payloadBytes))
	if err != nil {
		return s.sendErrorAndEnd(c, "构建测试请求失败")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := s.httpUpstream.Do(req, accountTestProxyURL(account), account.ID, account.Concurrency)
	if err != nil {
		return s.sendErrorAndEnd(c, fmt.Sprintf("上游连接失败: %s", err.Error()))
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if resp.StatusCode == http.StatusOK {
		s.sendEvent(c, TestEvent{Type: "content", Text: fmt.Sprintf("上游连接成功 (%s)", upstreamURL)})
		s.sendEvent(c, TestEvent{Type: "content", Text: fmt.Sprintf("API Key 有效 (HTTP %d)", resp.StatusCode)})
		s.sendEvent(c, TestEvent{Type: "test_complete", Success: true})
		return nil
	}
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return s.sendErrorAndEnd(c, fmt.Sprintf("上游认证失败 (HTTP %d)，请检查 API Key 是否正确", resp.StatusCode))
	}
	if resp.StatusCode == http.StatusBadRequest {
		s.sendEvent(c, TestEvent{Type: "content", Text: fmt.Sprintf("上游连接成功 (%s)", upstreamURL)})
		s.sendEvent(c, TestEvent{Type: "content", Text: fmt.Sprintf("API Key 有效（上游返回 %d，参数校验错误属正常）", resp.StatusCode)})
		s.sendEvent(c, TestEvent{Type: "test_complete", Success: true})
		return nil
	}
	return s.sendErrorAndEnd(c, fmt.Sprintf("上游返回异常 HTTP %d: %s", resp.StatusCode, truncateSoraErrorBody(respBody, 256)))
}

// testSoraAccountConnection 测试 Sora 账号的连接
// OAuth 类型：调用 /backend/me 接口验证 access_token 有效性
// APIKey 类型：向上游 base_url 发送轻量级 prompt-enhance 请求验证连通性
func (s *AccountTestService) testSoraAccountConnection(c *gin.Context, account *Account) error {
	if account.Type == AccountTypeAPIKey {
		return s.testSoraAPIKeyAccountConnection(c, account)
	}

	ctx := c.Request.Context()
	recorder := &soraProbeRecorder{}
	authToken := account.GetCredential("access_token")
	if authToken == "" {
		recorder.addStep("me", "failed", http.StatusUnauthorized, "missing_access_token", "No access token available")
		s.emitSoraProbeSummary(c, recorder)
		return s.sendErrorAndEnd(c, "No access token available")
	}

	s.prepareTestStream(c)
	if wait, ok := s.acquireSoraTestPermit(account.ID); !ok {
		msg := fmt.Sprintf("Sora 账号测试过于频繁，请 %d 秒后重试", ceilSeconds(wait))
		recorder.addStep("rate_limit", "failed", http.StatusTooManyRequests, "test_rate_limited", msg)
		s.emitSoraProbeSummary(c, recorder)
		return s.sendErrorAndEnd(c, msg)
	}

	s.sendEvent(c, TestEvent{Type: "test_start", Model: "sora"})
	proxyURL := accountTestProxyURL(account)
	soraTLSProfile := s.resolveSoraTLSProfile()

	meStatus, meHeaders, meBody, meErr := s.fetchSoraTestEndpoint(ctx, account, authToken, soraMeAPIURL, proxyURL, soraTLSProfile)
	if meErr != nil {
		recorder.addStep("me", "failed", 0, "network_error", meErr.Error())
		s.emitSoraProbeSummary(c, recorder)
		return s.sendErrorAndEnd(c, fmt.Sprintf("Request failed: %s", meErr.Error()))
	}
	if meStatus != http.StatusOK {
		return s.handleSoraMeFailure(c, recorder, account, proxyURL, meStatus, meHeaders, meBody)
	}
	recorder.addStep("me", "success", meStatus, "", "me endpoint ok")
	s.emitSoraMeSummary(c, meBody)

	s.runSoraSubscriptionProbe(c, ctx, recorder, account, authToken, proxyURL, soraTLSProfile)
	s.testSora2Capabilities(c, ctx, account, authToken, proxyURL, soraTLSProfile, recorder)

	s.emitSoraProbeSummary(c, recorder)
	s.sendEvent(c, TestEvent{Type: "test_complete", Success: true})
	return nil
}

func (s *AccountTestService) handleSoraMeFailure(
	c *gin.Context,
	recorder *soraProbeRecorder,
	account *Account,
	proxyURL string,
	statusCode int,
	headers http.Header,
	body []byte,
) error {
	if isCloudflareChallengeResponse(statusCode, headers, body) {
		recorder.addStep("me", "failed", statusCode, "cf_challenge", "Cloudflare challenge detected")
		s.emitSoraProbeSummary(c, recorder)
		s.logSoraCloudflareChallenge(account, proxyURL, soraMeAPIURL, headers, body)
		return s.sendErrorAndEnd(c, formatCloudflareChallengeMessage(fmt.Sprintf("Sora request blocked by Cloudflare challenge (HTTP %d). Please switch to a clean proxy/network and retry.", statusCode), headers, body))
	}

	upstreamCode, upstreamMessage := soraerror.ExtractUpstreamErrorCodeAndMessage(body)
	switch {
	case statusCode == http.StatusUnauthorized && strings.EqualFold(upstreamCode, "token_invalidated"):
		recorder.addStep("me", "failed", statusCode, "token_invalidated", "Sora token invalidated")
		s.emitSoraProbeSummary(c, recorder)
		return s.sendErrorAndEnd(c, "Sora token 已失效（token_invalidated），请重新授权账号")
	case strings.EqualFold(upstreamCode, "unsupported_country_code"):
		recorder.addStep("me", "failed", statusCode, "unsupported_country_code", "Sora is unavailable in current egress region")
		s.emitSoraProbeSummary(c, recorder)
		return s.sendErrorAndEnd(c, "Sora 在当前网络出口地区不可用（unsupported_country_code），请切换到支持地区后重试")
	case strings.TrimSpace(upstreamMessage) != "":
		recorder.addStep("me", "failed", statusCode, upstreamCode, upstreamMessage)
		s.emitSoraProbeSummary(c, recorder)
		return s.sendErrorAndEnd(c, fmt.Sprintf("Sora API returned %d: %s", statusCode, upstreamMessage))
	default:
		recorder.addStep("me", "failed", statusCode, upstreamCode, "Sora me endpoint failed")
		s.emitSoraProbeSummary(c, recorder)
		return s.sendErrorAndEnd(c, fmt.Sprintf("Sora API returned %d: %s", statusCode, truncateSoraErrorBody(body, 512)))
	}
}

func (s *AccountTestService) emitSoraMeSummary(c *gin.Context, body []byte) {
	var meResp map[string]any
	if err := json.Unmarshal(body, &meResp); err != nil {
		s.sendEvent(c, TestEvent{Type: "content", Text: "Sora connection OK (token valid)"})
		return
	}

	info := "Sora connection OK"
	if name, ok := meResp["name"].(string); ok && name != "" {
		info = fmt.Sprintf("Sora connection OK - User: %s", name)
	} else if email, ok := meResp["email"].(string); ok && email != "" {
		info = fmt.Sprintf("Sora connection OK - Email: %s", email)
	}
	s.sendEvent(c, TestEvent{Type: "content", Text: info})
}

func (s *AccountTestService) runSoraSubscriptionProbe(
	c *gin.Context,
	ctx context.Context,
	recorder *soraProbeRecorder,
	account *Account,
	authToken string,
	proxyURL string,
	tlsProfile *tlsfingerprint.Profile,
) {
	statusCode, headers, body, err := s.fetchSoraTestEndpoint(ctx, account, authToken, soraBillingAPIURL, proxyURL, tlsProfile)
	if err != nil {
		recorder.addStep("subscription", "failed", 0, "network_error", err.Error())
		s.sendEvent(c, TestEvent{Type: "content", Text: fmt.Sprintf("Subscription check skipped: %s", err.Error())})
		return
	}
	if statusCode == http.StatusOK {
		recorder.addStep("subscription", "success", statusCode, "", "subscription endpoint ok")
		if summary := parseSoraSubscriptionSummary(body); summary != "" {
			s.sendEvent(c, TestEvent{Type: "content", Text: summary})
		} else {
			s.sendEvent(c, TestEvent{Type: "content", Text: "Subscription check OK"})
		}
		return
	}
	if isCloudflareChallengeResponse(statusCode, headers, body) {
		recorder.addStep("subscription", "failed", statusCode, "cf_challenge", "Cloudflare challenge detected")
		s.logSoraCloudflareChallenge(account, proxyURL, soraBillingAPIURL, headers, body)
		s.sendEvent(c, TestEvent{Type: "content", Text: formatCloudflareChallengeMessage(fmt.Sprintf("Subscription check blocked by Cloudflare challenge (HTTP %d)", statusCode), headers, body)})
		return
	}
	upstreamCode, upstreamMessage := soraerror.ExtractUpstreamErrorCodeAndMessage(body)
	recorder.addStep("subscription", "failed", statusCode, upstreamCode, upstreamMessage)
	s.sendEvent(c, TestEvent{Type: "content", Text: fmt.Sprintf("Subscription check returned %d", statusCode)})
}

func (s *AccountTestService) testSora2Capabilities(
	c *gin.Context,
	ctx context.Context,
	account *Account,
	authToken string,
	proxyURL string,
	tlsProfile *tlsfingerprint.Profile,
	recorder *soraProbeRecorder,
) {
	inviteStatus, inviteHeader, inviteBody, err := s.fetchSoraTestEndpoint(ctx, account, authToken, soraInviteMineURL, proxyURL, tlsProfile)
	if err != nil {
		if recorder != nil {
			recorder.addStep("sora2_invite", "failed", 0, "network_error", err.Error())
		}
		s.sendEvent(c, TestEvent{Type: "content", Text: fmt.Sprintf("Sora2 invite check skipped: %s", err.Error())})
		return
	}

	if inviteStatus == http.StatusUnauthorized {
		bootstrapStatus, _, _, bootstrapErr := s.fetchSoraTestEndpoint(ctx, account, authToken, soraBootstrapURL, proxyURL, tlsProfile)
		if bootstrapErr == nil && bootstrapStatus == http.StatusOK {
			if recorder != nil {
				recorder.addStep("sora2_bootstrap", "success", bootstrapStatus, "", "bootstrap endpoint ok")
			}
			s.sendEvent(c, TestEvent{Type: "content", Text: "Sora2 bootstrap OK, retry invite check"})
			inviteStatus, inviteHeader, inviteBody, err = s.fetchSoraTestEndpoint(ctx, account, authToken, soraInviteMineURL, proxyURL, tlsProfile)
			if err != nil {
				if recorder != nil {
					recorder.addStep("sora2_invite", "failed", 0, "network_error", err.Error())
				}
				s.sendEvent(c, TestEvent{Type: "content", Text: fmt.Sprintf("Sora2 invite retry failed: %s", err.Error())})
				return
			}
		} else if recorder != nil {
			code := ""
			msg := ""
			if bootstrapErr != nil {
				code = "network_error"
				msg = bootstrapErr.Error()
			}
			recorder.addStep("sora2_bootstrap", "failed", bootstrapStatus, code, msg)
		}
	}

	if inviteStatus != http.StatusOK {
		if isCloudflareChallengeResponse(inviteStatus, inviteHeader, inviteBody) {
			if recorder != nil {
				recorder.addStep("sora2_invite", "failed", inviteStatus, "cf_challenge", "Cloudflare challenge detected")
			}
			s.logSoraCloudflareChallenge(account, proxyURL, soraInviteMineURL, inviteHeader, inviteBody)
			s.sendEvent(c, TestEvent{Type: "content", Text: formatCloudflareChallengeMessage(fmt.Sprintf("Sora2 invite check blocked by Cloudflare challenge (HTTP %d)", inviteStatus), inviteHeader, inviteBody)})
			return
		}
		upstreamCode, upstreamMessage := soraerror.ExtractUpstreamErrorCodeAndMessage(inviteBody)
		if recorder != nil {
			recorder.addStep("sora2_invite", "failed", inviteStatus, upstreamCode, upstreamMessage)
		}
		s.sendEvent(c, TestEvent{Type: "content", Text: fmt.Sprintf("Sora2 invite check returned %d", inviteStatus)})
		return
	}
	if recorder != nil {
		recorder.addStep("sora2_invite", "success", inviteStatus, "", "invite endpoint ok")
	}
	if summary := parseSoraInviteSummary(inviteBody); summary != "" {
		s.sendEvent(c, TestEvent{Type: "content", Text: summary})
	} else {
		s.sendEvent(c, TestEvent{Type: "content", Text: "Sora2 invite check OK"})
	}

	remainingStatus, remainingHeader, remainingBody, remainingErr := s.fetchSoraTestEndpoint(ctx, account, authToken, soraRemainingURL, proxyURL, tlsProfile)
	if remainingErr != nil {
		if recorder != nil {
			recorder.addStep("sora2_remaining", "failed", 0, "network_error", remainingErr.Error())
		}
		s.sendEvent(c, TestEvent{Type: "content", Text: fmt.Sprintf("Sora2 remaining check skipped: %s", remainingErr.Error())})
		return
	}
	if remainingStatus != http.StatusOK {
		if isCloudflareChallengeResponse(remainingStatus, remainingHeader, remainingBody) {
			if recorder != nil {
				recorder.addStep("sora2_remaining", "failed", remainingStatus, "cf_challenge", "Cloudflare challenge detected")
			}
			s.logSoraCloudflareChallenge(account, proxyURL, soraRemainingURL, remainingHeader, remainingBody)
			s.sendEvent(c, TestEvent{Type: "content", Text: formatCloudflareChallengeMessage(fmt.Sprintf("Sora2 remaining check blocked by Cloudflare challenge (HTTP %d)", remainingStatus), remainingHeader, remainingBody)})
			return
		}
		upstreamCode, upstreamMessage := soraerror.ExtractUpstreamErrorCodeAndMessage(remainingBody)
		if recorder != nil {
			recorder.addStep("sora2_remaining", "failed", remainingStatus, upstreamCode, upstreamMessage)
		}
		s.sendEvent(c, TestEvent{Type: "content", Text: fmt.Sprintf("Sora2 remaining check returned %d", remainingStatus)})
		return
	}
	if recorder != nil {
		recorder.addStep("sora2_remaining", "success", remainingStatus, "", "remaining endpoint ok")
	}
	if summary := parseSoraRemainingSummary(remainingBody); summary != "" {
		s.sendEvent(c, TestEvent{Type: "content", Text: summary})
	} else {
		s.sendEvent(c, TestEvent{Type: "content", Text: "Sora2 remaining check OK"})
	}
}

func (s *AccountTestService) fetchSoraTestEndpoint(
	ctx context.Context,
	account *Account,
	authToken string,
	endpointURL string,
	proxyURL string,
	tlsProfile *tlsfingerprint.Profile,
) (int, http.Header, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpointURL, nil)
	if err != nil {
		return 0, nil, nil, err
	}
	req.Header.Set("Authorization", "Bearer "+authToken)
	req.Header.Set("User-Agent", "Sora/1.2026.007 (Android 15; 24122RKC7C; build 2600700)")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Origin", "https://sora.chatgpt.com")
	req.Header.Set("Referer", "https://sora.chatgpt.com/")

	resp, err := s.httpUpstream.DoWithTLS(req, proxyURL, account.ID, account.Concurrency, tlsProfile)
	if err != nil {
		return 0, nil, nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return resp.StatusCode, resp.Header, nil, readErr
	}
	return resp.StatusCode, resp.Header, body, nil
}

func parseSoraSubscriptionSummary(body []byte) string {
	var subResp struct {
		Data []struct {
			Plan struct {
				ID    string `json:"id"`
				Title string `json:"title"`
			} `json:"plan"`
			EndTS string `json:"end_ts"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &subResp); err != nil || len(subResp.Data) == 0 {
		return ""
	}

	first := subResp.Data[0]
	parts := make([]string, 0, 3)
	if first.Plan.Title != "" {
		parts = append(parts, first.Plan.Title)
	}
	if first.Plan.ID != "" {
		parts = append(parts, first.Plan.ID)
	}
	if first.EndTS != "" {
		parts = append(parts, "end="+first.EndTS)
	}
	if len(parts) == 0 {
		return ""
	}
	return "Subscription: " + strings.Join(parts, " | ")
}

func parseSoraInviteSummary(body []byte) string {
	var inviteResp struct {
		InviteCode    string `json:"invite_code"`
		RedeemedCount int64  `json:"redeemed_count"`
		TotalCount    int64  `json:"total_count"`
	}
	if err := json.Unmarshal(body, &inviteResp); err != nil {
		return ""
	}

	parts := []string{"Sora2: supported"}
	if inviteResp.InviteCode != "" {
		parts = append(parts, "invite="+inviteResp.InviteCode)
	}
	if inviteResp.TotalCount > 0 {
		parts = append(parts, fmt.Sprintf("used=%d/%d", inviteResp.RedeemedCount, inviteResp.TotalCount))
	}
	return strings.Join(parts, " | ")
}

func parseSoraRemainingSummary(body []byte) string {
	var remainingResp struct {
		RateLimitAndCreditBalance struct {
			EstimatedNumVideosRemaining int64 `json:"estimated_num_videos_remaining"`
			RateLimitReached            bool  `json:"rate_limit_reached"`
			AccessResetsInSeconds       int64 `json:"access_resets_in_seconds"`
		} `json:"rate_limit_and_credit_balance"`
	}
	if err := json.Unmarshal(body, &remainingResp); err != nil {
		return ""
	}

	info := remainingResp.RateLimitAndCreditBalance
	parts := []string{fmt.Sprintf("Sora2 remaining: %d", info.EstimatedNumVideosRemaining)}
	if info.RateLimitReached {
		parts = append(parts, "rate_limited=true")
	}
	if info.AccessResetsInSeconds > 0 {
		parts = append(parts, fmt.Sprintf("reset_in=%ds", info.AccessResetsInSeconds))
	}
	return strings.Join(parts, " | ")
}

func (s *AccountTestService) resolveSoraTLSProfile() *tlsfingerprint.Profile {
	if s == nil || s.cfg == nil || !s.cfg.Sora.Client.DisableTLSFingerprint {
		return &tlsfingerprint.Profile{Name: "Built-in Default (Sora)"}
	}
	return nil
}

func isCloudflareChallengeResponse(statusCode int, headers http.Header, body []byte) bool {
	return soraerror.IsCloudflareChallengeResponse(statusCode, headers, body)
}

func formatCloudflareChallengeMessage(base string, headers http.Header, body []byte) string {
	return soraerror.FormatCloudflareChallengeMessage(base, headers, body)
}

func extractCloudflareRayID(headers http.Header, body []byte) string {
	return soraerror.ExtractCloudflareRayID(headers, body)
}

func extractSoraEgressIPHint(headers http.Header) string {
	if headers == nil {
		return "unknown"
	}
	for _, key := range []string{"x-openai-public-ip", "x-envoy-external-address", "cf-connecting-ip", "x-forwarded-for"} {
		if value := strings.TrimSpace(headers.Get(key)); value != "" {
			return value
		}
	}
	return "unknown"
}

func sanitizeProxyURLForLog(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return "<invalid_proxy_url>"
	}
	if parsed.User != nil {
		parsed.User = nil
	}
	return parsed.String()
}

func endpointPathForLog(endpoint string) string {
	parsed, err := url.Parse(strings.TrimSpace(endpoint))
	if err != nil || parsed.Path == "" {
		return endpoint
	}
	return parsed.Path
}

func (s *AccountTestService) logSoraCloudflareChallenge(account *Account, proxyURL, endpoint string, headers http.Header, body []byte) {
	accountID := int64(0)
	platform := ""
	proxyID := "none"
	if account != nil {
		accountID = account.ID
		platform = account.Platform
		if account.ProxyID != nil {
			proxyID = fmt.Sprintf("%d", *account.ProxyID)
		}
	}
	cfRay := extractCloudflareRayID(headers, body)
	if cfRay == "" {
		cfRay = "unknown"
	}
	log.Printf(
		"[SoraCFChallenge] account_id=%d platform=%s endpoint=%s path=%s proxy_id=%s proxy_url=%s cf_ray=%s egress_ip_hint=%s",
		accountID,
		platform,
		endpoint,
		endpointPathForLog(endpoint),
		proxyID,
		sanitizeProxyURLForLog(proxyURL),
		cfRay,
		extractSoraEgressIPHint(headers),
	)
}

func truncateSoraErrorBody(body []byte, max int) string {
	return soraerror.TruncateBody(body, max)
}
