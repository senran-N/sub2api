package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	antigravityStickySessionTTL = time.Hour
	antigravityMaxRetries       = 3
	antigravityRetryBaseDelay   = 1 * time.Second
	antigravityRetryMaxDelay    = 16 * time.Second

	// 限流相关常量
	// antigravityRateLimitThreshold 限流等待/切换阈值
	// - 智能重试：retryDelay < 此阈值时等待后重试，>= 此阈值时直接限流模型
	// - 预检查：剩余限流时间 < 此阈值时等待，>= 此阈值时切换账号
	antigravityRateLimitThreshold       = 7 * time.Second
	antigravitySmartRetryMinWait        = 1 * time.Second  // 智能重试最小等待时间
	antigravitySmartRetryMaxAttempts    = 1                // 智能重试最大次数（仅重试 1 次，防止重复限流/长期等待）
	antigravityDefaultRateLimitDuration = 30 * time.Second // 默认限流时间（无 retryDelay 时使用）

	// MODEL_CAPACITY_EXHAUSTED 专用重试参数
	// 模型容量不足时，所有账号共享同一容量池，切换账号无意义
	// 使用固定 1s 间隔重试，最多重试 60 次
	antigravityModelCapacityRetryMaxAttempts = 60
	antigravityModelCapacityRetryWait        = 1 * time.Second

	// Google RPC 状态和类型常量
	googleRPCStatusResourceExhausted      = "RESOURCE_EXHAUSTED"
	googleRPCStatusUnavailable            = "UNAVAILABLE"
	googleRPCTypeRetryInfo                = "type.googleapis.com/google.rpc.RetryInfo"
	googleRPCTypeErrorInfo                = "type.googleapis.com/google.rpc.ErrorInfo"
	googleRPCReasonModelCapacityExhausted = "MODEL_CAPACITY_EXHAUSTED"
	googleRPCReasonRateLimitExceeded      = "RATE_LIMIT_EXCEEDED"

	// 单账号 503 退避重试：Service 层原地重试的最大次数
	// 在 handleSmartRetry 中，对于 shouldRateLimitModel（长延迟 ≥ 7s）的情况，
	// 多账号模式下会设限流+切换账号；但单账号模式下改为原地等待+重试。
	antigravitySingleAccountSmartRetryMaxAttempts = 3

	// 单账号 503 退避重试：原地重试时单次最大等待时间
	// 防止上游返回过长的 retryDelay 导致请求卡住太久
	antigravitySingleAccountSmartRetryMaxWait = 15 * time.Second

	// 单账号 503 退避重试：原地重试的总累计等待时间上限
	// 超过此上限将不再重试，直接返回 503
	antigravitySingleAccountSmartRetryTotalMaxWait = 30 * time.Second

	// MODEL_CAPACITY_EXHAUSTED 全局去重：重试全部失败后的 cooldown 时间
	antigravityModelCapacityCooldown = 10 * time.Second
)

// antigravityPassthroughErrorMessages 透传给客户端的错误消息白名单（小写）
// 匹配时使用 strings.Contains，无需完全匹配
var antigravityPassthroughErrorMessages = []string{
	"prompt is too long",
}

// MODEL_CAPACITY_EXHAUSTED 全局去重：避免多个并发请求同时对同一模型进行容量耗尽重试
var (
	modelCapacityExhaustedMu    sync.RWMutex
	modelCapacityExhaustedUntil = make(map[string]time.Time) // modelName -> cooldown until
)

const (
	antigravityForwardBaseURLEnv  = "GATEWAY_ANTIGRAVITY_FORWARD_BASE_URL"
	antigravityFallbackSecondsEnv = "GATEWAY_ANTIGRAVITY_FALLBACK_COOLDOWN_SECONDS"
)

// AntigravityAccountSwitchError 账号切换信号
// 当账号限流时间超过阈值时，通知上层切换账号
type AntigravityAccountSwitchError struct {
	OriginalAccountID int64
	RateLimitedModel  string
	IsStickySession   bool // 是否为粘性会话切换（决定是否缓存计费）
}

func (e *AntigravityAccountSwitchError) Error() string {
	return fmt.Sprintf("account %d model %s rate limited, need switch",
		e.OriginalAccountID, e.RateLimitedModel)
}

// IsAntigravityAccountSwitchError 检查错误是否为账号切换信号
func IsAntigravityAccountSwitchError(err error) (*AntigravityAccountSwitchError, bool) {
	var switchErr *AntigravityAccountSwitchError
	if errors.As(err, &switchErr) {
		return switchErr, true
	}
	return nil, false
}

// PromptTooLongError 表示上游明确返回 prompt too long
type PromptTooLongError struct {
	StatusCode int
	RequestID  string
	Body       []byte
}

func (e *PromptTooLongError) Error() string {
	return fmt.Sprintf("prompt too long: status=%d", e.StatusCode)
}

