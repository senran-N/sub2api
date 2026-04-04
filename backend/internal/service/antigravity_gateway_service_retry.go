package service

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/senran-N/sub2api/internal/pkg/antigravity"
	"github.com/senran-N/sub2api/internal/pkg/logger"
)

// resolveAntigravityForwardBaseURL 解析转发用 base URL。
// 默认使用 daily（ForwardBaseURLs 的首个地址）；当环境变量为 prod 时使用第二个地址。
func resolveAntigravityForwardBaseURL() string {
	baseURLs := antigravity.ForwardBaseURLs()
	if len(baseURLs) == 0 {
		return ""
	}
	mode := strings.ToLower(strings.TrimSpace(os.Getenv(antigravityForwardBaseURLEnv)))
	if mode == "prod" && len(baseURLs) > 1 {
		return baseURLs[1]
	}
	return baseURLs[0]
}

// smartRetryAction 智能重试的处理结果
type smartRetryAction int

const (
	smartRetryActionContinue smartRetryAction = iota
	smartRetryActionBreakWithResp
	smartRetryActionContinueURL
)

// smartRetryResult 智能重试的结果
type smartRetryResult struct {
	action      smartRetryAction
	resp        *http.Response
	err         error
	switchError *AntigravityAccountSwitchError
}

// handleSmartRetry 处理 OAuth 账号的智能重试逻辑
// 将 429/503 限流处理逻辑抽取为独立函数，减少 antigravityRetryLoop 的复杂度
func (s *AntigravityGatewayService) handleSmartRetry(p antigravityRetryLoopParams, resp *http.Response, respBody []byte, baseURL string, urlIdx int, availableURLs []string) *smartRetryResult {
	if resp.StatusCode == http.StatusTooManyRequests && isURLLevelRateLimit(respBody) && urlIdx < len(availableURLs)-1 {
		logger.LegacyPrintf("service.antigravity_gateway", "%s URL fallback (429): %s -> %s", p.prefix, baseURL, availableURLs[urlIdx+1])
		return &smartRetryResult{action: smartRetryActionContinueURL}
	}

	category := antigravity429Unknown
	if resp.StatusCode == http.StatusTooManyRequests {
		category = classifyAntigravity429(respBody)
	}

	shouldSmartRetry, shouldRateLimitModel, waitDuration, modelName, isModelCapacityExhausted := shouldTriggerAntigravitySmartRetry(p.account, respBody)

	if resp.StatusCode == http.StatusTooManyRequests &&
		category == antigravity429QuotaExhausted &&
		p.account.IsOveragesEnabled() &&
		!p.account.isCreditsExhausted() {
		result := s.attemptCreditsOveragesRetry(p, baseURL, modelName, waitDuration, resp.StatusCode, respBody)
		if result.handled && result.resp != nil {
			return &smartRetryResult{
				action: smartRetryActionBreakWithResp,
				resp:   result.resp,
			}
		}
	}

	if shouldRateLimitModel {
		if resp.StatusCode == http.StatusServiceUnavailable && isSingleAccountRetry(p.ctx) {
			return s.handleSingleAccountRetryInPlace(p, resp, respBody, baseURL, waitDuration, modelName)
		}

		rateLimitDuration := waitDuration
		if rateLimitDuration <= 0 {
			rateLimitDuration = antigravityDefaultRateLimitDuration
		}
		logger.LegacyPrintf("service.antigravity_gateway", "%s status=%d oauth_long_delay model=%s account=%d upstream_retry_delay=%v body=%s (model rate limit, switch account)",
			p.prefix, resp.StatusCode, modelName, p.account.ID, rateLimitDuration, truncateForLog(respBody, 200))

		resetAt := time.Now().Add(rateLimitDuration)
		if !setModelRateLimitByModelName(p.ctx, p.accountRepo, p.account.ID, modelName, p.prefix, resp.StatusCode, resetAt, false) {
			p.handleError(p.ctx, p.prefix, p.account, resp.StatusCode, resp.Header, respBody, p.requestedModel, p.groupID, p.sessionHash, p.isStickySession)
			logger.LegacyPrintf("service.antigravity_gateway", "%s status=%d rate_limited account=%d (no model mapping)", p.prefix, resp.StatusCode, p.account.ID)
		} else {
			s.updateAccountModelRateLimitInCache(p.ctx, p.account, modelName, resetAt)
		}

		return &smartRetryResult{
			action: smartRetryActionBreakWithResp,
			switchError: &AntigravityAccountSwitchError{
				OriginalAccountID: p.account.ID,
				RateLimitedModel:  modelName,
				IsStickySession:   p.isStickySession,
			},
		}
	}

	if shouldSmartRetry {
		var lastRetryResp *http.Response
		var lastRetryBody []byte

		maxAttempts := antigravitySmartRetryMaxAttempts
		if isModelCapacityExhausted {
			maxAttempts = antigravityModelCapacityRetryMaxAttempts
			waitDuration = antigravityModelCapacityRetryWait

			if modelName != "" {
				modelCapacityExhaustedMu.RLock()
				cooldownUntil, exists := modelCapacityExhaustedUntil[modelName]
				modelCapacityExhaustedMu.RUnlock()
				if exists && time.Now().Before(cooldownUntil) {
					log.Printf("%s status=%d model_capacity_exhausted_dedup model=%s account=%d cooldown_until=%v (skip retry)",
						p.prefix, resp.StatusCode, modelName, p.account.ID, cooldownUntil.Format("15:04:05"))
					return &smartRetryResult{
						action: smartRetryActionBreakWithResp,
						resp: &http.Response{
							StatusCode: resp.StatusCode,
							Header:     resp.Header.Clone(),
							Body:       io.NopCloser(bytes.NewReader(respBody)),
						},
					}
				}
			}
		}

		for attempt := 1; attempt <= maxAttempts; attempt++ {
			log.Printf("%s status=%d oauth_smart_retry attempt=%d/%d delay=%v model=%s account=%d",
				p.prefix, resp.StatusCode, attempt, maxAttempts, waitDuration, modelName, p.account.ID)

			timer := time.NewTimer(waitDuration)
			select {
			case <-p.ctx.Done():
				timer.Stop()
				log.Printf("%s status=context_canceled_during_smart_retry", p.prefix)
				return &smartRetryResult{action: smartRetryActionBreakWithResp, err: p.ctx.Err()}
			case <-timer.C:
			}

			retryReq, err := antigravity.NewAPIRequestWithURL(p.ctx, baseURL, p.action, p.accessToken, p.body)
			if err != nil {
				logger.LegacyPrintf("service.antigravity_gateway", "%s status=smart_retry_request_build_failed error=%v", p.prefix, err)
				p.handleError(p.ctx, p.prefix, p.account, resp.StatusCode, resp.Header, respBody, p.requestedModel, p.groupID, p.sessionHash, p.isStickySession)
				return &smartRetryResult{
					action: smartRetryActionBreakWithResp,
					resp: &http.Response{
						StatusCode: resp.StatusCode,
						Header:     resp.Header.Clone(),
						Body:       io.NopCloser(bytes.NewReader(respBody)),
					},
				}
			}

			retryResp, retryErr := p.httpUpstream.Do(retryReq, p.proxyURL, p.account.ID, p.account.Concurrency)
			if retryErr == nil && retryResp != nil && retryResp.StatusCode != http.StatusTooManyRequests && retryResp.StatusCode != http.StatusServiceUnavailable {
				log.Printf("%s status=%d smart_retry_success attempt=%d/%d", p.prefix, retryResp.StatusCode, attempt, maxAttempts)
				if isModelCapacityExhausted && modelName != "" {
					modelCapacityExhaustedMu.Lock()
					delete(modelCapacityExhaustedUntil, modelName)
					modelCapacityExhaustedMu.Unlock()
				}
				return &smartRetryResult{action: smartRetryActionBreakWithResp, resp: retryResp}
			}

			if retryErr != nil || retryResp == nil {
				log.Printf("%s status=smart_retry_network_error attempt=%d/%d error=%v", p.prefix, attempt, maxAttempts, retryErr)
				continue
			}

			if lastRetryResp != nil {
				_ = lastRetryResp.Body.Close()
			}
			lastRetryResp = retryResp
			if retryResp != nil {
				lastRetryBody, _ = io.ReadAll(io.LimitReader(retryResp.Body, 8<<10))
				_ = retryResp.Body.Close()
			}

			if !isModelCapacityExhausted && attempt < maxAttempts && lastRetryBody != nil {
				newShouldRetry, _, newWaitDuration, _, _ := shouldTriggerAntigravitySmartRetry(p.account, lastRetryBody)
				if newShouldRetry && newWaitDuration > 0 {
					waitDuration = newWaitDuration
				}
			}
		}

		rateLimitDuration := waitDuration
		if rateLimitDuration <= 0 {
			rateLimitDuration = antigravityDefaultRateLimitDuration
		}
		retryBody := lastRetryBody
		if retryBody == nil {
			retryBody = respBody
		}

		if isModelCapacityExhausted {
			if modelName != "" {
				modelCapacityExhaustedMu.Lock()
				modelCapacityExhaustedUntil[modelName] = time.Now().Add(antigravityModelCapacityCooldown)
				modelCapacityExhaustedMu.Unlock()
			}
			log.Printf("%s status=%d smart_retry_exhausted_model_capacity attempts=%d model=%s account=%d body=%s (model capacity exhausted, not switching account)",
				p.prefix, resp.StatusCode, maxAttempts, modelName, p.account.ID, truncateForLog(retryBody, 200))
			return &smartRetryResult{
				action: smartRetryActionBreakWithResp,
				resp: &http.Response{
					StatusCode: resp.StatusCode,
					Header:     resp.Header.Clone(),
					Body:       io.NopCloser(bytes.NewReader(retryBody)),
				},
			}
		}

		if resp.StatusCode == http.StatusServiceUnavailable && isSingleAccountRetry(p.ctx) {
			logger.LegacyPrintf("service.antigravity_gateway", "%s status=%d smart_retry_exhausted_single_account attempts=%d model=%s account=%d body=%s (return 503 directly)",
				p.prefix, resp.StatusCode, antigravitySmartRetryMaxAttempts, modelName, p.account.ID, truncateForLog(retryBody, 200))
			return &smartRetryResult{
				action: smartRetryActionBreakWithResp,
				resp: &http.Response{
					StatusCode: resp.StatusCode,
					Header:     resp.Header.Clone(),
					Body:       io.NopCloser(bytes.NewReader(retryBody)),
				},
			}
		}

		log.Printf("%s status=%d smart_retry_exhausted attempts=%d model=%s account=%d upstream_retry_delay=%v body=%s (switch account)",
			p.prefix, resp.StatusCode, maxAttempts, modelName, p.account.ID, rateLimitDuration, truncateForLog(retryBody, 200))

		resetAt := time.Now().Add(rateLimitDuration)
		if p.accountRepo != nil && modelName != "" {
			if err := p.accountRepo.SetModelRateLimit(p.ctx, p.account.ID, modelName, resetAt); err != nil {
				logger.LegacyPrintf("service.antigravity_gateway", "%s status=%d model_rate_limit_failed model=%s error=%v", p.prefix, resp.StatusCode, modelName, err)
			} else {
				logger.LegacyPrintf("service.antigravity_gateway", "%s status=%d model_rate_limited_after_smart_retry model=%s account=%d reset_in=%v",
					p.prefix, resp.StatusCode, modelName, p.account.ID, rateLimitDuration)
				s.updateAccountModelRateLimitInCache(p.ctx, p.account, modelName, resetAt)
			}
		}

		if s.cache != nil && p.sessionHash != "" {
			_ = s.cache.DeleteSessionAccountID(p.ctx, p.groupID, p.sessionHash)
		}

		return &smartRetryResult{
			action: smartRetryActionBreakWithResp,
			switchError: &AntigravityAccountSwitchError{
				OriginalAccountID: p.account.ID,
				RateLimitedModel:  modelName,
				IsStickySession:   p.isStickySession,
			},
		}
	}

	return &smartRetryResult{action: smartRetryActionContinue}
}

