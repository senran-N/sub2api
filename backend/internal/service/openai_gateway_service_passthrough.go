package service

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/senran-N/sub2api/internal/pkg/logger"
	"github.com/senran-N/sub2api/internal/util/responseheaders"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"
)

func (s *OpenAIGatewayService) forwardOpenAIPassthrough(
	ctx context.Context,
	c *gin.Context,
	account *Account,
	body []byte,
	reqModel string,
	reasoningEffort *string,
	reqStream bool,
	startTime time.Time,
) (*OpenAIForwardResult, error) {
	if account != nil && account.Type == AccountTypeOAuth {
		forceCodexCLI := s != nil && s.cfg != nil && s.cfg.Gateway.ForceCodexCLI
		profile := GetCodexRequestProfile(c, body, forceCodexCLI)
		policy := NewCodexNativeMutationPolicy(profile)
		normalizedBody, normalized, err := normalizeOpenAIPassthroughOAuthBody(body, account, policy)
		if err != nil {
			return nil, err
		}
		if normalized {
			body = normalizedBody
		}
		reqStream = gjson.GetBytes(body, "stream").Bool()
	}

	sanitizedBody, sanitized, err := sanitizeEmptyBase64InputImagesInOpenAIBody(body)
	if err != nil {
		return nil, err
	}
	if sanitized {
		body = sanitizedBody
	}

	logger.LegacyPrintf("service.openai_gateway",
		"[OpenAI 自动透传] 命中自动透传分支: account=%d name=%s type=%s model=%s stream=%v",
		account.ID,
		account.Name,
		account.Type,
		reqModel,
		reqStream,
	)
	if reqStream && c != nil && c.Request != nil {
		if timeoutHeaders := collectOpenAIPassthroughTimeoutHeaders(c.Request.Header); len(timeoutHeaders) > 0 {
			streamWarnLogger := logger.FromContext(ctx).With(
				zap.String("component", "service.openai_gateway"),
				zap.Int64("account_id", account.ID),
				zap.Strings("timeout_headers", timeoutHeaders),
			)
			if s.isOpenAIPassthroughTimeoutHeadersAllowed() {
				streamWarnLogger.Warn("OpenAI passthrough 透传请求包含超时相关请求头，且当前配置为放行，可能导致上游提前断流")
			} else {
				streamWarnLogger.Warn("OpenAI passthrough 检测到超时相关请求头，将按配置过滤以降低断流风险")
			}
		}
	}

	token, _, err := s.GetAccessToken(ctx, account)
	if err != nil {
		return nil, err
	}

	upstreamCtx, releaseUpstreamCtx := detachStreamUpstreamContext(ctx, reqStream)
	upstreamReq, err := s.buildUpstreamRequestOpenAIPassthrough(upstreamCtx, c, account, body, token)
	releaseUpstreamCtx()
	if err != nil {
		return nil, err
	}

	proxyURL := ""
	if account.ProxyID != nil && account.Proxy != nil {
		proxyURL = account.Proxy.URL()
	}

	setOpsUpstreamRequestBody(c, body)
	if c != nil {
		c.Set("openai_passthrough", true)
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
			Passthrough:        true,
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
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		return nil, s.handleErrorResponsePassthrough(ctx, resp, c, account, body)
	}

	var usage *OpenAIUsage
	var firstTokenMs *int
	if reqStream {
		result, err := s.handleStreamingResponsePassthrough(ctx, resp, c, account, startTime)
		if err != nil {
			return nil, err
		}
		usage = result.usage
		firstTokenMs = result.firstTokenMs
	} else {
		usage, err = s.handleNonStreamingResponsePassthrough(ctx, resp, c)
		if err != nil {
			return nil, err
		}
	}

	if snapshot := ParseCodexRateLimitHeaders(resp.Header); snapshot != nil {
		s.updateCodexUsageSnapshot(ctx, account.ID, snapshot)
	}

	if usage == nil {
		usage = &OpenAIUsage{}
	}
	s.bindCodexSessionTransportFromBody(c, body, OpenAIUpstreamTransportHTTPSSE)

	return &OpenAIForwardResult{
		RequestID:       resp.Header.Get("x-request-id"),
		Usage:           *usage,
		Model:           reqModel,
		ServiceTier:     extractOpenAIServiceTierFromBody(body),
		ReasoningEffort: reasoningEffort,
		Stream:          reqStream,
		OpenAIWSMode:    false,
		Duration:        time.Since(startTime),
		FirstTokenMs:    firstTokenMs,
	}, nil
}

