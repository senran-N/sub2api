package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/senran-N/sub2api/internal/pkg/apicompat"
	"github.com/senran-N/sub2api/internal/pkg/claude"
	"github.com/senran-N/sub2api/internal/pkg/logger"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"go.uber.org/zap"
)

// CompatibleGatewayTextRuntime is the neutral execution seam for compatible
// text protocols shared by OpenAI-compatible platforms.
type CompatibleGatewayTextRuntime struct {
	openaiGatewayService *OpenAIGatewayService
}

func NewCompatibleGatewayTextRuntime(openaiGatewayService *OpenAIGatewayService) *CompatibleGatewayTextRuntime {
	return &CompatibleGatewayTextRuntime{openaiGatewayService: openaiGatewayService}
}

func ProvideCompatibleGatewayTextRuntime(openaiGatewayService *OpenAIGatewayService) *CompatibleGatewayTextRuntime {
	if openaiGatewayService == nil {
		return nil
	}
	return NewCompatibleGatewayTextRuntime(openaiGatewayService)
}

func (r *CompatibleGatewayTextRuntime) ForwardResponses(
	ctx context.Context,
	c *gin.Context,
	account *Account,
	body []byte,
	defaultMappedModel string,
) (result *OpenAIForwardResult, err error) {
	if r == nil || r.openaiGatewayService == nil {
		return nil, errors.New("compatible gateway text runtime is not configured")
	}
	s := r.openaiGatewayService
	startTime := time.Now()
	requestedModel := strings.TrimSpace(getOpenAIRequestMeta(c, body).Model)
	upstreamModelForFeedback := requestedModel
	statusCode := 0
	defer func() {
		r.persistRuntimeFeedback(
			ctx,
			account,
			requestedModel,
			upstreamModelForFeedback,
			CompatibleGatewayProtocolFamilyResponses,
			statusCode,
			result,
			err,
		)
	}()

	restrictionResult := s.detectCodexClientRestriction(c, account)
	apiKeyID := getAPIKeyIDFromContext(c)
	logCodexCLIOnlyDetection(ctx, c, account, apiKeyID, restrictionResult, body)
	if restrictionResult.Enabled && !restrictionResult.Matched {
		c.JSON(http.StatusForbidden, gin.H{
			"error": gin.H{
				"type":    "forbidden_error",
				"message": "This account only allows Codex official clients",
			},
		})
		return nil, errors.New("codex_cli_only restriction: only codex official clients are allowed")
	}

	originalBody := body
	forceCodexCLI := s != nil && s.cfg != nil && s.cfg.Gateway.ForceCodexCLI
	profile := GetCodexRequestProfile(c, body, forceCodexCLI)
	reqMeta := getOpenAIRequestMeta(c, body)
	reqModel, reqStream, promptCacheKey := reqMeta.Model, reqMeta.Stream, reqMeta.PromptCacheKey

	isCodexCLI := profile.NativeClient
	wsDecision := s.getOpenAIWSProtocolResolver().ResolveWithProfile(account, profile)
	clientTransport := GetOpenAIClientTransport(c)
	wsDecision = resolveOpenAIWSDecisionByClientTransport(wsDecision, clientTransport)
	if c != nil {
		c.Set("openai_ws_transport_decision", string(wsDecision.Transport))
		c.Set("openai_ws_transport_reason", wsDecision.Reason)
	}
	if wsDecision.Transport == OpenAIUpstreamTransportResponsesWebsocketV2 {
		logOpenAIWSModeDebug(
			"selected account_id=%d account_type=%s transport=%s reason=%s model=%s stream=%v",
			account.ID,
			account.Type,
			normalizeOpenAIWSLogValue(string(wsDecision.Transport)),
			normalizeOpenAIWSLogValue(wsDecision.Reason),
			reqModel,
			reqStream,
		)
	}
	if wsDecision.Transport == OpenAIUpstreamTransportResponsesWebsocket {
		if c != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"type":    "invalid_request_error",
					"message": "OpenAI WSv1 is temporarily unsupported. Please enable responses_websockets_v2.",
				},
			})
		}
		return nil, errors.New("openai ws v1 is temporarily unsupported; use ws v2")
	}
	passthroughEnabled := account.IsOpenAIPassthroughEnabled()
	if passthroughEnabled {
		forwardBody := originalBody
		mappedModel := resolveOpenAIForwardModel(account, reqModel, defaultMappedModel)
		mappedModel = normalizeOpenAIModelForUpstream(account, mappedModel)
		upstreamModelForFeedback = strings.TrimSpace(mappedModel)
		if mappedModel != "" && mappedModel != reqModel {
			var patchErr error
			forwardBody, patchErr = sjson.SetBytes(originalBody, "model", mappedModel)
			if patchErr != nil {
				return nil, fmt.Errorf("patch passthrough model: %w", patchErr)
			}
		}
		var reasoningEffort *string
		if reqMeta.ReasoningPresent {
			if reqMeta.ReasoningEffort == "" {
				reasoningEffort = nil
			} else {
				value := reqMeta.ReasoningEffort
				reasoningEffort = &value
			}
		} else if reqMeta.ReasoningEffort != "" {
			value := reqMeta.ReasoningEffort
			reasoningEffort = &value
		} else {
			reasoningEffort = extractOpenAIReasoningEffortFromBody(body, reqModel)
		}
		result, err = s.forwardOpenAIPassthrough(ctx, c, account, forwardBody, reqModel, reasoningEffort, reqStream, startTime)
		if err != nil {
			return nil, err
		}
		if result != nil && mappedModel != "" && mappedModel != reqModel {
			result.UpstreamModel = mappedModel
		}
		return result, nil
	}

	prepared, err := s.prepareOpenAIForwardRequest(c, account, body, reqModel, reqStream, promptCacheKey, defaultMappedModel, isCodexCLI, wsDecision)
	if err != nil {
		return nil, err
	}
	upstreamModelForFeedback = strings.TrimSpace(prepared.mappedModel)

	token, _, err := s.GetAccessToken(ctx, account)
	if err != nil {
		return nil, err
	}

	setOpsUpstreamRequestBody(c, prepared.body)

	if wsDecision.Transport == OpenAIUpstreamTransportResponsesWebsocketV2 {
		return s.forwardPreparedOpenAIWS(ctx, c, account, prepared, token, wsDecision, startTime, isCodexCLI)
	}

	return s.forwardPreparedOpenAIHTTP(ctx, c, account, prepared, token, startTime, isCodexCLI)
}

