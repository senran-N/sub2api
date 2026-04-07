package admin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestIfNoneMatchMatched(t *testing.T) {
	t.Parallel()

	require.True(t, ifNoneMatchMatched(`"abc"`, `"abc"`))
	require.True(t, ifNoneMatchMatched(`W/"abc"`, `"abc"`))
	require.True(t, ifNoneMatchMatched(`"other", "abc"`, `"abc"`))
	require.True(t, ifNoneMatchMatched(`*`, `"abc"`))
	require.False(t, ifNoneMatchMatched(`"other"`, `"abc"`))
	require.False(t, ifNoneMatchMatched(``, `"abc"`))
	require.False(t, ifNoneMatchMatched(`"abc"`, ``))
}

func TestRespondNotModifiedIfETagMatches(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	req := httptest.NewRequest(http.MethodGet, "/admin/test", nil)
	req.Header.Set("If-None-Match", `"etag-value"`)
	ctx.Request = req

	matched := respondNotModifiedIfETagMatches(ctx, `"etag-value"`)
	require.True(t, matched)
	require.Equal(t, http.StatusNotModified, ctx.Writer.Status())
	require.Equal(t, `"etag-value"`, recorder.Header().Get("ETag"))
	require.Equal(t, "If-None-Match", recorder.Header().Get("Vary"))
}
