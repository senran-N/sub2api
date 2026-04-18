package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const emptyClaudeCodePolicyLimitsETag = `W/"sub2api-claude-code-policy-limits-empty-v1"`

// RegisterCommonRoutes 注册通用路由（健康检查、状态等）
func RegisterCommonRoutes(r *gin.Engine) {
	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Claude Code 遥测日志（忽略，直接返回200）
	r.POST("/api/event_logging/batch", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Claude Code 组织策略轮询，返回空限制集以保持兼容。
	r.GET("/api/claude_code/policy_limits", func(c *gin.Context) {
		c.Header("ETag", emptyClaudeCodePolicyLimitsETag)
		if c.GetHeader("If-None-Match") == emptyClaudeCodePolicyLimitsETag {
			c.Status(http.StatusNotModified)
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"restrictions": gin.H{},
		})
	})

	// Claude Code 远程托管设置，返回空对象表示没有服务端覆盖配置。
	r.GET("/api/claude_code/settings", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{})
	})

	// Claude Code WebFetch 域名检查，允许抓取以避免前置阻断。
	r.GET("/api/web/domain_info", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"can_fetch": true,
		})
	})

	// Setup status endpoint (always returns needs_setup: false in normal mode)
	// This is used by the frontend to detect when the service has restarted after setup
	r.GET("/setup/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"data": gin.H{
				"needs_setup": false,
				"step":        "completed",
			},
		})
	})
}