func (r *CompatibleGatewayTextRuntime) ForwardChatCompletions(
	ctx context.Context,
	c *gin.Context,
	account *Account,
	body []byte,
	promptCacheKey string,
	defaultMappedModel string,
) (result *OpenAIForwardResult, err error) {
	if r == nil || r.openaiGatewayService == nil {
		return nil, errors.New("compatible gateway text runtime is not configured")
	}
	s := r.openaiGatewayService
	startTime := time.Now()
	var (
		originalModel            string
		upstreamModelForFeedback string
		statusCode               int
	)
	defer func() {
		r.persistRuntimeFeedback(
			ctx,
			account,
			originalModel,
			upstreamModelForFeedback,
			CompatibleGatewayProtocolFamilyChatCompletions,
			statusCode,
			result,
			err,
		)
	}()

	var chatReq apicompat.ChatCompletionsRequest
	if err := json.Unmarshal(body, &chatReq); err != nil {
		return nil, fmt.Errorf("parse chat completions request: %w", err)
	}
	originalModel = strings.TrimSpace(chatReq.Model)
	clientStream := chatReq.Stream
	includeUsage := chatReq.StreamOptions != nil && chatReq.StreamOptions.IncludeUsage

	mappedModel := resolveOpenAIForwardModel(account, originalModel, defaultMappedModel)
	mappedModel = normalizeOpenAIModelForUpstream(account, mappedModel)
	upstreamModelForFeedback = strings.TrimSpace(mappedModel)

	if account != nil && account.SupportsOpenAIPassthroughHTTP() {
		forwardBody := body
		if mappedModel != "" && mappedModel != originalModel {
			var patchErr error
			forwardBody, patchErr = sjson.SetBytes(body, "model", mappedModel)
			if patchErr != nil {
				return nil, fmt.Errorf("patch chat completions passthrough model: %w", patchErr)
			}
		}
		var reasoningEffort *string
		if value := normalizeOpenAIReasoningEffort(chatReq.ReasoningEffort); value != "" {
			reasoningEffort = &value
		}
		result, err = s.forwardOpenAIPassthrough(ctx, c, account, forwardBody, originalModel, reasoningEffort, clientStream, startTime)
		if err != nil {
			return nil, err
		}
		if result != nil && mappedModel != "" && mappedModel != originalModel {
			result.UpstreamModel = mappedModel
		}
		return result, nil
	}

	promptCacheKey = strings.TrimSpace(promptCacheKey)
	compatPromptCacheInjected := false
	if promptCacheKey == "" && account.Type == AccountTypeOAuth && shouldAutoInjectPromptCacheKeyForCompat(mappedModel) {
		promptCacheKey = deriveCompatPromptCacheKey(&chatReq, mappedModel)
		compatPromptCacheInjected = promptCacheKey != ""
	}

	isResponsesShape := !gjson.GetBytes(body, "messages").Exists() && gjson.GetBytes(body, "input").Exists()

	var (
		responsesReq  *apicompat.ResponsesRequest
		responsesBody []byte
	)
	if isResponsesShape {
		responsesBody, err = sjson.SetBytes(body, "model", mappedModel)
		if err != nil {
			return nil, fmt.Errorf("rewrite model in responses-shaped body: %w", err)
		}
		for _, field := range cursorResponsesUnsupportedFields {
			if stripped, deleteErr := sjson.DeleteBytes(responsesBody, field); deleteErr == nil {
				responsesBody = stripped
			}
		}
		responsesReq = &apicompat.ResponsesRequest{
			Model:       mappedModel,
			ServiceTier: gjson.GetBytes(responsesBody, "service_tier").String(),
		}
		if effort := gjson.GetBytes(responsesBody, "reasoning.effort").String(); effort != "" {
			responsesReq.Reasoning = &apicompat.ResponsesReasoning{Effort: effort}
		}
	} else {
		responsesReq, err = apicompat.ChatCompletionsToResponses(&chatReq)
		if err != nil {
			return nil, fmt.Errorf("convert chat completions to responses: %w", err)
		}
		responsesReq.Model = mappedModel
		responsesBody, err = json.Marshal(responsesReq)
		if err != nil {
			return nil, fmt.Errorf("marshal responses request: %w", err)
		}
	}

	logFields := []zap.Field{
		zap.Int64("account_id", account.ID),
		zap.String("original_model", originalModel),
		zap.String("mapped_model", mappedModel),
		zap.Bool("stream", clientStream),
		zap.Bool("responses_shape", isResponsesShape),
	}
	if compatPromptCacheInjected {
		logFields = append(logFields,
			zap.Bool("compat_prompt_cache_key_injected", true),
			zap.String("compat_prompt_cache_key_sha256", hashSensitiveValueForLog(promptCacheKey)),
		)
	}
	logger.L().Debug("openai chat_completions: model mapping applied", logFields...)

	if account.Type == AccountTypeOAuth {
		var reqBody map[string]any
		if err := json.Unmarshal(responsesBody, &reqBody); err != nil {
			return nil, fmt.Errorf("unmarshal for codex transform: %w", err)
		}
		codexResult := applyCodexOAuthTransform(reqBody, false, false)
		if codexResult.PromptCacheKey != "" {
			promptCacheKey = codexResult.PromptCacheKey
		} else if promptCacheKey != "" {
			reqBody["prompt_cache_key"] = promptCacheKey
		}
		if rewriteOpenAICodexBodyIdentityMap(account.ID, reqBody) {
			if value, ok := reqBody["prompt_cache_key"].(string); ok {
				promptCacheKey = strings.TrimSpace(value)
			}
		}
		responsesBody, err = json.Marshal(reqBody)
		if err != nil {
			return nil, fmt.Errorf("remarshal after codex transform: %w", err)
		}
	}

	token, _, err := s.GetAccessToken(ctx, account)
	if err != nil {
		return nil, fmt.Errorf("get access token: %w", err)
	}

	upstreamReq, err := s.buildUpstreamRequest(ctx, c, account, responsesBody, token, true, promptCacheKey)
	if err != nil {
		return nil, fmt.Errorf("build upstream request: %w", err)
	}

	setOpenAICompatPromptCacheSessionID(c, account, upstreamReq, promptCacheKey)

	proxyURL := ""
	if account.Proxy != nil {
		proxyURL = account.Proxy.URL()
	}
	resp, err := s.httpUpstream.Do(upstreamReq, proxyURL, account.ID, account.Concurrency)
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
		writeChatCompletionsError(c, passthroughRuleResult{
			StatusCode: http.StatusBadGateway,
			ErrType:    "upstream_error",
			ErrMessage: "Upstream request failed",
		})
		return nil, fmt.Errorf("upstream request failed: %s", safeErr)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		statusCode = resp.StatusCode
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
		_ = resp.Body.Close()
		resp.Body = io.NopCloser(bytes.NewReader(respBody))

		upstreamMsg := strings.TrimSpace(extractUpstreamErrorMessage(respBody))
		upstreamMsg = sanitizeUpstreamErrorMessage(upstreamMsg)
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
			if s.rateLimitService != nil {
				s.rateLimitService.HandleUpstreamError(ctx, account, resp.StatusCode, resp.Header, respBody)
			}
			return nil, &UpstreamFailoverError{
				StatusCode:             resp.StatusCode,
				ResponseBody:           respBody,
				FailureReason:          classifyOpenAIHTTPFailoverReason(resp.StatusCode),
				RetryableOnSameAccount: account.IsPoolMode() && (isOpenAIPoolModeRetryableStatus(resp.StatusCode) || isOpenAITransientProcessingError(resp.StatusCode, upstreamMsg, respBody)),
			}
		}
		return s.handleChatCompletionsErrorResponse(resp, c, account)
	}

	var handleErr error
	if clientStream {
		result, handleErr = s.handleChatStreamingResponse(resp, c, originalModel, mappedModel, includeUsage, startTime)
	} else {
		result, handleErr = s.handleChatBufferedStreamingResponse(resp, c, originalModel, mappedModel, startTime)
	}

	if handleErr == nil && result != nil {
		if responsesReq.ServiceTier != "" {
			st := responsesReq.ServiceTier
			result.ServiceTier = &st
		}
		if responsesReq.Reasoning != nil && responsesReq.Reasoning.Effort != "" {
			re := responsesReq.Reasoning.Effort
			result.ReasoningEffort = &re
		}
	}

	if handleErr == nil && account.Type == AccountTypeOAuth {
		if snapshot := ParseCodexRateLimitHeaders(resp.Header); snapshot != nil {
			s.updateCodexUsageSnapshot(ctx, account.ID, snapshot)
		}
	}

	return result, handleErr
}

