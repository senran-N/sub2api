package handler

import (
	"strings"

	"github.com/senran-N/sub2api/internal/pkg/ctxkey"
	middleware2 "github.com/senran-N/sub2api/internal/server/middleware"
	"github.com/senran-N/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

func resolveGatewayModelListingContext(c *gin.Context) (*int64, string) {
	if c == nil || c.Request == nil {
		return nil, ""
	}

	var (
		groupID  *int64
		platform string
	)

	apiKey, _ := middleware2.GetAPIKeyFromContext(c)
	if apiKey != nil && apiKey.Group != nil {
		groupID = &apiKey.Group.ID
		platform = apiKey.Group.Platform
	}
	if group, ok := c.Request.Context().Value(ctxkey.Group).(*service.Group); ok && service.IsGroupContextValid(group) {
		groupID = &group.ID
		platform = group.Platform
	}
	if forcedPlatform, ok := middleware2.GetForcePlatformFromContext(c); ok && strings.TrimSpace(forcedPlatform) != "" {
		platform = forcedPlatform
	}

	return groupID, platform
}
