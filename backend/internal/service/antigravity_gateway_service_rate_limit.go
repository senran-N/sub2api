package service

import (
	"context"
	"encoding/json"
	"log"
	mathrand "math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/senran-N/sub2api/internal/pkg/logger"
)

func (s *AntigravityGatewayService) shouldFailoverUpstreamError(statusCode int) bool {
	switch statusCode {
	case 401, 403, 429, 529:
		return true
	default:
		return statusCode >= 500
	}
}

// isGoogleProjectConfigError 判断（已提取的小写）错误消息是否属于 Google 服务端配置类问题。
// 只精确匹配已知的服务端侧错误，避免对客户端请求错误做无意义重试。
// 适用于所有走 Google 后端的平台（Antigravity、Gemini）。
func isGoogleProjectConfigError(lowerMsg string) bool {
	return strings.Contains(lowerMsg, "invalid project resource name")
}

const googleConfigErrorCooldown = 1 * time.Minute

// tempUnscheduleGoogleConfigError 对服务端配置类 400 错误触发临时封禁，
// 避免短时间内反复调度到同一个有问题的账号。
func tempUnscheduleGoogleConfigError(ctx context.Context, repo AccountRepository, accountID int64, logPrefix string) {
	until := time.Now().Add(googleConfigErrorCooldown)
	reason := "400: invalid project resource name (auto temp-unschedule 1m)"
	if err := repo.SetTempUnschedulable(ctx, accountID, until, reason); err != nil {
		log.Printf("%s temp_unschedule_failed account=%d error=%v", logPrefix, accountID, err)
	} else {
		log.Printf("%s temp_unscheduled account=%d until=%v reason=%q", logPrefix, accountID, until.Format("15:04:05"), reason)
	}
}

const emptyResponseCooldown = 1 * time.Minute

// tempUnscheduleEmptyResponse 对空流式响应触发临时封禁，
// 避免短时间内反复调度到同一个返回空响应的账号。
func tempUnscheduleEmptyResponse(ctx context.Context, repo AccountRepository, accountID int64, logPrefix string) {
	until := time.Now().Add(emptyResponseCooldown)
	reason := "empty stream response (auto temp-unschedule 1m)"
	if err := repo.SetTempUnschedulable(ctx, accountID, until, reason); err != nil {
		log.Printf("%s temp_unschedule_failed account=%d error=%v", logPrefix, accountID, err)
	} else {
		log.Printf("%s temp_unscheduled account=%d until=%v reason=%q", logPrefix, accountID, until.Format("15:04:05"), reason)
	}
}

// sleepAntigravityBackoffWithContext 带 context 取消检查的退避等待
// 返回 true 表示正常完成等待，false 表示 context 已取消
func sleepAntigravityBackoffWithContext(ctx context.Context, attempt int) bool {
	delay := antigravityRetryBaseDelay * time.Duration(1<<uint(attempt-1))
	if delay > antigravityRetryMaxDelay {
		delay = antigravityRetryMaxDelay
	}

	rng := mathrand.New(mathrand.NewSource(time.Now().UnixNano()))
	jitter := time.Duration(float64(delay) * 0.2 * (rng.Float64()*2 - 1))
	sleepFor := delay + jitter
	if sleepFor < 0 {
		sleepFor = 0
	}

	timer := time.NewTimer(sleepFor)
	select {
	case <-ctx.Done():
		timer.Stop()
		return false
	case <-timer.C:
		return true
	}
}

// isSingleAccountRetry 检查 context 中是否设置了单账号退避重试标记
func isSingleAccountRetry(ctx context.Context) bool {
	value, _ := SingleAccountRetryFromContext(ctx)
	return value
}

// setModelRateLimitByModelName 使用官方模型 ID 设置模型级限流
// 直接使用上游返回的模型 ID（如 claude-sonnet-4-5）作为限流 key
// 返回是否已成功设置（若模型名为空或 repo 为 nil 将返回 false）
func setModelRateLimitByModelName(ctx context.Context, repo AccountRepository, accountID int64, modelName, prefix string, statusCode int, resetAt time.Time, afterSmartRetry bool) bool {
	if repo == nil || modelName == "" {
		return false
	}
	if err := repo.SetModelRateLimit(ctx, accountID, modelName, resetAt); err != nil {
		logger.LegacyPrintf("service.antigravity_gateway", "%s status=%d model_rate_limit_failed model=%s error=%v", prefix, statusCode, modelName, err)
		return false
	}
	if afterSmartRetry {
		logger.LegacyPrintf("service.antigravity_gateway", "%s status=%d model_rate_limited_after_smart_retry model=%s account=%d reset_in=%v", prefix, statusCode, modelName, accountID, time.Until(resetAt).Truncate(time.Second))
	} else {
		logger.LegacyPrintf("service.antigravity_gateway", "%s status=%d model_rate_limited model=%s account=%d reset_in=%v", prefix, statusCode, modelName, accountID, time.Until(resetAt).Truncate(time.Second))
	}
	return true
}

