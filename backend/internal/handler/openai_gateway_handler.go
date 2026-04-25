package handler

import (
	"context"
	"strings"
	"time"

	"github.com/senran-N/sub2api/internal/config"
	"github.com/senran-N/sub2api/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// OpenAIGatewayHandler handles OpenAI API gateway requests
type OpenAIGatewayHandler struct {
	gatewayService          *service.OpenAIGatewayService
	billingCacheService     *service.BillingCacheService
	apiKeyService           *service.APIKeyService
	usageRecordWorkerPool   *service.UsageRecordWorkerPool
	errorPassthroughService *service.ErrorPassthroughService
	concurrencyHelper       *ConcurrencyHelper
	maxAccountSwitches      int
	cfg                     *config.Config
}

func resolveOpenAIForwardDefaultMappedModel(apiKey *service.APIKey, fallbackModel string) string {
	return service.ResolveOpenAIForwardDefaultMappedModel(apiKey, fallbackModel)
}

func resolveOpenAISelectionFallbackModel(
	c *gin.Context,
	gatewayService *service.OpenAIGatewayService,
	apiKey *service.APIKey,
	schedulingModel string,
	reqLog *zap.Logger,
	logEvent string,
) string {
	candidate := service.OpenAISelectionFallbackCandidate(apiKey, schedulingModel)
	if candidate == "" {
		return ""
	}

	ctx := context.Background()
	if c != nil && c.Request != nil {
		ctx = c.Request.Context()
	}
	if fallbackModel, ok := gatewayService.ResolveOpenAISelectionFallbackModel(ctx, apiKey, schedulingModel); ok {
		return fallbackModel
	}
	if reqLog != nil {
		reqLog.Info(logEvent,
			zap.String("default_mapped_model", candidate),
			zap.String("requested_model", schedulingModel),
			zap.String("reason", "model_fallback_disabled"),
		)
	}
	return ""
}

func compatibleGatewayMessagesDispatchPlatform(ctx context.Context, apiKey *service.APIKey) string {
	fallbackPlatform := service.PlatformOpenAI
	if apiKey != nil && apiKey.Group != nil {
		fallbackPlatform = apiKey.Group.Platform
	}
	return service.ResolveCompatibleGatewayPlatform(ctx, fallbackPlatform)
}

func compatibleGatewayUsesOpenAIMessagesDispatch(ctx context.Context, apiKey *service.APIKey) bool {
	return compatibleGatewayMessagesDispatchPlatform(ctx, apiKey) == service.PlatformOpenAI
}

func resolveOpenAIMessagesDispatchMappedModel(ctx context.Context, apiKey *service.APIKey, requestedModel string) string {
	if !compatibleGatewayUsesOpenAIMessagesDispatch(ctx, apiKey) {
		return ""
	}
	if apiKey == nil || apiKey.Group == nil {
		return ""
	}
	return strings.TrimSpace(apiKey.Group.ResolveMessagesDispatchModel(requestedModel))
}

// NewOpenAIGatewayHandler creates a new OpenAIGatewayHandler
func NewOpenAIGatewayHandler(
	gatewayService *service.OpenAIGatewayService,
	concurrencyService *service.ConcurrencyService,
	billingCacheService *service.BillingCacheService,
	apiKeyService *service.APIKeyService,
	usageRecordWorkerPool *service.UsageRecordWorkerPool,
	errorPassthroughService *service.ErrorPassthroughService,
	cfg *config.Config,
) *OpenAIGatewayHandler {
	pingInterval := time.Duration(0)
	maxAccountSwitches := 3
	if cfg != nil {
		pingInterval = time.Duration(cfg.Concurrency.PingInterval) * time.Second
		if cfg.Gateway.MaxAccountSwitches > 0 {
			maxAccountSwitches = cfg.Gateway.MaxAccountSwitches
		}
	}
	return &OpenAIGatewayHandler{
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

func (h *OpenAIGatewayHandler) acquireResponsesUserSlot(
	c *gin.Context,
	userID int64,
	userConcurrency int,
	reqStream bool,
	streamStarted *bool,
	reqLog *zap.Logger,
) (func(), bool) {
	return compatibleTextHandlerFromOpenAIHandler(h).acquireResponsesUserSlot(c, userID, userConcurrency, reqStream, streamStarted, reqLog)
}

func (h *OpenAIGatewayHandler) acquireResponsesAccountSlot(
	c *gin.Context,
	groupID *int64,
	sessionHash string,
	selection *service.AccountSelectionResult,
	reqStream bool,
	streamStarted *bool,
	reqLog *zap.Logger,
) (func(), bool) {
	return compatibleTextHandlerFromOpenAIHandler(h).acquireResponsesAccountSlot(c, groupID, sessionHash, selection, reqStream, streamStarted, reqLog)
}

func (h *OpenAIGatewayHandler) ensureResponsesDependencies(c *gin.Context, reqLog *zap.Logger) bool {
	return compatibleTextHandlerFromOpenAIHandler(h).ensureResponsesDependencies(c, reqLog)
}

func (h *OpenAIGatewayHandler) missingResponsesDependencies() []string {
	return compatibleTextHandlerFromOpenAIHandler(h).missingResponsesDependencies()
}

func getContextInt64(c *gin.Context, key string) (int64, bool) {
	if c == nil || key == "" {
		return 0, false
	}
	v, ok := c.Get(key)
	if !ok {
		return 0, false
	}
	switch t := v.(type) {
	case int64:
		return t, true
	case int:
		return int64(t), true
	case int32:
		return int64(t), true
	case float64:
		return int64(t), true
	default:
		return 0, false
	}
}

func (h *OpenAIGatewayHandler) submitUsageRecordTask(task service.UsageRecordTask) {
	compatibleTextHandlerFromOpenAIHandler(h).submitUsageRecordTask(task)
}

func (h *OpenAIGatewayHandler) submitUsageRecordTaskWithParent(parent context.Context, task service.UsageRecordTask) {
	compatibleTextHandlerFromOpenAIHandler(h).submitUsageRecordTaskWithParent(parent, task)
}

func (h *OpenAIGatewayHandler) handleFailoverExhausted(c *gin.Context, failoverErr *service.UpstreamFailoverError, streamStarted bool) {
	compatibleTextHandlerFromOpenAIHandler(h).handleFailoverExhausted(c, failoverErr, streamStarted)
}

// handleStreamingAwareError handles errors that may occur after streaming has started
func (h *OpenAIGatewayHandler) handleStreamingAwareError(c *gin.Context, status int, errType, message string, streamStarted bool) {
	compatibleTextHandlerFromOpenAIHandler(h).handleStreamingAwareError(c, status, errType, message, streamStarted)
}

// ensureForwardErrorResponse 在 Forward 返回错误但尚未写响应时补写统一错误响应。
func (h *OpenAIGatewayHandler) ensureForwardErrorResponse(c *gin.Context, streamStarted bool) bool {
	return compatibleTextHandlerFromOpenAIHandler(h).ensureForwardErrorResponse(c, streamStarted)
}

func shouldLogOpenAIForwardFailureAsWarn(c *gin.Context, wroteFallback bool) bool {
	if wroteFallback {
		return false
	}
	if c == nil || c.Writer == nil {
		return false
	}
	return c.Writer.Written()
}

// errorResponse returns OpenAI API format error response
func (h *OpenAIGatewayHandler) errorResponse(c *gin.Context, status int, errType, message string) {
	compatibleTextHandlerFromOpenAIHandler(h).errorResponse(c, status, errType, message)
}

func setOpenAIClientTransportHTTP(c *gin.Context) {
	service.SetOpenAIClientTransport(c, service.OpenAIClientTransportHTTP)
}

func setOpenAIClientTransportWS(c *gin.Context) {
	service.SetOpenAIClientTransport(c, service.OpenAIClientTransportWS)
}
