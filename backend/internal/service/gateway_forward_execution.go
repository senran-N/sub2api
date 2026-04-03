package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/senran-N/sub2api/internal/pkg/logger"
	"github.com/senran-N/sub2api/internal/pkg/tlsfingerprint"
)

func (s *GatewayService) executeForwardUpstream(
	ctx context.Context,
	c *gin.Context,
	account *Account,
	preparation *forwardRequestPreparation,
) (*http.Response, error) {
	body := preparation.body
	reqModel := preparation.reqModel
	reqStream := preparation.reqStream
	shouldMimicClaudeCode := preparation.shouldMimicClaudeCode
	token := preparation.token
	tokenType := preparation.tokenType
	proxyURL := preparation.proxyURL
	tlsProfile := preparation.tlsProfile

	var resp *http.Response
	retryStart := time.Now()
	for attempt := 1; attempt <= maxRetryAttempts; attempt++ {
		upstreamCtx, releaseUpstreamCtx := detachStreamUpstreamContext(ctx, reqStream)
		upstreamReq, err := s.buildUpstreamRequest(upstreamCtx, c, account, body, token, tokenType, reqModel, reqStream, shouldMimicClaudeCode)
		releaseUpstreamCtx()
		if err != nil {
			return nil, err
		}

		resp, err = s.httpUpstream.DoWithTLS(upstreamReq, proxyURL, account.ID, account.Concurrency, tlsProfile)
		if err != nil {
			if resp != nil && resp.Body != nil {
				_ = resp.Body.Close()
			}
			safeErr := sanitizeUpstreamErrorMessage(err.Error())
			setOpsUpstreamError(c, 0, safeErr, "")
			appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
				Platform:           account.Platform,
				AccountID:          account.ID,
				AccountName:        account.Name,
				UpstreamStatusCode: 0,
				UpstreamURL:        safeUpstreamURL(upstreamReq.URL.String()),
				Kind:               "request_error",
				Message:            safeErr,
			})
			c.JSON(http.StatusBadGateway, gin.H{
				"type": "error",
				"error": gin.H{
					"type":    "upstream_error",
					"message": "Upstream request failed",
				},
			})
			return nil, fmt.Errorf("upstream request failed: %s", safeErr)
		}

		if resp.StatusCode == http.StatusBadRequest {
			handledResp, handled := s.maybeRetryForwardBadRequest(
				ctx,
				c,
				account,
				upstreamReq,
				resp,
				body,
				reqModel,
				reqStream,
				token,
				tokenType,
				proxyURL,
				tlsProfile,
				shouldMimicClaudeCode,
				retryStart,
			)
			if handled {
				resp = handledResp
			}
		}

		if resp.StatusCode >= 400 && resp.StatusCode != http.StatusBadRequest && s.shouldRetryUpstreamError(account, resp.StatusCode) {
			if attempt < maxRetryAttempts {
				elapsed := time.Since(retryStart)
				if elapsed >= maxRetryElapsed {
					break
				}

				delay := retryBackoffDelay(attempt)
				remaining := maxRetryElapsed - elapsed
				if delay > remaining {
					delay = remaining
				}
				if delay <= 0 {
					break
				}

				respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
				_ = resp.Body.Close()
				appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
					Platform:           account.Platform,
					AccountID:          account.ID,
					AccountName:        account.Name,
					UpstreamStatusCode: resp.StatusCode,
					UpstreamRequestID:  resp.Header.Get("x-request-id"),
					UpstreamURL:        safeUpstreamURL(upstreamReq.URL.String()),
					Kind:               "retry",
					Message:            extractUpstreamErrorMessage(respBody),
					Detail: func() string {
						if s.cfg != nil && s.cfg.Gateway.LogUpstreamErrorBody {
							return truncateString(string(respBody), s.cfg.Gateway.LogUpstreamErrorBodyMaxBytes)
						}
						return ""
					}(),
				})
				logger.LegacyPrintf("service.gateway", "Account %d: upstream error %d, retry %d/%d after %v (elapsed=%v/%v)",
					account.ID, resp.StatusCode, attempt, maxRetryAttempts, delay, elapsed, maxRetryElapsed)
				if err := sleepWithContext(ctx, delay); err != nil {
					return nil, err
				}
				continue
			}
			break
		}

		if account.Platform == PlatformGemini && resp.StatusCode < 400 && s.cfg != nil && s.cfg.Gateway.GeminiDebugResponseHeaders {
			logger.LegacyPrintf("service.gateway", "[DEBUG] Gemini API Response Headers for account %d:", account.ID)
			for k, v := range resp.Header {
				logger.LegacyPrintf("service.gateway", "[DEBUG]   %s: %v", k, v)
			}
		}
		break
	}

	if resp == nil || resp.Body == nil {
		return nil, errors.New("upstream request failed: empty response")
	}
	return resp, nil
}