func antigravityFallbackCooldownSeconds() (time.Duration, bool) {
	raw := strings.TrimSpace(os.Getenv(antigravityFallbackSecondsEnv))
	if raw == "" {
		return 0, false
	}
	seconds, err := strconv.Atoi(raw)
	if err != nil || seconds <= 0 {
		return 0, false
	}
	return time.Duration(seconds) * time.Second, true
}

// antigravitySmartRetryInfo 智能重试所需的信息
type antigravitySmartRetryInfo struct {
	RetryDelay               time.Duration
	ModelName                string
	IsModelCapacityExhausted bool
}

// parseAntigravitySmartRetryInfo 解析 Google RPC RetryInfo 和 ErrorInfo 信息
func parseAntigravitySmartRetryInfo(body []byte) *antigravitySmartRetryInfo {
	var parsed map[string]any
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil
	}

	errObj, ok := parsed["error"].(map[string]any)
	if !ok {
		return nil
	}

	status, _ := errObj["status"].(string)
	isResourceExhausted := status == googleRPCStatusResourceExhausted
	isUnavailable := status == googleRPCStatusUnavailable
	if !isResourceExhausted && !isUnavailable {
		return nil
	}

	details, ok := errObj["details"].([]any)
	if !ok {
		return nil
	}

	var retryDelay time.Duration
	var modelName string
	var hasRateLimitExceeded bool
	var hasModelCapacityExhausted bool

	for _, detail := range details {
		detailMap, ok := detail.(map[string]any)
		if !ok {
			continue
		}

		atType, _ := detailMap["@type"].(string)
		if atType == googleRPCTypeErrorInfo {
			if meta, ok := detailMap["metadata"].(map[string]any); ok {
				if model, ok := meta["model"].(string); ok {
					modelName = model
				}
			}
			if reason, ok := detailMap["reason"].(string); ok {
				if reason == googleRPCReasonModelCapacityExhausted {
					hasModelCapacityExhausted = true
				}
				if reason == googleRPCReasonRateLimitExceeded {
					hasRateLimitExceeded = true
				}
			}
			continue
		}

		if atType == googleRPCTypeRetryInfo {
			delay, ok := detailMap["retryDelay"].(string)
			if !ok || delay == "" {
				continue
			}
			dur, err := time.ParseDuration(delay)
			if err != nil {
				logger.LegacyPrintf("service.antigravity_gateway", "[Antigravity] failed to parse retryDelay: %s error=%v", delay, err)
				continue
			}
			retryDelay = dur
		}
	}

	if isResourceExhausted && !hasRateLimitExceeded {
		return nil
	}
	if isUnavailable && !hasModelCapacityExhausted {
		return nil
	}
	if modelName == "" {
		return nil
	}
	if retryDelay <= 0 {
		retryDelay = antigravityDefaultRateLimitDuration
	}

	return &antigravitySmartRetryInfo{
		RetryDelay:               retryDelay,
		ModelName:                modelName,
		IsModelCapacityExhausted: hasModelCapacityExhausted,
	}
}

// shouldTriggerAntigravitySmartRetry 判断是否应该触发智能重试
func shouldTriggerAntigravitySmartRetry(account *Account, respBody []byte) (shouldRetry bool, shouldRateLimitModel bool, waitDuration time.Duration, modelName string, isModelCapacityExhausted bool) {
	if account.Platform != PlatformAntigravity {
		return false, false, 0, "", false
	}

	info := parseAntigravitySmartRetryInfo(respBody)
	if info == nil {
		return false, false, 0, "", false
	}

	if info.IsModelCapacityExhausted {
		return true, false, antigravityModelCapacityRetryWait, info.ModelName, true
	}

	if info.RetryDelay >= antigravityRateLimitThreshold {
		return false, true, info.RetryDelay, info.ModelName, false
	}

	waitDuration = info.RetryDelay
	if waitDuration < antigravitySmartRetryMinWait {
		waitDuration = antigravitySmartRetryMinWait
	}

	return true, false, waitDuration, info.ModelName, false
}