func (s *OpenAIGatewayService) buildUpstreamRequestOpenAIPassthrough(
	ctx context.Context,
	c *gin.Context,
	account *Account,
	body []byte,
	token string,
) (*http.Request, error) {
	forceCodexCLI := s != nil && s.cfg != nil && s.cfg.Gateway.ForceCodexCLI
	profile := GetCodexRequestProfile(c, body, forceCodexCLI)
	policy := NewCodexNativeMutationPolicy(profile)

	targetURL := openaiPlatformAPIURL
	upstreamTarget := newCompatibleResponsesUpstreamTarget(openaiPlatformAPIURL)
	switch account.Type {
	case AccountTypeOAuth:
		targetURL = chatgptCodexURL
	case AccountTypeAPIKey, AccountTypeUpstream:
		baseURL := account.GetOpenAIBaseURL()
		if baseURL == "" {
			baseURL = openaiPlatformAPIURL
		}
		validatedURL, err := s.validateUpstreamBaseURL(baseURL)
		if err != nil {
			return nil, err
		}
		rawRequestPath := ""
		if c != nil && c.Request != nil && c.Request.URL != nil {
			rawRequestPath = c.Request.URL.Path
		}
		upstreamTarget = newCompatiblePassthroughUpstreamTargetWithOptions(
			validatedURL,
			rawRequestPath,
			account.GetCompatibleAuthMode(""),
			account.GetCompatibleEndpointOverride("responses"),
			account.GetCompatibleEndpointOverride("chat_completions"),
		)
		targetURL = upstreamTarget.URL
		if c != nil && c.Request != nil && c.Request.URL != nil && c.Request.URL.RawQuery != "" {
			if parsedTarget, err := url.Parse(targetURL); err == nil {
				parsedTarget.RawQuery = c.Request.URL.RawQuery
				targetURL = parsedTarget.String()
			}
		}
	}
	if account.Type != AccountTypeAPIKey && account.Type != AccountTypeUpstream {
		targetURL = appendOpenAIResponsesRequestPathSuffix(targetURL, openAIResponsesRequestPathSuffix(c))
	}

	method := http.MethodPost
	if c != nil && c.Request != nil && strings.TrimSpace(c.Request.Method) != "" {
		method = c.Request.Method
	}
	req, err := http.NewRequestWithContext(ctx, method, targetURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	allowTimeoutHeaders := s.isOpenAIPassthroughTimeoutHeadersAllowed()
	if c != nil && c.Request != nil {
		for key, values := range c.Request.Header {
			lower := strings.ToLower(strings.TrimSpace(key))
			if !isOpenAIPassthroughAllowedRequestHeader(lower, allowTimeoutHeaders) {
				continue
			}
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}
	}

	req.Header.Del("authorization")
	req.Header.Del("api-key")
	req.Header.Del("x-api-key")
	req.Header.Del("x-goog-api-key")
	upstreamTarget.ApplyAuthHeader(req.Header, token)

	if account.Type == AccountTypeOAuth {
		promptCacheKey := strings.TrimSpace(gjson.GetBytes(body, "prompt_cache_key").String())
		req.Host = "chatgpt.com"
		req.Header.Del("accept-language")
		if chatgptAccountID := account.GetChatGPTAccountID(); chatgptAccountID != "" {
			req.Header.Set("chatgpt-account-id", chatgptAccountID)
		}
		sessionResolution := policy.ResolveOAuthSessionHeaders(promptCacheKey, resolveOpenAICompactSessionID(c), !profile.NativeClient)
		if profile.CompactPath {
			req.Header.Set("accept", policy.ResolveAccept("application/json"))
			req.Header.Set("version", resolveOpenAICodexUpstreamVersion(account))
		} else {
			req.Header.Set("accept", policy.ResolveAccept("text/event-stream"))
		}
		req.Header.Set("OpenAI-Beta", openAICodexHTTPBetaValue)
		req.Header.Set("originator", resolveOpenAICodexUpstreamOriginator(account))
		isolatedSessionID := ""
		if sessionResolution.SessionID != "" {
			isolatedSessionID = resolveOpenAICodexUpstreamSessionHeaderValue(account.ID, sessionResolution.SessionID, sessionResolution.SessionSource)
			req.Header.Set("session_id", isolatedSessionID)
		}
		if sessionResolution.ConversationID != "" {
			req.Header.Set("conversation_id", resolveOpenAICodexUpstreamSessionHeaderValue(account.ID, sessionResolution.ConversationID, sessionResolution.ConversationSource))
		}
		if requestID := resolveOpenAICodexUpstreamClientRequestID(account.ID, profile.Headers.ClientRequestID, isolatedSessionID); requestID != "" {
			req.Header.Set("x-client-request-id", requestID)
		}
		if betaFeatures := resolveOpenAICodexUpstreamBetaFeatures(account); betaFeatures != "" {
			req.Header.Set("x-codex-beta-features", betaFeatures)
		} else {
			req.Header.Del("x-codex-beta-features")
		}
		if turnMetadata := resolveOpenAICodexUpstreamTurnMetadata(account.ID, profile.Headers.TurnMetadata); turnMetadata != "" {
			req.Header.Set(openAIWSTurnMetadataHeader, turnMetadata)
		} else {
			req.Header.Del(openAIWSTurnMetadataHeader)
		}
		if acceptLanguage := resolveOpenAICodexUpstreamAcceptLanguage(account); acceptLanguage != "" {
			req.Header.Set("accept-language", acceptLanguage)
		}
		if windowID := resolveOpenAICodexUpstreamWindowID(account.ID); windowID != "" {
			req.Header.Set(openAICodexMetadataWindowIDKey, windowID)
		}
		if subagent := resolveOpenAICodexUpstreamSubagent(profile, body); subagent != "" {
			req.Header.Set(openAICodexMetadataSubagentKey, subagent)
		} else {
			req.Header.Del(openAICodexMetadataSubagentKey)
		}
		if parentThreadID := resolveOpenAICodexUpstreamParentThreadID(account.ID, profile, body); parentThreadID != "" {
			req.Header.Set(openAICodexMetadataParentThreadIDKey, parentThreadID)
		} else {
			req.Header.Del(openAICodexMetadataParentThreadIDKey)
		}
		if profile.CompactPath {
			if installationID := resolveOpenAICodexUpstreamInstallationID(account.ID); installationID != "" {
				req.Header.Set(openAICodexMetadataInstallationIDKey, installationID)
			}
		} else {
			req.Header.Del(openAICodexMetadataInstallationIDKey)
		}
	}

	userAgent := ""
	if account != nil && account.Type == AccountTypeOAuth {
		userAgent = resolveOpenAICodexUpstreamUserAgent(account)
	} else {
		userAgent = policy.ResolveUserAgent(account, forceCodexCLI, true)
	}
	if userAgent != "" {
		req.Header.Set("user-agent", userAgent)
	}

	if req.Header.Get("content-type") == "" {
		req.Header.Set("content-type", "application/json")
	}

	return req, nil
}

func (s *OpenAIGatewayService) handleErrorResponsePassthrough(
	ctx context.Context,
	resp *http.Response,
	c *gin.Context,
	account *Account,
	requestBody []byte,
) error {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))

	upstreamMsg := strings.TrimSpace(extractUpstreamErrorMessage(body))
	upstreamMsg = sanitizeUpstreamErrorMessage(upstreamMsg)
	upstreamDetail := ""
	if s.cfg != nil && s.cfg.Gateway.LogUpstreamErrorBody {
		maxBytes := s.cfg.Gateway.LogUpstreamErrorBodyMaxBytes
		if maxBytes <= 0 {
			maxBytes = 2048
		}
		upstreamDetail = truncateString(string(body), maxBytes)
	}
	setOpsUpstreamError(c, resp.StatusCode, upstreamMsg, upstreamDetail)
	logOpenAIInstructionsRequiredDebug(ctx, c, account, resp.StatusCode, upstreamMsg, requestBody, body)
	shouldDisable := false
	if s.rateLimitService != nil {
		shouldDisable = s.rateLimitService.HandleUpstreamError(ctx, account, resp.StatusCode, resp.Header, body)
	}
	kind := "http_error"
	if shouldDisable {
		kind = "failover"
	}
	appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
		Platform:             account.Platform,
		AccountID:            account.ID,
		AccountName:          account.Name,
		UpstreamStatusCode:   resp.StatusCode,
		UpstreamRequestID:    resp.Header.Get("x-request-id"),
		Passthrough:          true,
		Kind:                 kind,
		Message:              upstreamMsg,
		Detail:               upstreamDetail,
		UpstreamResponseBody: upstreamDetail,
	})
	if shouldDisable {
		return &UpstreamFailoverError{
			StatusCode:      resp.StatusCode,
			ResponseBody:    body,
			ResponseHeaders: resp.Header.Clone(),
		}
	}

	writeOpenAIPassthroughResponseHeaders(c.Writer.Header(), resp.Header, s.responseHeaderFilter)
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/json"
	}
	c.Data(resp.StatusCode, contentType, body)

	if upstreamMsg == "" {
		return fmt.Errorf("upstream error: %d", resp.StatusCode)
	}
	return fmt.Errorf("upstream error: %d message=%s", resp.StatusCode, upstreamMsg)
}

