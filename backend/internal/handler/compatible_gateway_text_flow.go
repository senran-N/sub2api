package handler

import (
	"context"
	"errors"
	"net/http"
	"strings"
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

	maxAccountSwitches := h.maxAccountSwitches
	switchCount := 0
	failedAccountIDs := make(map[int64]struct{})
	sameAccountRetryCount := make(map[int64]int)
	var lastFailoverErr *service.UpstreamFailoverError
	codexSchedulingObserved := false

	for {
		if flow.fallbackContextKey != "" {
			c.Set(flow.fallbackContextKey, "")
		}
		selectionModel := flow.selectionModel()
		flow.reqLog.Debug(flow.logPrefix+".account_selecting", zap.Int("excluded_account_count", len(failedAccountIDs)))
		selection, scheduleDecision, err := h.gatewayService.SelectAccountWithScheduler(
			flow.schedulerCtx,
			flow.apiKey.GroupID,
			flow.previousResponseID,
			flow.sessionHash,
			selectionModel,
			failedAccountIDs,
			service.OpenAIUpstreamTransportAny,
		)
		if err != nil {
			flow.reqLog.Warn(flow.logPrefix+".account_select_failed",
				zap.Error(err),
				zap.Int("excluded_account_count", len(failedAccountIDs)),
			)
			if !h.tryCompatibleGatewayTextSelectionFallback(c, flow, selectionModel, failedAccountIDs, &selection, &scheduleDecision, &err) {
				h.handleTextFlowSelectionError(c, flow, err, lastFailoverErr, failedAccountIDs)
				return
			}
		}
		if selection == nil || selection.Account == nil {
			h.handleTextFlowNoAvailableAccount(c, flow)
			return
		}

		if flow.route == compatibleGatewayTextRouteResponses && flow.previousResponseID != "" {
			flow.reqLog.Debug("openai.account_selected_with_previous_response_id", zap.Int64("account_id", selection.Account.ID))
		}
		h.logCompatibleGatewayTextScheduleDecision(flow, scheduleDecision, &codexSchedulingObserved)

		account := selection.Account
		flow.sessionHash = ensureOpenAIPoolModeSessionHash(flow.sessionHash, account)
		flow.reqLog.Debug(flow.logPrefix+".account_selected", zap.Int64("account_id", account.ID), zap.String("account_name", account.Name))
		setOpsSelectedAccount(c, account.ID, account.Platform)

		accountReleaseFunc, acquired := h.acquireResponsesAccountSlot(c, flow.apiKey.GroupID, flow.sessionHash, selection, flow.reqStream, flow.streamStarted, flow.reqLog)
		if !acquired {
			return
		}

		service.SetOpsLatencyMs(c, service.OpsRoutingLatencyMsKey, time.Since(flow.routingStart).Milliseconds())
		forwardStart := time.Now()
		defaultMappedModel := flow.forwardDefaultMappedModel(c)
		result, err := flow.forward(c.Request.Context(), c, account, defaultMappedModel)

		forwardDurationMs := time.Since(forwardStart).Milliseconds()
		if accountReleaseFunc != nil {
			accountReleaseFunc()
		}
		setCompatibleGatewayTextResponseLatency(c, forwardDurationMs, result)

		if err != nil {
			var failoverErr *service.UpstreamFailoverError
			if errors.As(err, &failoverErr) {
				h.gatewayService.ReportOpenAIAccountScheduleResult(account.ID, false, nil)
				lastFailoverErr = failoverErr
				decision := h.applyCompatibleGatewayTextFailoverPolicy(c, flow, account, failoverErr, sameAccountRetryCount, failedAccountIDs, &switchCount, maxAccountSwitches)
				if decision.SameAccountRetry {
					flow.reqLog.Warn(flow.logPrefix+".pool_mode_same_account_retry",
						zap.Int64("account_id", account.ID),
						zap.Int("upstream_status", failoverErr.StatusCode),
						zap.Int("retry_limit", decision.RetryLimit),
						zap.Int("retry_count", decision.RetryCount),
					)
					if !sleepWithContext(c.Request.Context(), sameAccountRetryDelay) {
						return
					}
					continue
				}
				if decision.Action == FailoverExhausted {
					h.handleTextFlowFailoverExhausted(c, flow, failoverErr)
					return
				}
				flow.reqLog.Warn(flow.logPrefix+".upstream_failover_switching",
					zap.Int64("account_id", account.ID),
					zap.Int("upstream_status", failoverErr.StatusCode),
					zap.Int("switch_count", decision.SwitchCount),
					zap.Int("max_switches", maxAccountSwitches),
				)
				continue
			}
			h.gatewayService.ReportOpenAIAccountScheduleResult(account.ID, false, nil)
			h.handleTextFlowForwardError(c, flow, account, err)
			return
		}

		h.reportCompatibleGatewayTextSuccess(c, flow, account, result)
		flow.reqLog.Debug(flow.logPrefix+".request_completed",
			zap.Int64("account_id", account.ID),
			zap.Int("switch_count", switchCount),
		)
		return
	}
}

func (flow compatibleGatewayTextFlow) selectionModel() string {
	if preferred := strings.TrimSpace(flow.preferredMappedModel); preferred != "" {
		return preferred
	}
	return flow.schedulingModel
}

