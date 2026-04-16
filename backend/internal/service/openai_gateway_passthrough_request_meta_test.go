package service

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestGetOpenAICompatiblePassthroughRequestMeta_JSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/images/generations", strings.NewReader(`{"model":"grok-2-image","stream":true}`))
	c.Request.Header.Set("Content-Type", "application/json")

	meta := GetOpenAICompatiblePassthroughRequestMeta(c, []byte(`{"model":"grok-2-image","stream":true}`))
	require.Equal(t, "grok-2-image", meta.Model)
	require.True(t, meta.Stream)
	require.True(t, meta.JSONBody)
}

func TestGetOpenAICompatiblePassthroughRequestMeta_Multipart(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	body := strings.Join([]string{
		"--boundary123",
		`Content-Disposition: form-data; name="model"`,
		"",
		"grok-4-voice",
		"--boundary123",
		`Content-Disposition: form-data; name="file"; filename="audio.wav"`,
		"Content-Type: audio/wav",
		"",
		"RIFF",
		"--boundary123--",
		"",
	}, "\r\n")
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/stt", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "multipart/form-data; boundary=boundary123")

	meta := GetOpenAICompatiblePassthroughRequestMeta(c, []byte(body))
	require.Equal(t, "grok-4-voice", meta.Model)
	require.False(t, meta.JSONBody)
}