func isOpenAIPassthroughAllowedRequestHeader(lowerKey string, allowTimeoutHeaders bool) bool {
	if lowerKey == "" {
		return false
	}
	if isOpenAIPassthroughTimeoutHeader(lowerKey) {
		return allowTimeoutHeaders
	}
	return openaiPassthroughAllowedHeaders[lowerKey]
}

func isOpenAIPassthroughTimeoutHeader(lowerKey string) bool {
	switch lowerKey {
	case "x-stainless-timeout", "x-stainless-read-timeout", "x-stainless-connect-timeout", "x-request-timeout", "request-timeout", "grpc-timeout":
		return true
	default:
		return false
	}
}

func (s *OpenAIGatewayService) isOpenAIPassthroughTimeoutHeadersAllowed() bool {
	return s != nil && s.cfg != nil && s.cfg.Gateway.OpenAIPassthroughAllowTimeoutHeaders
}

func collectOpenAIPassthroughTimeoutHeaders(headers http.Header) []string {
	if headers == nil {
		return nil
	}
	var matched []string
	for key, values := range headers {
		lowerKey := strings.ToLower(strings.TrimSpace(key))
		if isOpenAIPassthroughTimeoutHeader(lowerKey) {
			entry := lowerKey
			if len(values) > 0 {
				entry = fmt.Sprintf("%s=%s", lowerKey, strings.Join(values, "|"))
			}
			matched = append(matched, entry)
		}
	}
	sort.Strings(matched)
	return matched
}

