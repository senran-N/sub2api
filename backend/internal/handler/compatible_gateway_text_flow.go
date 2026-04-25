package handler

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/senran-N/sub2api/internal/pkg/ip"
	"github.com/senran-N/sub2api/internal/pkg/logger"
	middleware2 "github.com/senran-N/sub2api/internal/server/middleware"
	"github.com/senran-N/sub2api/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type compatibleGatewayTextRouteKind int

const (
	compatibleGatewayTextRouteResponses compatibleGatewayTextRouteKind = iota
	compatibleGatewayTextRouteChatCompletions
	compatibleGatewayTextRouteMessages
)

type compatibleGatewayTextForwardFunc func(context.Context, *gin.Context, *service.Account, string) (*service.OpenAIForwardResult, error)

type compatibleGatewayTextFlow struct {
	route                     compatibleGatewayTextRouteKind
	logPrefix                 string
	recordUsageComponent      string
	body                      []byte
	reqModel                  string
	reqStream                 bool
	schedulingModel           string
	preferredMappedModel      string
	previousResponseID        string
	sessionHash               string
	channelUsage              service.ChannelUsageFields
	apiKey                    *service.APIKey
	subject                   middleware2.AuthSubject
	subscription              *service.UserSubscription
	reqLog                    *zap.Logger
	routingStart              time.Time
	schedulerCtx              context.Context
	streamStarted             *bool
	codexProfile              *service.CodexRequestProfile
	useSelectionFallback      bool
	fallbackContextKey        string
	fallbackSkippedLog        string
	includeRequestPayloadHash bool
	submitUsageWithParent     bool
	forward                   compatibleGatewayTextForwardFunc
}

func (h *CompatibleGatewayTextHandler) compatibleTextRuntime() *service.CompatibleGatewayTextRuntime {
	if h == nil {
		return nil
	}
	if h.textRuntime != nil {
		return h.textRuntime
	}
	if h.gatewayService == nil {
		return nil
	}
	return h.gatewayService.CompatibleTextRuntime()
}

func (h *CompatibleGatewayTextHandler) executeCompatibleGatewayTextFlow(c *gin.Context, flow compatibleGatewayTextFlow) {
	if h == nil || h.gatewayService == nil || flow.forward == nil {
		h.handleTextFlowForwardError(c, flow, nil, errors.New("compatible gateway text flow is not configured"))
		return
	}
	if flow.schedulerCtx == nil {
		flow.schedulerCtx = c.Request.Context()
	}
	if flow.streamStarted == nil {
		started := false
		flow.streamStarted = &started
	}
	if flow.logPrefix == "" {
		flow.logPrefix = "openai"
	}
	if flow.recordUsageComponent == "" {
		flow.recordUsageComponent = "handler.openai_gateway.responses"
	}

	codexSchedulingObserved := false

	result := service.NewCompatibleTextExecutionKernel(h.gatewayService).Execute(c.Request.Context(), service.CompatibleTextExecutionRequest{
		ProtocolFamily:            flow.protocolFamily(),
		Provider:                  service.PlatformOpenAI,
		APIKey:                    flow.apiKey,
		Body:                      flow.body,
		ReqModel:                  flow.reqModel,
		ReqStream:                 flow.reqStream,
		SchedulingModel:           flow.schedulingModel,
		PreferredMappedModel:      flow.preferredMappedModel,
		PreviousResponseID:        flow.previousResponseID,
		SessionHash:               flow.sessionHash,
		ChannelUsage:              flow.channelUsage,
		RoutingStart:              flow.routingStart,
		SchedulerContext:          flow.schedulerCtx,
		CodexProfile:              flow.codexProfile,
		UseSelectionFallback:      flow.useSelectionFallback,
		IncludeRequestPayloadHash: flow.includeRequestPayloadHash,
		MaxAccountSwitches:        h.maxAccountSwitches,
		Forward: func(ctx context.Context, account *service.Account, defaultMappedModel string) (*service.OpenAIForwardResult, error) {
			return flow.forward(ctx, c, account, defaultMappedModel)
		},
		AcquireAccountSlot: func(_ context.Context, sessionHash string, selection *service.AccountSelectionResult) (func(), bool) {
			return h.acquireResponsesAccountSlot(c, flow.apiKey.GroupID, sessionHash, selection, flow.reqStream, flow.streamStarted, flow.reqLog)
		},
		Hooks: h.compatibleTextExecutionHooks(c, flow, &codexSchedulingObserved),
	})

	flow.sessionHash = result.SessionHash
	switch result.Outcome {
	case service.CompatibleTextExecutionSucceeded:
		h.recordCompatibleGatewayTextUsage(c, flow, result.Account, result.ForwardResult)
	case service.CompatibleTextExecutionSelectionError:
		h.handleTextFlowSelectionError(c, flow, result.InitialSelectionErr, result.Err, result.LastFailoverErr, result.FailedAccountIDs)
	case service.CompatibleTextExecutionNoAvailable:
		h.handleTextFlowNoAvailableAccount(c, flow)
	case service.CompatibleTextExecutionAcquireBlocked, service.CompatibleTextExecutionCanceled:
		return
	case service.CompatibleTextExecutionFailoverExhausted:
		h.handleTextFlowFailoverExhausted(c, flow, result.FailoverErr)
	case service.CompatibleTextExecutionForwardError:
		h.handleTextFlowForwardError(c, flow, result.Account, result.Err)
	default:
		h.handleTextFlowForwardError(c, flow, result.Account, result.Err)
	}
}

