package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/senran-N/sub2api/internal/pkg/antigravity"
	"github.com/senran-N/sub2api/internal/pkg/logger"
)

type antigravityStreamResult struct {
	usage            *ClaudeUsage
	firstTokenMs     *int
	clientDisconnect bool // 客户端是否在流式传输过程中断开
}

// antigravityClientWriter 封装流式响应的客户端写入，自动检测断开并标记。
// 断开后所有写入操作变为 no-op，调用方通过 Disconnected() 判断是否继续 drain 上游。
type antigravityClientWriter struct {
	w            gin.ResponseWriter
	flusher      http.Flusher
	disconnected bool
	prefix       string // 日志前缀，标识来源方法
}

func newAntigravityClientWriter(w gin.ResponseWriter, flusher http.Flusher, prefix string) *antigravityClientWriter {
	return &antigravityClientWriter{w: w, flusher: flusher, prefix: prefix}
}

// Write 写入数据到客户端，写入失败时标记断开并返回 false
func (cw *antigravityClientWriter) Write(p []byte) bool {
	if cw.disconnected {
		return false
	}
	if _, err := cw.w.Write(p); err != nil {
		cw.markDisconnected()
		return false
	}
	cw.flusher.Flush()
	return true
}

// Fprintf 格式化写入数据到客户端，写入失败时标记断开并返回 false
func (cw *antigravityClientWriter) Fprintf(format string, args ...any) bool {
	if cw.disconnected {
		return false
	}
	if _, err := fmt.Fprintf(cw.w, format, args...); err != nil {
		cw.markDisconnected()
		return false
	}
	cw.flusher.Flush()
	return true
}

func (cw *antigravityClientWriter) Disconnected() bool { return cw.disconnected }

func (cw *antigravityClientWriter) markDisconnected() {
	cw.disconnected = true
	logger.LegacyPrintf("service.antigravity_gateway", "Client disconnected during streaming (%s), continuing to drain upstream for billing", cw.prefix)
}

// handleStreamReadError 处理上游读取错误的通用逻辑。
// 返回 (clientDisconnect, handled)：handled=true 表示错误已处理，调用方应返回已收集的 usage。
func handleStreamReadError(err error, clientDisconnected bool, prefix string) (disconnect bool, handled bool) {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		logger.LegacyPrintf("service.antigravity_gateway", "Context canceled during streaming (%s), returning collected usage", prefix)
		return true, true
	}
	if clientDisconnected {
		logger.LegacyPrintf("service.antigravity_gateway", "Upstream read error after client disconnect (%s): %v, returning collected usage", prefix, err)
		return true, true
	}
	return false, false
}

func convertAntigravityClaudeUsage(agUsage *antigravity.ClaudeUsage) *ClaudeUsage {
	if agUsage == nil {
		return &ClaudeUsage{}
	}

	return &ClaudeUsage{
		InputTokens:              agUsage.InputTokens,
		OutputTokens:             agUsage.OutputTokens,
		CacheCreationInputTokens: agUsage.CacheCreationInputTokens,
		CacheReadInputTokens:     agUsage.CacheReadInputTokens,
	}
}