// handleModelRateLimitParams 模型级限流处理参数
type handleModelRateLimitParams struct {
	ctx             context.Context
	prefix          string
	account         *Account
	statusCode      int
	body            []byte
	cache           GatewayCache
	groupID         int64
	sessionHash     string
	isStickySession bool
}

// handleModelRateLimitResult 模型级限流处理结果
type handleModelRateLimitResult struct {
	Handled      bool
	ShouldRetry  bool
	WaitDuration time.Duration
	SwitchError  *AntigravityAccountSwitchError
}

// handleModelRateLimit 处理模型级限流（在原有逻辑之前调用）
func (s *AntigravityGatewayService) handleModelRateLimit(p *handleModelRateLimitParams) *handleModelRateLimitResult {
	if p.statusCode != 429 && p.statusCode != 503 {
		return &handleModelRateLimitResult{Handled: false}
	}

	info := parseAntigravitySmartRetryInfo(p.body)
	if info == nil || info.ModelName == "" {
		return &handleModelRateLimitResult{Handled: false}
	}

	if info.IsModelCapacityExhausted {
		log.Printf("%s status=%d model_capacity_exhausted model=%s (not switching account, retry handled by smart retry)",
			p.prefix, p.statusCode, info.ModelName)
		return &handleModelRateLimitResult{Handled: true}
	}

	if info.RetryDelay < antigravityRateLimitThreshold {
		logger.LegacyPrintf("service.antigravity_gateway", "%s status=%d model_rate_limit_wait model=%s wait=%v",
			p.prefix, p.statusCode, info.ModelName, info.RetryDelay)
		return &handleModelRateLimitResult{
			Handled:      true,
			ShouldRetry:  true,
			WaitDuration: info.RetryDelay,
		}
	}

	s.setModelRateLimitAndClearSession(p, info)
	return &handleModelRateLimitResult{
		Handled: true,
		SwitchError: &AntigravityAccountSwitchError{
			OriginalAccountID: p.account.ID,
			RateLimitedModel:  info.ModelName,
			IsStickySession:   p.isStickySession,
		},
	}
}

// setModelRateLimitAndClearSession 设置模型限流并清除粘性会话
func (s *AntigravityGatewayService) setModelRateLimitAndClearSession(p *handleModelRateLimitParams, info *antigravitySmartRetryInfo) {
	resetAt := time.Now().Add(info.RetryDelay)
	logger.LegacyPrintf("service.antigravity_gateway", "%s status=%d model_rate_limited model=%s account=%d reset_in=%v",
		p.prefix, p.statusCode, info.ModelName, p.account.ID, info.RetryDelay)

	if err := s.accountRepo.SetModelRateLimit(p.ctx, p.account.ID, info.ModelName, resetAt); err != nil {
		logger.LegacyPrintf("service.antigravity_gateway", "%s model_rate_limit_failed model=%s error=%v", p.prefix, info.ModelName, err)
	}

	s.updateAccountModelRateLimitInCache(p.ctx, p.account, info.ModelName, resetAt)

	if p.cache != nil && p.sessionHash != "" {
		_ = p.cache.DeleteSessionAccountID(p.ctx, p.groupID, p.sessionHash)
	}
}

// updateAccountModelRateLimitInCache 立即更新 Redis 中账号的模型限流状态
func (s *AntigravityGatewayService) updateAccountModelRateLimitInCache(ctx context.Context, account *Account, modelKey string, resetAt time.Time) {
	if s.schedulerSnapshot == nil || account == nil || modelKey == "" {
		return
	}

	if account.Extra == nil {
		account.Extra = make(map[string]any)
	}

	limits, _ := account.Extra["model_rate_limits"].(map[string]any)
	if limits == nil {
		limits = make(map[string]any)
		account.Extra["model_rate_limits"] = limits
	}

	limits[modelKey] = map[string]any{
		"rate_limited_at":     time.Now().UTC().Format(time.RFC3339),
		"rate_limit_reset_at": resetAt.UTC().Format(time.RFC3339),
	}

	if err := s.schedulerSnapshot.UpdateAccountInCache(ctx, account); err != nil {
		logger.LegacyPrintf("service.antigravity_gateway", "[antigravity-Forward] cache_update_failed account=%d model=%s err=%v", account.ID, modelKey, err)
	}
}

