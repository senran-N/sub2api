package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/senran-N/sub2api/internal/config"
	"github.com/senran-N/sub2api/internal/pkg/httputil"
	"github.com/senran-N/sub2api/internal/pkg/ip"
	"github.com/senran-N/sub2api/internal/pkg/logger"
	middleware2 "github.com/senran-N/sub2api/internal/server/middleware"
	"github.com/senran-N/sub2api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"
)

// CompatibleGatewayTextHandler owns the shared text-route orchestration for the
// OpenAI-compatible protocol family.
type CompatibleGatewayTextHandler struct {
	gatewayService          *service.OpenAIGatewayService
	billingCacheService     *service.BillingCacheService
	apiKeyService           *service.APIKeyService
	usageRecordWorkerPool   *service.UsageRecordWorkerPool
	errorPassthroughService *service.ErrorPassthroughService
	concurrencyHelper       *ConcurrencyHelper
	maxAccountSwitches      int
	cfg                     *config.Config
}

func NewCompatibleGatewayTextHandler(
	gatewayService *service.OpenAIGatewayService,
	concurrencyService *service.ConcurrencyService,
	billingCacheService *service.BillingCacheService,
	apiKeyService *service.APIKeyService,
	usageRecordWorkerPool *service.UsageRecordWorkerPool,
	errorPassthroughService *service.ErrorPassthroughService,
	cfg *config.Config,
) *CompatibleGatewayTextHandler {
	pingInterval := time.Duration(0)
	maxAccountSwitches := 3
	if cfg != nil {
		pingInterval = time.Duration(cfg.Concurrency.PingInterval) * time.Second
		if cfg.Gateway.MaxAccountSwitches > 0 {
			maxAccountSwitches = cfg.Gateway.MaxAccountSwitches
		}
	}
	return &CompatibleGatewayTextHandler{
		gatewayService:          gatewayService,
		billingCacheService:     billingCacheService,
		apiKeyService:           apiKeyService,
		usageRecordWorkerPool:   usageRecordWorkerPool,
		errorPassthroughService: errorPassthroughService,
		concurrencyHelper:       NewConcurrencyHelper(concurrencyService, SSEPingFormatComment, pingInterval),
		maxAccountSwitches:      maxAccountSwitches,
		cfg:                     cfg,
	}
}

func compatibleTextHandlerFromOpenAIHandler(h *OpenAIGatewayHandler) *CompatibleGatewayTextHandler {
	if h == nil {
		return nil
	}
	return &CompatibleGatewayTextHandler{
		gatewayService:          h.gatewayService,
		billingCacheService:     h.billingCacheService,
		apiKeyService:           h.apiKeyService,
		usageRecordWorkerPool:   h.usageRecordWorkerPool,
		errorPassthroughService: h.errorPassthroughService,
		concurrencyHelper:       h.concurrencyHelper,
		maxAccountSwitches:      h.maxAccountSwitches,
		cfg:                     h.cfg,
	}
}