// handleSingleAccountRetryInPlace 单账号 503 退避重试的原地重试逻辑。
func (s *AntigravityGatewayService) handleSingleAccountRetryInPlace(
	p antigravityRetryLoopParams,
	resp *http.Response,
	respBody []byte,
	baseURL string,
	waitDuration time.Duration,
	modelName string,
) *smartRetryResult {
	if waitDuration > antigravitySingleAccountSmartRetryMaxWait {
		waitDuration = antigravitySingleAccountSmartRetryMaxWait
	}
	if waitDuration < antigravitySmartRetryMinWait {
		waitDuration = antigravitySmartRetryMinWait
	}

	logger.LegacyPrintf("service.antigravity_gateway", "%s status=%d single_account_503_retry_in_place model=%s account=%d upstream_retry_delay=%v (retrying in-place instead of rate-limiting)",
		p.prefix, resp.StatusCode, modelName, p.account.ID, waitDuration)

	var lastRetryResp *http.Response
	var lastRetryBody []byte
	totalWaited := time.Duration(0)

	for attempt := 1; attempt <= antigravitySingleAccountSmartRetryMaxAttempts; attempt++ {
		if totalWaited+waitDuration > antigravitySingleAccountSmartRetryTotalMaxWait {
			remaining := antigravitySingleAccountSmartRetryTotalMaxWait - totalWaited
			if remaining <= 0 {
				logger.LegacyPrintf("service.antigravity_gateway", "%s single_account_503_retry: total_wait_exceeded total=%v max=%v, giving up",
					p.prefix, totalWaited, antigravitySingleAccountSmartRetryTotalMaxWait)
				break
			}
			waitDuration = remaining
		}

		logger.LegacyPrintf("service.antigravity_gateway", "%s status=%d single_account_503_retry attempt=%d/%d delay=%v total_waited=%v model=%s account=%d",
			p.prefix, resp.StatusCode, attempt, antigravitySingleAccountSmartRetryMaxAttempts, waitDuration, totalWaited, modelName, p.account.ID)

		timer := time.NewTimer(waitDuration)
		select {
		case <-p.ctx.Done():
			timer.Stop()
			logger.LegacyPrintf("service.antigravity_gateway", "%s status=context_canceled_during_single_account_retry", p.prefix)
			return &smartRetryResult{action: smartRetryActionBreakWithResp, err: p.ctx.Err()}
		case <-timer.C:
		}
		totalWaited += waitDuration

		retryReq, err := antigravity.NewAPIRequestWithURL(p.ctx, baseURL, p.action, p.accessToken, p.body)
		if err != nil {
			logger.LegacyPrintf("service.antigravity_gateway", "%s single_account_503_retry: request_build_failed error=%v", p.prefix, err)
			break
		}

		retryResp, retryErr := p.httpUpstream.Do(retryReq, p.proxyURL, p.account.ID, p.account.Concurrency)
		if retryErr == nil && retryResp != nil && retryResp.StatusCode != http.StatusTooManyRequests && retryResp.StatusCode != http.StatusServiceUnavailable {
			logger.LegacyPrintf("service.antigravity_gateway", "%s status=%d single_account_503_retry_success attempt=%d/%d total_waited=%v",
				p.prefix, retryResp.StatusCode, attempt, antigravitySingleAccountSmartRetryMaxAttempts, totalWaited)
			if lastRetryResp != nil {
				_ = lastRetryResp.Body.Close()
			}
			return &smartRetryResult{action: smartRetryActionBreakWithResp, resp: retryResp}
		}

		if retryErr != nil || retryResp == nil {
			logger.LegacyPrintf("service.antigravity_gateway", "%s single_account_503_retry: network_error attempt=%d/%d error=%v",
				p.prefix, attempt, antigravitySingleAccountSmartRetryMaxAttempts, retryErr)
			continue
		}

		if lastRetryResp != nil {
			_ = lastRetryResp.Body.Close()
		}
		lastRetryResp = retryResp
		lastRetryBody, _ = io.ReadAll(io.LimitReader(retryResp.Body, 8<<10))
		_ = retryResp.Body.Close()

		if attempt < antigravitySingleAccountSmartRetryMaxAttempts && lastRetryBody != nil {
			_, _, newWaitDuration, _, _ := shouldTriggerAntigravitySmartRetry(p.account, lastRetryBody)
			if newWaitDuration > 0 {
				waitDuration = newWaitDuration
				if waitDuration > antigravitySingleAccountSmartRetryMaxWait {
					waitDuration = antigravitySingleAccountSmartRetryMaxWait
				}
				if waitDuration < antigravitySmartRetryMinWait {
					waitDuration = antigravitySmartRetryMinWait
				}
			}
		}
	}

	retryBody := lastRetryBody
	if retryBody == nil {
		retryBody = respBody
	}
	logger.LegacyPrintf("service.antigravity_gateway", "%s status=%d single_account_503_retry_exhausted attempts=%d total_waited=%v model=%s account=%d body=%s (return 503 directly)",
		p.prefix, resp.StatusCode, antigravitySingleAccountSmartRetryMaxAttempts, totalWaited, modelName, p.account.ID, truncateForLog(retryBody, 200))

	return &smartRetryResult{
		action: smartRetryActionBreakWithResp,
		resp: &http.Response{
			StatusCode: resp.StatusCode,
			Header:     resp.Header.Clone(),
			Body:       io.NopCloser(bytes.NewReader(retryBody)),
		},
	}
}

