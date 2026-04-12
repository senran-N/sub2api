package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/senran-N/sub2api/internal/pkg/logger"
	"github.com/tidwall/sjson"
)

type openAIForwardPreparedRequest struct {
	body           []byte
	reqBody        map[string]any
	originalModel  string
	mappedModel    string
	promptCacheKey string
	reqStream      bool
}

func (s *OpenAIGatewayService) prepareOpenAIForwardRequest(
	c *gin.Context,
	account *Account,
	body []byte,
	reqModel string,
	reqStream bool,
	promptCacheKey string,
	isCodexCLI bool,
	wsDecision OpenAIWSProtocolDecision,
) (*openAIForwardPreparedRequest, error) {
	reqBody, err := getOpenAIRequestBodyMap(c, body)
	if err != nil {
		return nil, err
	}

	originalModel := reqModel
	if value, ok := reqBody["model"].(string); ok {
		reqModel = value
		originalModel = reqModel
	}
	if value, ok := reqBody["stream"].(bool); ok {
		reqStream = value
	}
	if promptCacheKey == "" {
		if value, ok := reqBody["prompt_cache_key"].(string); ok {
			promptCacheKey = strings.TrimSpace(value)
		}
	}

	bodyModified := false
	patchDisabled := false
	patchHasOp := false
	patchDelete := false
	patchPath := ""
	var patchValue any

	markPatchSet := func(path string, value any) {
		if strings.TrimSpace(path) == "" {
			patchDisabled = true
			return
		}
		if patchDisabled {
			return
		}
		if !patchHasOp {
			patchHasOp = true
			patchDelete = false
			patchPath = path
			patchValue = value
			return
		}
		if patchDelete || patchPath != path {
			patchDisabled = true
			return
		}
		patchValue = value
	}
	markPatchDelete := func(path string) {
		if strings.TrimSpace(path) == "" {
			patchDisabled = true
			return
		}
		if patchDisabled {
			return
		}
		if !patchHasOp {
			patchHasOp = true
			patchDelete = true
			patchPath = path
			return
		}
		if !patchDelete || patchPath != path {
			patchDisabled = true
		}
	}
	disablePatch := func() {
		patchDisabled = true
	}

	if isInstructionsEmpty(reqBody) {
		reqBody["instructions"] = defaultOpenAICodexInstructions
		bodyModified = true
		markPatchSet("instructions", defaultOpenAICodexInstructions)
	}

	if sanitizeEmptyBase64InputImagesInOpenAIRequestBodyMap(reqBody) {
		bodyModified = true
		disablePatch()
	}

	mappedModel := account.GetMappedModel(reqModel)
	if mappedModel != reqModel {
		logger.LegacyPrintf("service.openai_gateway", "[OpenAI] Model mapping applied: %s -> %s (account: %s, isCodexCLI: %v)", reqModel, mappedModel, account.Name, isCodexCLI)
		reqBody["model"] = mappedModel
		bodyModified = true
		markPatchSet("model", mappedModel)
	}

	if model, ok := reqBody["model"].(string); ok {
		normalizedModel := normalizeOpenAIModelForUpstream(account, model)
		if normalizedModel != "" && normalizedModel != model {
			logger.LegacyPrintf("service.openai_gateway", "[OpenAI] Codex model normalization: %s -> %s (account: %s, type: %s, isCodexCLI: %v)",
				model, normalizedModel, account.Name, account.Type, isCodexCLI)
			reqBody["model"] = normalizedModel
			mappedModel = normalizedModel
			bodyModified = true
			markPatchSet("model", normalizedModel)
		}

		if !SupportsVerbosity(normalizedModel) {
			if text, ok := reqBody["text"].(map[string]any); ok {
				delete(text, "verbosity")
			}
		}
	}

	if reasoning, ok := reqBody["reasoning"].(map[string]any); ok {
		if effort, ok := reasoning["effort"].(string); ok && effort == "minimal" {
			reasoning["effort"] = "none"
			bodyModified = true
			markPatchSet("reasoning.effort", "none")
			logger.LegacyPrintf("service.openai_gateway", "[OpenAI] Normalized reasoning.effort: minimal -> none (account: %s)", account.Name)
		}
	}

	if account.Type == AccountTypeOAuth {
		codexResult := applyCodexOAuthTransform(reqBody, isCodexCLI, isOpenAIResponsesCompactPath(c))
		if codexResult.Modified {
			bodyModified = true
			disablePatch()
		}
		if codexResult.NormalizedModel != "" {
			mappedModel = codexResult.NormalizedModel
		}
		if codexResult.PromptCacheKey != "" {
			promptCacheKey = codexResult.PromptCacheKey
		}
	}

	if !isCodexCLI {
		if maxOutputTokens, hasMaxOutputTokens := reqBody["max_output_tokens"]; hasMaxOutputTokens {
			switch account.Platform {
			case PlatformOpenAI:
				if account.Type == AccountTypeAPIKey {
					delete(reqBody, "max_output_tokens")
					bodyModified = true
					markPatchDelete("max_output_tokens")
				}
			case PlatformAnthropic:
				delete(reqBody, "max_output_tokens")
				markPatchDelete("max_output_tokens")
				if _, hasMaxTokens := reqBody["max_tokens"]; !hasMaxTokens {
					reqBody["max_tokens"] = maxOutputTokens
					disablePatch()
				}
				bodyModified = true
			case PlatformGemini:
				delete(reqBody, "max_output_tokens")
				bodyModified = true
				markPatchDelete("max_output_tokens")
			default:
				delete(reqBody, "max_output_tokens")
				bodyModified = true
				markPatchDelete("max_output_tokens")
			}
		}

		if _, hasMaxCompletionTokens := reqBody["max_completion_tokens"]; hasMaxCompletionTokens {
			if account.Type == AccountTypeAPIKey || account.Platform != PlatformOpenAI {
				delete(reqBody, "max_completion_tokens")
				bodyModified = true
				markPatchDelete("max_completion_tokens")
			}
		}

		for _, unsupportedField := range []string{"prompt_cache_retention", "safety_identifier", "user"} {
			if _, has := reqBody[unsupportedField]; has {
				delete(reqBody, unsupportedField)
				bodyModified = true
				markPatchDelete(unsupportedField)
			}
		}
	}

	if bodyModified {
		serializedByPatch := false
		if !patchDisabled && patchHasOp {
			var patchErr error
			if patchDelete {
				body, patchErr = sjson.DeleteBytes(body, patchPath)
			} else {
				body, patchErr = sjson.SetBytes(body, patchPath, patchValue)
			}
			if patchErr == nil {
				serializedByPatch = true
			}
		}
		if !serializedByPatch {
			body, err = json.Marshal(reqBody)
			if err != nil {
				return nil, fmt.Errorf("serialize request body: %w", err)
			}
		}
	}

	return &openAIForwardPreparedRequest{
		body:           body,
		reqBody:        reqBody,
		originalModel:  originalModel,
		mappedModel:    mappedModel,
		promptCacheKey: promptCacheKey,
		reqStream:      reqStream,
	}, nil
}