// Responses handles OpenAI Responses API endpoint.
// POST /openai/v1/responses
func (h *CompatibleGatewayTextHandler) Responses(c *gin.Context) {
	streamStarted := false
	defer h.recoverResponsesPanic(c, &streamStarted)
	compactStartedAt := time.Now()
	defer h.logOpenAIRemoteCompactOutcome(c, compactStartedAt)
	setOpenAIClientTransportHTTP(c)

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
	reqLog := requestLogger(
		c,
		"handler.openai_gateway.responses",
		zap.Int64("user_id", subject.UserID),
		zap.Int64("api_key_id", apiKey.ID),
		zap.Any("group_id", apiKey.GroupID),
	)
	attachRequestAccountLoadCache(c)
	if !h.ensureResponsesDependencies(c, reqLog) {
		return
	}

	body, err := httputil.ReadRequestBodyWithPrealloc(c.Request)
	if err != nil {
		if maxErr, ok := extractMaxBytesError(err); ok {
			h.errorResponse(c, http.StatusRequestEntityTooLarge, "invalid_request_error", buildBodyTooLargeMessage(maxErr.Limit))
			return
		}
		h.errorResponse(c, http.StatusBadRequest, "invalid_request_error", "Failed to read request body")
		return
	}

	if len(body) == 0 {
		h.errorResponse(c, http.StatusBadRequest, "invalid_request_error", "Request body is empty")
		return
	}

	setOpsRequestContext(c, "", false, body)
	sessionHashBody := body
	if service.IsOpenAIResponsesCompactPathForTest(c) {
		if compactSeed := strings.TrimSpace(gjson.GetBytes(body, "prompt_cache_key").String()); compactSeed != "" {
			c.Set(service.OpenAICompactSessionSeedKeyForTest(), compactSeed)
		}
		normalizedCompactBody, normalizedCompact, compactErr := service.NormalizeOpenAICompactRequestBodyForTest(body)
		if compactErr != nil {
			h.errorResponse(c, http.StatusBadRequest, "invalid_request_error", "Failed to normalize compact request body")
			return
		}
		if normalizedCompact {
			body = normalizedCompactBody
		}
	}

	if !gjson.ValidBytes(body) {
		h.errorResponse(c, http.StatusBadRequest, "invalid_request_error", "Failed to parse request body")
		return
	}

	profile := service.GetCodexRequestProfile(c, body, h != nil && h.cfg != nil && h.cfg.Gateway.ForceCodexCLI)
	service.ObserveOpenAICodexRequestProfile(profile)
	reqMeta := service.GetOpenAIRequestMeta(c, body)
	if !reqMeta.ModelExists || reqMeta.ModelType != gjson.String || profile.Body.Model == "" {
		h.errorResponse(c, http.StatusBadRequest, "invalid_request_error", "model is required")
		return
	}
	reqModel := profile.Body.Model

	if reqMeta.StreamExists && reqMeta.StreamType != gjson.True && reqMeta.StreamType != gjson.False {
		h.errorResponse(c, http.StatusBadRequest, "invalid_request_error", "invalid stream field type")
		return
	}
	reqStream := reqMeta.Stream
	reqLog = reqLog.With(zap.String("model", reqModel), zap.Bool("stream", reqStream))
	channelMapping, restricted := h.gatewayService.ResolveChannelMappingAndRestrict(c.Request.Context(), apiKey.GroupID, reqModel)
	if restricted {
		h.handleStreamingAwareError(c, http.StatusBadRequest, "invalid_request_error", "Requested model is not allowed for this channel", streamStarted)
		return
	}
	channelUsage := channelMapping.ToUsageFields(reqModel, "")
	schedulingModel := reqModel
	if channelMapping.MappedModel != "" {
		schedulingModel = channelMapping.MappedModel
	}
	previousResponseID := profile.Body.PreviousResponseID
	if previousResponseID != "" {
		previousResponseIDKind := profile.Continuation.PreviousResponseIDKind
		reqLog = reqLog.With(
			zap.Bool("has_previous_response_id", true),
			zap.String("previous_response_id_kind", previousResponseIDKind),
			zap.Int("previous_response_id_len", len(previousResponseID)),
			zap.String("codex_wire_api", string(profile.WireAPI)),
			zap.Bool("codex_official_client", profile.OfficialClient),
		)
		if previousResponseIDKind == service.OpenAIPreviousResponseIDKindMessageID {
			reqLog.Warn("openai.request_validation_failed",
				zap.String("reason", "previous_response_id_looks_like_message_id"),
			)
			h.errorResponse(c, http.StatusBadRequest, "invalid_request_error", "previous_response_id must be a response.id (resp_*), not a message id")
			return
		}
	}

	setOpsRequestContext(c, reqModel, reqStream, body)
	setOpsEndpointContext(c, "", int16(service.RequestTypeFromLegacy(reqStream, false)))

	if !h.validateFunctionCallOutputRequest(c, body, reqLog) {
		return
	}

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
		reqLog.Info("openai.billing_eligibility_check_failed", zap.Error(err))
		status, code, message := billingErrorDetails(err)
		h.handleStreamingAwareError(c, status, code, message, streamStarted)
		return
	}

	sessionHash := h.gatewayService.GenerateSessionHash(c, sessionHashBody)
	service.AttachOpenAIResolvedSessionHash(c, sessionHash)

	maxAccountSwitches := h.maxAccountSwitches
	switchCount := 0
	failedAccountIDs := make(map[int64]struct{})
	sameAccountRetryCount := make(map[int64]int)
	var lastFailoverErr *service.UpstreamFailoverError
	codexSchedulingObserved := false
	schedulerCtx := service.WithOpenAICodexTransportPreference(c.Request.Context(), profile.NativeClient)

	for {
		c.Set("openai_responses_fallback_model", "")
		reqLog.Debug("openai.account_selecting", zap.Int("excluded_account_count", len(failedAccountIDs)))
		selection, scheduleDecision, err := h.gatewayService.SelectAccountWithScheduler(
			schedulerCtx,
			apiKey.GroupID,
			previousResponseID,
			sessionHash,
			schedulingModel,
			failedAccountIDs,
			service.OpenAIUpstreamTransportAny,
		)
		if err != nil {
			reqLog.Warn("openai.account_select_failed",
				zap.Error(err),
				zap.Int("excluded_account_count", len(failedAccountIDs)),
			)
			if len(failedAccountIDs) == 0 {
				initialSelectionErr := err
				defaultModel := resolveOpenAISelectionFallbackModel(
					c,
					h.gatewayService,
					apiKey,
					schedulingModel,
					reqLog,
					"openai.fallback_to_default_model_skipped",
				)
				if defaultModel != "" && defaultModel != schedulingModel {
					reqLog.Info("openai.fallback_to_default_model",
						zap.String("default_mapped_model", defaultModel),
					)
					selection, scheduleDecision, err = h.gatewayService.SelectAccountWithScheduler(
						schedulerCtx,
						apiKey.GroupID,
						previousResponseID,
						sessionHash,
						defaultModel,
						failedAccountIDs,
						service.OpenAIUpstreamTransportAny,
					)
					if err == nil && selection != nil {
						c.Set("openai_responses_fallback_model", defaultModel)
					}
				}
				if err != nil {
					status, code, message := openAISelectionErrorResponseAfterDefaultFallback(initialSelectionErr, err)
					h.handleStreamingAwareError(c, status, code, message, streamStarted)
					return
				}
			}
			if lastFailoverErr != nil {
				h.handleFailoverExhausted(c, lastFailoverErr, streamStarted)
			} else {
				h.handleFailoverExhaustedSimple(c, 502, streamStarted)
			}
			return
		}
		if selection == nil || selection.Account == nil {
			h.handleStreamingAwareError(c, http.StatusServiceUnavailable, "api_error", "No available accounts", streamStarted)
			return
		}
		if previousResponseID != "" && selection != nil && selection.Account != nil {
			reqLog.Debug("openai.account_selected_with_previous_response_id", zap.Int64("account_id", selection.Account.ID))
		}
		reqLog.Debug("openai.account_schedule_decision",
			zap.String("layer", scheduleDecision.Layer),
			zap.Bool("sticky_previous_hit", scheduleDecision.StickyPreviousHit),
			zap.Bool("sticky_session_hit", scheduleDecision.StickySessionHit),
			zap.Int("candidate_count", scheduleDecision.CandidateCount),
			zap.Int("top_k", scheduleDecision.TopK),
			zap.Int64("latency_ms", scheduleDecision.LatencyMs),
			zap.Float64("load_skew", scheduleDecision.LoadSkew),
		)
		if !codexSchedulingObserved {
			service.ObserveOpenAICodexSchedulingDecision(profile, scheduleDecision)
			codexSchedulingObserved = true
		}
		account := selection.Account
		sessionHash = ensureOpenAIPoolModeSessionHash(sessionHash, account)
		reqLog.Debug("openai.account_selected", zap.Int64("account_id", account.ID), zap.String("account_name", account.Name))
		setOpsSelectedAccount(c, account.ID, account.Platform)

		accountReleaseFunc, acquired := h.acquireResponsesAccountSlot(c, apiKey.GroupID, sessionHash, selection, reqStream, &streamStarted, reqLog)
		if !acquired {
			return
		}

		service.SetOpsLatencyMs(c, service.OpsRoutingLatencyMsKey, time.Since(routingStart).Milliseconds())
		forwardStart := time.Now()
		fallbackModel := c.GetString("openai_responses_fallback_model")
		if fallbackModel == "" && channelUsage.ChannelMappedModel != "" {
			fallbackModel = channelUsage.ChannelMappedModel
		}
		defaultMappedModel := resolveOpenAIForwardDefaultMappedModel(apiKey, fallbackModel)
		forwardModelHint := defaultMappedModel
		if forwardModelHint == "" {
			forwardModelHint = schedulingModel
		}
		result, err := h.gatewayService.Forward(c.Request.Context(), c, account, body, defaultMappedModel)
		forwardDurationMs := time.Since(forwardStart).Milliseconds()
		if accountReleaseFunc != nil {
			accountReleaseFunc()
		}
		upstreamLatencyMs, _ := getContextInt64(c, service.OpsUpstreamLatencyMsKey)
		responseLatencyMs := forwardDurationMs
		if upstreamLatencyMs > 0 && forwardDurationMs > upstreamLatencyMs {
			responseLatencyMs = forwardDurationMs - upstreamLatencyMs
		}
		service.SetOpsLatencyMs(c, service.OpsResponseLatencyMsKey, responseLatencyMs)
		if err == nil && result != nil && result.FirstTokenMs != nil {
			service.SetOpsLatencyMs(c, service.OpsTimeToFirstTokenMsKey, int64(*result.FirstTokenMs))
		}
		if err != nil {
			var failoverErr *service.UpstreamFailoverError
			if errors.As(err, &failoverErr) {
				h.gatewayService.ReportOpenAIAccountScheduleResult(account.ID, false, nil)
				h.persistCompatibleGatewayRuntimeFeedback(c.Request.Context(), compatibleGatewayRuntimeFeedbackInput{
					Account:        account,
					RequestedModel: reqModel,
					UpstreamModel:  forwardModelHint,
					StatusCode:     failoverErr.StatusCode,
					ProtocolFamily: service.CompatibleGatewayProtocolFamilyResponses,
					Err:            failoverErr,
				})
				lastFailoverErr = failoverErr
				codexFailoverDecision := service.CodexRecoveryDecision{}
				if profile.NativeClient {
					codexFailoverDecision = h.gatewayService.ResolveCodexFailoverRecovery(c, account, failoverErr, sessionHash != "" || previousResponseID != "")
				}
				decision := applyOpenAIPoolFailoverPolicy(
					account,
					failoverErr,
					codexFailoverDecision,
					sessionHash != "" || previousResponseID != "",
					sameAccountRetryCount,
					failedAccountIDs,
					&switchCount,
					maxAccountSwitches,
					func() {
						h.gatewayService.TempUnscheduleRetryableError(c.Request.Context(), account.ID, failoverErr)
					},
					func() {
						h.gatewayService.RecordOpenAIAccountSwitch()
						if profile.NativeClient && codexFailoverDecision.SwitchAccount {
							h.gatewayService.RecordCodexRecoveryAccountSwitch(c, account, failoverErr, true)
						}
					},
				)
				if decision.SameAccountRetry {
					reqLog.Warn("openai.pool_mode_same_account_retry",
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
					h.handleFailoverExhausted(c, failoverErr, streamStarted)
					return
				}
				reqLog.Warn("openai.upstream_failover_switching",
					zap.Int64("account_id", account.ID),
					zap.Int("upstream_status", failoverErr.StatusCode),
					zap.Int("switch_count", decision.SwitchCount),
					zap.Int("max_switches", maxAccountSwitches),
				)
				continue
			}
			h.gatewayService.ReportOpenAIAccountScheduleResult(account.ID, false, nil)
			wroteFallback := h.ensureForwardErrorResponse(c, streamStarted)
			h.persistCompatibleGatewayRuntimeFeedback(c.Request.Context(), compatibleGatewayRuntimeFeedbackInput{
				Account:        account,
				RequestedModel: reqModel,
				UpstreamModel:  forwardModelHint,
				StatusCode:     c.Writer.Status(),
				ProtocolFamily: service.CompatibleGatewayProtocolFamilyResponses,
				Err:            err,
			})
			fields := []zap.Field{
				zap.Int64("account_id", account.ID),
				zap.Bool("fallback_error_response_written", wroteFallback),
				zap.Error(err),
			}
			if shouldLogOpenAIForwardFailureAsWarn(c, wroteFallback) {
				reqLog.Warn("openai.forward_failed", fields...)
				return
			}
			reqLog.Error("openai.forward_failed", fields...)
			return
		}
		if result != nil {
			if account.Type == service.AccountTypeOAuth {
				h.gatewayService.UpdateCodexUsageSnapshotFromHeaders(c.Request.Context(), account.ID, result.ResponseHeaders)
			}
			h.gatewayService.ReportOpenAIAccountScheduleResult(account.ID, true, result.FirstTokenMs)
			h.persistCompatibleGatewayRuntimeFeedback(c.Request.Context(), compatibleGatewayRuntimeFeedbackInput{
				Account:        account,
				RequestedModel: reqModel,
				UpstreamModel:  forwardModelHint,
				Result:         result,
				StatusCode:     c.Writer.Status(),
				ProtocolFamily: service.CompatibleGatewayProtocolFamilyResponses,
			})
		} else {
			h.gatewayService.ReportOpenAIAccountScheduleResult(account.ID, true, nil)
			h.persistCompatibleGatewayRuntimeFeedback(c.Request.Context(), compatibleGatewayRuntimeFeedbackInput{
				Account:        account,
				RequestedModel: reqModel,
				UpstreamModel:  forwardModelHint,
				StatusCode:     c.Writer.Status(),
				ProtocolFamily: service.CompatibleGatewayProtocolFamilyResponses,
			})
		}

		userAgent := c.GetHeader("User-Agent")
		clientIP := ip.GetClientIP(c)
		requestPayloadHash := service.HashUsageRequestPayload(body)

		h.submitUsageRecordTaskWithParent(c.Request.Context(), func(ctx context.Context) {
			if err := h.gatewayService.RecordUsage(ctx, &service.OpenAIRecordUsageInput{
				Result:             result,
				APIKey:             apiKey,
				User:               apiKey.User,
				Account:            account,
				Subscription:       subscription,
				InboundEndpoint:    GetInboundEndpoint(c),
				UpstreamEndpoint:   GetUpstreamEndpoint(c, account.Platform),
				UserAgent:          userAgent,
				IPAddress:          clientIP,
				RequestPayloadHash: requestPayloadHash,
				APIKeyService:      h.apiKeyService,
			}); err != nil {
				logger.L().With(
					zap.String("component", "handler.openai_gateway.responses"),
					zap.Int64("user_id", subject.UserID),
					zap.Int64("api_key_id", apiKey.ID),
					zap.Any("group_id", apiKey.GroupID),
					zap.String("model", reqModel),
					zap.Int64("account_id", account.ID),
				).Error("openai.record_usage_failed", zap.Error(err))
			}
		})
		reqLog.Debug("openai.request_completed",
			zap.Int64("account_id", account.ID),
			zap.Int("switch_count", switchCount),
		)
		return
	}
}

