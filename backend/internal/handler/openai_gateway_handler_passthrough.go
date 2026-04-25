package handler

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	pkghttputil "github.com/senran-N/sub2api/internal/pkg/httputil"
	"github.com/senran-N/sub2api/internal/pkg/ip"
	"github.com/senran-N/sub2api/internal/pkg/logger"
	middleware2 "github.com/senran-N/sub2api/internal/server/middleware"
	"github.com/senran-N/sub2api/internal/service"
	"go.uber.org/zap"
)

// Passthrough handles OpenAI-compatible HTTP endpoints that should be proxied
// to passthrough-capable upstreams such as xAI/Grok.
func (h *OpenAIGatewayHandler) Passthrough(c *gin.Context) {
	streamStarted := false
	defer h.recoverResponsesPanic(c, &streamStarted)

	requestStart := time.Now()

	apiKey, ok := middleware2.GetAPIKeyFromContext(c)
	if !ok {
		h.errorResponse(c, http.StatusUnauthorized, "authentication_error", "Invalid API key")
		return
	}

	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		h.errorResponse(c, http.StatusInternalServerError, "api_error", "User context not found")
		return
	}

	inboundEndpoint := GetInboundEndpoint(c)
	reqLog := requestLogger(
		c,
		"handler.openai_gateway.passthrough",
		zap.Int64("user_id", subject.UserID),
		zap.Int64("api_key_id", apiKey.ID),
		zap.Any("group_id", apiKey.GroupID),
		zap.String("method", c.Request.Method),
		zap.String("endpoint", inboundEndpoint),
	)
	attachRequestAccountLoadCache(c)

	if !h.ensureResponsesDependencies(c, reqLog) {
		return
	}

	body, err := pkghttputil.ReadRequestBodyWithPrealloc(c.Request)
	if err != nil {
		if maxErr, ok := extractMaxBytesError(err); ok {
			h.errorResponse(c, http.StatusRequestEntityTooLarge, "invalid_request_error", buildBodyTooLargeMessage(maxErr.Limit))
			return
		}
		h.errorResponse(c, http.StatusBadRequest, "invalid_request_error", "Failed to read request body")
		return
	}

	reqMeta := service.GetOpenAICompatiblePassthroughRequestMeta(c, body)
	reqModel := reqMeta.Model
	reqStream := reqMeta.Stream || strings.Contains(strings.ToLower(c.GetHeader("Accept")), "text/event-stream")
	reqLog = reqLog.With(zap.String("model", reqModel), zap.Bool("stream", reqStream))

	setOpsRequestContext(c, reqModel, reqStream, body)
	setOpsEndpointContext(c, "", int16(service.RequestTypeFromLegacy(reqStream, false)))

	if h.errorPassthroughService != nil {
		service.BindErrorPassthroughService(c, h.errorPassthroughService)
	}

	subscription, _ := middleware2.GetSubscriptionFromContext(c)

	service.SetOpsLatencyMs(c, service.OpsAuthLatencyMsKey, time.Since(requestStart).Milliseconds())
	routingStart := time.Now()

	userReleaseFunc, acquired := h.acquireResponsesUserSlot(c, subject.UserID, subject.Concurrency, reqStream, &streamStarted, reqLog)
	if !acquired {
		return
	}
	if userReleaseFunc != nil {
		defer userReleaseFunc()
	}

	if err := h.billingCacheService.CheckBillingEligibility(c.Request.Context(), apiKey.User, apiKey, apiKey.Group, subscription); err != nil {
		reqLog.Info("openai_passthrough.billing_eligibility_check_failed", zap.Error(err))
		status, code, message := billingErrorDetails(err)
		h.handleStreamingAwareError(c, status, code, message, streamStarted)
		return
	}

	sessionHash := h.gatewayService.GenerateSessionHash(c, body)
	channelUsage := service.ChannelUsageFields{}
	schedulingModel := reqModel
	if reqModel != "" {
		channelMapping, restricted := h.gatewayService.ResolveChannelMappingAndRestrict(c.Request.Context(), apiKey.GroupID, reqModel)
		if restricted {
			h.handleStreamingAwareError(c, http.StatusBadRequest, "invalid_request_error", "Requested model is not allowed for this channel", streamStarted)
			return
		}
		channelUsage = channelMapping.ToUsageFields(reqModel, "")
		if channelMapping.MappedModel != "" {
			schedulingModel = channelMapping.MappedModel
		}
	}

	result := service.NewCompatiblePassthroughExecutionKernel(h.gatewayService).Execute(c.Request.Context(), service.CompatiblePassthroughExecutionRequest{
		APIKey:               apiKey,
		ReqModel:             reqModel,
		ReqStream:            reqStream,
		SchedulingModel:      schedulingModel,
		SessionHash:          sessionHash,
		ChannelUsage:         channelUsage,
		RoutingStart:         routingStart,
		SchedulerContext:     c.Request.Context(),
		UseSelectionFallback: schedulingModel != "",
		MaxAccountSwitches:   h.maxAccountSwitches,
		RequiredTransport:    service.OpenAIUpstreamTransportAny,
		ValidateAccount:      (*service.Account).SupportsOpenAIPassthroughHTTP,
		AcquireAccountSlot: func(_ context.Context, sessionHash string, selection *service.AccountSelectionResult) (func(), bool) {
			return h.acquireResponsesAccountSlot(c, apiKey.GroupID, sessionHash, selection, reqStream, &streamStarted, reqLog)
		},
		Forward: func(ctx context.Context, account *service.Account, defaultMappedModel string) (*service.OpenAIForwardResult, error) {
			return h.gatewayService.ForwardCompatiblePassthrough(ctx, c, account, body, reqMeta, defaultMappedModel)
		},
		Hooks: service.CompatiblePassthroughExecutionHooks{
			ResetFallbackModel: func() {
				c.Set("openai_passthrough_fallback_model", "")
			},
			ResolveSelectionFallback: func(_ context.Context, selectionModel string, _ error) string {
				return resolveOpenAISelectionFallbackModel(
					c,
					h.gatewayService,
					apiKey,
					selectionModel,
					reqLog,
					"openai_passthrough.fallback_to_default_model_skipped",
				)
			},
			OnFallbackSelected: func(defaultModel string) {
				c.Set("openai_passthrough_fallback_model", defaultModel)
				reqLog.Info("openai_passthrough.fallback_to_default_model", zap.String("default_mapped_model", defaultModel))
			},
			OnAccountSelecting: func(_ string, excludedCount int) {
				reqLog.Debug("openai_passthrough.account_selecting", zap.Int("excluded_account_count", excludedCount))
			},
			OnAccountSelectFailed: func(err error, excludedCount int) {
				reqLog.Warn("openai_passthrough.account_select_failed",
					zap.Error(err),
					zap.Int("excluded_account_count", excludedCount),
				)
			},
			OnAccountSelected: func(account *service.Account) {
				setOpsSelectedAccount(c, account.ID, account.Platform)
			},
			OnAccountRejected: func(account *service.Account) {
				reqLog.Warn("openai_passthrough.account_not_compatible",
					zap.Int64("account_id", account.ID),
					zap.String("account_name", account.Name),
					zap.String("account_type", account.Type),
				)
			},
			OnRoutingLatency: func(d time.Duration) {
				service.SetOpsLatencyMs(c, service.OpsRoutingLatencyMsKey, d.Milliseconds())
			},
			OnForwardLatency: func(d time.Duration, result *service.OpenAIForwardResult) {
				setCompatibleGatewayTextResponseLatency(c, d.Milliseconds(), result)
			},
			OnFailoverAttemptFailed: func(attempt service.CompatibleTextExecutionAttemptFeedback) {
				if attempt.FailoverErr == nil {
					return
				}
				h.persistCompatibleGatewayRuntimeFeedback(c.Request.Context(), compatibleGatewayRuntimeFeedbackInput{
					Account:        attempt.Account,
					RequestedModel: reqModel,
					UpstreamModel:  attempt.ForwardModelHint,
					StatusCode:     attempt.FailoverErr.StatusCode,
					Endpoint:       inboundEndpoint,
					Err:            attempt.FailoverErr,
				})
			},
			RecordCodexRecoverySwitch: func(account *service.Account, failoverErr *service.UpstreamFailoverError, trackMetrics bool) {
				h.gatewayService.RecordCodexRecoveryAccountSwitch(c, account, failoverErr, trackMetrics)
			},
			OnSameAccountRetry: func(account *service.Account, failoverErr *service.UpstreamFailoverError, decision service.OpenAIPoolFailoverDecision) {
				reqLog.Warn("openai_passthrough.pool_mode_same_account_retry",
					zap.Int64("account_id", account.ID),
					zap.Int("upstream_status", failoverErr.StatusCode),
					zap.Int("retry_limit", decision.RetryLimit),
					zap.Int("retry_count", decision.RetryCount),
				)
			},
			OnFailoverSwitch: func(account *service.Account, failoverErr *service.UpstreamFailoverError, decision service.OpenAIPoolFailoverDecision) {
				reqLog.Warn("openai_passthrough.upstream_failover_switching",
					zap.Int64("account_id", account.ID),
					zap.Int("upstream_status", failoverErr.StatusCode),
					zap.Int("switch_count", decision.SwitchCount),
					zap.Int("max_switches", h.maxAccountSwitches),
				)
			},
		},
	})

	switch result.Outcome {
	case service.CompatibleTextExecutionSucceeded:
		account := result.Account
		if account == nil {
			h.handleStreamingAwareError(c, http.StatusBadGateway, "api_error", "Upstream request failed", streamStarted)
			return
		}
		h.persistCompatibleGatewayRuntimeFeedback(c.Request.Context(), compatibleGatewayRuntimeFeedbackInput{
			Account:        account,
			RequestedModel: reqModel,
			UpstreamModel:  result.ForwardModelHint,
			Result:         result.ForwardResult,
			StatusCode:     c.Writer.Status(),
			Endpoint:       inboundEndpoint,
		})
		userAgent := c.GetHeader("User-Agent")
		clientIP := ip.GetClientIP(c)
		h.submitUsageRecordTask(func(ctx context.Context) {
			if err := h.gatewayService.RecordUsage(ctx, &service.OpenAIRecordUsageInput{
				Result:             result.ForwardResult,
				APIKey:             apiKey,
				User:               apiKey.User,
				Account:            account,
				Subscription:       subscription,
				InboundEndpoint:    inboundEndpoint,
				UpstreamEndpoint:   GetUpstreamEndpoint(c, account.Platform),
				UserAgent:          userAgent,
				IPAddress:          clientIP,
				APIKeyService:      h.apiKeyService,
				ChannelUsageFields: channelUsage,
			}); err != nil {
				logger.L().With(
					zap.String("component", "handler.openai_gateway.passthrough"),
					zap.Int64("user_id", subject.UserID),
					zap.Int64("api_key_id", apiKey.ID),
					zap.Any("group_id", apiKey.GroupID),
					zap.String("endpoint", inboundEndpoint),
					zap.String("model", reqModel),
					zap.Int64("account_id", account.ID),
				).Error("openai_passthrough.record_usage_failed", zap.Error(err))
			}
		})
		reqLog.Debug("openai_passthrough.request_completed",
			zap.Int64("account_id", account.ID),
			zap.Int("switch_count", result.SwitchCount),
		)
	case service.CompatibleTextExecutionSelectionError:
		if len(result.FailedAccountIDs) == 0 {
			status, code, message := openAISelectionErrorResponseAfterDefaultFallback(result.InitialSelectionErr, result.Err)
			h.handleStreamingAwareError(c, status, code, message, streamStarted)
			return
		}
		if result.LastFailoverErr != nil {
			h.handleFailoverExhausted(c, result.LastFailoverErr, streamStarted)
			return
		}
		h.handleStreamingAwareError(c, http.StatusServiceUnavailable, "api_error", "No passthrough-capable accounts available", streamStarted)
	case service.CompatibleTextExecutionNoAvailable:
		h.handleStreamingAwareError(c, http.StatusServiceUnavailable, "api_error", "No available accounts", streamStarted)
	case service.CompatibleTextExecutionAcquireBlocked, service.CompatibleTextExecutionCanceled:
		return
	case service.CompatibleTextExecutionFailoverExhausted:
		if result.FailoverErr == nil {
			h.handleStreamingAwareError(c, http.StatusBadGateway, "api_error", "Upstream request failed", streamStarted)
			return
		}
		h.handleFailoverExhausted(c, result.FailoverErr, streamStarted)
	case service.CompatibleTextExecutionForwardError:
		wroteFallback := h.ensureForwardErrorResponse(c, streamStarted)
		h.persistCompatibleGatewayRuntimeFeedback(c.Request.Context(), compatibleGatewayRuntimeFeedbackInput{
			Account:        result.Account,
			RequestedModel: reqModel,
			UpstreamModel:  result.ForwardModelHint,
			StatusCode:     c.Writer.Status(),
			Endpoint:       inboundEndpoint,
			Err:            result.Err,
		})
		fields := []zap.Field{
			zap.Bool("fallback_error_response_written", wroteFallback),
			zap.Error(result.Err),
		}
		if result.Account != nil {
			fields = append([]zap.Field{zap.Int64("account_id", result.Account.ID)}, fields...)
		}
		reqLog.Warn("openai_passthrough.forward_failed", fields...)
	default:
		h.handleStreamingAwareError(c, http.StatusBadGateway, "api_error", "Upstream request failed", streamStarted)
		return
	}
}
