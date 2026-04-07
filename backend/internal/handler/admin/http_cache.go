package admin

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/senran-N/sub2api/internal/pkg/response"
)

func writeConditionalETagHeaders(c *gin.Context, etag string) {
	if etag == "" {
		return
	}

	c.Header("ETag", etag)
	c.Header("Vary", "If-None-Match")
}

func ifNoneMatchMatched(ifNoneMatch, etag string) bool {
	if etag == "" || ifNoneMatch == "" {
		return false
	}

	for _, token := range strings.Split(ifNoneMatch, ",") {
		candidate := strings.TrimSpace(token)
		if candidate == "*" {
			return true
		}
		if candidate == etag {
			return true
		}
		if strings.HasPrefix(candidate, "W/") && strings.TrimPrefix(candidate, "W/") == etag {
			return true
		}
	}

	return false
}

func respondNotModifiedIfETagMatches(c *gin.Context, etag string) bool {
	writeConditionalETagHeaders(c, etag)
	if !ifNoneMatchMatched(c.GetHeader("If-None-Match"), etag) {
		return false
	}

	c.Status(http.StatusNotModified)
	return true
}

func respondSnapshotCacheEntry(c *gin.Context, entry snapshotCacheEntry, hit bool) {
	if respondNotModifiedIfETagMatches(c, entry.ETag) {
		return
	}

	c.Header("X-Snapshot-Cache", cacheStatusValue(hit))
	response.Success(c, entry.Payload)
}