type openaiStreamingResultPassthrough struct {
	usage        *OpenAIUsage
	firstTokenMs *int
}

func (s *OpenAIGatewayService) handleStreamingResponsePassthrough(
	ctx context.Context,
	resp *http.Response,
	c *gin.Context,
	account *Account,
	startTime time.Time,
) (*openaiStreamingResultPassthrough, error) {
	writeOpenAIPassthroughResponseHeaders(c.Writer.Header(), resp.Header, s.responseHeaderFilter)

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	if requestID := resp.Header.Get("x-request-id"); requestID != "" {
		c.Header("x-request-id", requestID)
	}

	writer := c.Writer
	flusher, ok := writer.(http.Flusher)
	if !ok {
		return nil, errors.New("streaming not supported")
	}

	usage := &OpenAIUsage{}
	var firstTokenMs *int
	clientDisconnected := false
	sawDone := false
	sawTerminalEvent := false
	sawFailedEvent := false
	failedMessage := ""
	clientOutputStarted := false
	upstreamRequestID := strings.TrimSpace(resp.Header.Get("x-request-id"))

	scanner := bufio.NewScanner(resp.Body)
	maxLineSize := defaultMaxLineSize
	if s.cfg != nil && s.cfg.Gateway.MaxLineSize > 0 {
		maxLineSize = s.cfg.Gateway.MaxLineSize
	}
	scanBuf := getSSEScannerBuf64K()
	scanner.Buffer(scanBuf[:0], maxLineSize)
	defer putSSEScannerBuf64K(scanBuf)

	pendingLines := make([]string, 0, 4)
	writePendingLines := func() bool {
		if len(pendingLines) == 0 {
			return true
		}
		for _, pendingLine := range pendingLines {
			if _, err := fmt.Fprintln(writer, pendingLine); err != nil {
				clientDisconnected = true
				logger.LegacyPrintf("service.openai_gateway", "[OpenAI passthrough] Client disconnected during streaming, continue draining upstream for usage: account=%d", account.ID)
				return false
			}
		}
		flusher.Flush()
		clientOutputStarted = true
		pendingLines = pendingLines[:0]
		return true
	}

	for scanner.Scan() {
		line := scanner.Text()
		lineStartsClientOutput := false
		forceFlushFailedEvent := false
		if data, ok := extractOpenAISSEDataLine(line); ok {
			dataBytes := []byte(data)
			trimmedData := strings.TrimSpace(data)
			eventType := strings.TrimSpace(gjson.GetBytes(dataBytes, "type").String())
			if eventType == "response.failed" {
				failedMessage = extractOpenAISSEErrorMessage(dataBytes)
				if !openAIStreamClientOutputStarted(c, clientOutputStarted) && openAIStreamFailedEventShouldFailover(dataBytes, failedMessage) {
					return &openaiStreamingResultPassthrough{usage: usage, firstTokenMs: firstTokenMs},
						s.newOpenAIStreamFailoverError(c, account, true, upstreamRequestID, dataBytes, failedMessage)
				}
				forceFlushFailedEvent = true
				sawFailedEvent = true
			}
			if trimmedData == "[DONE]" {
				sawDone = true
			}
			if openAIStreamEventIsTerminal(trimmedData) {
				sawTerminalEvent = true
			}
			lineStartsClientOutput = forceFlushFailedEvent || openAIStreamDataStartsClientOutput(trimmedData, eventType)
			if firstTokenMs == nil && lineStartsClientOutput && trimmedData != "[DONE]" {
				ms := int(time.Since(startTime).Milliseconds())
				firstTokenMs = &ms
			}
			s.parseSSEUsageBytes(dataBytes, usage)
		}

		if clientDisconnected {
			continue
		}
		if !clientOutputStarted && !lineStartsClientOutput {
			pendingLines = append(pendingLines, line)
			continue
		}
		if !clientOutputStarted && len(pendingLines) > 0 {
			if !writePendingLines() {
				continue
			}
		}
		if _, err := fmt.Fprintln(writer, line); err != nil {
			clientDisconnected = true
			logger.LegacyPrintf("service.openai_gateway", "[OpenAI passthrough] Client disconnected during streaming, continue draining upstream for usage: account=%d", account.ID)
			continue
		}
		clientOutputStarted = true
		flusher.Flush()
	}
	if err := scanner.Err(); err != nil {
		if sawTerminalEvent && !sawFailedEvent {
			return &openaiStreamingResultPassthrough{usage: usage, firstTokenMs: firstTokenMs}, nil
		}
		if sawFailedEvent {
			return &openaiStreamingResultPassthrough{usage: usage, firstTokenMs: firstTokenMs}, fmt.Errorf("upstream response failed: %s", failedMessage)
		}
		if clientDisconnected {
			return &openaiStreamingResultPassthrough{usage: usage, firstTokenMs: firstTokenMs}, fmt.Errorf("stream usage incomplete after disconnect: %w", err)
		}
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return &openaiStreamingResultPassthrough{usage: usage, firstTokenMs: firstTokenMs}, fmt.Errorf("stream usage incomplete: %w", err)
		}
		if errors.Is(err, bufio.ErrTooLong) {
			logger.LegacyPrintf("service.openai_gateway", "[OpenAI passthrough] SSE line too long: account=%d max_size=%d error=%v", account.ID, maxLineSize, err)
			return &openaiStreamingResultPassthrough{usage: usage, firstTokenMs: firstTokenMs}, err
		}
		if !openAIStreamClientOutputStarted(c, clientOutputStarted) {
			msg := "OpenAI stream disconnected before completion"
			if errText := strings.TrimSpace(err.Error()); errText != "" {
				msg += ": " + errText
			}
			return &openaiStreamingResultPassthrough{usage: usage, firstTokenMs: firstTokenMs},
				s.newOpenAIStreamFailoverError(c, account, true, upstreamRequestID, nil, msg)
		}
		logger.LegacyPrintf(
			"service.openai_gateway",
			"[OpenAI passthrough] 流读取异常中断: account=%d request_id=%s err=%v",
			account.ID,
			upstreamRequestID,
			err,
		)
		return &openaiStreamingResultPassthrough{usage: usage, firstTokenMs: firstTokenMs}, fmt.Errorf("stream read error: %w", err)
	}
	if sawFailedEvent {
		return &openaiStreamingResultPassthrough{usage: usage, firstTokenMs: firstTokenMs}, fmt.Errorf("upstream response failed: %s", failedMessage)
	}
	if !clientDisconnected && !sawDone && !sawTerminalEvent && ctx.Err() == nil {
		logger.FromContext(ctx).With(
			zap.String("component", "service.openai_gateway"),
			zap.Int64("account_id", account.ID),
			zap.String("upstream_request_id", upstreamRequestID),
		).Info("OpenAI passthrough 上游流在未收到 [DONE] 时结束，疑似断流")
		if !openAIStreamClientOutputStarted(c, clientOutputStarted) {
			return &openaiStreamingResultPassthrough{usage: usage, firstTokenMs: firstTokenMs},
				s.newOpenAIStreamFailoverError(c, account, true, upstreamRequestID, nil, "OpenAI stream ended before a terminal event")
		}
		return &openaiStreamingResultPassthrough{usage: usage, firstTokenMs: firstTokenMs}, errors.New("stream usage incomplete: missing terminal event")
	}

	return &openaiStreamingResultPassthrough{usage: usage, firstTokenMs: firstTokenMs}, nil
}