// ChatCompletions handles OpenAI Chat Completions API requests.
// POST /v1/chat/completions
func (h *CompatibleGatewayTextHandler) ChatCompletions(c *gin.Context) {
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
	reqLog := requestLogger(
		c,
		"handler.openai_gateway.chat_completions",
		zap.Int64("user_id", subject.UserID),
		zap.Int64("api_key_id", apiKey.ID),
		zap.Any("group_id", apiKey.GroupID),
	)
	attachRequestAccountLoadCache(c)

	if !h.ensureResponsesDependencies(c, reqLog) {
		return
	}

	body, err := httputil.ReadRequestBodyWithPrealloc(c.Request)
	if err != nil {
		if maxErr, ok := extractMaxBytesError(err); ok {
			h.errorResponse(c, http.StatusRequestEntityTooLarge, "invalid_request_error", buildBodyTooLargeMessage(maxErr.Limit))
			return
		}
		h.errorResponse(c, http.StatusBadRequest, "invalid_request_error", "Failed to read request body")
		return
	}
	if len(body) == 0 {
		h.errorResponse(c, http.StatusBadRequest, "invalid_request_error", "Request body is empty")
		return
	}

	if !gjson.ValidBytes(body) {
		h.errorResponse(c, http.StatusBadRequest, "invalid_request_error", "Failed to parse request body")
		return
	}

	modelResult := gjson.GetBytes(body, "model")
	if !modelResult.Exists() || modelResult.Type != gjson.String || modelResult.String() == "" {
		h.errorResponse(c, http.StatusBadRequest, "invalid_request_error", "model is required")
		return
	}
	reqModel := modelResult.String()
	reqStream := gjson.GetBytes(body, "stream").Bool()

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
		reqLog.Info("openai_chat_completions.billing_eligibility_check_failed", zap.Error(err))
		status, code, message := billingErrorDetails(err)
		h.handleStreamingAwareError(c, status, code, message, streamStarted)
		return
	}

	sessionHash := h.gatewayService.GenerateSessionHash(c, body)
	promptCacheKey := h.gatewayService.ExtractSessionID(c, body)
	channelMapping, restricted := h.gatewayService.ResolveChannelMappingAndRestrict(c.Request.Context(), apiKey.GroupID, reqModel)
	if restricted {
		h.handleStreamingAwareError(c, http.StatusBadRequest, "invalid_request_error", "Requested model is not allowed for this channel", streamStarted)
		return
	}
	channelUsage := channelMapping.ToUsageFields(reqModel, "")
	schedulingModel := reqModel
	if channelMapping.MappedModel != "" {
		schedulingModel = channelMapping.MappedModel
	}

	maxAccountSwitches := h.maxAccountSwitches
	switchCount := 0
	failedAccountIDs := make(map[int64]struct{})
	sameAccountRetryCount := make(map[int64]int)
	var lastFailoverErr *service.UpstreamFailoverError

	for {
		c.Set("openai_chat_completions_fallback_model", "")
		reqLog.Debug("openai_chat_completions.account_selecting", zap.Int("excluded_account_count", len(failedAccountIDs)))
		selection, scheduleDecision, err := h.gatewayService.SelectAccountWithScheduler(
			c.Request.Context(),
			apiKey.GroupID,
			"",
			sessionHash,
			schedulingModel,
			failedAccountIDs,
			service.OpenAIUpstreamTransportAny,
		)
		if err != nil {
			reqLog.Warn("openai_chat_completions.account_select_failed",
				zap.Error(err),
				zap.Int("excluded_account_count", len(failedAccountIDs)),
			)
			if len(failedAccountIDs) == 0 {
				initialSelectionErr := err
				defaultModel := resolveOpenAISelectionFallbackModel(
					c,
					h.gatewayService,
					apiKey,
					schedulingModel,
					reqLog,
					"openai_chat_completions.fallback_to_default_model_skipped",
				)
				if defaultModel != "" && defaultModel != schedulingModel {
					reqLog.Info("openai_chat_completions.fallback_to_default_model",
						zap.String("default_mapped_model", defaultModel),
					)
					selection, scheduleDecision, err = h.gatewayService.SelectAccountWithScheduler(
						c.Request.Context(),
						apiKey.GroupID,
						"",
						sessionHash,
						defaultModel,
						failedAccountIDs,
						service.OpenAIUpstreamTransportAny,
					)
					if err == nil && selection != nil {
						c.Set("openai_chat_completions_fallback_model", defaultModel)
					}
				}
				if err != nil {
					status, code, message := openAISelectionErrorResponseAfterDefaultFallback(initialSelectionErr, err)
					h.handleStreamingAwareError(c, status, code, message, streamStarted)
					return
				}
			} else {
				if lastFailoverErr != nil {
					h.handleFailoverExhausted(c, lastFailoverErr, streamStarted)
				} else {
					h.handleStreamingAwareError(c, http.StatusBadGateway, "api_error", "Upstream request failed", streamStarted)
				}
				return
			}
		}
		if selection == nil || selection.Account == nil {
			h.handleStreamingAwareError(c, http.StatusServiceUnavailable, "api_error", "No available accounts", streamStarted)
			return
		}
		account := selection.Account
		sessionHash = ensureOpenAIPoolModeSessionHash(sessionHash, account)
		reqLog.Debug("openai_chat_completions.account_selected", zap.Int64("account_id", account.ID), zap.String("account_name", account.Name))
		_ = scheduleDecision
		setOpsSelectedAccount(c, account.ID, account.Platform)

		accountReleaseFunc, acquired := h.acquireResponsesAccountSlot(c, apiKey.GroupID, sessionHash, selection, reqStream, &streamStarted, reqLog)
		if !acquired {
			return
		}

		service.SetOpsLatencyMs(c, service.OpsRoutingLatencyMsKey, time.Since(routingStart).Milliseconds())
		forwardStart := time.Now()

		fallbackModel := c.GetString("openai_chat_completions_fallback_model")
		if fallbackModel == "" && channelUsage.ChannelMappedModel != "" {
			fallbackModel = channelUsage.ChannelMappedModel
		}
		defaultMappedModel := resolveOpenAIForwardDefaultMappedModel(apiKey, fallbackModel)
		forwardModelHint := defaultMappedModel
		if forwardModelHint == "" {
			forwardModelHint = schedulingModel
		}
		result, err := h.gatewayService.ForwardAsChatCompletions(c.Request.Context(), c, account, body, promptCacheKey, defaultMappedModel)

		forwardDurationMs := time.Since(forwardStart).Milliseconds()
		if accountReleaseFunc != nil {
			accountReleaseFunc()
		}
		upstreamLatencyMs, _ := getContextInt64(c, service.OpsUpstreamLatencyMsKey)
		responseLatencyMs := forwardDurationMs
		if upstreamLatencyMs > 0 && forwardDurationMs > upstreamLatencyMs {
			responseLatencyMs = forwardDurationMs - upstreamLatencyMs
		}
		service.SetOpsLatencyMs(c, service.OpsResponseLatencyMsKey, responseLatencyMs)
		if err == nil && result != nil && result.FirstTokenMs != nil {
			service.SetOpsLatencyMs(c, service.OpsTimeToFirstTokenMsKey, int64(*result.FirstTokenMs))
		}
		if err != nil {
			var failoverErr *service.UpstreamFailoverError
			if errors.As(err, &failoverErr) {
				h.gatewayService.ReportOpenAIAccountScheduleResult(account.ID, false, nil)
				h.persistCompatibleGatewayRuntimeFeedback(c.Request.Context(), compatibleGatewayRuntimeFeedbackInput{
					Account:        account,
					RequestedModel: reqModel,
					UpstreamModel:  forwardModelHint,
					StatusCode:     failoverErr.StatusCode,
					ProtocolFamily: service.CompatibleGatewayProtocolFamilyChatCompletions,
					Err:            failoverErr,
				})
				lastFailoverErr = failoverErr
				decision := applyOpenAIPoolFailoverPolicy(
					account,
					failoverErr,
					service.CodexRecoveryDecision{},
					sessionHash != "",
					sameAccountRetryCount,
					failedAccountIDs,
					&switchCount,
					maxAccountSwitches,
					func() {
						h.gatewayService.TempUnscheduleRetryableError(c.Request.Context(), account.ID, failoverErr)
					},
					func() {
						h.gatewayService.RecordOpenAIAccountSwitch()
						h.gatewayService.RecordCodexRecoveryAccountSwitch(c, account, failoverErr, false)
					},
				)
				if decision.SameAccountRetry {
					reqLog.Warn("openai_chat_completions.pool_mode_same_account_retry",
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
					h.handleFailoverExhausted(c, failoverErr, streamStarted)
					return
				}
				reqLog.Warn("openai_chat_completions.upstream_failover_switching",
					zap.Int64("account_id", account.ID),
					zap.Int("upstream_status", failoverErr.StatusCode),
					zap.Int("switch_count", decision.SwitchCount),
					zap.Int("max_switches", maxAccountSwitches),
				)
				continue
			}
			h.gatewayService.ReportOpenAIAccountScheduleResult(account.ID, false, nil)
			wroteFallback := h.ensureForwardErrorResponse(c, streamStarted)
			h.persistCompatibleGatewayRuntimeFeedback(c.Request.Context(), compatibleGatewayRuntimeFeedbackInput{
				Account:        account,
				RequestedModel: reqModel,
				UpstreamModel:  forwardModelHint,
				StatusCode:     c.Writer.Status(),
				ProtocolFamily: service.CompatibleGatewayProtocolFamilyChatCompletions,
				Err:            err,
			})
			reqLog.Warn("openai_chat_completions.forward_failed",
				zap.Int64("account_id", account.ID),
				zap.Bool("fallback_error_response_written", wroteFallback),
				zap.Error(err),
			)
			return
		}
		if result != nil {
			h.gatewayService.ReportOpenAIAccountScheduleResult(account.ID, true, result.FirstTokenMs)
			h.persistCompatibleGatewayRuntimeFeedback(c.Request.Context(), compatibleGatewayRuntimeFeedbackInput{
				Account:        account,
				RequestedModel: reqModel,
				UpstreamModel:  forwardModelHint,
				Result:         result,
				StatusCode:     c.Writer.Status(),
				ProtocolFamily: service.CompatibleGatewayProtocolFamilyChatCompletions,
			})
		} else {
			h.gatewayService.ReportOpenAIAccountScheduleResult(account.ID, true, nil)
			h.persistCompatibleGatewayRuntimeFeedback(c.Request.Context(), compatibleGatewayRuntimeFeedbackInput{
				Account:        account,
				RequestedModel: reqModel,
				UpstreamModel:  forwardModelHint,
				StatusCode:     c.Writer.Status(),
				ProtocolFamily: service.CompatibleGatewayProtocolFamilyChatCompletions,
			})
		}

		userAgent := c.GetHeader("User-Agent")
		clientIP := ip.GetClientIP(c)

		h.submitUsageRecordTask(func(ctx context.Context) {
			if err := h.gatewayService.RecordUsage(ctx, &service.OpenAIRecordUsageInput{
				Result:             result,
				APIKey:             apiKey,
				User:               apiKey.User,
				Account:            account,
				Subscription:       subscription,
				InboundEndpoint:    GetInboundEndpoint(c),
				UpstreamEndpoint:   GetUpstreamEndpoint(c, account.Platform),
				UserAgent:          userAgent,
				IPAddress:          clientIP,
				APIKeyService:      h.apiKeyService,
				ChannelUsageFields: channelUsage,
			}); err != nil {
				logger.L().With(
					zap.String("component", "handler.openai_gateway.chat_completions"),
					zap.Int64("user_id", subject.UserID),
					zap.Int64("api_key_id", apiKey.ID),
					zap.Any("group_id", apiKey.GroupID),
					zap.String("model", reqModel),
					zap.Int64("account_id", account.ID),
				).Error("openai_chat_completions.record_usage_failed", zap.Error(err))
			}
		})
		reqLog.Debug("openai_chat_completions.request_completed",
			zap.Int64("account_id", account.ID),
			zap.Int("switch_count", switchCount),
		)
		return
	}
}

