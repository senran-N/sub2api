package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/senran-N/sub2api/internal/pkg/logger"
)

func cloneHTTPResponseWithBody(resp *http.Response, body []byte) *http.Response {
	return &http.Response{
		StatusCode: resp.StatusCode,
		Header:     resp.Header.Clone(),
		Body:       io.NopCloser(bytes.NewReader(body)),
	}
}

func (s *AntigravityGatewayService) retryGeminiThoughtSignature(
	ctx context.Context,
	c *gin.Context,
	account *Account,
	prefix string,
	transport *antigravityForwardTransportContext,
	originalModel string,
	mappedModel string,
	injectedBody []byte,
	isStickySession bool,
	resp *http.Response,
	respBody []byte,
) (*http.Response, []byte, string, error) {
	signatureCheckBody := respBody
	if unwrapped, unwrapErr := s.unwrapV1InternalResponse(respBody); unwrapErr == nil && len(unwrapped) > 0 {
		signatureCheckBody = unwrapped
	}

	if resp.StatusCode != http.StatusBadRequest ||
		s.settingService == nil ||
		!s.settingService.IsSignatureRectifierEnabled(ctx) ||
		!isSignatureRelatedError(signatureCheckBody) ||
		!bytes.Contains(injectedBody, []byte(`"thoughtSignature"`)) {
		return resp, respBody, resp.Header.Get("Content-Type"), nil
	}

	upstreamMsg := sanitizeUpstreamErrorMessage(strings.TrimSpace(extractAntigravityErrorMessage(signatureCheckBody)))
	upstreamDetail := s.getUpstreamErrorDetail(signatureCheckBody)
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

	logger.LegacyPrintf("service.antigravity_gateway", "Antigravity Gemini account %d: detected signature-related 400, retrying with cleaned thought signatures", account.ID)

	cleanedInjectedBody := CleanGeminiNativeThoughtSignatures(injectedBody)
	retryWrappedBody, wrapErr := s.wrapV1InternalRequest(transport.projectID, mappedModel, cleanedInjectedBody)
	if wrapErr != nil {
		logger.LegacyPrintf("service.antigravity_gateway", "Antigravity Gemini account %d: signature retry wrap failed: %v", account.ID, wrapErr)
		return resp, respBody, resp.Header.Get("Content-Type"), nil
	}

	retryResult, retryErr := s.antigravityRetryLoop(s.newRetryLoopParams(
		ctx, prefix, account, transport, antigravityUpstreamStreamAction, retryWrappedBody, c, originalModel, isStickySession, s.handleUpstreamError,
	))
	if retryErr != nil {
		if switchErr, ok := IsAntigravityAccountSwitchError(retryErr); ok {
			appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
				Platform:           account.Platform,
				AccountID:          account.ID,
				AccountName:        account.Name,
				UpstreamStatusCode: http.StatusServiceUnavailable,
				Kind:               "failover",
				Message:            sanitizeUpstreamErrorMessage(retryErr.Error()),
			})
			return nil, nil, "", &UpstreamFailoverError{
				StatusCode:        http.StatusServiceUnavailable,
				ForceCacheBilling: switchErr.IsStickySession,
			}
		}

		appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
			Platform:           account.Platform,
			AccountID:          account.ID,
			AccountName:        account.Name,
			UpstreamStatusCode: 0,
			Kind:               "signature_retry_request_error",
			Message:            sanitizeUpstreamErrorMessage(retryErr.Error()),
		})
		logger.LegacyPrintf("service.antigravity_gateway", "Antigravity Gemini account %d: signature retry request failed: %v", account.ID, retryErr)
		return resp, respBody, resp.Header.Get("Content-Type"), nil
	}

	retryResp := retryResult.resp
	if retryResp.StatusCode < 400 {
		return retryResp, nil, retryResp.Header.Get("Content-Type"), nil
	}

	retryRespBody, _ := io.ReadAll(io.LimitReader(retryResp.Body, 2<<20))
	_ = retryResp.Body.Close()
	retryOpsBody := retryRespBody
	if retryUnwrapped, unwrapErr := s.unwrapV1InternalResponse(retryRespBody); unwrapErr == nil && len(retryUnwrapped) > 0 {
		retryOpsBody = retryUnwrapped
	}
	appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
		Platform:           account.Platform,
		AccountID:          account.ID,
		AccountName:        account.Name,
		UpstreamStatusCode: retryResp.StatusCode,
		UpstreamRequestID:  retryResp.Header.Get("x-request-id"),
		Kind:               "signature_retry",
		Message:            sanitizeUpstreamErrorMessage(strings.TrimSpace(extractAntigravityErrorMessage(retryOpsBody))),
		Detail:             s.getUpstreamErrorDetail(retryOpsBody),
	})

	return cloneHTTPResponseWithBody(retryResp, retryRespBody), retryRespBody, retryResp.Header.Get("Content-Type"), nil
}