func (s *OpenAIGatewayService) handleNonStreamingResponsePassthrough(
	ctx context.Context,
	resp *http.Response,
	c *gin.Context,
) (*OpenAIUsage, error) {
	maxBytes := resolveUpstreamResponseReadLimit(s.cfg)
	body, err := readUpstreamResponseBodyLimited(resp.Body, maxBytes)
	if err != nil {
		if errors.Is(err, ErrUpstreamResponseBodyTooLarge) {
			setOpsUpstreamError(c, http.StatusBadGateway, "upstream response too large", "")
			c.JSON(http.StatusBadGateway, gin.H{
				"error": gin.H{
					"type":    "upstream_error",
					"message": "Upstream response too large",
				},
			})
		}
		return nil, err
	}

	if isEventStreamResponse(resp.Header) {
		return s.handlePassthroughSSEToJSON(resp, c, body)
	}

	usage := &OpenAIUsage{}
	usageParsed := false
	if len(body) > 0 {
		if parsedUsage, ok := extractOpenAIUsageFromJSONBytes(body); ok {
			*usage = parsedUsage
			usageParsed = true
		}
	}
	if !usageParsed {
		usage = s.parseSSEUsageFromBody(string(body))
	}

	writeOpenAIPassthroughResponseHeaders(c.Writer.Header(), resp.Header, s.responseHeaderFilter)

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/json"
	}
	c.Data(resp.StatusCode, contentType, body)
	return usage, nil
}