// Messages handles Anthropic Messages API requests routed to OpenAI platform.
// POST /v1/messages (when group platform is OpenAI)
func (h *CompatibleGatewayTextHandler) Messages(c *gin.Context) {
	streamStarted := false
	defer h.recoverAnthropicMessagesPanic(c, &streamStarted)

	requestStart := time.Now()

	apiKey, ok := middleware2.GetAPIKeyFromContext(c)
	if !ok {
		h.anthropicErrorResponse(c, http.StatusUnauthorized, "authentication_error", "Invalid API key")
		return
	}

	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		h.anthropicErrorResponse(c, http.StatusInternalServerError, "api_error", "User context not found")
		return
	}
	reqLog := requestLogger(
		c,
		"handler.openai_gateway.messages",
		zap.Int64("user_id", subject.UserID),
		zap.Int64("api_key_id", apiKey.ID),
		zap.Any("group_id", apiKey.GroupID),
	)
	attachRequestAccountLoadCache(c)

	if compatibleGatewayUsesOpenAIMessagesDispatch(c.Request.Context(), apiKey) && apiKey.Group != nil && !apiKey.Group.AllowMessagesDispatch {
		h.anthropicErrorResponse(c, http.StatusForbidden, "permission_error",
			"This group does not allow /v1/messages dispatch")
		return
	}

	if !h.ensureResponsesDependencies(c, reqLog) {
		return
	}

	body, err := httputil.ReadRequestBodyWithPrealloc(c.Request)
	if err != nil {
		if maxErr, ok := extractMaxBytesError(err); ok {
			h.anthropicErrorResponse(c, http.StatusRequestEntityTooLarge, "invalid_request_error", buildBodyTooLargeMessage(maxErr.Limit))
			return
		}
		h.anthropicErrorResponse(c, http.StatusBadRequest, "invalid_request_error", "Failed to read request body")
		return
	}
	if len(body) == 0 {
		h.anthropicErrorResponse(c, http.StatusBadRequest, "invalid_request_error", "Request body is empty")
		return
	}

	if !gjson.ValidBytes(body) {
		h.anthropicErrorResponse(c, http.StatusBadRequest, "invalid_request_error", "Failed to parse request body")
		return
	}

	reqMeta := service.GetOpenAIRequestMeta(c, body)
	if !reqMeta.ModelExists || reqMeta.ModelType != gjson.String || reqMeta.Model == "" {
		h.anthropicErrorResponse(c, http.StatusBadRequest, "invalid_request_error", "model is required")
		return
	}
	reqModel := reqMeta.Model
	routingModel := service.NormalizeOpenAICompatRequestedModel(reqModel)
	preferredMappedModel := resolveOpenAIMessagesDispatchMappedModel(c.Request.Context(), apiKey, reqModel)
	reqStream := reqMeta.Stream
	channelMapping, restricted := h.gatewayService.ResolveChannelMappingAndRestrict(c.Request.Context(), apiKey.GroupID, reqModel)
	if restricted {
		h.anthropicErrorResponse(c, http.StatusBadRequest, "invalid_request_error", "Requested model is not allowed for this channel")
		return
	}
	channelUsage := channelMapping.ToUsageFields(reqModel, "")
	if channelMapping.MappedModel != "" {
		routingModel = service.NormalizeOpenAICompatRequestedModel(channelMapping.MappedModel)
	}

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
		reqLog.Info("openai_messages.billing_eligibility_check_failed", zap.Error(err))
		status, code, message := billingErrorDetails(err)
		h.anthropicStreamingAwareError(c, status, code, message, streamStarted)
		return
	}

	sessionHash := h.gatewayService.GenerateSessionHash(c, body)
	promptCacheKey := h.gatewayService.ExtractSessionID(c, body)

	if sessionHash == "" || promptCacheKey == "" {
		if userID := strings.TrimSpace(gjson.GetBytes(body, "metadata.user_id").String()); userID != "" {
			seed := reqModel + "-" + userID
			if promptCacheKey == "" {
				promptCacheKey = service.GenerateSessionUUID(seed)
			}
			if sessionHash == "" {
				sessionHash = service.DeriveSessionHashFromSeed(seed)
			}
		}
	}

	maxAccountSwitches := h.maxAccountSwitches
	switchCount := 0
	failedAccountIDs := make(map[int64]struct{})
	sameAccountRetryCount := make(map[int64]int)
	var lastFailoverErr *service.UpstreamFailoverError

	for {
		effectiveMappedModel := strings.TrimSpace(preferredMappedModel)
		currentRoutingModel := routingModel
		if effectiveMappedModel != "" {
			currentRoutingModel = effectiveMappedModel
		}
		reqLog.Debug("openai_messages.account_selecting", zap.Int("excluded_account_count", len(failedAccountIDs)))
		selection, scheduleDecision, err := h.gatewayService.SelectAccountWithScheduler(
			c.Request.Context(),
			apiKey.GroupID,
			"",
			sessionHash,
			currentRoutingModel,
			failedAccountIDs,
			service.OpenAIUpstreamTransportAny,
		)
		if err != nil {
			reqLog.Warn("openai_messages.account_select_failed",
				zap.Error(err),
				zap.Int("excluded_account_count", len(failedAccountIDs)),
			)
			if len(failedAccountIDs) == 0 {
				h.anthropicStreamingAwareError(c, http.StatusServiceUnavailable, "api_error", "No available accounts", streamStarted)
				return
			} else {
				if lastFailoverErr != nil {
					h.handleAnthropicFailoverExhausted(c, lastFailoverErr, streamStarted)
				} else {
					h.anthropicStreamingAwareError(c, http.StatusBadGateway, "api_error", "Upstream request failed", streamStarted)
				}
				return
			}
		}
		if selection == nil || selection.Account == nil {
			h.anthropicStreamingAwareError(c, http.StatusServiceUnavailable, "api_error", "No available accounts", streamStarted)
			return
		}
		account := selection.Account
		sessionHash = ensureOpenAIPoolModeSessionHash(sessionHash, account)
		reqLog.Debug("openai_messages.account_selected", zap.Int64("account_id", account.ID), zap.String("account_name", account.Name))
		_ = scheduleDecision
		setOpsSelectedAccount(c, account.ID, account.Platform)

		accountReleaseFunc, acquired := h.acquireResponsesAccountSlot(c, apiKey.GroupID, sessionHash, selection, reqStream, &streamStarted, reqLog)
		if !acquired {
			return
		}

		service.SetOpsLatencyMs(c, service.OpsRoutingLatencyMsKey, time.Since(routingStart).Milliseconds())
		forwardStart := time.Now()

		defaultMappedModel := strings.TrimSpace(effectiveMappedModel)
		forwardModelHint := defaultMappedModel
		if forwardModelHint == "" {
			forwardModelHint = currentRoutingModel
		}
		forwardBody := body
		if channelMapping.Mapped {
			forwardBody = h.gatewayService.ReplaceModelInBody(body, channelMapping.MappedModel)
		}
		result, err := h.gatewayService.ForwardAsAnthropic(c.Request.Context(), c, account, forwardBody, promptCacheKey, defaultMappedModel)

		forwardDurationMs := time.Since(forwardStart).Milliseconds()
		if accountReleaseFunc != nil {
			accountReleaseFunc()
		}
		upstreamLatencyMs, _ := getContextInt64(c, service.OpsUpstreamLatencyMsKey)
		responseLatencyMs := forwardDurationMs
		if upstreamLatencyMs > 0 && forwardDurationMs > upstreamLatencyMs {
			responseLatencyMs = forwardDurationMs - upstreamLatencyMs
		}
		service.SetOpsLatencyMs(c, service.OpsResponseLatencyMsKey, responseLatencyMs)
		if err == nil && result != nil && result.FirstTokenMs != nil {
			service.SetOpsLatencyMs(c, service.OpsTimeToFirstTokenMsKey, int64(*result.FirstTokenMs))
		}
		if err != nil {
			var failoverErr *service.UpstreamFailoverError
			if errors.As(err, &failoverErr) {
				h.gatewayService.ReportOpenAIAccountScheduleResult(account.ID, false, nil)
				h.persistCompatibleGatewayRuntimeFeedback(c.Request.Context(), compatibleGatewayRuntimeFeedbackInput{
					Account:        account,
					RequestedModel: reqModel,
					UpstreamModel:  forwardModelHint,
					StatusCode:     failoverErr.StatusCode,
					ProtocolFamily: service.CompatibleGatewayProtocolFamilyMessages,
					Err:            failoverErr,
				})
				lastFailoverErr = failoverErr
				decision := applyOpenAIPoolFailoverPolicy(
					account,
					failoverErr,
					service.CodexRecoveryDecision{},
					sessionHash != "",
					sameAccountRetryCount,
					failedAccountIDs,
					&switchCount,
					maxAccountSwitches,
					func() {
						h.gatewayService.TempUnscheduleRetryableError(c.Request.Context(), account.ID, failoverErr)
					},
					func() {
						h.gatewayService.RecordOpenAIAccountSwitch()
						h.gatewayService.RecordCodexRecoveryAccountSwitch(c, account, failoverErr, false)
					},
				)
				if decision.SameAccountRetry {
					reqLog.Warn("openai_messages.pool_mode_same_account_retry",
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
					h.handleAnthropicFailoverExhausted(c, failoverErr, streamStarted)
					return
				}
				reqLog.Warn("openai_messages.upstream_failover_switching",
					zap.Int64("account_id", account.ID),
					zap.Int("upstream_status", failoverErr.StatusCode),
					zap.Int("switch_count", decision.SwitchCount),
					zap.Int("max_switches", maxAccountSwitches),
				)
				continue
			}
			h.gatewayService.ReportOpenAIAccountScheduleResult(account.ID, false, nil)
			wroteFallback := h.ensureAnthropicErrorResponse(c, streamStarted)
			h.persistCompatibleGatewayRuntimeFeedback(c.Request.Context(), compatibleGatewayRuntimeFeedbackInput{
				Account:        account,
				RequestedModel: reqModel,
				UpstreamModel:  forwardModelHint,
				StatusCode:     c.Writer.Status(),
				ProtocolFamily: service.CompatibleGatewayProtocolFamilyMessages,
				Err:            err,
			})
			reqLog.Warn("openai_messages.forward_failed",
				zap.Int64("account_id", account.ID),
				zap.Bool("fallback_error_response_written", wroteFallback),
				zap.Error(err),
			)
			return
		}
		if result != nil {
			h.gatewayService.ReportOpenAIAccountScheduleResult(account.ID, true, result.FirstTokenMs)
			h.persistCompatibleGatewayRuntimeFeedback(c.Request.Context(), compatibleGatewayRuntimeFeedbackInput{
				Account:        account,
				RequestedModel: reqModel,
				UpstreamModel:  forwardModelHint,
				Result:         result,
				StatusCode:     c.Writer.Status(),
				ProtocolFamily: service.CompatibleGatewayProtocolFamilyMessages,
			})
		} else {
			h.gatewayService.ReportOpenAIAccountScheduleResult(account.ID, true, nil)
			h.persistCompatibleGatewayRuntimeFeedback(c.Request.Context(), compatibleGatewayRuntimeFeedbackInput{
				Account:        account,
				RequestedModel: reqModel,
				UpstreamModel:  forwardModelHint,
				StatusCode:     c.Writer.Status(),
				ProtocolFamily: service.CompatibleGatewayProtocolFamilyMessages,
			})
		}

		userAgent := c.GetHeader("User-Agent")
		clientIP := ip.GetClientIP(c)
		requestPayloadHash := service.HashUsageRequestPayload(body)

		h.submitUsageRecordTaskWithParent(c.Request.Context(), func(ctx context.Context) {
			if err := h.gatewayService.RecordUsage(ctx, &service.OpenAIRecordUsageInput{
				Result:             result,
				APIKey:             apiKey,
				User:               apiKey.User,
				Account:            account,
				Subscription:       subscription,
				InboundEndpoint:    GetInboundEndpoint(c),
				UpstreamEndpoint:   GetUpstreamEndpoint(c, account.Platform),
				UserAgent:          userAgent,
				IPAddress:          clientIP,
				RequestPayloadHash: requestPayloadHash,
				APIKeyService:      h.apiKeyService,
				ChannelUsageFields: channelUsage,
			}); err != nil {
				logger.L().With(
					zap.String("component", "handler.openai_gateway.messages"),
					zap.Int64("user_id", subject.UserID),
					zap.Int64("api_key_id", apiKey.ID),
					zap.Any("group_id", apiKey.GroupID),
					zap.String("model", reqModel),
					zap.Int64("account_id", account.ID),
				).Error("openai_messages.record_usage_failed", zap.Error(err))
			}
		})
		reqLog.Debug("openai_messages.request_completed",
			zap.Int64("account_id", account.ID),
			zap.Int("switch_count", switchCount),
		)
		return
	}
}