func (s *OpenAIGatewayService) forwardPreparedOpenAIWS(
	ctx context.Context,
	c *gin.Context,
	account *Account,
	prepared *openAIForwardPreparedRequest,
	token string,
	wsDecision OpenAIWSProtocolDecision,
	startTime time.Time,
	isCodexCLI bool,
) (*OpenAIForwardResult, error) {
	wsReqBody := prepared.reqBody
	if len(prepared.reqBody) > 0 {
		wsReqBody = make(map[string]any, len(prepared.reqBody))
		for key, value := range prepared.reqBody {
			wsReqBody[key] = value
		}
	}
	_, hasPreviousResponseID := wsReqBody["previous_response_id"]
	logOpenAIWSModeDebug(
		"forward_start account_id=%d account_type=%s model=%s stream=%v has_previous_response_id=%v",
		account.ID,
		account.Type,
		prepared.mappedModel,
		prepared.reqStream,
		hasPreviousResponseID,
	)

	maxAttempts := openAIWSReconnectRetryLimit + 1
	wsAttempts := 0
	var wsResult *OpenAIForwardResult
	var wsErr error
	wsLastFailureReason := ""
	wsPrevResponseRecoveryTried := false
	wsInvalidEncryptedContentRecoveryTried := false

	recoverPrevResponseNotFound := func(attempt int) bool {
		if wsPrevResponseRecoveryTried {
			return false
		}
		previousResponseID := openAIWSPayloadString(wsReqBody, "previous_response_id")
		if previousResponseID == "" {
			logOpenAIWSModeInfo(
				"reconnect_prev_response_recovery_skip account_id=%d attempt=%d reason=missing_previous_response_id previous_response_id_present=false",
				account.ID,
				attempt,
			)
			return false
		}
		if HasFunctionCallOutput(wsReqBody) {
			logOpenAIWSModeInfo(
				"reconnect_prev_response_recovery_skip account_id=%d attempt=%d reason=has_function_call_output previous_response_id_present=true",
				account.ID,
				attempt,
			)
			return false
		}
		delete(wsReqBody, "previous_response_id")
		wsPrevResponseRecoveryTried = true
		logOpenAIWSModeInfo(
			"reconnect_prev_response_recovery account_id=%d attempt=%d action=drop_previous_response_id retry=1 previous_response_id=%s previous_response_id_kind=%s",
			account.ID,
			attempt,
			truncateOpenAIWSLogValue(previousResponseID, openAIWSIDValueMaxLen),
			normalizeOpenAIWSLogValue(ClassifyOpenAIPreviousResponseIDKind(previousResponseID)),
		)
		return true
	}

	recoverInvalidEncryptedContent := func(attempt int) bool {
		if wsInvalidEncryptedContentRecoveryTried {
			return false
		}
		removedReasoningItems := trimOpenAIEncryptedReasoningItems(wsReqBody)
		if !removedReasoningItems {
			logOpenAIWSModeInfo(
				"reconnect_invalid_encrypted_content_recovery_skip account_id=%d attempt=%d reason=missing_encrypted_reasoning_items",
				account.ID,
				attempt,
			)
			return false
		}
		previousResponseID := openAIWSPayloadString(wsReqBody, "previous_response_id")
		hasFunctionCallOutput := HasFunctionCallOutput(wsReqBody)
		if previousResponseID != "" && !hasFunctionCallOutput {
			delete(wsReqBody, "previous_response_id")
		}
		wsInvalidEncryptedContentRecoveryTried = true
		logOpenAIWSModeInfo(
			"reconnect_invalid_encrypted_content_recovery account_id=%d attempt=%d action=drop_encrypted_reasoning_items retry=1 previous_response_id_present=%v previous_response_id=%s previous_response_id_kind=%s has_function_call_output=%v dropped_previous_response_id=%v",
			account.ID,
			attempt,
			previousResponseID != "",
			truncateOpenAIWSLogValue(previousResponseID, openAIWSIDValueMaxLen),
			normalizeOpenAIWSLogValue(ClassifyOpenAIPreviousResponseIDKind(previousResponseID)),
			hasFunctionCallOutput,
			previousResponseID != "" && !hasFunctionCallOutput,
		)
		return true
	}

	retryBudget := s.openAIWSRetryTotalBudget()
	retryStartedAt := time.Now()

wsRetryLoop:
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		wsAttempts = attempt
		wsResult, wsErr = s.forwardOpenAIWSV2(
			ctx,
			c,
			account,
			wsReqBody,
			token,
			wsDecision,
			isCodexCLI,
			prepared.reqStream,
			prepared.originalModel,
			prepared.mappedModel,
			startTime,
			attempt,
			wsLastFailureReason,
		)
		if wsErr == nil {
			break
		}
		if c != nil && c.Writer != nil && c.Writer.Written() {
			break
		}

		reason, retryable := classifyOpenAIWSReconnectReason(wsErr)
		if reason != "" {
			wsLastFailureReason = reason
		}
		if reason == "previous_response_not_found" && recoverPrevResponseNotFound(attempt) {
			continue
		}
		if reason == "invalid_encrypted_content" && recoverInvalidEncryptedContent(attempt) {
			continue
		}
		if retryable && attempt < maxAttempts {
			backoff := s.openAIWSRetryBackoff(attempt)
			if retryBudget > 0 && time.Since(retryStartedAt)+backoff > retryBudget {
				s.recordOpenAIWSRetryExhausted()
				logOpenAIWSModeInfo(
					"reconnect_budget_exhausted account_id=%d attempts=%d max_retries=%d reason=%s elapsed_ms=%d budget_ms=%d",
					account.ID,
					attempt,
					openAIWSReconnectRetryLimit,
					normalizeOpenAIWSLogValue(reason),
					time.Since(retryStartedAt).Milliseconds(),
					retryBudget.Milliseconds(),
				)
				break
			}
			s.recordOpenAIWSRetryAttempt(backoff)
			logOpenAIWSModeInfo(
				"reconnect_retry account_id=%d retry=%d max_retries=%d reason=%s backoff_ms=%d",
				account.ID,
				attempt,
				openAIWSReconnectRetryLimit,
				normalizeOpenAIWSLogValue(reason),
				backoff.Milliseconds(),
			)
			if backoff > 0 {
				timer := time.NewTimer(backoff)
				select {
				case <-ctx.Done():
					if !timer.Stop() {
						<-timer.C
					}
					wsErr = wrapOpenAIWSFallback("retry_backoff_canceled", ctx.Err())
					break wsRetryLoop
				case <-timer.C:
				}
			}
			continue
		}
		if retryable {
			s.recordOpenAIWSRetryExhausted()
			logOpenAIWSModeInfo(
				"reconnect_exhausted account_id=%d attempts=%d max_retries=%d reason=%s",
				account.ID,
				attempt,
				openAIWSReconnectRetryLimit,
				normalizeOpenAIWSLogValue(reason),
			)
		} else if reason != "" {
			s.recordOpenAIWSNonRetryableFastFallback()
			logOpenAIWSModeInfo(
				"reconnect_stop account_id=%d attempt=%d reason=%s",
				account.ID,
				attempt,
				normalizeOpenAIWSLogValue(reason),
			)
		}
		break
	}

	if wsErr == nil {
		firstTokenMs := int64(0)
		hasFirstTokenMs := wsResult != nil && wsResult.FirstTokenMs != nil
		if hasFirstTokenMs {
			firstTokenMs = int64(*wsResult.FirstTokenMs)
		}
		requestID := ""
		if wsResult != nil {
			requestID = strings.TrimSpace(wsResult.RequestID)
		}
		logOpenAIWSModeDebug(
			"forward_succeeded account_id=%d request_id=%s stream=%v has_first_token_ms=%v first_token_ms=%d ws_attempts=%d",
			account.ID,
			requestID,
			prepared.reqStream,
			hasFirstTokenMs,
			firstTokenMs,
			wsAttempts,
		)
		wsResult.UpstreamModel = prepared.mappedModel
		return wsResult, nil
	}

	s.writeOpenAIWSFallbackErrorResponse(c, account, wsErr)
	return nil, wsErr
}

