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

	fs := NewFailoverState(h.maxAccountSwitches, false)
	for {
		selection, _, err := h.selectRuntimeAccount(c.Request.Context(), service.SelectionRequest{
			Provider:    service.PlatformAnthropic,
			Protocol:    req.Protocol,
			Model:       req.RequestedModel,
			GroupID:     req.APIKey.GroupID,
			SessionHash: sessionHash,
			ExcludedIDs: fs.FailedAccountIDs,
		})
		if err != nil {
			selectionFailure := fs.HandleSelectionError(c.Request.Context(), err)
			switch selectionFailure.Outcome {
			case service.RuntimeSelectionFailureInitialUnavailable:
				req.WriteError(http.StatusServiceUnavailable, "api_error", "No available accounts: "+selectionFailure.Err.Error())
				return
			case service.RuntimeSelectionFailureRetry:
				continue
			case service.RuntimeSelectionFailureCanceled:
				return
			default:
				if selectionFailure.FailoverErr != nil {
					req.HandleExhausted(selectionFailure.FailoverErr, *req.StreamStarted)
				} else {
					req.WriteError(http.StatusBadGateway, "server_error", "All available accounts exhausted")
				}
				return
			}
		}

		account := selection.Account
		setOpsSelectedAccount(c, account.ID, account.Platform)

		slot := h.acquireRuntimeAccountSlot(c.Request.Context(), c, selection, nil, "", req.Stream, req.StreamStarted, false)
		switch slot.Outcome {
		case service.RuntimeAccountSlotSucceeded:
		case service.RuntimeAccountSlotQueueFull:
			reqLog.Info(req.LogPrefix+".account_wait_queue_full",
				zap.Int64("account_id", account.ID),
				zap.Int("max_waiting", runtimeSlotMaxWaiting(slot)),
			)
			req.WriteError(http.StatusTooManyRequests, "rate_limit_error", "Too many pending requests, please retry later")
			return
		case service.RuntimeAccountSlotAcquireError:
			reqLog.Warn(req.LogPrefix+".account_slot_acquire_failed", zap.Int64("account_id", account.ID), zap.Error(slot.Err))
			h.handleConcurrencyError(c, slot.Err, "account", *req.StreamStarted)
			return
		case service.RuntimeAccountSlotWaitAcquireError:
			reqLog.Warn(req.LogPrefix+".account_slot_acquire_failed_after_wait", zap.Int64("account_id", account.ID), zap.Error(slot.Err))
			h.handleConcurrencyError(c, slot.Err, "account", *req.StreamStarted)
			return
		default:
			req.WriteError(http.StatusServiceUnavailable, "api_error", "No available accounts")
			return
		}

		accountReleaseFunc := wrapReleaseOnDone(c.Request.Context(), slot.ReleaseFunc)
		if !h.tryReserveAccountRPMForForward(c.Request.Context(), reqLog, account, sessionBoundAccountID, req.APIKey.GroupID, sessionHash) {
			service.CleanupRuntimeAdmissionDenied(service.RuntimeAdmissionCleanupRequest{
				Account:          account,
				FailedAccountIDs: fs.FailedAccountIDs,
				AccountRelease:   accountReleaseFunc,
			})
			continue
		}

		billingModel := parsedReq.Model
		if req.ChannelMapping.Mapped {
			billingModel = req.ChannelMapping.MappedModel
		}
		windowCostReservation, windowCostAllowed := h.tryReserveAccountWindowCostForForward(
			c.Request.Context(),
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
			service.CleanupRuntimeAdmissionDenied(service.RuntimeAdmissionCleanupRequest{
				Account:          account,
				FailedAccountIDs: fs.FailedAccountIDs,
				AccountRelease:   accountReleaseFunc,
			})
			continue
		}

		forwardBody := req.Body
		if req.ChannelMapping.Mapped {
			forwardBody = h.gatewayService.ReplaceModelInBody(req.Body, req.ChannelMapping.MappedModel)
		}
		attempt := h.executeRuntimeForwardAttempt(c.Request.Context(), reqLog, service.RuntimeForwardAttemptRequest{
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
			AccountRelease:        accountReleaseFunc,
			WindowCostReservation: windowCostReservation,
		})
		result, err := attempt.Result, attempt.Err

		if err != nil {
			if failover := fs.HandleForwardError(c.Request.Context(), h.gatewayService, account.ID, account.Platform, err, attempt.ResponseStarted); failover.Handled {
				switch failover.Action {
				case FailoverContinue:
					continue
				case FailoverExhausted:
					req.HandleExhausted(failover.FailoverErr, *req.StreamStarted || failover.ResponseStarted)
					return
				case FailoverCanceled:
					return
				}
			}
			h.ensureForwardErrorResponse(c, *req.StreamStarted)
			reqLog.Error(req.LogPrefix+".forward_failed",
				zap.Int64("account_id", account.ID),
				zap.Error(err),
			)
			return
		}

		h.recordAnthropicCompatibleTextUsage(c, reqLog, req, account, result)
		return
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