func (s *AntigravityGatewayService) finalizeClaudeForwardError(
	ctx context.Context,
	c *gin.Context,
	account *Account,
	prefix string,
	originalModel string,
	isStickySession bool,
	resp *http.Response,
	respBody []byte,
) error {
	if resp.StatusCode == http.StatusBadRequest && isPromptTooLongError(respBody) {
		upstreamMsg := strings.TrimSpace(extractAntigravityErrorMessage(respBody))
		upstreamMsg = sanitizeUpstreamErrorMessage(upstreamMsg)
		upstreamDetail := s.getUpstreamErrorDetail(respBody)
		logBody, maxBytes := s.getLogConfig()
		if logBody {
			logger.LegacyPrintf("service.antigravity_gateway", "%s status=400 prompt_too_long=true upstream_message=%q request_id=%s body=%s", prefix, upstreamMsg, resp.Header.Get("x-request-id"), truncateForLog(respBody, maxBytes))
		}
		appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
			Platform:           account.Platform,
			AccountID:          account.ID,
			AccountName:        account.Name,
			UpstreamStatusCode: resp.StatusCode,
			UpstreamRequestID:  resp.Header.Get("x-request-id"),
			Kind:               "prompt_too_long",
			Message:            upstreamMsg,
			Detail:             upstreamDetail,
		})
		return &PromptTooLongError{
			StatusCode: resp.StatusCode,
			RequestID:  resp.Header.Get("x-request-id"),
			Body:       respBody,
		}
	}

	s.handleUpstreamError(ctx, prefix, account, resp.StatusCode, resp.Header, respBody, originalModel, 0, "", isStickySession)

	if resp.StatusCode == http.StatusBadRequest {
		msg := strings.ToLower(strings.TrimSpace(extractAntigravityErrorMessage(respBody)))
		if isGoogleProjectConfigError(msg) {
			upstreamMsg := sanitizeUpstreamErrorMessage(strings.TrimSpace(extractAntigravityErrorMessage(respBody)))
			upstreamDetail := s.getUpstreamErrorDetail(respBody)
			log.Printf("%s status=400 google_config_error failover=true upstream_message=%q account=%d", prefix, upstreamMsg, account.ID)
			appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
				Platform:           account.Platform,
				AccountID:          account.ID,
				AccountName:        account.Name,
				UpstreamStatusCode: resp.StatusCode,
				UpstreamRequestID:  resp.Header.Get("x-request-id"),
				Kind:               "failover",
				Message:            upstreamMsg,
				Detail:             upstreamDetail,
			})
			return &UpstreamFailoverError{StatusCode: resp.StatusCode, ResponseBody: respBody, RetryableOnSameAccount: true}
		}
	}

	if s.shouldFailoverUpstreamError(resp.StatusCode) {
		upstreamMsg := strings.TrimSpace(extractAntigravityErrorMessage(respBody))
		upstreamMsg = sanitizeUpstreamErrorMessage(upstreamMsg)
		upstreamDetail := s.getUpstreamErrorDetail(respBody)
		appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
			Platform:           account.Platform,
			AccountID:          account.ID,
			AccountName:        account.Name,
			UpstreamStatusCode: resp.StatusCode,
			UpstreamRequestID:  resp.Header.Get("x-request-id"),
			Kind:               "failover",
			Message:            upstreamMsg,
			Detail:             upstreamDetail,
		})
		return &UpstreamFailoverError{StatusCode: resp.StatusCode, ResponseBody: respBody}
	}

	return s.writeMappedClaudeError(c, account, resp.StatusCode, resp.Header.Get("x-request-id"), respBody)
}