func (s *AntigravityGatewayService) handleUpstreamError(
	ctx context.Context, prefix string, account *Account,
	statusCode int, headers http.Header, body []byte,
	requestedModel string,
	groupID int64, sessionHash string, isStickySession bool,
) *handleModelRateLimitResult {
	if !account.ShouldHandleErrorCode(statusCode) {
		return nil
	}

	result := s.handleModelRateLimit(&handleModelRateLimitParams{
		ctx:             ctx,
		prefix:          prefix,
		account:         account,
		statusCode:      statusCode,
		body:            body,
		cache:           s.cache,
		groupID:         groupID,
		sessionHash:     sessionHash,
		isStickySession: isStickySession,
	})
	if result.Handled {
		return result
	}

	if statusCode == 503 {
		return nil
	}

	if statusCode == 429 {
		if logBody, maxBytes := s.getLogConfig(); logBody {
			logger.LegacyPrintf("service.antigravity_gateway", "[Antigravity-Debug] 429 response body: %s", truncateString(string(body), maxBytes))
		}

		resetAt := ParseGeminiRateLimitResetTime(body)
		defaultDur := s.getDefaultRateLimitDuration()

		modelKey := resolveFinalAntigravityModelKey(ctx, account, requestedModel)
		if strings.TrimSpace(modelKey) == "" {
			modelKey = resolveAntigravityModelKey(requestedModel)
		}
		if modelKey != "" {
			resetTime := s.resolveResetTime(resetAt, defaultDur)
			if err := s.accountRepo.SetModelRateLimit(ctx, account.ID, modelKey, resetTime); err != nil {
				logger.LegacyPrintf("service.antigravity_gateway", "%s status=429 model_rate_limit_set_failed model=%s error=%v", prefix, modelKey, err)
			} else {
				logger.LegacyPrintf("service.antigravity_gateway", "%s status=429 model_rate_limited model=%s account=%d reset_at=%v reset_in=%v",
					prefix, modelKey, account.ID, resetTime.Format("15:04:05"), time.Until(resetTime).Truncate(time.Second))
				s.updateAccountModelRateLimitInCache(ctx, account, modelKey, resetTime)
			}
			return nil
		}

		resetTime := s.resolveResetTime(resetAt, defaultDur)
		logger.LegacyPrintf("service.antigravity_gateway", "%s status=429 rate_limited account=%d reset_at=%v reset_in=%v (fallback)",
			prefix, account.ID, resetTime.Format("15:04:05"), time.Until(resetTime).Truncate(time.Second))
		if err := s.accountRepo.SetRateLimited(ctx, account.ID, resetTime); err != nil {
			logger.LegacyPrintf("service.antigravity_gateway", "%s status=429 rate_limit_set_failed account=%d error=%v", prefix, account.ID, err)
		}
		return nil
	}

	if s.rateLimitService == nil {
		return nil
	}
	shouldDisable := s.rateLimitService.HandleUpstreamError(ctx, account, statusCode, headers, body)
	if shouldDisable {
		logger.LegacyPrintf("service.antigravity_gateway", "%s status=%d marked_error", prefix, statusCode)
	}
	return nil
}

// getDefaultRateLimitDuration 获取默认限流时间
func (s *AntigravityGatewayService) getDefaultRateLimitDuration() time.Duration {
	defaultDur := antigravityDefaultRateLimitDuration
	if s.settingService != nil && s.settingService.cfg != nil && s.settingService.cfg.Gateway.AntigravityFallbackCooldownMinutes > 0 {
		defaultDur = time.Duration(s.settingService.cfg.Gateway.AntigravityFallbackCooldownMinutes) * time.Minute
	}
	if override, ok := antigravityFallbackCooldownSeconds(); ok {
		defaultDur = override
	}
	return defaultDur
}

// resolveResetTime 根据解析的重置时间或默认时长计算重置时间点
func (s *AntigravityGatewayService) resolveResetTime(resetAt *int64, defaultDur time.Duration) time.Time {
	if resetAt != nil {
		return time.Unix(*resetAt, 0)
	}
	return time.Now().Add(defaultDur)
}