func (flow compatibleGatewayTextFlow) forwardDefaultMappedModel(c *gin.Context) string {
	if flow.route == compatibleGatewayTextRouteMessages {
		return strings.TrimSpace(flow.preferredMappedModel)
	}
	fallbackModel := ""
	if flow.fallbackContextKey != "" {
		fallbackModel = c.GetString(flow.fallbackContextKey)
	}
	if fallbackModel == "" && flow.channelUsage.ChannelMappedModel != "" {
		fallbackModel = flow.channelUsage.ChannelMappedModel
	}
	return resolveOpenAIForwardDefaultMappedModel(flow.apiKey, fallbackModel)
}

func (h *CompatibleGatewayTextHandler) tryCompatibleGatewayTextSelectionFallback(
	c *gin.Context,
	flow compatibleGatewayTextFlow,
	selectionModel string,
	failedAccountIDs map[int64]struct{},
	selection **service.AccountSelectionResult,
	scheduleDecision *service.OpenAIAccountScheduleDecision,
	selectionErr *error,
) bool {
	if !flow.useSelectionFallback || len(failedAccountIDs) != 0 {
		return false
	}
	initialSelectionErr := *selectionErr
	defaultModel := resolveOpenAISelectionFallbackModel(
		c,
		h.gatewayService,
		flow.apiKey,
		selectionModel,
		flow.reqLog,
		flow.fallbackSkippedLog,
	)
	if defaultModel == "" || defaultModel == selectionModel {
		status, code, message := openAISelectionErrorResponseAfterDefaultFallback(initialSelectionErr, *selectionErr)
		h.handleStreamingAwareError(c, status, code, message, *flow.streamStarted)
		return true
	}
	flow.reqLog.Info(flow.logPrefix+".fallback_to_default_model",
		zap.String("default_mapped_model", defaultModel),
	)
	selected, decision, err := h.gatewayService.SelectAccountWithScheduler(
		flow.schedulerCtx,
		flow.apiKey.GroupID,
		flow.previousResponseID,
		flow.sessionHash,
		defaultModel,
		failedAccountIDs,
		service.OpenAIUpstreamTransportAny,
	)
	if err != nil {
		status, code, message := openAISelectionErrorResponseAfterDefaultFallback(initialSelectionErr, err)
		h.handleStreamingAwareError(c, status, code, message, *flow.streamStarted)
		return true
	}
	if selected != nil && flow.fallbackContextKey != "" {
		c.Set(flow.fallbackContextKey, defaultModel)
	}
	*selection = selected
	*scheduleDecision = decision
	*selectionErr = nil
	return true
}

func (h *CompatibleGatewayTextHandler) handleTextFlowSelectionError(c *gin.Context, flow compatibleGatewayTextFlow, err error, lastFailoverErr *service.UpstreamFailoverError, failedAccountIDs map[int64]struct{}) {
	if len(failedAccountIDs) == 0 {
		if flow.route == compatibleGatewayTextRouteMessages {
			h.anthropicStreamingAwareError(c, http.StatusServiceUnavailable, "api_error", "No available accounts", *flow.streamStarted)
			return
		}
		status, code, message := openAISelectionErrorResponseAfterDefaultFallback(err, err)
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

func (h *CompatibleGatewayTextHandler) applyCompatibleGatewayTextFailoverPolicy(
	c *gin.Context,
	flow compatibleGatewayTextFlow,
	account *service.Account,
	failoverErr *service.UpstreamFailoverError,
	sameAccountRetryCount map[int64]int,
	failedAccountIDs map[int64]struct{},
	switchCount *int,
	maxAccountSwitches int,
) openAIPoolFailoverDecision {
	codexFailoverDecision := service.CodexRecoveryDecision{}
	stickyContext := flow.sessionHash != ""
	if flow.route == compatibleGatewayTextRouteResponses {
		stickyContext = stickyContext || flow.previousResponseID != ""
		if flow.codexProfile != nil && flow.codexProfile.NativeClient {
			codexFailoverDecision = h.gatewayService.ResolveCodexFailoverRecovery(c, account, failoverErr, stickyContext)
		}
	}
	return applyOpenAIPoolFailoverPolicy(
		account,
		failoverErr,
		codexFailoverDecision,
		stickyContext,
		sameAccountRetryCount,
		failedAccountIDs,
		switchCount,
		maxAccountSwitches,
		func() {
			h.gatewayService.TempUnscheduleRetryableError(c.Request.Context(), account.ID, failoverErr)
		},
		func() {
			h.gatewayService.RecordOpenAIAccountSwitch()
			if flow.route == compatibleGatewayTextRouteResponses {
				if flow.codexProfile != nil && flow.codexProfile.NativeClient && codexFailoverDecision.SwitchAccount {
					h.gatewayService.RecordCodexRecoveryAccountSwitch(c, account, failoverErr, true)
				}
				return
			}
			h.gatewayService.RecordCodexRecoveryAccountSwitch(c, account, failoverErr, false)
		},
	)
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

func (h *CompatibleGatewayTextHandler) reportCompatibleGatewayTextSuccess(c *gin.Context, flow compatibleGatewayTextFlow, account *service.Account, result *service.OpenAIForwardResult) {
	if result != nil {
		if flow.route == compatibleGatewayTextRouteResponses && account.Type == service.AccountTypeOAuth {
			h.gatewayService.UpdateCodexUsageSnapshotFromHeaders(c.Request.Context(), account.ID, result.ResponseHeaders)
		}
		h.gatewayService.ReportOpenAIAccountScheduleResult(account.ID, true, result.FirstTokenMs)
	} else {
		h.gatewayService.ReportOpenAIAccountScheduleResult(account.ID, true, nil)
	}

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