func (s *AntigravityGatewayService) finalizeGeminiForwardError(
	ctx context.Context,
	c *gin.Context,
	account *Account,
	prefix string,
	originalModel string,
	isStickySession bool,
	resp *http.Response,
	respBody []byte,
	contentType string,
) error {
	requestID := setAntigravityRequestIDHeader(c, resp)

	unwrapped, unwrapErr := s.unwrapV1InternalResponse(respBody)
	unwrappedForOps := unwrapped
	if unwrapErr != nil || len(unwrappedForOps) == 0 {
		unwrappedForOps = respBody
	}

	s.handleUpstreamError(ctx, prefix, account, resp.StatusCode, resp.Header, respBody, originalModel, 0, "", isStickySession)
	upstreamMsg := strings.TrimSpace(extractAntigravityErrorMessage(unwrappedForOps))
	upstreamMsg = sanitizeUpstreamErrorMessage(upstreamMsg)
	upstreamDetail := s.getUpstreamErrorDetail(unwrappedForOps)

	setOpsUpstreamError(c, resp.StatusCode, upstreamMsg, upstreamDetail)

	if resp.StatusCode == http.StatusBadRequest && isGoogleProjectConfigError(strings.ToLower(upstreamMsg)) {
		log.Printf("%s status=400 google_config_error failover=true upstream_message=%q account=%d", prefix, upstreamMsg, account.ID)
		appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
			Platform:           account.Platform,
			AccountID:          account.ID,
			AccountName:        account.Name,
			UpstreamStatusCode: resp.StatusCode,
			UpstreamRequestID:  requestID,
			Kind:               "failover",
			Message:            upstreamMsg,
			Detail:             upstreamDetail,
		})
		return &UpstreamFailoverError{StatusCode: resp.StatusCode, ResponseBody: unwrappedForOps, RetryableOnSameAccount: true}
	}

	if s.shouldFailoverUpstreamError(resp.StatusCode) {
		appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
			Platform:           account.Platform,
			AccountID:          account.ID,
			AccountName:        account.Name,
			UpstreamStatusCode: resp.StatusCode,
			UpstreamRequestID:  requestID,
			Kind:               "failover",
			Message:            upstreamMsg,
			Detail:             upstreamDetail,
		})
		return &UpstreamFailoverError{StatusCode: resp.StatusCode, ResponseBody: unwrappedForOps}
	}

	if contentType == "" {
		contentType = "application/json"
	}
	appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
		Platform:           account.Platform,
		AccountID:          account.ID,
		AccountName:        account.Name,
		UpstreamStatusCode: resp.StatusCode,
		UpstreamRequestID:  requestID,
		Kind:               "http_error",
		Message:            upstreamMsg,
		Detail:             upstreamDetail,
	})
	logger.LegacyPrintf("service.antigravity_gateway", "[antigravity-Forward] upstream error status=%d body=%s", resp.StatusCode, truncateForLog(unwrappedForOps, 500))
	c.Data(resp.StatusCode, contentType, unwrappedForOps)
	return fmt.Errorf("antigravity upstream error: %d", resp.StatusCode)
}