func (h *CompatibleGatewayTextHandler) validateFunctionCallOutputRequest(c *gin.Context, body []byte, reqLog *zap.Logger) bool {
	if !gjson.GetBytes(body, `input.#(type=="function_call_output")`).Exists() {
		return true
	}

	var reqBody map[string]any
	if err := json.Unmarshal(body, &reqBody); err != nil {
		return true
	}

	c.Set(service.OpenAIParsedRequestBodyKey, reqBody)
	service.CacheOpenAIRequestMetaFromBodyMap(c, reqBody)
	validation := service.ValidateFunctionCallOutputContext(reqBody)
	if !validation.HasFunctionCallOutput {
		return true
	}

	previousResponseID, _ := reqBody["previous_response_id"].(string)
	if strings.TrimSpace(previousResponseID) != "" || validation.HasToolCallContext {
		return true
	}

	if validation.HasFunctionCallOutputMissingCallID {
		reqLog.Warn("openai.request_validation_failed",
			zap.String("reason", "function_call_output_missing_call_id"),
		)
		h.errorResponse(c, http.StatusBadRequest, "invalid_request_error", "function_call_output requires call_id or previous_response_id; if relying on history, ensure store=true and reuse previous_response_id")
		return false
	}
	if validation.HasItemReferenceForAllCallIDs {
		return true
	}

	reqLog.Warn("openai.request_validation_failed",
		zap.String("reason", "function_call_output_missing_item_reference"),
	)
	h.errorResponse(c, http.StatusBadRequest, "invalid_request_error", "function_call_output requires item_reference ids matching each call_id, or previous_response_id/tool_call context; if relying on history, ensure store=true and reuse previous_response_id")
	return false
}