func (flow compatibleGatewayTextFlow) protocolFamily() service.CompatibleGatewayProtocolFamily {
	switch flow.route {
	case compatibleGatewayTextRouteChatCompletions:
		return service.CompatibleGatewayProtocolFamilyChatCompletions
	case compatibleGatewayTextRouteMessages:
		return service.CompatibleGatewayProtocolFamilyMessages
	default:
		return service.CompatibleGatewayProtocolFamilyResponses
	}
}

func (h *CompatibleGatewayTextHandler) compatibleTextExecutionHooks(
	c *gin.Context,
	flow compatibleGatewayTextFlow,
	codexSchedulingObserved *bool,
) service.CompatibleTextExecutionHooks {
	return service.CompatibleTextExecutionHooks{
		ResetFallbackModel: func() {
			if flow.fallbackContextKey != "" {
				c.Set(flow.fallbackContextKey, "")
			}
		},
		ResolveSelectionFallback: func(_ context.Context, selectionModel string, _ error) string {
			return resolveOpenAISelectionFallbackModel(
				c,
				h.gatewayService,
				flow.apiKey,
				selectionModel,
				flow.reqLog,
				flow.fallbackSkippedLog,
			)
		},
		OnFallbackSelected: func(defaultModel string) {
			if flow.fallbackContextKey != "" {
				c.Set(flow.fallbackContextKey, defaultModel)
			}
			flow.reqLog.Info(flow.logPrefix+".fallback_to_default_model",
				zap.String("default_mapped_model", defaultModel),
			)
		},
		OnAccountSelecting: func(_ string, excludedCount int) {
			flow.reqLog.Debug(flow.logPrefix+".account_selecting", zap.Int("excluded_account_count", excludedCount))
		},
		OnAccountSelectFailed: func(err error, excludedCount int) {
			flow.reqLog.Warn(flow.logPrefix+".account_select_failed",
				zap.Error(err),
				zap.Int("excluded_account_count", excludedCount),
			)
		},
		OnPreviousResponseSelected: func(account *service.Account) {
			flow.reqLog.Debug("openai.account_selected_with_previous_response_id", zap.Int64("account_id", account.ID))
		},
		OnScheduleDecision: func(decision service.OpenAIAccountScheduleDecision) {
			h.logCompatibleGatewayTextScheduleDecision(flow, decision, codexSchedulingObserved)
		},
		OnAccountSelected: func(account *service.Account) {
			flow.reqLog.Debug(flow.logPrefix+".account_selected", zap.Int64("account_id", account.ID), zap.String("account_name", account.Name))
			setOpsSelectedAccount(c, account.ID, account.Platform)
		},
		OnRoutingLatency: func(d time.Duration) {
			service.SetOpsLatencyMs(c, service.OpsRoutingLatencyMsKey, d.Milliseconds())
		},
		OnForwardLatency: func(d time.Duration, result *service.OpenAIForwardResult) {
			setCompatibleGatewayTextResponseLatency(c, d.Milliseconds(), result)
		},
		ResolveCodexFailoverRecovery: func(account *service.Account, failoverErr *service.UpstreamFailoverError, stickyContext bool) service.CodexRecoveryDecision {
			return h.gatewayService.ResolveCodexFailoverRecovery(c, account, failoverErr, stickyContext)
		},
		RecordCodexRecoverySwitch: func(account *service.Account, failoverErr *service.UpstreamFailoverError, trackMetrics bool) {
			h.gatewayService.RecordCodexRecoveryAccountSwitch(c, account, failoverErr, trackMetrics)
		},
		OnSameAccountRetry: func(account *service.Account, failoverErr *service.UpstreamFailoverError, decision service.OpenAIPoolFailoverDecision) {
			flow.reqLog.Warn(flow.logPrefix+".pool_mode_same_account_retry",
				zap.Int64("account_id", account.ID),
				zap.Int("upstream_status", failoverErr.StatusCode),
				zap.Int("retry_limit", decision.RetryLimit),
				zap.Int("retry_count", decision.RetryCount),
			)
		},
		OnFailoverSwitch: func(account *service.Account, failoverErr *service.UpstreamFailoverError, decision service.OpenAIPoolFailoverDecision) {
			flow.reqLog.Warn(flow.logPrefix+".upstream_failover_switching",
				zap.Int64("account_id", account.ID),
				zap.Int("upstream_status", failoverErr.StatusCode),
				zap.Int("switch_count", decision.SwitchCount),
				zap.Int("max_switches", h.maxAccountSwitches),
			)
		},
		OnCompleted: func(account *service.Account, switchCount int) {
			flow.reqLog.Debug(flow.logPrefix+".request_completed",
				zap.Int64("account_id", account.ID),
				zap.Int("switch_count", switchCount),
			)
		},
	}
}

