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

// ForwardGemini 转发 Gemini 协议请求
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
func (s *AntigravityGatewayService) ForwardGemini(ctx context.Context, c *gin.Context, account *Account, originalModel string, action string, stream bool, body []byte, isStickySession bool) (*ForwardResult, error) {
	startTime := time.Now()
	sessionID := getSessionID(c)
	prefix := logPrefix(sessionID, account.Name)

	if strings.TrimSpace(originalModel) == "" {
		return nil, s.writeGoogleError(c, http.StatusBadRequest, "Missing model in URL")
	}
	if strings.TrimSpace(action) == "" {
		return nil, s.writeGoogleError(c, http.StatusBadRequest, "Missing action in URL")
	}
	if len(body) == 0 {
		return nil, s.writeGoogleError(c, http.StatusBadRequest, "Request body is empty")
	}

	imageSize := s.extractImageSize(body)

	switch action {
	case "generateContent", "streamGenerateContent":
	case "countTokens":
		c.JSON(http.StatusOK, map[string]any{"totalTokens": 0})
		return &ForwardResult{
			RequestID:    "",
			Usage:        ClaudeUsage{},
			Model:        originalModel,
			Stream:       false,
			Duration:     time.Since(startTime),
			FirstTokenMs: nil,
		}, nil
	default:
		return nil, s.writeGoogleError(c, http.StatusNotFound, "Unsupported action: "+action)
	}

	mappedModel, billingModel := s.resolveGeminiForwardModels(account, originalModel)
	if mappedModel == "" {
		return nil, s.writeGoogleError(c, http.StatusForbidden, fmt.Sprintf("model %s not in whitelist", originalModel))
	}

	transport, err := s.resolveForwardTransportContext(ctx, account)
	if err != nil {
		if errors.Is(err, errAntigravityTokenProviderNotConfigured) {
			return nil, s.writeGoogleError(c, http.StatusBadGateway, "Antigravity token provider not configured")
		}
		return nil, &UpstreamFailoverError{
			StatusCode:   http.StatusBadGateway,
			ResponseBody: []byte(`{"error":{"message":"Failed to get upstream access token","status":"UNAVAILABLE"}}`),
		}
	}

	injectedBody, wrappedBody, err := s.buildGeminiUpstreamBody(account, body, transport.projectID, mappedModel)
	if err != nil {
		return nil, s.writeGoogleError(c, http.StatusBadRequest, "Invalid request body")
	}

	result, err := s.antigravityRetryLoop(s.newRetryLoopParams(
		ctx, prefix, account, transport, antigravityUpstreamStreamAction, wrappedBody, c, originalModel, isStickySession, s.handleUpstreamError,
	))
	if err != nil {
		return nil, s.mapGeminiRetryLoopError(c, err)
	}

	resp := result.resp
	defer func() {
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
		contentType := resp.Header.Get("Content-Type")
		_ = resp.Body.Close()
		resp.Body = io.NopCloser(bytes.NewReader(respBody))

		if s.settingService != nil && s.settingService.IsModelFallbackEnabled(ctx) &&
			isModelNotFoundError(resp.StatusCode, respBody) {
			fallbackModel := s.settingService.GetFallbackModel(ctx, PlatformAntigravity)
			if fallbackModel != "" && fallbackModel != mappedModel {
				logger.LegacyPrintf("service.antigravity_gateway", "[Antigravity] Model not found (%s), retrying with fallback model %s (account: %s)", mappedModel, fallbackModel, account.Name)

				fallbackWrapped, err := s.wrapV1InternalRequest(transport.projectID, fallbackModel, injectedBody)
				if err == nil {
					fallbackReq, err := antigravity.NewAPIRequest(ctx, antigravityUpstreamStreamAction, transport.accessToken, fallbackWrapped)
					if err == nil {
						fallbackResp, err := s.httpUpstream.Do(fallbackReq, transport.proxyURL, account.ID, account.Concurrency)
						if err == nil && fallbackResp.StatusCode < 400 {
							_ = resp.Body.Close()
							resp = fallbackResp
						} else if fallbackResp != nil {
							_ = fallbackResp.Body.Close()
						}
					}
				}
			}
		}

		resp, respBody, contentType, err = s.retryGeminiThoughtSignature(
			ctx, c, account, prefix, transport, originalModel, mappedModel, injectedBody, isStickySession, resp, respBody,
		)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode < 400 {
			goto handleSuccess
		}

		return nil, s.finalizeGeminiForwardError(ctx, c, account, prefix, originalModel, isStickySession, resp, respBody, contentType)
	}

handleSuccess:
	return s.completeGeminiForwardSuccess(c, resp, startTime, prefix, originalModel, billingModel, stream, imageSize)
}
