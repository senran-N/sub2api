package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	pkghttputil "github.com/senran-N/sub2api/internal/pkg/httputil"
	middleware2 "github.com/senran-N/sub2api/internal/server/middleware"
	"github.com/senran-N/sub2api/internal/service"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"
)

// ChatCompletions handles OpenAI Chat Completions API endpoint for Anthropic platform groups.
// POST /v1/chat/completions
// This converts Chat Completions requests to Anthropic format (via Responses format chain),
// forwards to Anthropic upstream, and converts responses back to Chat Completions format.
func (h *GatewayHandler) ChatCompletions(c *gin.Context) {
	streamStarted := false

	requestStart := time.Now()

	apiKey, ok := middleware2.GetAPIKeyFromContext(c)
	if !ok {
		h.chatCompletionsErrorResponse(c, http.StatusUnauthorized, "authentication_error", "Invalid API key")
		return
	}

	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		h.chatCompletionsErrorResponse(c, http.StatusInternalServerError, "api_error", "User context not found")
		return
	}
	reqLog := requestLogger(
		c,
		"handler.gateway.chat_completions",
		zap.Int64("user_id", subject.UserID),
		zap.Int64("api_key_id", apiKey.ID),
		zap.Any("group_id", apiKey.GroupID),
	)
	attachRequestAccountLoadCache(c)

	// Read request body
	body, err := pkghttputil.ReadRequestBodyWithPrealloc(c.Request)
	if err != nil {
		if maxErr, ok := extractMaxBytesError(err); ok {
			h.chatCompletionsErrorResponse(c, http.StatusRequestEntityTooLarge, "invalid_request_error", buildBodyTooLargeMessage(maxErr.Limit))
			return
		}
		h.chatCompletionsErrorResponse(c, http.StatusBadRequest, "invalid_request_error", "Failed to read request body")
		return
	}

	if len(body) == 0 {
		h.chatCompletionsErrorResponse(c, http.StatusBadRequest, "invalid_request_error", "Request body is empty")
		return
	}

	setOpsRequestContext(c, "", false, body)

	// Validate JSON
	if !gjson.ValidBytes(body) {
		h.chatCompletionsErrorResponse(c, http.StatusBadRequest, "invalid_request_error", "Failed to parse request body")
		return
	}

	// Extract model and stream
	modelResult := gjson.GetBytes(body, "model")
	if !modelResult.Exists() || modelResult.Type != gjson.String || modelResult.String() == "" {
		h.chatCompletionsErrorResponse(c, http.StatusBadRequest, "invalid_request_error", "model is required")
		return
	}
	reqModel := modelResult.String()
	reqStream := gjson.GetBytes(body, "stream").Bool()
	reqLog = reqLog.With(zap.String("model", reqModel), zap.Bool("stream", reqStream))

	setOpsRequestContext(c, reqModel, reqStream, body)
	setOpsEndpointContext(c, "", int16(service.RequestTypeFromLegacy(reqStream, false)))

	channelMapping, _ := h.gatewayService.ResolveChannelMappingAndRestrict(c.Request.Context(), apiKey.GroupID, reqModel)

	// Claude Code only restriction
	if apiKey.Group != nil && apiKey.Group.ClaudeCodeOnly {
		h.chatCompletionsErrorResponse(c, http.StatusForbidden, "permission_error",
			"This group is restricted to Claude Code clients (/v1/messages only)")
		return
	}

	// Error passthrough binding
	if h.errorPassthroughService != nil {
		service.BindErrorPassthroughService(c, h.errorPassthroughService)
	}

	subscription, _ := middleware2.GetSubscriptionFromContext(c)

	service.SetOpsLatencyMs(c, service.OpsAuthLatencyMsKey, time.Since(requestStart).Milliseconds())

	// 1. Acquire user concurrency slot
	maxWait := service.CalculateMaxWait(subject.Concurrency)
	queueResult, err := h.concurrencyHelper.AcquireUserSlotOrQueue(c.Request.Context(), subject.UserID, subject.Concurrency, maxWait)
	if err != nil {
		reqLog.Warn("gateway.cc.user_slot_acquire_failed", zap.Error(err))
		h.handleConcurrencyError(c, err, "user", streamStarted)
		return
	}
	if !queueResult.Acquired && !queueResult.QueueAllowed {
		h.chatCompletionsErrorResponse(c, http.StatusTooManyRequests, "rate_limit_error", "Too many pending requests, please retry later")
		return
	}
	waitCounted := queueResult.WaitCounted
	defer func() {
		if waitCounted {
			h.concurrencyHelper.DecrementWaitCount(c.Request.Context(), subject.UserID)
		}
	}()

	var userReleaseFunc func()
	if queueResult.Acquired {
		userReleaseFunc = queueResult.ReleaseFunc
	} else {
		userReleaseFunc, err = h.concurrencyHelper.AcquireUserSlotAfterQueueing(c, subject.UserID, subject.Concurrency, reqStream, &streamStarted)
		if err != nil {
			reqLog.Warn("gateway.cc.user_slot_acquire_failed_after_wait", zap.Error(err))
			h.handleConcurrencyError(c, err, "user", streamStarted)
			return
		}
	}
	if waitCounted {
		h.concurrencyHelper.DecrementWaitCount(c.Request.Context(), subject.UserID)
		waitCounted = false
	}
	userReleaseFunc = wrapReleaseOnDone(c.Request.Context(), userReleaseFunc)
	if userReleaseFunc != nil {
		defer userReleaseFunc()
	}

	// 2. Re-check billing
	if err := h.billingCacheService.CheckBillingEligibility(c.Request.Context(), apiKey.User, apiKey, apiKey.Group, subscription); err != nil {
		reqLog.Info("gateway.cc.billing_check_failed", zap.Error(err))
		status, code, message := billingErrorDetails(err)
		h.chatCompletionsErrorResponse(c, status, code, message)
		return
	}

	h.runAnthropicCompatibleTextFlow(c, reqLog, anthropicCompatibleTextFlowRequest{
		Protocol:       service.GatewayProtocolChatCompletions,
		LogPrefix:      "gateway.cc",
		Body:           body,
		RequestedModel: reqModel,
		Stream:         reqStream,
		APIKey:         apiKey,
		Subscription:   subscription,
		ChannelMapping: channelMapping,
		StreamStarted:  &streamStarted,
		WriteError: func(status int, codeOrType, message string) {
			h.chatCompletionsErrorResponse(c, status, codeOrType, message)
		},
		HandleExhausted: func(lastErr *service.UpstreamFailoverError, exhaustedStreamStarted bool) {
			h.handleCCFailoverExhausted(c, lastErr, exhaustedStreamStarted)
		},
	})
}

// chatCompletionsErrorResponse writes an error in OpenAI Chat Completions format.
func (h *GatewayHandler) chatCompletionsErrorResponse(c *gin.Context, status int, errType, message string) {
	c.JSON(status, gin.H{
		"error": gin.H{
			"type":    errType,
			"message": message,
		},
	})
}

// handleCCFailoverExhausted writes a failover-exhausted error in CC format.
func (h *GatewayHandler) handleCCFailoverExhausted(c *gin.Context, lastErr *service.UpstreamFailoverError, streamStarted bool) {
	if streamStarted {
		return
	}
	statusCode := http.StatusBadGateway
	if lastErr != nil && lastErr.StatusCode > 0 {
		statusCode = lastErr.StatusCode
	}
	h.chatCompletionsErrorResponse(c, statusCode, "server_error", "All available accounts exhausted")
}
