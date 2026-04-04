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
	"github.com/senran-N/sub2api/internal/pkg/antigravity"
	"github.com/senran-N/sub2api/internal/pkg/logger"
)

// Forward 转发 Claude 协议请求（Claude → Gemini 转换）
//
// 限流处理流程:
//
//	请求 → antigravityRetryLoop → 预检查(remaining>0? → 切换账号) → 发送上游
//	  ├─ 成功 → 正常返回
//	  └─ 429/503 → handleSmartRetry
//	      ├─ retryDelay >= 7s → 设置模型限流 + 清除粘性绑定 → 切换账号
//	      └─ retryDelay <  7s → 等待后重试 1 次
//	          ├─ 成功 → 正常返回
//	          └─ 失败 → 设置模型限流 + 清除粘性绑定 → 切换账号
func (s *AntigravityGatewayService) Forward(ctx context.Context, c *gin.Context, account *Account, body []byte, isStickySession bool) (*ForwardResult, error) {
	if account.Type == AccountTypeUpstream {
		return s.ForwardUpstream(ctx, c, account, body)
	}

	startTime := time.Now()
	sessionID := getSessionID(c)
	prefix := logPrefix(sessionID, account.Name)

	claudeReq, err := parseClaudeForwardRequest(body)
	if err != nil {
		return nil, s.writeClaudeError(c, http.StatusBadRequest, "invalid_request_error", "Invalid request body")
	}
	if strings.TrimSpace(claudeReq.Model) == "" {
		return nil, s.writeClaudeError(c, http.StatusBadRequest, "invalid_request_error", "Missing model")
	}

	originalModel, mappedModel, billingModel := s.resolveClaudeForwardModels(account, &claudeReq)
	if mappedModel == "" {
		return nil, s.writeClaudeError(c, http.StatusForbidden, "permission_error", fmt.Sprintf("model %s not in whitelist", claudeReq.Model))
	}

	transport, err := s.resolveForwardTransportContext(ctx, account)
	if err != nil {
		if errors.Is(err, errAntigravityTokenProviderNotConfigured) {
			return nil, s.writeClaudeError(c, http.StatusBadGateway, "api_error", "Antigravity token provider not configured")
		}
		return nil, &UpstreamFailoverError{
			StatusCode:   http.StatusBadGateway,
			ResponseBody: []byte(`{"error":{"type":"authentication_error","message":"Failed to get upstream access token"},"type":"error"}`),
		}
	}

	geminiBody, _, err := s.buildClaudeUpstreamBody(ctx, &claudeReq, transport.projectID, mappedModel)
	if err != nil {
		return nil, s.writeClaudeError(c, http.StatusBadRequest, "invalid_request_error", "Invalid request")
	}

	result, err := s.antigravityRetryLoop(s.newRetryLoopParams(
		ctx, prefix, account, transport, antigravityUpstreamStreamAction, geminiBody, c, originalModel, isStickySession, s.handleUpstreamError,
	))
	if err != nil {
		return nil, s.mapClaudeRetryLoopError(c, err)
	}

	resp := result.resp
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))

		if resp.StatusCode == http.StatusBadRequest && isSignatureRelatedError(respBody) && s.settingService.IsSignatureRectifierEnabled(ctx) {
			upstreamMsg := strings.TrimSpace(extractAntigravityErrorMessage(respBody))
			upstreamMsg = sanitizeUpstreamErrorMessage(upstreamMsg)
			logBody, maxBytes := s.getLogConfig()
			upstreamDetail := s.getUpstreamErrorDetail(respBody)
			appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
				Platform:           account.Platform,
				AccountID:          account.ID,
				AccountName:        account.Name,
				UpstreamStatusCode: resp.StatusCode,
				UpstreamRequestID:  resp.Header.Get("x-request-id"),
				Kind:               "signature_error",
				Message:            upstreamMsg,
				Detail:             upstreamDetail,
			})

			retryStages := []struct {
				name  string
				strip func(*antigravity.ClaudeRequest) (bool, error)
			}{
				{name: "thinking-only", strip: stripThinkingFromClaudeRequest},
				{name: "thinking+tools", strip: stripSignatureSensitiveBlocksFromClaudeRequest},
			}

			for _, stage := range retryStages {
				retryClaudeReq := claudeReq
				retryClaudeReq.Messages = append([]antigravity.ClaudeMessage(nil), claudeReq.Messages...)

				stripped, stripErr := stage.strip(&retryClaudeReq)
				if stripErr != nil || !stripped {
					continue
				}

				logger.LegacyPrintf("service.antigravity_gateway", "Antigravity account %d: detected signature-related 400, retrying once (%s)", account.ID, stage.name)

				retryGeminiBody, _, txErr := s.buildClaudeUpstreamBody(ctx, &retryClaudeReq, transport.projectID, mappedModel)
				if txErr != nil {
					continue
				}
				retryResult, retryErr := s.antigravityRetryLoop(s.newRetryLoopParams(
					ctx, prefix, account, transport, antigravityUpstreamStreamAction, retryGeminiBody, c, originalModel, isStickySession, s.handleUpstreamError,
				))
				if retryErr != nil {
					appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
						Platform:           account.Platform,
						AccountID:          account.ID,
						AccountName:        account.Name,
						UpstreamStatusCode: 0,
						Kind:               "signature_retry_request_error",
						Message:            sanitizeUpstreamErrorMessage(retryErr.Error()),
					})
					logger.LegacyPrintf("service.antigravity_gateway", "Antigravity account %d: signature retry request failed (%s): %v", account.ID, stage.name, retryErr)
					continue
				}

				retryResp := retryResult.resp
				if retryResp.StatusCode < 400 {
					_ = resp.Body.Close()
					resp = retryResp
					respBody = nil
					break
				}

				retryBody, _ := io.ReadAll(io.LimitReader(retryResp.Body, 8<<10))
				_ = retryResp.Body.Close()
				if retryResp.StatusCode == http.StatusTooManyRequests {
					retryBaseURL := ""
					if retryResp.Request != nil && retryResp.Request.URL != nil {
						retryBaseURL = retryResp.Request.URL.Scheme + "://" + retryResp.Request.URL.Host
					}
					logger.LegacyPrintf("service.antigravity_gateway", "%s status=429 rate_limited base_url=%s retry_stage=%s body=%s", prefix, retryBaseURL, stage.name, truncateForLog(retryBody, 200))
				}

				kind := "signature_retry"
				if strings.TrimSpace(stage.name) != "" {
					kind = "signature_retry_" + strings.ReplaceAll(stage.name, "+", "_")
				}
				retryUpstreamMsg := strings.TrimSpace(extractAntigravityErrorMessage(retryBody))
				retryUpstreamMsg = sanitizeUpstreamErrorMessage(retryUpstreamMsg)
				retryUpstreamDetail := ""
				if logBody {
					retryUpstreamDetail = truncateString(string(retryBody), maxBytes)
				}
				appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
					Platform:           account.Platform,
					AccountID:          account.ID,
					AccountName:        account.Name,
					UpstreamStatusCode: retryResp.StatusCode,
					UpstreamRequestID:  retryResp.Header.Get("x-request-id"),
					Kind:               kind,
					Message:            retryUpstreamMsg,
					Detail:             retryUpstreamDetail,
				})

				if retryResp.StatusCode != http.StatusBadRequest || !isSignatureRelatedError(retryBody) {
					respBody = retryBody
					resp = &http.Response{
						StatusCode: retryResp.StatusCode,
						Header:     retryResp.Header.Clone(),
						Body:       io.NopCloser(bytes.NewReader(retryBody)),
					}
					break
				}

				respBody = retryBody
				resp = &http.Response{
					StatusCode: retryResp.StatusCode,
					Header:     retryResp.Header.Clone(),
					Body:       io.NopCloser(bytes.NewReader(retryBody)),
				}
			}
		}

		if resp.StatusCode == http.StatusBadRequest && respBody != nil && !isSignatureRelatedError(respBody) {
			errMsg := strings.TrimSpace(extractAntigravityErrorMessage(respBody))
			if isThinkingBudgetConstraintError(errMsg) && s.settingService.IsBudgetRectifierEnabled(ctx) {
				appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
					Platform:           account.Platform,
					AccountID:          account.ID,
					AccountName:        account.Name,
					UpstreamStatusCode: resp.StatusCode,
					UpstreamRequestID:  resp.Header.Get("x-request-id"),
					Kind:               "budget_constraint_error",
					Message:            errMsg,
					Detail:             s.getUpstreamErrorDetail(respBody),
				})

				if claudeReq.Thinking == nil || claudeReq.Thinking.Type != "adaptive" {
					retryClaudeReq := claudeReq
					retryClaudeReq.Messages = append([]antigravity.ClaudeMessage(nil), claudeReq.Messages...)
					retryClaudeReq.Thinking = &antigravity.ThinkingConfig{
						Type:         "enabled",
						BudgetTokens: BudgetRectifyBudgetTokens,
					}
					if retryClaudeReq.MaxTokens < BudgetRectifyMinMaxTokens {
						retryClaudeReq.MaxTokens = BudgetRectifyMaxTokens
					}

					logger.LegacyPrintf("service.antigravity_gateway", "Antigravity account %d: detected budget_tokens constraint error, retrying with rectified budget (budget_tokens=%d, max_tokens=%d)", account.ID, BudgetRectifyBudgetTokens, BudgetRectifyMaxTokens)

					retryGeminiBody, _, txErr := s.buildClaudeUpstreamBody(ctx, &retryClaudeReq, transport.projectID, mappedModel)
					if txErr == nil {
						retryResult, retryErr := s.antigravityRetryLoop(s.newRetryLoopParams(
							ctx, prefix, account, transport, antigravityUpstreamStreamAction, retryGeminiBody, c, originalModel, isStickySession, s.handleUpstreamError,
						))
						if retryErr == nil {
							retryResp := retryResult.resp
							if retryResp.StatusCode < 400 {
								_ = resp.Body.Close()
								resp = retryResp
								respBody = nil
							} else {
								retryBody, _ := io.ReadAll(io.LimitReader(retryResp.Body, 2<<20))
								_ = retryResp.Body.Close()
								respBody = retryBody
								resp = &http.Response{
									StatusCode: retryResp.StatusCode,
									Header:     retryResp.Header.Clone(),
									Body:       io.NopCloser(bytes.NewReader(retryBody)),
								}
							}
						} else {
							logger.LegacyPrintf("service.antigravity_gateway", "Antigravity account %d: budget rectifier retry failed: %v", account.ID, retryErr)
						}
					}
				}
			}
		}

		if resp.StatusCode >= 400 {
			return nil, s.finalizeClaudeForwardError(ctx, c, account, prefix, originalModel, isStickySession, resp, respBody)
		}
	}

	return s.completeClaudeForwardSuccess(c, resp, startTime, prefix, originalModel, billingModel, claudeReq.Stream)
}