func (h *CompatibleGatewayTextHandler) acquireResponsesUserSlot(
	c *gin.Context,
	userID int64,
	userConcurrency int,
	reqStream bool,
	streamStarted *bool,
	reqLog *zap.Logger,
) (func(), bool) {
	ctx := c.Request.Context()
	maxWait := service.CalculateMaxWait(userConcurrency)
	queueResult, err := h.concurrencyHelper.AcquireUserSlotOrQueue(ctx, userID, userConcurrency, maxWait)
	if err != nil {
		reqLog.Warn("openai.user_slot_acquire_failed", zap.Error(err))
		h.handleConcurrencyError(c, err, "user", *streamStarted)
		return nil, false
	}
	if queueResult.Acquired {
		return wrapReleaseOnDone(ctx, queueResult.ReleaseFunc), true
	}
	if !queueResult.QueueAllowed {
		reqLog.Info("openai.user_wait_queue_full", zap.Int("max_wait", maxWait))
		h.errorResponse(c, http.StatusTooManyRequests, "rate_limit_error", "Too many pending requests, please retry later")
		return nil, false
	}

	waitCounted := queueResult.WaitCounted
	defer func() {
		if waitCounted {
			h.concurrencyHelper.DecrementWaitCount(ctx, userID)
		}
	}()

	userReleaseFunc, err := h.concurrencyHelper.AcquireUserSlotAfterQueueing(c, userID, userConcurrency, reqStream, streamStarted)
	if err != nil {
		reqLog.Warn("openai.user_slot_acquire_failed_after_wait", zap.Error(err))
		h.handleConcurrencyError(c, err, "user", *streamStarted)
		return nil, false
	}

	if waitCounted {
		h.concurrencyHelper.DecrementWaitCount(ctx, userID)
		waitCounted = false
	}
	return wrapReleaseOnDone(ctx, userReleaseFunc), true
}