func (s *OpenAIGatewayService) handlePassthroughSSEToJSON(
	resp *http.Response,
	c *gin.Context,
	body []byte,
) (*OpenAIUsage, error) {
	bodyText := string(body)
	finalResponse, ok := extractCodexFinalResponse(bodyText)

	usage := &OpenAIUsage{}
	if ok {
		if parsedUsage, parsed := extractOpenAIUsageFromJSONBytes(finalResponse); parsed {
			*usage = parsedUsage
		}
		body = s.correctToolCallsInResponseBody(finalResponse)
	} else {
		terminalType, terminalPayload, terminalOK := extractOpenAISSETerminalEvent(bodyText)
		if terminalOK && terminalType == "response.failed" {
			message := extractOpenAISSEErrorMessage(terminalPayload)
			if message == "" {
				message = "Upstream compact response failed"
			}
			return nil, s.writeOpenAINonStreamingProtocolError(resp, c, message)
		}
		usage = s.parseSSEUsageFromBody(bodyText)
	}

	writeOpenAIPassthroughResponseHeaders(c.Writer.Header(), resp.Header, s.responseHeaderFilter)
	contentType := "application/json; charset=utf-8"
	if !ok {
		contentType = resp.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "text/event-stream"
		}
	}
	c.Writer.Header().Set("Content-Type", contentType)
	c.Data(resp.StatusCode, contentType, body)
	return usage, nil
}