func (s *GatewayService) maybeRetryForwardBadRequest(
	ctx context.Context,
	c *gin.Context,
	account *Account,
	upstreamReq *http.Request,
	resp *http.Response,
	body []byte,
	reqModel string,
	reqStream bool,
	token string,
	tokenType string,
	proxyURL string,
	tlsProfile *tlsfingerprint.Profile,
	shouldMimicClaudeCode bool,
	retryStart time.Time,
) (*http.Response, bool) {
	respBody, readErr := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if readErr != nil {
		return resp, false
	}
	_ = resp.Body.Close()

	if s.shouldRectifySignatureError(ctx, account, respBody) {
		appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
			Platform:           account.Platform,
			AccountID:          account.ID,
			AccountName:        account.Name,
			UpstreamStatusCode: resp.StatusCode,
			UpstreamRequestID:  resp.Header.Get("x-request-id"),
			UpstreamURL:        safeUpstreamURL(upstreamReq.URL.String()),
			Kind:               "signature_error",
			Message:            extractUpstreamErrorMessage(respBody),
			Detail: func() string {
				if s.cfg != nil && s.cfg.Gateway.LogUpstreamErrorBody {
					return truncateString(string(respBody), s.cfg.Gateway.LogUpstreamErrorBodyMaxBytes)
				}
				return ""
			}(),
		})

		if time.Since(retryStart) >= maxRetryElapsed {
			resp.Body = io.NopCloser(bytes.NewReader(respBody))
			return resp, true
		}
		logger.LegacyPrintf("service.gateway", "[warn] Account %d: thinking blocks have invalid signature, retrying with filtered blocks", account.ID)

		filteredBody := FilterThinkingBlocksForRetry(body)
		retryCtx, releaseRetryCtx := detachStreamUpstreamContext(ctx, reqStream)
		retryReq, buildErr := s.buildUpstreamRequest(retryCtx, c, account, filteredBody, token, tokenType, reqModel, reqStream, shouldMimicClaudeCode)
		releaseRetryCtx()
		if buildErr == nil {
			retryResp, retryErr := s.httpUpstream.DoWithTLS(retryReq, proxyURL, account.ID, account.Concurrency, tlsProfile)
			if retryErr == nil {
				if retryResp.StatusCode < 400 {
					logger.LegacyPrintf("service.gateway", "Account %d: thinking block retry succeeded (blocks downgraded)", account.ID)
					return retryResp, true
				}

				retryRespBody, retryReadErr := io.ReadAll(io.LimitReader(retryResp.Body, 2<<20))
				_ = retryResp.Body.Close()
				if retryReadErr == nil && retryResp.StatusCode == http.StatusBadRequest && s.isSignatureErrorPattern(ctx, account, retryRespBody) {
					appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
						Platform:           account.Platform,
						AccountID:          account.ID,
						AccountName:        account.Name,
						UpstreamStatusCode: retryResp.StatusCode,
						UpstreamRequestID:  retryResp.Header.Get("x-request-id"),
						UpstreamURL:        safeUpstreamURL(retryReq.URL.String()),
						Kind:               "signature_retry_thinking",
						Message:            extractUpstreamErrorMessage(retryRespBody),
						Detail: func() string {
							if s.cfg != nil && s.cfg.Gateway.LogUpstreamErrorBody {
								return truncateString(string(retryRespBody), s.cfg.Gateway.LogUpstreamErrorBodyMaxBytes)
							}
							return ""
						}(),
					})
					msg2 := extractUpstreamErrorMessage(retryRespBody)
					if looksLikeToolSignatureError(msg2) && time.Since(retryStart) < maxRetryElapsed {
						logger.LegacyPrintf("service.gateway", "Account %d: signature retry still failing and looks tool-related, retrying with tool blocks downgraded", account.ID)
						filteredBody2 := FilterSignatureSensitiveBlocksForRetry(body)
						retryCtx2, releaseRetryCtx2 := detachStreamUpstreamContext(ctx, reqStream)
						retryReq2, buildErr2 := s.buildUpstreamRequest(retryCtx2, c, account, filteredBody2, token, tokenType, reqModel, reqStream, shouldMimicClaudeCode)
						releaseRetryCtx2()
						if buildErr2 == nil {
							retryResp2, retryErr2 := s.httpUpstream.DoWithTLS(retryReq2, proxyURL, account.ID, account.Concurrency, tlsProfile)
							if retryErr2 == nil {
								return retryResp2, true
							}
							if retryResp2 != nil && retryResp2.Body != nil {
								_ = retryResp2.Body.Close()
							}
							appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
								Platform:           account.Platform,
								AccountID:          account.ID,
								AccountName:        account.Name,
								UpstreamStatusCode: 0,
								UpstreamURL:        safeUpstreamURL(retryReq2.URL.String()),
								Kind:               "signature_retry_tools_request_error",
								Message:            sanitizeUpstreamErrorMessage(retryErr2.Error()),
							})
							logger.LegacyPrintf("service.gateway", "Account %d: tool-downgrade signature retry failed: %v", account.ID, retryErr2)
						} else {
							logger.LegacyPrintf("service.gateway", "Account %d: tool-downgrade signature retry build failed: %v", account.ID, buildErr2)
						}
					}
				}

				return &http.Response{
					StatusCode: retryResp.StatusCode,
					Header:     retryResp.Header.Clone(),
					Body:       io.NopCloser(bytes.NewReader(retryRespBody)),
				}, true
			}
			if retryResp != nil && retryResp.Body != nil {
				_ = retryResp.Body.Close()
			}
			logger.LegacyPrintf("service.gateway", "Account %d: signature error retry failed: %v", account.ID, retryErr)
		} else {
			logger.LegacyPrintf("service.gateway", "Account %d: signature error retry build request failed: %v", account.ID, buildErr)
		}

		resp.Body = io.NopCloser(bytes.NewReader(respBody))
		return resp, true
	}

	errMsg := extractUpstreamErrorMessage(respBody)
	if isThinkingBudgetConstraintError(errMsg) && s.settingService.IsBudgetRectifierEnabled(ctx) {
		appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
			Platform:           account.Platform,
			AccountID:          account.ID,
			AccountName:        account.Name,
			UpstreamStatusCode: resp.StatusCode,
			UpstreamRequestID:  resp.Header.Get("x-request-id"),
			UpstreamURL:        safeUpstreamURL(upstreamReq.URL.String()),
			Kind:               "budget_constraint_error",
			Message:            errMsg,
			Detail: func() string {
				if s.cfg != nil && s.cfg.Gateway.LogUpstreamErrorBody {
					return truncateString(string(respBody), s.cfg.Gateway.LogUpstreamErrorBodyMaxBytes)
				}
				return ""
			}(),
		})

		rectifiedBody, applied := RectifyThinkingBudget(body)
		if applied && time.Since(retryStart) < maxRetryElapsed {
			logger.LegacyPrintf("service.gateway", "Account %d: detected budget_tokens constraint error, retrying with rectified budget (budget_tokens=%d, max_tokens=%d)", account.ID, BudgetRectifyBudgetTokens, BudgetRectifyMaxTokens)
			budgetRetryCtx, releaseBudgetRetryCtx := detachStreamUpstreamContext(ctx, reqStream)
			budgetRetryReq, buildErr := s.buildUpstreamRequest(budgetRetryCtx, c, account, rectifiedBody, token, tokenType, reqModel, reqStream, shouldMimicClaudeCode)
			releaseBudgetRetryCtx()
			if buildErr == nil {
				budgetRetryResp, retryErr := s.httpUpstream.DoWithTLS(budgetRetryReq, proxyURL, account.ID, account.Concurrency, tlsProfile)
				if retryErr == nil {
					return budgetRetryResp, true
				}
				if budgetRetryResp != nil && budgetRetryResp.Body != nil {
					_ = budgetRetryResp.Body.Close()
				}
				logger.LegacyPrintf("service.gateway", "Account %d: budget rectifier retry failed: %v", account.ID, retryErr)
			} else {
				logger.LegacyPrintf("service.gateway", "Account %d: budget rectifier retry build failed: %v", account.ID, buildErr)
			}
		}
	}

	resp.Body = io.NopCloser(bytes.NewReader(respBody))
	return resp, true
}

func looksLikeToolSignatureError(msg string) bool {
	m := strings.ToLower(msg)
	return strings.Contains(m, "tool_use") ||
		strings.Contains(m, "tool_result") ||
		strings.Contains(m, "functioncall") ||
		strings.Contains(m, "function_call") ||
		strings.Contains(m, "functionresponse") ||
		strings.Contains(m, "function_response")
}