func (h *CompatibleGatewayTextHandler) acquireResponsesAccountSlot(
	c *gin.Context,
	groupID *int64,
	sessionHash string,
	selection *service.AccountSelectionResult,
	reqStream bool,
	streamStarted *bool,
	reqLog *zap.Logger,
) (func(), bool) {
	if selection == nil || selection.Account == nil {
		h.handleStreamingAwareError(c, http.StatusServiceUnavailable, "api_error", "No available accounts", *streamStarted)
		return nil, false
	}

	ctx := c.Request.Context()
	account := selection.Account
	if selection.Acquired {
		return wrapReleaseOnDone(ctx, selection.ReleaseFunc), true
	}
	if selection.WaitPlan == nil {
		h.handleStreamingAwareError(c, http.StatusServiceUnavailable, "api_error", "No available accounts", *streamStarted)
		return nil, false
	}

	queueResult, err := h.concurrencyHelper.AcquireAccountSlotOrQueue(
		ctx,
		account.ID,
		selection.WaitPlan.MaxConcurrency,
		selection.WaitPlan.MaxWaiting,
	)
	if err != nil {
		reqLog.Warn("openai.account_slot_quick_acquire_failed", zap.Int64("account_id", account.ID), zap.Error(err))
		h.handleConcurrencyError(c, err, "account", *streamStarted)
		return nil, false
	}
	if queueResult.Acquired {
		if err := h.gatewayService.BindStickySession(ctx, groupID, sessionHash, account.ID); err != nil {
			reqLog.Warn("openai.bind_sticky_session_failed", zap.Int64("account_id", account.ID), zap.Error(err))
		}
		return wrapReleaseOnDone(ctx, queueResult.ReleaseFunc), true
	}
	if !queueResult.QueueAllowed {
		reqLog.Info("openai.account_wait_queue_full",
			zap.Int64("account_id", account.ID),
			zap.Int("max_waiting", selection.WaitPlan.MaxWaiting),
		)
		h.handleStreamingAwareError(c, http.StatusTooManyRequests, "rate_limit_error", "Too many pending requests, please retry later", *streamStarted)
		return nil, false
	}

	accountWaitCounted := queueResult.WaitCounted
	releaseWait := func() {
		if accountWaitCounted {
			h.concurrencyHelper.DecrementAccountWaitCount(ctx, account.ID)
			accountWaitCounted = false
		}
	}
	defer releaseWait()

	accountReleaseFunc, err := h.concurrencyHelper.AcquireAccountSlotAfterQueueingWithWaitTimeout(
		c,
		account.ID,
		selection.WaitPlan.MaxConcurrency,
		selection.WaitPlan.Timeout,
		reqStream,
		streamStarted,
	)
	if err != nil {
		reqLog.Warn("openai.account_slot_acquire_failed", zap.Int64("account_id", account.ID), zap.Error(err))
		h.handleConcurrencyError(c, err, "account", *streamStarted)
		return nil, false
	}

	releaseWait()
	if err := h.gatewayService.BindStickySession(ctx, groupID, sessionHash, account.ID); err != nil {
		reqLog.Warn("openai.bind_sticky_session_failed", zap.Int64("account_id", account.ID), zap.Error(err))
	}
	return wrapReleaseOnDone(ctx, accountReleaseFunc), true
}

func (h *CompatibleGatewayTextHandler) ensureResponsesDependencies(c *gin.Context, reqLog *zap.Logger) bool {
	missing := h.missingResponsesDependencies()
	if len(missing) == 0 {
		return true
	}

	if reqLog == nil {
		reqLog = requestLogger(c, "handler.openai_gateway.responses")
	}
	reqLog.Error("openai.handler_dependencies_missing", zap.Strings("missing_dependencies", missing))

	if c != nil && c.Writer != nil && !c.Writer.Written() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": gin.H{
				"type":    "api_error",
				"message": "Service temporarily unavailable",
			},
		})
	}
	return false
}

func (h *CompatibleGatewayTextHandler) missingResponsesDependencies() []string {
	missing := make([]string, 0, 5)
	if h == nil {
		return append(missing, "handler")
	}
	if h.gatewayService == nil {
		missing = append(missing, "gatewayService")
	}
	if h.billingCacheService == nil {
		missing = append(missing, "billingCacheService")
	}
	if h.apiKeyService == nil {
		missing = append(missing, "apiKeyService")
	}
	if h.concurrencyHelper == nil || h.concurrencyHelper.concurrencyService == nil {
		missing = append(missing, "concurrencyHelper")
	}
	return missing
}

func (h *CompatibleGatewayTextHandler) submitUsageRecordTask(task service.UsageRecordTask) {
	h.submitUsageRecordTaskWithParent(context.TODO(), task)
}

func (h *CompatibleGatewayTextHandler) submitUsageRecordTaskWithParent(parent context.Context, task service.UsageRecordTask) {
	if task == nil {
		return
	}
	if h.usageRecordWorkerPool != nil {
		h.usageRecordWorkerPool.Submit(task)
		return
	}
	ctx, cancel := newDetachedTimeoutContext(parent, usageRecordFallbackTaskTimeout)
	defer cancel()
	defer func() {
		if recovered := recover(); recovered != nil {
			logger.L().With(
				zap.String("component", "handler.openai_gateway.responses"),
				zap.Any("panic", recovered),
			).Error("openai.usage_record_task_panic_recovered")
		}
	}()
	task(ctx)
}

func (h *CompatibleGatewayTextHandler) handleConcurrencyError(c *gin.Context, err error, slotType string, streamStarted bool) {
	h.handleStreamingAwareError(c, http.StatusTooManyRequests, "rate_limit_error",
		fmt.Sprintf("Concurrency limit exceeded for %s, please retry later", slotType), streamStarted)
}

func (h *CompatibleGatewayTextHandler) handleFailoverExhausted(c *gin.Context, failoverErr *service.UpstreamFailoverError, streamStarted bool) {
	statusCode := failoverErr.StatusCode
	responseBody := failoverErr.ResponseBody

	if h.errorPassthroughService != nil && len(responseBody) > 0 {
		if rule := h.errorPassthroughService.MatchRule("openai", statusCode, responseBody); rule != nil {
			respCode := statusCode
			if !rule.PassthroughCode && rule.ResponseCode != nil {
				respCode = *rule.ResponseCode
			}

			msg := service.ExtractUpstreamErrorMessage(responseBody)
			if !rule.PassthroughBody && rule.CustomMessage != nil {
				msg = *rule.CustomMessage
			}

			if rule.SkipMonitoring {
				c.Set(service.OpsSkipPassthroughKey, true)
			}

			h.handleStreamingAwareError(c, respCode, "upstream_error", msg, streamStarted)
			return
		}
	}

	upstreamMsg := service.ExtractUpstreamErrorMessage(responseBody)
	service.SetOpsUpstreamError(c, statusCode, upstreamMsg, "")

	status, errType, errMsg := h.mapUpstreamError(statusCode)
	h.handleStreamingAwareError(c, status, errType, errMsg, streamStarted)
}

func (h *CompatibleGatewayTextHandler) handleFailoverExhaustedSimple(c *gin.Context, statusCode int, streamStarted bool) {
	status, errType, errMsg := h.mapUpstreamError(statusCode)
	service.SetOpsUpstreamError(c, statusCode, errMsg, "")
	h.handleStreamingAwareError(c, status, errType, errMsg, streamStarted)
}

