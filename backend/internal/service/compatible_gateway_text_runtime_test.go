package service

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestResolveCompatibleGatewayRuntimeStatusCode(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("prefers explicit status", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)
		c.Writer.WriteHeader(http.StatusTooManyRequests)
		c.Writer.WriteHeaderNow()

		require.Equal(t, http.StatusBadGateway, resolveCompatibleGatewayRuntimeStatusCode(c, http.StatusBadGateway))
	})

	t.Run("uses written gin status", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)
		c.Writer.WriteHeader(http.StatusTooManyRequests)
		c.Writer.WriteHeaderNow()

		require.Equal(t, http.StatusTooManyRequests, resolveCompatibleGatewayRuntimeStatusCode(c, 0))
	})

	t.Run("keeps unset status when response is not written", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)

		require.Zero(t, resolveCompatibleGatewayRuntimeStatusCode(c, 0))
	})
}