// antigravityRetryLoopParams 重试循环的参数
type antigravityRetryLoopParams struct {
	ctx             context.Context
	prefix          string
	account         *Account
	proxyURL        string
	accessToken     string
	action          string
	body            []byte
	c               *gin.Context
	httpUpstream    HTTPUpstream
	settingService  *SettingService
	accountRepo     AccountRepository // 用于智能重试的模型级别限流
	handleError     func(ctx context.Context, prefix string, account *Account, statusCode int, headers http.Header, body []byte, requestedModel string, groupID int64, sessionHash string, isStickySession bool) *handleModelRateLimitResult
	requestedModel  string // 用于限流检查的原始请求模型
	isStickySession bool   // 是否为粘性会话（用于账号切换时的缓存计费判断）
	groupID         int64  // 用于模型级限流时清除粘性会话
	sessionHash     string // 用于模型级限流时清除粘性会话
}

// antigravityRetryLoopResult 重试循环的结果
type antigravityRetryLoopResult struct {
	resp *http.Response
}

// getSessionID 从 gin.Context 获取 session_id（用于日志追踪）
func getSessionID(c *gin.Context) string {
	if c == nil {
		return ""
	}
	return c.GetHeader("session_id")
}

// logPrefix 生成统一的日志前缀
func logPrefix(sessionID, accountName string) string {
	if sessionID != "" {
		return fmt.Sprintf("[antigravity-Forward] session=%s account=%s", sessionID, accountName)
	}
	return fmt.Sprintf("[antigravity-Forward] account=%s", accountName)
}

// AntigravityGatewayService 处理 Antigravity 平台的 API 转发
type AntigravityGatewayService struct {
	accountRepo       AccountRepository
	tokenProvider     *AntigravityTokenProvider
	rateLimitService  *RateLimitService
	httpUpstream      HTTPUpstream
	settingService    *SettingService
	cache             GatewayCache // 用于模型级限流时清除粘性会话绑定
	schedulerSnapshot *SchedulerSnapshotService
	internal500Cache  Internal500CounterCache // INTERNAL 500 渐进惩罚计数器
}

func NewAntigravityGatewayService(
	accountRepo AccountRepository,
	cache GatewayCache,
	schedulerSnapshot *SchedulerSnapshotService,
	tokenProvider *AntigravityTokenProvider,
	rateLimitService *RateLimitService,
	httpUpstream HTTPUpstream,
	settingService *SettingService,
	internal500Cache Internal500CounterCache,
) *AntigravityGatewayService {
	return &AntigravityGatewayService{
		accountRepo:       accountRepo,
		tokenProvider:     tokenProvider,
		rateLimitService:  rateLimitService,
		httpUpstream:      httpUpstream,
		settingService:    settingService,
		cache:             cache,
		schedulerSnapshot: schedulerSnapshot,
		internal500Cache:  internal500Cache,
	}
}

// GetTokenProvider 返回 token provider
func (s *AntigravityGatewayService) GetTokenProvider() *AntigravityTokenProvider {
	return s.tokenProvider
}

// getLogConfig 获取上游错误日志配置
// 返回是否记录日志体和最大字节数
func (s *AntigravityGatewayService) getLogConfig() (logBody bool, maxBytes int) {
	maxBytes = 2048 // 默认值
	if s.settingService == nil || s.settingService.cfg == nil {
		return false, maxBytes
	}
	cfg := s.settingService.cfg.Gateway
	if cfg.LogUpstreamErrorBodyMaxBytes > 0 {
		maxBytes = cfg.LogUpstreamErrorBodyMaxBytes
	}
	return cfg.LogUpstreamErrorBody, maxBytes
}

// getUpstreamErrorDetail 获取上游错误详情（用于日志记录）
func (s *AntigravityGatewayService) getUpstreamErrorDetail(body []byte) string {
	logBody, maxBytes := s.getLogConfig()
	if !logBody {
		return ""
	}
	return truncateString(string(body), maxBytes)
}

// checkErrorPolicy nil 安全的包装
func (s *AntigravityGatewayService) checkErrorPolicy(ctx context.Context, account *Account, statusCode int, body []byte) ErrorPolicyResult {
	if s.rateLimitService == nil {
		return ErrorPolicyNone
	}
	return s.rateLimitService.CheckErrorPolicy(ctx, account, statusCode, body)
}

// applyErrorPolicy 应用错误策略结果，返回是否应终止当前循环及应返回的状态码。
// ErrorPolicySkipped 时 outStatus 为 500（前端约定：未命中的错误返回 500）。
func (s *AntigravityGatewayService) applyErrorPolicy(p antigravityRetryLoopParams, statusCode int, headers http.Header, respBody []byte) (handled bool, outStatus int, retErr error) {
	switch s.checkErrorPolicy(p.ctx, p.account, statusCode, respBody) {
	case ErrorPolicySkipped:
		return true, http.StatusInternalServerError, nil
	case ErrorPolicyMatched:
		_ = p.handleError(p.ctx, p.prefix, p.account, statusCode, headers, respBody,
			p.requestedModel, p.groupID, p.sessionHash, p.isStickySession)
		return true, statusCode, nil
	case ErrorPolicyTempUnscheduled:
		slog.Info("temp_unschedulable_matched",
			"prefix", p.prefix, "status_code", statusCode, "account_id", p.account.ID)
		return true, statusCode, &AntigravityAccountSwitchError{OriginalAccountID: p.account.ID, RateLimitedModel: p.requestedModel, IsStickySession: p.isStickySession}
	}
	return false, statusCode, nil
}