func (s *OpenAIGatewayService) forwardPreparedOpenAIHTTP(
	ctx context.Context,
	c *gin.Context,
	account *Account,
	prepared *openAIForwardPreparedRequest,
	token string,
	startTime time.Time,
	isCodexCLI bool,
) (*OpenAIForwardResult, error) {
	httpInvalidEncryptedContentRetryTried := false

	for {
		upstreamCtx, releaseUpstreamCtx := detachStreamUpstreamContext(ctx, prepared.reqStream)
		upstreamReq, err := s.buildUpstreamRequest(upstreamCtx, c, account, prepared.body, token, prepared.reqStream, prepared.promptCacheKey, isCodexCLI)
		releaseUpstreamCtx()
		if err != nil {
			return nil, err
		}

		proxyURL := ""
		if account.ProxyID != nil && account.Proxy != nil {
			proxyURL = account.Proxy.URL()
		}

		upstreamStart := time.Now()
		resp, err := s.httpUpstream.Do(upstreamReq, proxyURL, account.ID, account.Concurrency)
		SetOpsLatencyMs(c, OpsUpstreamLatencyMsKey, time.Since(upstreamStart).Milliseconds())
		if err != nil {
			safeErr := sanitizeUpstreamErrorMessage(err.Error())
			setOpsUpstreamError(c, 0, safeErr, "")
			appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
				Platform:           account.Platform,
				AccountID:          account.ID,
				AccountName:        account.Name,
				UpstreamStatusCode: 0,
				Kind:               "request_error",
				Message:            safeErr,
			})
			c.JSON(http.StatusBadGateway, gin.H{
				"error": gin.H{
					"type":    "upstream_error",
					"message": "Upstream request failed",
				},
			})
			return nil, fmt.Errorf("upstream request failed: %s", safeErr)
		}

		if resp.StatusCode >= 400 {
			respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
			_ = resp.Body.Close()
			resp.Body = io.NopCloser(bytes.NewReader(respBody))

			upstreamMsg := strings.TrimSpace(extractUpstreamErrorMessage(respBody))
			upstreamMsg = sanitizeUpstreamErrorMessage(upstreamMsg)
			upstreamCode := extractUpstreamErrorCode(respBody)
			if !httpInvalidEncryptedContentRetryTried && resp.StatusCode == http.StatusBadRequest && upstreamCode == "invalid_encrypted_content" {
				if trimOpenAIEncryptedReasoningItems(prepared.reqBody) {
					prepared.body, err = json.Marshal(prepared.reqBody)
					if err != nil {
						return nil, fmt.Errorf("serialize invalid_encrypted_content retry body: %w", err)
					}
					setOpsUpstreamRequestBody(c, prepared.body)
					httpInvalidEncryptedContentRetryTried = true
					logger.LegacyPrintf("service.openai_gateway", "[OpenAI] Retrying non-WSv2 request once after invalid_encrypted_content (account: %s)", account.Name)
					continue
				}
				logger.LegacyPrintf("service.openai_gateway", "[OpenAI] Skip non-WSv2 invalid_encrypted_content retry because encrypted reasoning items are missing (account: %s)", account.Name)
			}
			if s.shouldFailoverOpenAIUpstreamResponse(resp.StatusCode, upstreamMsg, respBody) {
				upstreamDetail := ""
				if s.cfg != nil && s.cfg.Gateway.LogUpstreamErrorBody {
					maxBytes := s.cfg.Gateway.LogUpstreamErrorBodyMaxBytes
					if maxBytes <= 0 {
						maxBytes = 2048
					}
					upstreamDetail = truncateString(string(respBody), maxBytes)
				}
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

				s.handleFailoverSideEffects(ctx, resp, account)
				return nil, &UpstreamFailoverError{
					StatusCode:             resp.StatusCode,
					ResponseBody:           respBody,
					RetryableOnSameAccount: account.IsPoolMode() && (isOpenAIPoolModeRetryableStatus(resp.StatusCode) || isOpenAITransientProcessingError(resp.StatusCode, upstreamMsg, respBody)),
				}
			}
			return s.handleErrorResponse(ctx, resp, c, account, prepared.body)
		}
		defer func() { _ = resp.Body.Close() }()

		var usage *OpenAIUsage
		var firstTokenMs *int
		if prepared.reqStream {
			streamResult, err := s.handleStreamingResponse(ctx, resp, c, account, startTime, prepared.originalModel, prepared.mappedModel)
			if err != nil {
				return nil, err
			}
			usage = streamResult.usage
			firstTokenMs = streamResult.firstTokenMs
		} else {
			usage, err = s.handleNonStreamingResponse(ctx, resp, c, account, prepared.originalModel, prepared.mappedModel)
			if err != nil {
				return nil, err
			}
		}

		if account.Type == AccountTypeOAuth {
			if snapshot := ParseCodexRateLimitHeaders(resp.Header); snapshot != nil {
				s.updateCodexUsageSnapshot(ctx, account.ID, snapshot)
			}
		}

		if usage == nil {
			usage = &OpenAIUsage{}
		}

		reasoningEffort := extractOpenAIReasoningEffort(prepared.reqBody, prepared.originalModel)
		serviceTier := extractOpenAIServiceTier(prepared.reqBody)

		return &OpenAIForwardResult{
			RequestID:       resp.Header.Get("x-request-id"),
			Usage:           *usage,
			Model:           prepared.originalModel,
			UpstreamModel:   prepared.mappedModel,
			ServiceTier:     serviceTier,
			ReasoningEffort: reasoningEffort,
			Stream:          prepared.reqStream,
			OpenAIWSMode:    false,
			Duration:        time.Since(startTime),
			FirstTokenMs:    firstTokenMs,
		}, nil
	}
}