func (r *CompatibleGatewayTextRuntime) ForwardMessages(
	ctx context.Context,
	c *gin.Context,
	account *Account,
	body []byte,
	promptCacheKey string,
	defaultMappedModel string,
) (result *OpenAIForwardResult, err error) {
	if r == nil || r.openaiGatewayService == nil {
		return nil, errors.New("compatible gateway text runtime is not configured")
	}
	s := r.openaiGatewayService
	startTime := time.Now()
	var (
		originalModel            string
		upstreamModelForFeedback string
		statusCode               int
	)
	defer func() {
		r.persistRuntimeFeedback(
			ctx,
			account,
			originalModel,
			upstreamModelForFeedback,
			CompatibleGatewayProtocolFamilyMessages,
			statusCode,
			result,
			err,
		)
	}()

	var anthropicReq apicompat.AnthropicRequest
	if err := json.Unmarshal(body, &anthropicReq); err != nil {
		return nil, fmt.Errorf("parse anthropic request: %w", err)
	}
	originalModel = strings.TrimSpace(anthropicReq.Model)
	applyOpenAICompatModelNormalization(&anthropicReq)
	normalizedModel := anthropicReq.Model
	clientStream := anthropicReq.Stream

	responsesReq, err := apicompat.AnthropicToResponses(&anthropicReq)
	if err != nil {
		return nil, fmt.Errorf("convert anthropic to responses: %w", err)
	}

	responsesReq.Stream = true
	isStream := true

	if containsBetaToken(c.GetHeader("anthropic-beta"), claude.FastModeBetaToken()) {
		responsesReq.ServiceTier = "priority"
	}

	billingModel := resolveOpenAIForwardModel(account, normalizedModel, defaultMappedModel)
	upstreamModel := normalizeOpenAIModelForUpstream(account, billingModel)
	upstreamModelForFeedback = strings.TrimSpace(upstreamModel)
	responsesReq.Model = upstreamModel

	logger.L().Debug("openai messages: model mapping applied",
		zap.Int64("account_id", account.ID),
		zap.String("original_model", originalModel),
		zap.String("normalized_model", normalizedModel),
		zap.String("billing_model", billingModel),
		zap.String("upstream_model", upstreamModel),
		zap.Bool("stream", isStream),
	)

	responsesBody, err := json.Marshal(responsesReq)
	if err != nil {
		return nil, fmt.Errorf("marshal responses request: %w", err)
	}
	responsesBody, err = ensureOpenAIResponsesPromptCacheKey(responsesBody, promptCacheKey)
	if err != nil {
		return nil, err
	}

	if account.Type == AccountTypeOAuth {
		var reqBody map[string]any
		if err := json.Unmarshal(responsesBody, &reqBody); err != nil {
			return nil, fmt.Errorf("unmarshal for codex transform: %w", err)
		}
		codexResult := applyCodexOAuthTransform(reqBody, false, false)
		if codexResult.PromptCacheKey != "" {
			promptCacheKey = codexResult.PromptCacheKey
		}
		if promptCacheKey != "" {
			reqBody["prompt_cache_key"] = promptCacheKey
		}
		if rewriteOpenAICodexBodyIdentityMap(account.ID, reqBody) {
			if value, ok := reqBody["prompt_cache_key"].(string); ok {
				promptCacheKey = strings.TrimSpace(value)
			}
		}
		if s.cfg != nil {
			if forcedTemplateText := strings.TrimSpace(s.cfg.Gateway.ForcedCodexInstructionsTemplate); forcedTemplateText != "" {
				existingInstructions, _ := reqBody["instructions"].(string)
				if _, err := applyForcedCodexInstructionsTemplate(reqBody, forcedTemplateText, forcedCodexInstructionsTemplateData{
					ExistingInstructions: existingInstructions,
					OriginalModel:        originalModel,
					NormalizedModel:      normalizedModel,
					BillingModel:         billingModel,
					UpstreamModel:        upstreamModel,
				}); err != nil {
					return nil, err
				}
			}
		}
		isStream = true
		responsesBody, err = json.Marshal(reqBody)
		if err != nil {
			return nil, fmt.Errorf("remarshal after codex transform: %w", err)
		}
	}

	token, _, err := s.GetAccessToken(ctx, account)
	if err != nil {
		return nil, fmt.Errorf("get access token: %w", err)
	}

	upstreamReq, err := s.buildUpstreamRequest(ctx, c, account, responsesBody, token, isStream, promptCacheKey)
	if err != nil {
		return nil, fmt.Errorf("build upstream request: %w", err)
	}

	setOpenAICompatPromptCacheSessionID(c, account, upstreamReq, promptCacheKey)

	proxyURL := ""
	if account.Proxy != nil {
		proxyURL = account.Proxy.URL()
	}
	resp, err := s.httpUpstream.Do(upstreamReq, proxyURL, account.ID, account.Concurrency)
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
		writeAnthropicError(c, passthroughRuleResult{
			StatusCode: http.StatusBadGateway,
			ErrType:    "api_error",
			ErrMessage: "Upstream request failed",
		})
		return nil, fmt.Errorf("upstream request failed: %s", safeErr)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		statusCode = resp.StatusCode
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
		_ = resp.Body.Close()
		resp.Body = io.NopCloser(bytes.NewReader(respBody))

		upstreamMsg := strings.TrimSpace(extractUpstreamErrorMessage(respBody))
		upstreamMsg = sanitizeUpstreamErrorMessage(upstreamMsg)
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
			if s.rateLimitService != nil {
				s.rateLimitService.HandleUpstreamError(ctx, account, resp.StatusCode, resp.Header, respBody)
			}
			return nil, &UpstreamFailoverError{
				StatusCode:             resp.StatusCode,
				ResponseBody:           respBody,
				FailureReason:          classifyOpenAIHTTPFailoverReason(resp.StatusCode),
				RetryableOnSameAccount: account.IsPoolMode() && (isOpenAIPoolModeRetryableStatus(resp.StatusCode) || isOpenAITransientProcessingError(resp.StatusCode, upstreamMsg, respBody)),
			}
		}
		return s.handleAnthropicErrorResponse(resp, c, account)
	}

	var handleErr error
	if clientStream {
		result, handleErr = s.handleAnthropicStreamingResponse(resp, c, originalModel, billingModel, upstreamModel, startTime)
	} else {
		result, handleErr = s.handleAnthropicBufferedStreamingResponse(resp, c, originalModel, billingModel, upstreamModel, startTime)
	}

	if handleErr == nil && result != nil {
		if responsesReq.ServiceTier != "" {
			st := responsesReq.ServiceTier
			result.ServiceTier = &st
		}
		if responsesReq.Reasoning != nil && responsesReq.Reasoning.Effort != "" {
			re := responsesReq.Reasoning.Effort
			result.ReasoningEffort = &re
		}
	}

	if handleErr == nil && account.Type == AccountTypeOAuth {
		if snapshot := ParseCodexRateLimitHeaders(resp.Header); snapshot != nil {
			s.updateCodexUsageSnapshot(ctx, account.ID, snapshot)
		}
	}

	return result, handleErr
}

func (r *CompatibleGatewayTextRuntime) persistRuntimeFeedback(
	ctx context.Context,
	account *Account,
	requestedModel string,
	upstreamModel string,
	protocolFamily CompatibleGatewayProtocolFamily,
	statusCode int,
	result *OpenAIForwardResult,
	runtimeErr error,
) {
	if r == nil || r.openaiGatewayService == nil || account == nil {
		return
	}
	r.openaiGatewayService.PersistCompatibleGatewayRuntimeFeedback(ctx, CompatibleGatewayRuntimeFeedbackInput{
		Account:        account,
		RequestedModel: strings.TrimSpace(requestedModel),
		UpstreamModel:  strings.TrimSpace(upstreamModel),
		Result:         result,
		StatusCode:     statusCode,
		ProtocolFamily: protocolFamily,
		Err:            runtimeErr,
	})
}