func (h *CompatibleGatewayTextHandler) handleTextFlowSelectionError(
	c *gin.Context,
	flow compatibleGatewayTextFlow,
	initialErr error,
	err error,
	lastFailoverErr *service.UpstreamFailoverError,
	failedAccountIDs map[int64]struct{},
) {
	if len(failedAccountIDs) == 0 {
		if flow.route == compatibleGatewayTextRouteMessages {
			h.anthropicStreamingAwareError(c, http.StatusServiceUnavailable, "api_error", "No available accounts", *flow.streamStarted)
			return
		}
		if initialErr == nil {
			initialErr = err
		}
		status, code, message := openAISelectionErrorResponseAfterDefaultFallback(initialErr, err)
		h.handleStreamingAwareError(c, status, code, message, *flow.streamStarted)
		return
	}
	if lastFailoverErr != nil {
		h.handleTextFlowFailoverExhausted(c, flow, lastFailoverErr)
		return
	}
	if flow.route == compatibleGatewayTextRouteMessages {
		h.anthropicStreamingAwareError(c, http.StatusBadGateway, "api_error", "Upstream request failed", *flow.streamStarted)
		return
	}
	h.handleFailoverExhaustedSimple(c, http.StatusBadGateway, *flow.streamStarted)
}

func (h *CompatibleGatewayTextHandler) handleTextFlowNoAvailableAccount(c *gin.Context, flow compatibleGatewayTextFlow) {
	if flow.route == compatibleGatewayTextRouteMessages {
		h.anthropicStreamingAwareError(c, http.StatusServiceUnavailable, "api_error", "No available accounts", *flow.streamStarted)
		return
	}
	h.handleStreamingAwareError(c, http.StatusServiceUnavailable, "api_error", "No available accounts", *flow.streamStarted)
}

func (h *CompatibleGatewayTextHandler) handleTextFlowFailoverExhausted(c *gin.Context, flow compatibleGatewayTextFlow, failoverErr *service.UpstreamFailoverError) {
	if failoverErr == nil {
		if flow.route == compatibleGatewayTextRouteMessages {
			h.anthropicStreamingAwareError(c, http.StatusBadGateway, "api_error", "Upstream request failed", *flow.streamStarted)
			return
		}
		h.handleFailoverExhaustedSimple(c, http.StatusBadGateway, *flow.streamStarted)
		return
	}
	if flow.route == compatibleGatewayTextRouteMessages {
		h.handleAnthropicFailoverExhausted(c, failoverErr, *flow.streamStarted)
		return
	}
	h.handleFailoverExhausted(c, failoverErr, *flow.streamStarted)
}

