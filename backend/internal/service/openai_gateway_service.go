package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/senran-N/sub2api/internal/config"
	"github.com/senran-N/sub2api/internal/pkg/openai"
	"github.com/senran-N/sub2api/internal/util/responseheaders"
	"github.com/tidwall/sjson"
)

const (
	// ChatGPT internal API for OAuth accounts
	chatgptCodexURL = "https://chatgpt.com/backend-api/codex/responses"
	// OpenAI Platform API for API Key accounts (fallback)
	openaiPlatformAPIURL   = "https://api.openai.com/v1/responses"
	openaiStickySessionTTL = time.Hour // 粘性会话TTL
	codexCLIUserAgent      = "codex_cli_rs/0.104.0"
	// codex_cli_only 拒绝时单个请求头日志长度上限（字符）
	codexCLIOnlyHeaderValueMaxBytes = 256

	// OpenAIParsedRequestBodyKey 缓存 handler 侧已解析的请求体，避免重复解析。
	OpenAIParsedRequestBodyKey = "openai_parsed_request_body"
	// OpenAIParsedRequestMetaKey 缓存请求热路径的轻量元信息，避免 handler/service 重复扫描 JSON。
	OpenAIParsedRequestMetaKey = "openai_parsed_request_meta"
	// OpenAIParsedRequestMetaCacheKey 记录 meta 缓存是否与特定 body 绑定。
	OpenAIParsedRequestMetaCacheKey = "openai_parsed_request_meta_cache"
	// OpenAI WS Mode 失败后的重连次数上限（不含首次尝试）。
	// 与 Codex 客户端保持一致：失败后最多重连 5 次。
	openAIWSReconnectRetryLimit = 5
	// OpenAI WS Mode 重连退避默认值（可由配置覆盖）。
	openAIWSRetryBackoffInitialDefault = 120 * time.Millisecond
	openAIWSRetryBackoffMaxDefault     = 2 * time.Second
	openAIWSRetryJitterRatioDefault    = 0.2
	openAICompactSessionSeedKey        = "openai_compact_session_seed"
	codexCLIVersion                    = "0.104.0"
	// Codex 限额快照仅用于后台展示/诊断，不需要每个成功请求都立即落库。
	openAICodexSnapshotPersistMinInterval = 30 * time.Second
)

// OpenAI allowed headers whitelist (for non-passthrough).
var openaiAllowedHeaders = map[string]bool{
	"accept-language":       true,
	"content-type":          true,
	"conversation_id":       true,
	"user-agent":            true,
	"originator":            true,
	"session_id":            true,
	"x-codex-turn-state":    true,
	"x-codex-turn-metadata": true,
}

// OpenAI passthrough allowed headers whitelist.
// 透传模式下仅放行这些低风险请求头，避免将非标准/环境噪声头传给上游触发风控。
var openaiPassthroughAllowedHeaders = map[string]bool{
	"accept":                true,
	"accept-language":       true,
	"content-type":          true,
	"conversation_id":       true,
	"openai-beta":           true,
	"user-agent":            true,
	"originator":            true,
	"session_id":            true,
	"x-codex-turn-state":    true,
	"x-codex-turn-metadata": true,
}

// codex_cli_only 拒绝时记录的请求头白名单（仅用于诊断日志，不参与上游透传）
var codexCLIOnlyDebugHeaderWhitelist = []string{
	"User-Agent",
	"Content-Type",
	"Accept",
	"Accept-Language",
	"OpenAI-Beta",
	"Originator",
	"Session_ID",
	"Conversation_ID",
	"X-Request-ID",
	"X-Client-Request-ID",
	"X-Forwarded-For",
	"X-Real-IP",
}