// antigravityRetryLoop 执行带 URL fallback 的重试循环
func (s *AntigravityGatewayService) antigravityRetryLoop(p antigravityRetryLoopParams) (*antigravityRetryLoopResult, error) {
	overagesInjected := false
	if p.requestedModel != "" && p.account.Platform == PlatformAntigravity &&
		p.account.IsOveragesEnabled() && !p.account.isCreditsExhausted() &&
		p.account.isModelRateLimitedWithContext(p.ctx, p.requestedModel) {
		if creditsBody := injectEnabledCreditTypes(p.body); creditsBody != nil {
			p.body = creditsBody
			overagesInjected = true
			logger.LegacyPrintf("service.antigravity_gateway", "%s pre_check: model_rate_limited_credits_inject model=%s account=%d (injecting enabledCreditTypes)",
				p.prefix, p.requestedModel, p.account.ID)
		}
	}

	if p.requestedModel != "" {
		if remaining := p.account.GetRateLimitRemainingTimeWithContext(p.ctx, p.requestedModel); remaining > 0 {
			if overagesInjected {
				logger.LegacyPrintf("service.antigravity_gateway", "%s pre_check: credits_injected_ignore_rate_limit remaining=%v model=%s account=%d",
					p.prefix, remaining.Truncate(time.Millisecond), p.requestedModel, p.account.ID)
			} else if isSingleAccountRetry(p.ctx) {
				logger.LegacyPrintf("service.antigravity_gateway", "%s pre_check: single_account_retry skipping rate_limit remaining=%v model=%s account=%d (will retry in-place if 503)",
					p.prefix, remaining.Truncate(time.Millisecond), p.requestedModel, p.account.ID)
			} else {
				logger.LegacyPrintf("service.antigravity_gateway", "%s pre_check: rate_limit_switch remaining=%v model=%s account=%d",
					p.prefix, remaining.Truncate(time.Millisecond), p.requestedModel, p.account.ID)
				return nil, &AntigravityAccountSwitchError{
					OriginalAccountID: p.account.ID,
					RateLimitedModel:  p.requestedModel,
					IsStickySession:   p.isStickySession,
				}
			}
		}
	}

	baseURL := resolveAntigravityForwardBaseURL()
	if baseURL == "" {
		return nil, errors.New("no antigravity forward base url configured")
	}
	availableURLs := []string{baseURL}

	var resp *http.Response
	var usedBaseURL string
	logBody := p.settingService != nil && p.settingService.cfg != nil && p.settingService.cfg.Gateway.LogUpstreamErrorBody
	maxBytes := 2048
	if p.settingService != nil && p.settingService.cfg != nil && p.settingService.cfg.Gateway.LogUpstreamErrorBodyMaxBytes > 0 {
		maxBytes = p.settingService.cfg.Gateway.LogUpstreamErrorBodyMaxBytes
	}
	getUpstreamDetail := func(body []byte) string {
		if !logBody {
			return ""
		}
		return truncateString(string(body), maxBytes)
	}

urlFallbackLoop:
	for urlIdx, baseURL := range availableURLs {
		usedBaseURL = baseURL
		allAttemptsInternal500 := true
		for attempt := 1; attempt <= antigravityMaxRetries; attempt++ {
			select {
			case <-p.ctx.Done():
				logger.LegacyPrintf("service.antigravity_gateway", "%s status=context_canceled error=%v", p.prefix, p.ctx.Err())
				return nil, p.ctx.Err()
			default:
			}

			upstreamReq, err := antigravity.NewAPIRequestWithURL(p.ctx, baseURL, p.action, p.accessToken, p.body)
			if err != nil {
				return nil, err
			}

			if p.c != nil && len(p.body) > 0 {
				p.c.Set(OpsUpstreamRequestBodyKey, string(p.body))
			}

			resp, err = p.httpUpstream.Do(upstreamReq, p.proxyURL, p.account.ID, p.account.Concurrency)
			if err == nil && resp == nil {
				err = errors.New("upstream returned nil response")
			}
			if err != nil {
				safeErr := sanitizeUpstreamErrorMessage(err.Error())
				appendOpsUpstreamError(p.c, OpsUpstreamErrorEvent{
					Platform:           p.account.Platform,
					AccountID:          p.account.ID,
					AccountName:        p.account.Name,
					UpstreamStatusCode: 0,
					UpstreamURL:        safeUpstreamURL(upstreamReq.URL.String()),
					Kind:               "request_error",
					Message:            safeErr,
				})
				if shouldAntigravityFallbackToNextURL(err, 0) && urlIdx < len(availableURLs)-1 {
					logger.LegacyPrintf("service.antigravity_gateway", "%s URL fallback (connection error): %s -> %s", p.prefix, baseURL, availableURLs[urlIdx+1])
					continue urlFallbackLoop
				}
				if attempt < antigravityMaxRetries {
					logger.LegacyPrintf("service.antigravity_gateway", "%s status=request_failed retry=%d/%d error=%v", p.prefix, attempt, antigravityMaxRetries, err)
					if !sleepAntigravityBackoffWithContext(p.ctx, attempt) {
						logger.LegacyPrintf("service.antigravity_gateway", "%s status=context_canceled_during_backoff", p.prefix)
						return nil, p.ctx.Err()
					}
					continue
				}
				logger.LegacyPrintf("service.antigravity_gateway", "%s status=request_failed retries_exhausted error=%v", p.prefix, err)
				setOpsUpstreamError(p.c, 0, safeErr, "")
				return nil, fmt.Errorf("upstream request failed after retries: %w", err)
			}

			if resp.StatusCode >= 400 {
				respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
				_ = resp.Body.Close()

				if overagesInjected && shouldMarkCreditsExhausted(resp, respBody, nil) {
					modelKey := resolveCreditsOveragesModelKey(p.ctx, p.account, "", p.requestedModel)
					s.handleCreditsRetryFailure(p.ctx, p.prefix, modelKey, p.account, &http.Response{
						StatusCode: resp.StatusCode,
						Header:     resp.Header.Clone(),
						Body:       io.NopCloser(bytes.NewReader(respBody)),
					}, nil)
				}

				if handled, outStatus, policyErr := s.applyErrorPolicy(p, resp.StatusCode, resp.Header, respBody); handled {
					if policyErr != nil {
						return nil, policyErr
					}
					resp = &http.Response{
						StatusCode: outStatus,
						Header:     resp.Header.Clone(),
						Body:       io.NopCloser(bytes.NewReader(respBody)),
					}
					break urlFallbackLoop
				}

				if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusServiceUnavailable {
					smartResult := s.handleSmartRetry(p, resp, respBody, baseURL, urlIdx, availableURLs)
					switch smartResult.action {
					case smartRetryActionContinueURL:
						continue urlFallbackLoop
					case smartRetryActionBreakWithResp:
						if smartResult.err != nil {
							return nil, smartResult.err
						}
						if smartResult.switchError != nil {
							return nil, smartResult.switchError
						}
						resp = smartResult.resp
						break urlFallbackLoop
					}

					if attempt < antigravityMaxRetries {
						upstreamMsg := strings.TrimSpace(extractAntigravityErrorMessage(respBody))
						upstreamMsg = sanitizeUpstreamErrorMessage(upstreamMsg)
						appendOpsUpstreamError(p.c, OpsUpstreamErrorEvent{
							Platform:           p.account.Platform,
							AccountID:          p.account.ID,
							AccountName:        p.account.Name,
							UpstreamStatusCode: resp.StatusCode,
							UpstreamRequestID:  resp.Header.Get("x-request-id"),
							UpstreamURL:        safeUpstreamURL(upstreamReq.URL.String()),
							Kind:               "retry",
							Message:            upstreamMsg,
							Detail:             getUpstreamDetail(respBody),
						})
						logger.LegacyPrintf("service.antigravity_gateway", "%s status=%d retry=%d/%d body=%s", p.prefix, resp.StatusCode, attempt, antigravityMaxRetries, truncateForLog(respBody, 200))
						if !sleepAntigravityBackoffWithContext(p.ctx, attempt) {
							logger.LegacyPrintf("service.antigravity_gateway", "%s status=context_canceled_during_backoff", p.prefix)
							return nil, p.ctx.Err()
						}
						continue
					}

					p.handleError(p.ctx, p.prefix, p.account, resp.StatusCode, resp.Header, respBody, p.requestedModel, p.groupID, p.sessionHash, p.isStickySession)
					logger.LegacyPrintf("service.antigravity_gateway", "%s status=%d rate_limited base_url=%s body=%s", p.prefix, resp.StatusCode, baseURL, truncateForLog(respBody, 200))
					resp = &http.Response{
						StatusCode: resp.StatusCode,
						Header:     resp.Header.Clone(),
						Body:       io.NopCloser(bytes.NewReader(respBody)),
					}
					break urlFallbackLoop
				}

				if shouldRetryAntigravityError(resp.StatusCode) {
					if attempt < antigravityMaxRetries {
						upstreamMsg := strings.TrimSpace(extractAntigravityErrorMessage(respBody))
						upstreamMsg = sanitizeUpstreamErrorMessage(upstreamMsg)
						appendOpsUpstreamError(p.c, OpsUpstreamErrorEvent{
							Platform:           p.account.Platform,
							AccountID:          p.account.ID,
							AccountName:        p.account.Name,
							UpstreamStatusCode: resp.StatusCode,
							UpstreamRequestID:  resp.Header.Get("x-request-id"),
							UpstreamURL:        safeUpstreamURL(upstreamReq.URL.String()),
							Kind:               "retry",
							Message:            upstreamMsg,
							Detail:             getUpstreamDetail(respBody),
						})
						logger.LegacyPrintf("service.antigravity_gateway", "%s status=%d retry=%d/%d body=%s", p.prefix, resp.StatusCode, attempt, antigravityMaxRetries, truncateForLog(respBody, 500))
						if !sleepAntigravityBackoffWithContext(p.ctx, attempt) {
							logger.LegacyPrintf("service.antigravity_gateway", "%s status=context_canceled_during_backoff", p.prefix)
							return nil, p.ctx.Err()
						}
						if !isAntigravityInternalServerError(resp.StatusCode, respBody) {
							allAttemptsInternal500 = false
						}
						continue
					}
				}

				if allAttemptsInternal500 && isAntigravityInternalServerError(resp.StatusCode, respBody) {
					s.handleInternal500RetryExhausted(p.ctx, p.prefix, p.account)
				}

				resp = &http.Response{
					StatusCode: resp.StatusCode,
					Header:     resp.Header.Clone(),
					Body:       io.NopCloser(bytes.NewReader(respBody)),
				}
				break urlFallbackLoop
			}

			break urlFallbackLoop
		}
	}

	if resp != nil && resp.StatusCode < 400 && usedBaseURL != "" {
		antigravity.DefaultURLAvailability.MarkSuccess(usedBaseURL)
	}
	if resp != nil && resp.StatusCode < 400 {
		s.resetInternal500Counter(p.ctx, p.prefix, p.account.ID)
	}

	return &antigravityRetryLoopResult{resp: resp}, nil
}

// shouldRetryAntigravityError 判断是否应该重试
func shouldRetryAntigravityError(statusCode int) bool {
	switch statusCode {
	case 429, 500, 502, 503, 504, 529:
		return true
	default:
		return false
	}
}

// isURLLevelRateLimit 判断是否为 URL 级别的限流（应切换 URL 重试）
func isURLLevelRateLimit(body []byte) bool {
	bodyStr := string(body)
	return strings.Contains(bodyStr, "Resource has been exhausted") &&
		!strings.Contains(bodyStr, "capacity on this model")
}

// isAntigravityConnectionError 判断是否为连接错误（网络超时、DNS 失败、连接拒绝）
func isAntigravityConnectionError(err error) bool {
	if err == nil {
		return false
	}

	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}

	var opErr *net.OpError
	return errors.As(err, &opErr)
}

// shouldAntigravityFallbackToNextURL 判断是否应切换到下一个 URL
func shouldAntigravityFallbackToNextURL(err error, statusCode int) bool {
	if isAntigravityConnectionError(err) {
		return true
	}
	return statusCode == http.StatusTooManyRequests
}