func (h *CompatibleGatewayTextHandler) handleTextFlowForwardError(c *gin.Context, flow compatibleGatewayTextFlow, account *service.Account, err error) {
	wroteFallback := false
	if flow.route == compatibleGatewayTextRouteMessages {
		wroteFallback = h.ensureAnthropicErrorResponse(c, *flow.streamStarted)
	} else {
		wroteFallback = h.ensureForwardErrorResponse(c, *flow.streamStarted)
	}
	fields := []zap.Field{
		zap.Bool("fallback_error_response_written", wroteFallback),
		zap.Error(err),
	}
	if account != nil {
		fields = append([]zap.Field{zap.Int64("account_id", account.ID)}, fields...)
	}
	if flow.route == compatibleGatewayTextRouteResponses && !shouldLogOpenAIForwardFailureAsWarn(c, wroteFallback) {
		flow.reqLog.Error(flow.logPrefix+".forward_failed", fields...)
		return
	}
	flow.reqLog.Warn(flow.logPrefix+".forward_failed", fields...)
}

func (h *CompatibleGatewayTextHandler) logCompatibleGatewayTextScheduleDecision(flow compatibleGatewayTextFlow, decision service.OpenAIAccountScheduleDecision, observed *bool) {
	if flow.route != compatibleGatewayTextRouteResponses {
		return
	}
	flow.reqLog.Debug("openai.account_schedule_decision",
		zap.String("layer", decision.Layer),
		zap.Bool("sticky_previous_hit", decision.StickyPreviousHit),
		zap.Bool("sticky_session_hit", decision.StickySessionHit),
		zap.Int("candidate_count", decision.CandidateCount),
		zap.Int("top_k", decision.TopK),
		zap.Int64("latency_ms", decision.LatencyMs),
		zap.Float64("load_skew", decision.LoadSkew),
	)
	if observed == nil || *observed || flow.codexProfile == nil {
		return
	}
	service.ObserveOpenAICodexSchedulingDecision(*flow.codexProfile, decision)
	*observed = true
}

func setCompatibleGatewayTextResponseLatency(c *gin.Context, forwardDurationMs int64, result *service.OpenAIForwardResult) {
	upstreamLatencyMs, _ := getContextInt64(c, service.OpsUpstreamLatencyMsKey)
	responseLatencyMs := forwardDurationMs
	if upstreamLatencyMs > 0 && forwardDurationMs > upstreamLatencyMs {
		responseLatencyMs = forwardDurationMs - upstreamLatencyMs
	}
	service.SetOpsLatencyMs(c, service.OpsResponseLatencyMsKey, responseLatencyMs)
	if result != nil && result.FirstTokenMs != nil {
		service.SetOpsLatencyMs(c, service.OpsTimeToFirstTokenMsKey, int64(*result.FirstTokenMs))
	}
}

func (h *CompatibleGatewayTextHandler) recordCompatibleGatewayTextUsage(c *gin.Context, flow compatibleGatewayTextFlow, account *service.Account, result *service.OpenAIForwardResult) {
	userAgent := c.GetHeader("User-Agent")
	clientIP := ip.GetClientIP(c)
	requestPayloadHash := ""
	if flow.includeRequestPayloadHash {
		requestPayloadHash = service.HashUsageRequestPayload(flow.body)
	}
	task := func(ctx context.Context) {
		if err := h.gatewayService.RecordUsage(ctx, &service.OpenAIRecordUsageInput{
			Result:             result,
			APIKey:             flow.apiKey,
			User:               flow.apiKey.User,
			Account:            account,
			Subscription:       flow.subscription,
			InboundEndpoint:    GetInboundEndpoint(c),
			UpstreamEndpoint:   GetUpstreamEndpoint(c, account.Platform),
			UserAgent:          userAgent,
			IPAddress:          clientIP,
			RequestPayloadHash: requestPayloadHash,
			APIKeyService:      h.apiKeyService,
			ChannelUsageFields: flow.channelUsage,
		}); err != nil {
			logger.L().With(
				zap.String("component", flow.recordUsageComponent),
				zap.Int64("user_id", flow.subject.UserID),
				zap.Int64("api_key_id", flow.apiKey.ID),
				zap.Any("group_id", flow.apiKey.GroupID),
				zap.String("model", flow.reqModel),
				zap.Int64("account_id", account.ID),
			).Error(flow.logPrefix+".record_usage_failed", zap.Error(err))
		}
	}
	if flow.submitUsageWithParent {
		h.submitUsageRecordTaskWithParent(c.Request.Context(), task)
		return
	}
	h.submitUsageRecordTask(task)
}
