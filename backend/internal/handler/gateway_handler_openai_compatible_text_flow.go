package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/senran-N/sub2api/internal/pkg/ip"
	"github.com/senran-N/sub2api/internal/service"
	"go.uber.org/zap"
)

// anthropicCompatibleTextFlowRequest carries the route-specific pieces for
// OpenAI-compatible text protocols that are executed by the native Anthropic
// runtime.
type anthropicCompatibleTextFlowRequest struct {
	Protocol        service.GatewayProtocol
	LogPrefix       string
	Body            []byte
	RequestedModel  string
	Stream          bool
	APIKey          *service.APIKey
	Subscription    *service.UserSubscription
	ChannelMapping  service.ChannelMappingResult
	StreamStarted   *bool
	WriteError      func(status int, codeOrType, message string)
	HandleExhausted func(lastErr *service.UpstreamFailoverError, streamStarted bool)
}

func (h *GatewayHandler) runAnthropicCompatibleTextFlow(c *gin.Context, reqLog *zap.Logger, req anthropicCompatibleTextFlowRequest) {
	session := h.gatewayService.PrepareRuntimeSession(c.Request.Context(), service.RuntimeSessionPrepareRequest{
		Body:                 req.Body,
		ParseProtocol:        string(req.Protocol),
		Model:                req.RequestedModel,
		Stream:               req.Stream,
		ClientIP:             ip.GetClientIP(c),
		UserAgent:            c.GetHeader("User-Agent"),
		APIKeyID:             req.APIKey.ID,
		GroupID:              req.APIKey.GroupID,
		BridgeLegacyMetadata: h.metadataBridgeEnabled(),
	})
	c.Request = c.Request.WithContext(session.Context)
	parsedReq := session.Parsed
	sessionHash := session.SessionHash
	sessionBoundAccountID := session.BoundAccountID

	pipelineResult := service.NewRuntimePipeline().Execute(c.Request.Context(), service.RuntimePipelineRequest{
		Subject: service.RuntimeSubject{
			UserID:   req.APIKey.UserID,
			APIKeyID: req.APIKey.ID,
			GroupID:  req.APIKey.GroupID,
			Provider: service.PlatformAnthropic,
			Protocol: req.Protocol,
		},
		Session: service.RuntimeSessionState{
			SessionHash:    sessionHash,
			BoundAccountID: sessionBoundAccountID,
			HasBound:       sessionBoundAccountID > 0,
		},
		Selection: service.SelectionRequest{
			Provider:    service.PlatformAnthropic,
			Protocol:    req.Protocol,
			Model:       req.RequestedModel,
			GroupID:     req.APIKey.GroupID,
			SessionHash: sessionHash,
		},
		MaxSwitches:     h.maxAccountSwitches,
		TempUnscheduler: h.gatewayService,
		Select: func(ctx context.Context, selectionReq service.SelectionRequest) (*service.AccountSelectionResult, service.SelectionDecision, error) {
			return h.selectRuntimeAccount(ctx, selectionReq)
		},
		AcquireSlot: func(ctx context.Context, state *service.RuntimePipelineState) service.RuntimeAccountSlotResult {
			slot := h.acquireRuntimeAccountSlot(ctx, c, state.Selection, nil, "", req.Stream, req.StreamStarted, false)
			slot.ReleaseFunc = wrapReleaseOnDone(ctx, slot.ReleaseFunc)
			return slot
		},
		Admit: func(ctx context.Context, state *service.RuntimePipelineState) service.RuntimeAdmissionResult {
			account := state.Account
			if !h.tryReserveAccountRPMForForward(ctx, reqLog, account, sessionBoundAccountID, req.APIKey.GroupID, sessionHash) {
				return service.RuntimeAdmissionResult{
					Outcome: service.RuntimeAdmissionRPMDenied,
					Account: account,
					Cleanup: service.RuntimeAdmissionCleanupRequest{
						AccountRelease: state.AccountRelease,
					},
				}
			}

			billingModel := parsedReq.Model
			if req.ChannelMapping.Mapped {
				billingModel = req.ChannelMapping.MappedModel
			}
			windowCostReservation, windowCostAllowed := h.tryReserveAccountWindowCostForForward(
				ctx,
				reqLog,
				account,
				req.APIKey,
				parsedReq,
				billingModel,
				sessionBoundAccountID,
				req.APIKey.GroupID,
				sessionHash,
			)
			if !windowCostAllowed {
				return service.RuntimeAdmissionResult{
					Outcome: service.RuntimeAdmissionWindowCostDenied,
					Account: account,
					Cleanup: service.RuntimeAdmissionCleanupRequest{
						AccountRelease: state.AccountRelease,
					},
				}
			}
			return service.RuntimeAdmissionResult{
				Outcome:               service.RuntimeAdmissionSucceeded,
				Account:               account,
				WindowCostReservation: windowCostReservation,
			}
		},
		Forward: func(ctx context.Context, state *service.RuntimePipelineState) service.RuntimeForwardResult {
			account := state.Account
			forwardBody := req.Body
			if req.ChannelMapping.Mapped {
				forwardBody = h.gatewayService.ReplaceModelInBody(req.Body, req.ChannelMapping.MappedModel)
			}
			attempt := h.executeRuntimeForwardAttempt(ctx, reqLog, service.RuntimeForwardAttemptRequest{
				Account: account,
				Forward: func(ctx context.Context) (*service.ForwardResult, error) {
					return h.nativeGatewayRuntime().Forward(ctx, service.NativeGatewayForwardRequest{
						Provider:   service.PlatformAnthropic,
						Protocol:   req.Protocol,
						Account:    account,
						GinContext: c,
						Body:       forwardBody,
						Parsed:     parsedReq,
					})
				},
				WriterSize:            c.Writer.Size,
				AccountRelease:        state.AccountRelease,
				WindowCostReservation: state.WindowCostReservation,
			})
			return service.RuntimeForwardResult{
				Result:          attempt.Result,
				Err:             attempt.Err,
				ResponseStarted: attempt.ResponseStarted,
				Attempt:         attempt,
			}
		},
		Hooks: service.RuntimePipelineHooks{
			OnAccountSelected: func(_ context.Context, state *service.RuntimePipelineState) service.RuntimePipelineHookResult {
				setOpsSelectedAccount(c, state.Account.ID, state.Account.Platform)
				return service.RuntimePipelineHookResult{}
			},
		},
	})

	switch pipelineResult.Outcome {
	case service.RuntimePipelineSucceeded:
		h.recordAnthropicCompatibleTextUsage(c, reqLog, req, pipelineResult.Account, pipelineResult.ForwardResult)
	case service.RuntimePipelineSelectionInitialUnavailable:
		message := "No available accounts"
		if pipelineResult.SelectionFailure.Err != nil {
			message += ": " + pipelineResult.SelectionFailure.Err.Error()
		}
		req.WriteError(http.StatusServiceUnavailable, "api_error", message)
	case service.RuntimePipelineSelectionRetryCanceled, service.RuntimePipelineFailoverCanceled:
		return
	case service.RuntimePipelineSelectionExhausted:
		if pipelineResult.SelectionFailure.FailoverErr != nil {
			req.HandleExhausted(pipelineResult.SelectionFailure.FailoverErr, *req.StreamStarted)
			return
		}
		req.WriteError(http.StatusBadGateway, "server_error", "All available accounts exhausted")
	case service.RuntimePipelineAccountSlotQueueFull:
		reqLog.Info(req.LogPrefix+".account_wait_queue_full",
			zap.Int64("account_id", runtimeSlotAccountID(pipelineResult.Slot)),
			zap.Int("max_waiting", runtimeSlotMaxWaiting(pipelineResult.Slot)),
		)
		req.WriteError(http.StatusTooManyRequests, "rate_limit_error", "Too many pending requests, please retry later")
	case service.RuntimePipelineAccountSlotAcquireError:
		reqLog.Warn(req.LogPrefix+".account_slot_acquire_failed", zap.Int64("account_id", runtimeSlotAccountID(pipelineResult.Slot)), zap.Error(pipelineResult.Slot.Err))
		h.handleConcurrencyError(c, pipelineResult.Slot.Err, "account", *req.StreamStarted)
	case service.RuntimePipelineAccountSlotWaitAcquireError:
		reqLog.Warn(req.LogPrefix+".account_slot_acquire_failed_after_wait", zap.Int64("account_id", runtimeSlotAccountID(pipelineResult.Slot)), zap.Error(pipelineResult.Slot.Err))
		h.handleConcurrencyError(c, pipelineResult.Slot.Err, "account", *req.StreamStarted)
	case service.RuntimePipelineAccountSlotUnavailable, service.RuntimePipelineNoAvailableAccount:
		req.WriteError(http.StatusServiceUnavailable, "api_error", "No available accounts")
	case service.RuntimePipelineFailoverExhausted:
		req.HandleExhausted(pipelineResult.Failover.FailoverErr, *req.StreamStarted || pipelineResult.Failover.ResponseStarted)
	case service.RuntimePipelineForwardError, service.RuntimePipelineMisconfigured, service.RuntimePipelineForwardHookAborted:
		h.ensureForwardErrorResponse(c, *req.StreamStarted)
		fields := []zap.Field{zap.Error(pipelineResult.Err)}
		if pipelineResult.Account != nil {
			fields = append([]zap.Field{zap.Int64("account_id", pipelineResult.Account.ID)}, fields...)
		}
		reqLog.Error(req.LogPrefix+".forward_failed", fields...)
	default:
		h.ensureForwardErrorResponse(c, *req.StreamStarted)
		reqLog.Error(req.LogPrefix+".forward_failed", zap.Error(pipelineResult.Err))
	}
}