// OpenAIUsage represents OpenAI API response usage
type OpenAIUsage struct {
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens,omitempty"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens,omitempty"`
	ImageOutputTokens        int `json:"image_output_tokens,omitempty"`
}

// OpenAIForwardResult represents the result of forwarding
type OpenAIForwardResult struct {
	RequestID string
	Usage     OpenAIUsage
	Model     string // 原始模型（用于响应和日志显示）
	// BillingModel is the model used for cost calculation.
	// When non-empty, CalculateCost uses this instead of Model.
	// This is set by the Anthropic Messages conversion path where
	// the mapped upstream model differs from the client-facing model.
	BillingModel string
	// UpstreamModel is the actual model sent to the upstream provider after mapping.
	// Empty when no mapping was applied (requested model was used as-is).
	UpstreamModel string
	// ServiceTier records the OpenAI Responses API service tier, e.g. "priority" / "flex".
	// Nil means the request did not specify a recognized tier.
	ServiceTier *string
	// ReasoningEffort is extracted from request body (reasoning.effort) or derived from model suffix.
	// Stored for usage records display; nil means not provided / not applicable.
	ReasoningEffort *string
	Stream          bool
	OpenAIWSMode    bool
	ResponseHeaders http.Header
	Duration        time.Duration
	FirstTokenMs    *int
}

type OpenAIWSRetryMetricsSnapshot struct {
	RetryAttemptsTotal            int64 `json:"retry_attempts_total"`
	RetryBackoffMsTotal           int64 `json:"retry_backoff_ms_total"`
	RetryExhaustedTotal           int64 `json:"retry_exhausted_total"`
	NonRetryableFastFallbackTotal int64 `json:"non_retryable_fast_fallback_total"`
}

type OpenAICompatibilityFallbackMetricsSnapshot struct {
	SessionHashLegacyReadFallbackTotal int64   `json:"session_hash_legacy_read_fallback_total"`
	SessionHashLegacyReadFallbackHit   int64   `json:"session_hash_legacy_read_fallback_hit"`
	SessionHashLegacyDualWriteTotal    int64   `json:"session_hash_legacy_dual_write_total"`
	SessionHashLegacyReadHitRate       float64 `json:"session_hash_legacy_read_hit_rate"`

	MetadataLegacyFallbackIsMaxTokensOneHaikuTotal int64 `json:"metadata_legacy_fallback_is_max_tokens_one_haiku_total"`
	MetadataLegacyFallbackThinkingEnabledTotal     int64 `json:"metadata_legacy_fallback_thinking_enabled_total"`
	MetadataLegacyFallbackPrefetchedStickyAccount  int64 `json:"metadata_legacy_fallback_prefetched_sticky_account_total"`
	MetadataLegacyFallbackPrefetchedStickyGroup    int64 `json:"metadata_legacy_fallback_prefetched_sticky_group_total"`
	MetadataLegacyFallbackSingleAccountRetryTotal  int64 `json:"metadata_legacy_fallback_single_account_retry_total"`
	MetadataLegacyFallbackAccountSwitchCountTotal  int64 `json:"metadata_legacy_fallback_account_switch_count_total"`
	MetadataLegacyFallbackTotal                    int64 `json:"metadata_legacy_fallback_total"`
}

type openAIWSRetryMetrics struct {
	retryAttempts            atomic.Int64
	retryBackoffMs           atomic.Int64
	retryExhausted           atomic.Int64
	nonRetryableFastFallback atomic.Int64
}

type accountWriteThrottle struct {
	minInterval time.Duration
	mu          sync.Mutex
	lastByID    map[int64]time.Time
}

func newAccountWriteThrottle(minInterval time.Duration) *accountWriteThrottle {
	return &accountWriteThrottle{
		minInterval: minInterval,
		lastByID:    make(map[int64]time.Time),
	}
}

func (t *accountWriteThrottle) Allow(id int64, now time.Time) bool {
	if t == nil || id <= 0 || t.minInterval <= 0 {
		return true
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	if last, ok := t.lastByID[id]; ok && now.Sub(last) < t.minInterval {
		return false
	}
	t.lastByID[id] = now

	if len(t.lastByID) > 4096 {
		cutoff := now.Add(-4 * t.minInterval)
		for accountID, writtenAt := range t.lastByID {
			if writtenAt.Before(cutoff) {
				delete(t.lastByID, accountID)
			}
		}
	}

	return true
}

var defaultOpenAICodexSnapshotPersistThrottle = newAccountWriteThrottle(openAICodexSnapshotPersistMinInterval)

// OpenAIGatewayService handles OpenAI API gateway operations
type OpenAIGatewayService struct {
	accountRepo           AccountRepository
	usageLogRepo          UsageLogRepository
	usageBillingRepo      UsageBillingRepository
	userRepo              UserRepository
	userSubRepo           UserSubscriptionRepository
	cache                 GatewayCache
	cfg                   *config.Config
	codexDetector         CodexClientRestrictionDetector
	schedulerSnapshot     *SchedulerSnapshotService
	concurrencyService    *ConcurrencyService
	billingService        *BillingService
	rateLimitService      *RateLimitService
	billingCacheService   *BillingCacheService
	userGroupRateResolver *userGroupRateResolver
	httpUpstream          HTTPUpstream
	deferredService       *DeferredService
	openAITokenProvider   *OpenAITokenProvider
	toolCorrector         *CodexToolCorrector
	openaiWSResolver      OpenAIWSProtocolResolver
	resolver              *ModelPricingResolver
	channelService        *ChannelService

	openaiWSPoolOnce              sync.Once
	openaiWSStateStoreOnce        sync.Once
	openaiSchedulerOnce           sync.Once
	openaiWSPassthroughDialerOnce sync.Once
	openaiWSPool                  *openAIWSConnPool
	openaiWSStateStore            OpenAIWSStateStore
	openaiScheduler               OpenAIAccountScheduler
	openaiWSPassthroughDialer     openAIWSClientDialer
	openaiAccountStats            *openAIAccountRuntimeStats

	openaiWSFallbackUntil sync.Map // key: int64(accountID), value: time.Time
	openaiWSRetryMetrics  openAIWSRetryMetrics
	responseHeaderFilter  *responseheaders.CompiledHeaderFilter
	codexSnapshotThrottle *accountWriteThrottle
}

// NewOpenAIGatewayService creates a new OpenAIGatewayService
func NewOpenAIGatewayService(
	accountRepo AccountRepository,
	usageLogRepo UsageLogRepository,
	usageBillingRepo UsageBillingRepository,
	userRepo UserRepository,
	userSubRepo UserSubscriptionRepository,
	userGroupRateRepo UserGroupRateRepository,
	cache GatewayCache,
	cfg *config.Config,
	schedulerSnapshot *SchedulerSnapshotService,
	concurrencyService *ConcurrencyService,
	billingService *BillingService,
	rateLimitService *RateLimitService,
	billingCacheService *BillingCacheService,
	httpUpstream HTTPUpstream,
	deferredService *DeferredService,
	openAITokenProvider *OpenAITokenProvider,
	resolver *ModelPricingResolver,
	channelService *ChannelService,
) *OpenAIGatewayService {
	svc := &OpenAIGatewayService{
		accountRepo:         accountRepo,
		usageLogRepo:        usageLogRepo,
		usageBillingRepo:    usageBillingRepo,
		userRepo:            userRepo,
		userSubRepo:         userSubRepo,
		cache:               cache,
		cfg:                 cfg,
		codexDetector:       NewOpenAICodexClientRestrictionDetector(cfg),
		schedulerSnapshot:   schedulerSnapshot,
		concurrencyService:  concurrencyService,
		billingService:      billingService,
		rateLimitService:    rateLimitService,
		billingCacheService: billingCacheService,
		userGroupRateResolver: newUserGroupRateResolver(
			userGroupRateRepo,
			nil,
			resolveUserGroupRateCacheTTL(cfg),
			nil,
			"service.openai_gateway",
		),
		httpUpstream:          httpUpstream,
		deferredService:       deferredService,
		openAITokenProvider:   openAITokenProvider,
		toolCorrector:         NewCodexToolCorrector(),
		openaiWSResolver:      NewOpenAIWSProtocolResolver(cfg),
		resolver:              resolver,
		channelService:        channelService,
		responseHeaderFilter:  compileResponseHeaderFilter(cfg),
		codexSnapshotThrottle: newAccountWriteThrottle(openAICodexSnapshotPersistMinInterval),
	}
	svc.logOpenAIWSModeBootstrap()
	return svc
}

func (s *OpenAIGatewayService) getCodexSnapshotThrottle() *accountWriteThrottle {
	if s != nil && s.codexSnapshotThrottle != nil {
		return s.codexSnapshotThrottle
	}
	return defaultOpenAICodexSnapshotPersistThrottle
}

func (s *OpenAIGatewayService) ResolveChannelMappingAndRestrict(ctx context.Context, groupID *int64, model string) (ChannelMappingResult, bool) {
	if s == nil || s.channelService == nil {
		return ChannelMappingResult{MappedModel: model}, false
	}
	return s.channelService.ResolveChannelMappingAndRestrict(ctx, groupID, model)
}

func (s *OpenAIGatewayService) billingDeps() *billingDeps {
	return &billingDeps{
		accountRepo:         s.accountRepo,
		userRepo:            s.userRepo,
		userSubRepo:         s.userSubRepo,
		billingCacheService: s.billingCacheService,
		deferredService:     s.deferredService,
	}
}

// CloseOpenAIWSPool 关闭 OpenAI WebSocket 连接池的后台 worker 和空闲连接。
// 应在应用优雅关闭时调用。
func (s *OpenAIGatewayService) CloseOpenAIWSPool() {
	if s != nil && s.openaiWSPool != nil {
		s.openaiWSPool.Close()
	}
}

func (s *OpenAIGatewayService) getCodexClientRestrictionDetector() CodexClientRestrictionDetector {
	if s != nil && s.codexDetector != nil {
		return s.codexDetector
	}
	var cfg *config.Config
	if s != nil {
		cfg = s.cfg
	}
	return NewOpenAICodexClientRestrictionDetector(cfg)
}

func SnapshotOpenAICompatibilityFallbackMetrics() OpenAICompatibilityFallbackMetricsSnapshot {
	legacyReadFallbackTotal, legacyReadFallbackHit, legacyDualWriteTotal := openAIStickyCompatStats()
	isMaxTokensOneHaiku, thinkingEnabled, prefetchedStickyAccount, prefetchedStickyGroup, singleAccountRetry, accountSwitchCount := RequestMetadataFallbackStats()

	readHitRate := float64(0)
	if legacyReadFallbackTotal > 0 {
		readHitRate = float64(legacyReadFallbackHit) / float64(legacyReadFallbackTotal)
	}
	metadataFallbackTotal := isMaxTokensOneHaiku + thinkingEnabled + prefetchedStickyAccount + prefetchedStickyGroup + singleAccountRetry + accountSwitchCount

	return OpenAICompatibilityFallbackMetricsSnapshot{
		SessionHashLegacyReadFallbackTotal: legacyReadFallbackTotal,
		SessionHashLegacyReadFallbackHit:   legacyReadFallbackHit,
		SessionHashLegacyDualWriteTotal:    legacyDualWriteTotal,
		SessionHashLegacyReadHitRate:       readHitRate,

		MetadataLegacyFallbackIsMaxTokensOneHaikuTotal: isMaxTokensOneHaiku,
		MetadataLegacyFallbackThinkingEnabledTotal:     thinkingEnabled,
		MetadataLegacyFallbackPrefetchedStickyAccount:  prefetchedStickyAccount,
		MetadataLegacyFallbackPrefetchedStickyGroup:    prefetchedStickyGroup,
		MetadataLegacyFallbackSingleAccountRetryTotal:  singleAccountRetry,
		MetadataLegacyFallbackAccountSwitchCountTotal:  accountSwitchCount,
		MetadataLegacyFallbackTotal:                    metadataFallbackTotal,
	}
}

// Forward forwards request to OpenAI API
func (s *OpenAIGatewayService) Forward(ctx context.Context, c *gin.Context, account *Account, body []byte, defaultMappedModel string) (*OpenAIForwardResult, error) {
	startTime := time.Now()

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
	reqMeta := getOpenAIRequestMeta(c, body)
	reqModel, reqStream, promptCacheKey := reqMeta.Model, reqMeta.Stream, reqMeta.PromptCacheKey

	isCodexCLI := openai.IsCodexOfficialClientByHeaders(c.GetHeader("User-Agent"), c.GetHeader("originator")) || (s.cfg != nil && s.cfg.Gateway.ForceCodexCLI)
	wsDecision := s.getOpenAIWSProtocolResolver().Resolve(account)
	clientTransport := GetOpenAIClientTransport(c)
	// 仅允许 WS 入站请求走 WS 上游，避免出现 HTTP -> WS 协议混用。
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
	// 当前仅支持 WSv2；WSv1 命中时直接返回错误，避免出现“配置可开但行为不确定”。
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
		if mappedModel != "" && mappedModel != reqModel {
			var patchErr error
			forwardBody, patchErr = sjson.SetBytes(originalBody, "model", mappedModel)
			if patchErr != nil {
				return nil, fmt.Errorf("patch passthrough model: %w", patchErr)
			}
		}
		// 透传分支只需要轻量提取字段，避免热路径全量 Unmarshal。
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
		result, err := s.forwardOpenAIPassthrough(ctx, c, account, forwardBody, reqModel, reasoningEffort, reqStream, startTime)
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

	// Get access token
	token, _, err := s.GetAccessToken(ctx, account)
	if err != nil {
		return nil, err
	}

	// Capture upstream request body for ops retry of this attempt.
	setOpsUpstreamRequestBody(c, prepared.body)

	// 命中 WS 时仅走 WebSocket Mode；不再自动回退 HTTP。
	if wsDecision.Transport == OpenAIUpstreamTransportResponsesWebsocketV2 {
		return s.forwardPreparedOpenAIWS(ctx, c, account, prepared, token, wsDecision, startTime, isCodexCLI)
	}

	return s.forwardPreparedOpenAIHTTP(ctx, c, account, prepared, token, startTime, isCodexCLI)
}

func IsOpenAIResponsesCompactPathForTest(c *gin.Context) bool {
	return isOpenAIResponsesCompactPath(c)
}

func OpenAICompactSessionSeedKeyForTest() string {
	return openAICompactSessionSeedKey
}

func NormalizeOpenAICompactRequestBodyForTest(body []byte) ([]byte, bool, error) {
	return normalizeOpenAICompactRequestBody(body)
}
