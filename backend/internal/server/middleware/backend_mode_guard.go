package middleware

import (
	stdpath "path"
	"strings"

	"github.com/senran-N/sub2api/internal/pkg/response"
	"github.com/senran-N/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// BackendModeUserGuard blocks non-admin users from accessing user routes when backend mode is enabled.
// Must be placed AFTER JWT auth middleware so that the user role is available in context.
func BackendModeUserGuard(settingService *service.SettingService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if settingService == nil || !settingService.IsBackendModeEnabled(c.Request.Context()) {
			c.Next()
			return
		}
		role, _ := GetUserRoleFromContext(c)
		if role == "admin" {
			c.Next()
			return
		}
		response.Forbidden(c, "Backend mode is active. User self-service is disabled.")
		c.Abort()
	}
}

// BackendModeAuthGuard selectively blocks auth endpoints when backend mode is enabled.
// Allows: login, login/2fa, logout, refresh (admin needs these).
// Blocks: register, forgot-password, reset-password, OAuth, etc.
func BackendModeAuthGuard(settingService *service.SettingService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if settingService == nil || !settingService.IsBackendModeEnabled(c.Request.Context()) {
			c.Next()
			return
		}
		path := normalizeBackendModeAuthPath(c.Request.URL.Path)
		// Allow login, 2FA, logout, refresh, public settings
		allowedSuffixes := []string{"/auth/login", "/auth/login/2fa", "/auth/logout", "/auth/refresh"}
		for _, suffix := range allowedSuffixes {
			if strings.HasSuffix(path, suffix) {
				c.Next()
				return
			}
		}
		response.Forbidden(c, "Backend mode is active. Registration and self-service auth flows are disabled.")
		c.Abort()
	}
}

func normalizeBackendModeAuthPath(path string) string {
	if path == "" {
		return "/"
	}

	normalized := stdpath.Clean(path)
	if normalized == "." {
		return "/"
	}
	return normalized
}