func (h *GatewayHandler) recordAnthropicCompatibleTextUsage(
	c *gin.Context,
	reqLog *zap.Logger,
	req anthropicCompatibleTextFlowRequest,
	account *service.Account,
	result *service.ForwardResult,
) {
	userAgent := c.GetHeader("User-Agent")
	clientIP := ip.GetClientIP(c)
	requestPayloadHash := service.HashUsageRequestPayload(req.Body)
	inboundEndpoint := GetInboundEndpoint(c)
	upstreamEndpoint := GetUpstreamEndpoint(c, account.Platform)

	h.submitUsageRecordTask(func(ctx context.Context) {
		if err := h.gatewayService.RecordUsage(ctx, &service.RecordUsageInput{
			Result:             result,
			APIKey:             req.APIKey,
			User:               req.APIKey.User,
			Account:            account,
			Subscription:       req.Subscription,
			InboundEndpoint:    inboundEndpoint,
			UpstreamEndpoint:   upstreamEndpoint,
			UserAgent:          userAgent,
			IPAddress:          clientIP,
			RequestPayloadHash: requestPayloadHash,
			APIKeyService:      h.apiKeyService,
			ChannelUsageFields: req.ChannelMapping.ToUsageFields(req.RequestedModel, result.UpstreamModel),
		}); err != nil {
			reqLog.Error(req.LogPrefix+".record_usage_failed",
				zap.Int64("account_id", account.ID),
				zap.Error(err),
			)
		}
	})
}