func writeOpenAIPassthroughResponseHeaders(dst http.Header, src http.Header, filter *responseheaders.CompiledHeaderFilter) {
	if dst == nil || src == nil {
		return
	}
	if filter != nil {
		responseheaders.WriteFilteredHeaders(dst, src, filter)
	} else {
		if value := strings.TrimSpace(src.Get("Content-Type")); value != "" {
			dst.Set("Content-Type", value)
		}
	}

	getCaseInsensitiveValues := func(headers http.Header, want string) []string {
		if headers == nil {
			return nil
		}
		for key, values := range headers {
			if strings.EqualFold(key, want) {
				return values
			}
		}
		return nil
	}

	for _, rawKey := range []string{
		"x-codex-primary-used-percent",
		"x-codex-primary-reset-after-seconds",
		"x-codex-primary-window-minutes",
		"x-codex-secondary-used-percent",
		"x-codex-secondary-reset-after-seconds",
		"x-codex-secondary-window-minutes",
		"x-codex-primary-over-secondary-limit-percent",
	} {
		values := getCaseInsensitiveValues(src, rawKey)
		if len(values) == 0 {
			continue
		}
		key := http.CanonicalHeaderKey(rawKey)
		dst.Del(key)
		for _, value := range values {
			dst.Add(key, value)
		}
	}
}