func (h *CompatibleGatewayTextHandler) mapUpstreamError(statusCode int) (int, string, string) {
	switch statusCode {
	case 401:
		return http.StatusBadGateway, "upstream_error", "Upstream authentication failed, please contact administrator"
	case 403:
		return http.StatusBadGateway, "upstream_error", "Upstream access forbidden, please contact administrator"
	case 429:
		return http.StatusTooManyRequests, "rate_limit_error", "Upstream rate limit exceeded, please retry later"
	case 529:
		return http.StatusServiceUnavailable, "upstream_error", "Upstream service overloaded, please retry later"
	case 500, 502, 503, 504:
		return http.StatusBadGateway, "upstream_error", "Upstream service temporarily unavailable"
	default:
		return http.StatusBadGateway, "upstream_error", "Upstream request failed"
	}
}

func (h *CompatibleGatewayTextHandler) handleStreamingAwareError(c *gin.Context, status int, errType, message string, streamStarted bool) {
	if streamStarted {
		flusher, ok := c.Writer.(http.Flusher)
		if ok {
			errorEvent := "event: error\ndata: " + `{"error":{"type":` + strconv.Quote(errType) + `,"message":` + strconv.Quote(message) + `}}` + "\n\n"
			if _, err := fmt.Fprint(c.Writer, errorEvent); err != nil {
				if ginErr := c.Error(err); ginErr != nil {
					_ = ginErr.SetType(gin.ErrorTypePrivate)
				}
			}
			flusher.Flush()
		}
		return
	}

	h.errorResponse(c, status, errType, message)
}

func (h *CompatibleGatewayTextHandler) ensureForwardErrorResponse(c *gin.Context, streamStarted bool) bool {
	if c == nil || c.Writer == nil || c.Writer.Written() {
		return false
	}
	h.handleStreamingAwareError(c, http.StatusBadGateway, "upstream_error", "Upstream request failed", streamStarted)
	return true
}

func (h *CompatibleGatewayTextHandler) errorResponse(c *gin.Context, status int, errType, message string) {
	c.JSON(status, gin.H{
		"error": gin.H{
			"type":    errType,
			"message": message,
		},
	})
}

func (h *CompatibleGatewayTextHandler) anthropicErrorResponse(c *gin.Context, status int, errType, message string) {
	c.JSON(status, gin.H{
		"type": "error",
		"error": gin.H{
			"type":    errType,
			"message": message,
		},
	})
}

func (h *CompatibleGatewayTextHandler) anthropicStreamingAwareError(c *gin.Context, status int, errType, message string, streamStarted bool) {
	if streamStarted {
		flusher, ok := c.Writer.(http.Flusher)
		if ok {
			errPayload, err := json.Marshal(gin.H{
				"type": "error",
				"error": gin.H{
					"type":    errType,
					"message": message,
				},
			})
			if err != nil {
				errPayload = []byte(`{"type":"error","error":{"type":"internal_error","message":"failed to encode error payload"}}`)
			}
			if _, writeErr := fmt.Fprintf(c.Writer, "event: error\ndata: %s\n\n", errPayload); writeErr == nil {
				flusher.Flush()
			}
		}
		return
	}
	h.anthropicErrorResponse(c, status, errType, message)
}

func (h *CompatibleGatewayTextHandler) handleAnthropicFailoverExhausted(c *gin.Context, failoverErr *service.UpstreamFailoverError, streamStarted bool) {
	status, errType, errMsg := h.mapUpstreamError(failoverErr.StatusCode)
	h.anthropicStreamingAwareError(c, status, errType, errMsg, streamStarted)
}

func (h *CompatibleGatewayTextHandler) ensureAnthropicErrorResponse(c *gin.Context, streamStarted bool) bool {
	if c == nil || c.Writer == nil || c.Writer.Written() {
		return false
	}
	h.anthropicStreamingAwareError(c, http.StatusBadGateway, "api_error", "Upstream request failed", streamStarted)
	return true
}

func (h *CompatibleGatewayTextHandler) logOpenAIRemoteCompactOutcome(c *gin.Context, startedAt time.Time) {
	if !isOpenAIRemoteCompactPath(c) {
		return
	}

	var (
		ctx    = context.Background()
		path   string
		status int
	)
	if c != nil {
		if c.Request != nil {
			ctx = c.Request.Context()
			if c.Request.URL != nil {
				path = strings.TrimSpace(c.Request.URL.Path)
			}
		}
		if c.Writer != nil {
			status = c.Writer.Status()
		}
	}

	outcome := "failed"
	if status >= 200 && status < 300 {
		outcome = "succeeded"
	}
	latencyMs := time.Since(startedAt).Milliseconds()
	if latencyMs < 0 {
		latencyMs = 0
	}

	fields := []zap.Field{
		zap.String("component", "handler.openai_gateway.responses"),
		zap.Bool("remote_compact", true),
		zap.String("compact_outcome", outcome),
		zap.Int("status_code", status),
		zap.Int64("latency_ms", latencyMs),
		zap.String("path", path),
		zap.Bool("force_codex_cli", h != nil && h.cfg != nil && h.cfg.Gateway.ForceCodexCLI),
	}

	if c != nil {
		if userAgent := strings.TrimSpace(c.GetHeader("User-Agent")); userAgent != "" {
			fields = append(fields, zap.String("request_user_agent", userAgent))
		}
		if v, ok := c.Get(opsModelKey); ok {
			if model, ok := v.(string); ok && strings.TrimSpace(model) != "" {
				fields = append(fields, zap.String("request_model", strings.TrimSpace(model)))
			}
		}
		if v, ok := c.Get(opsAccountIDKey); ok {
			if accountID, ok := v.(int64); ok && accountID > 0 {
				fields = append(fields, zap.Int64("account_id", accountID))
			}
		}
		if c.Writer != nil {
			if upstreamRequestID := strings.TrimSpace(c.Writer.Header().Get("x-request-id")); upstreamRequestID != "" {
				fields = append(fields, zap.String("upstream_request_id", upstreamRequestID))
			} else if upstreamRequestID := strings.TrimSpace(c.Writer.Header().Get("X-Request-Id")); upstreamRequestID != "" {
				fields = append(fields, zap.String("upstream_request_id", upstreamRequestID))
			}
		}
	}

	log := logger.FromContext(ctx).With(fields...)
	if outcome == "succeeded" {
		log.Info("codex.remote_compact.succeeded")
		return
	}
	log.Warn("codex.remote_compact.failed")
}

func (h *CompatibleGatewayTextHandler) recoverResponsesPanic(c *gin.Context, streamStarted *bool) {
	recovered := recover()
	if recovered == nil {
		return
	}

	h.handleRecoveredResponsesPanic(c, streamStarted, recovered)
}

func (h *CompatibleGatewayTextHandler) handleRecoveredResponsesPanic(c *gin.Context, streamStarted *bool, recovered any) {
	if recovered == nil {
		return
	}

	started := false
	if streamStarted != nil {
		started = *streamStarted
	}
	wroteFallback := h.ensureForwardErrorResponse(c, started)
	requestLogger(c, "handler.openai_gateway.responses").Error(
		"openai.responses_panic_recovered",
		zap.Bool("fallback_error_response_written", wroteFallback),
		zap.Any("panic", recovered),
		zap.ByteString("stack", debug.Stack()),
	)
}

func (h *CompatibleGatewayTextHandler) recoverAnthropicMessagesPanic(c *gin.Context, streamStarted *bool) {
	recovered := recover()
	if recovered == nil {
		return
	}

	h.handleRecoveredAnthropicMessagesPanic(c, streamStarted, recovered)
}

func (h *CompatibleGatewayTextHandler) handleRecoveredAnthropicMessagesPanic(c *gin.Context, streamStarted *bool, recovered any) {
	if recovered == nil {
		return
	}

	started := streamStarted != nil && *streamStarted
	requestLogger(c, "handler.openai_gateway.messages").Error(
		"openai.messages_panic_recovered",
		zap.Bool("stream_started", started),
		zap.Any("panic", recovered),
		zap.ByteString("stack", debug.Stack()),
	)
	if !started {
		h.anthropicErrorResponse(c, http.StatusInternalServerError, "api_error", "Internal server error")
	}
}
