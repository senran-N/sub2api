package service

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

var benchmarkOpenAIRequestMetaSink openAIRequestMeta

func TestExtractOpenAIRequestMetaFromBody(t *testing.T) {
	tests := []struct {
		name          string
		body          []byte
		wantModel     string
		wantStream    bool
		wantPromptKey string
	}{
		{
			name:          "完整字段",
			body:          []byte(`{"model":"gpt-5","stream":true,"prompt_cache_key":" ses-1 "}`),
			wantModel:     "gpt-5",
			wantStream:    true,
			wantPromptKey: "ses-1",
		},
		{
			name:          "缺失可选字段",
			body:          []byte(`{"model":"gpt-4"}`),
			wantModel:     "gpt-4",
			wantStream:    false,
			wantPromptKey: "",
		},
		{
			name:          "空请求体",
			body:          nil,
			wantModel:     "",
			wantStream:    false,
			wantPromptKey: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model, stream, promptKey := extractOpenAIRequestMetaFromBody(tt.body)
			require.Equal(t, tt.wantModel, model)
			require.Equal(t, tt.wantStream, stream)
			require.Equal(t, tt.wantPromptKey, promptKey)
		})
	}
}

func TestExtractOpenAIReasoningEffortFromBody(t *testing.T) {
	tests := []struct {
		name      string
		body      []byte
		model     string
		wantNil   bool
		wantValue string
	}{
		{
			name:      "优先读取 reasoning.effort",
			body:      []byte(`{"reasoning":{"effort":"medium"}}`),
			model:     "gpt-5-high",
			wantNil:   false,
			wantValue: "medium",
		},
		{
			name:      "兼容 reasoning_effort",
			body:      []byte(`{"reasoning_effort":"x-high"}`),
			model:     "",
			wantNil:   false,
			wantValue: "xhigh",
		},
		{
			name:    "minimal 归一化为空",
			body:    []byte(`{"reasoning":{"effort":"minimal"}}`),
			model:   "gpt-5-high",
			wantNil: true,
		},
		{
			name:      "缺失字段时从模型后缀推导",
			body:      []byte(`{"input":"hi"}`),
			model:     "gpt-5-high",
			wantNil:   false,
			wantValue: "high",
		},
		{
			name:    "未知后缀不返回",
			body:    []byte(`{"input":"hi"}`),
			model:   "gpt-5-unknown",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractOpenAIReasoningEffortFromBody(tt.body, tt.model)
			if tt.wantNil {
				require.Nil(t, got)
				return
			}
			require.NotNil(t, got)
			require.Equal(t, tt.wantValue, *got)
		})
	}
}

func TestGetOpenAIRequestBodyMap_UsesContextCache(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)

	cached := map[string]any{"model": "cached-model", "stream": true}
	c.Set(OpenAIParsedRequestBodyKey, cached)

	got, err := getOpenAIRequestBodyMap(c, []byte(`{invalid-json`))
	require.NoError(t, err)
	require.Equal(t, cached, got)
}

func TestGetOpenAIRequestBodyMap_ParseErrorWithoutCache(t *testing.T) {
	_, err := getOpenAIRequestBodyMap(nil, []byte(`{invalid-json`))
	require.Error(t, err)
	require.Contains(t, err.Error(), "parse request")
}

func TestGetOpenAIRequestBodyMap_WriteBackContextCache(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)

	got, err := getOpenAIRequestBodyMap(c, []byte(`{"model":"gpt-5","stream":true}`))
	require.NoError(t, err)
	require.Equal(t, "gpt-5", got["model"])

	cached, ok := c.Get(OpenAIParsedRequestBodyKey)
	require.True(t, ok)
	cachedMap, ok := cached.(map[string]any)
	require.True(t, ok)
	require.Equal(t, got, cachedMap)
}

func TestGetOpenAIRequestMeta_CachesBodyExtraction(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)

	body := []byte(`{"model":" gpt-5 ","stream":true,"prompt_cache_key":" ses-1 ","previous_response_id":" resp_123 ","reasoning":{"effort":"x-high"}}`)
	meta := GetOpenAIRequestMeta(c, body)
	require.True(t, meta.ModelExists)
	require.Equal(t, gjson.String, meta.ModelType)
	require.Equal(t, "gpt-5", meta.Model)
	require.True(t, meta.StreamExists)
	require.Equal(t, gjson.True, meta.StreamType)
	require.True(t, meta.Stream)
	require.Equal(t, "ses-1", meta.PromptCacheKey)
	require.Equal(t, "resp_123", meta.PreviousResponseID)
	require.Equal(t, "xhigh", meta.ReasoningEffort)

	cached, ok := c.Get(OpenAIParsedRequestMetaKey)
	require.True(t, ok)
	cachedMeta, ok := cached.(openAIRequestMeta)
	require.True(t, ok)
	require.Equal(t, meta, cachedMeta)
}

func TestCacheOpenAIRequestMetaFromBodyMap(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)

	reqBody := map[string]any{
		"model":                "gpt-5-high",
		"stream":               false,
		"prompt_cache_key":     "seed-1",
		"previous_response_id": "resp_prev",
		"reasoning_effort":     "medium",
	}
	CacheOpenAIRequestMetaFromBodyMap(c, reqBody)

	meta := GetOpenAIRequestMeta(c, []byte(`{invalid-json`))
	require.True(t, meta.ModelExists)
	require.Equal(t, "gpt-5-high", meta.Model)
	require.True(t, meta.StreamExists)
	require.False(t, meta.Stream)
	require.Equal(t, gjson.False, meta.StreamType)
	require.Equal(t, "seed-1", meta.PromptCacheKey)
	require.Equal(t, "resp_prev", meta.PreviousResponseID)
	require.Equal(t, "medium", meta.ReasoningEffort)
}

func TestGetOpenAIRequestMeta_InvalidatesBodyBoundCacheWhenBodyChanges(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)

	first := []byte(`{"model":"gpt-5","prompt_cache_key":"seed-1"}`)
	second := []byte(`{"model":"gpt-5"}`)

	meta1 := GetOpenAIRequestMeta(c, first)
	require.Equal(t, "seed-1", meta1.PromptCacheKey)

	meta2 := GetOpenAIRequestMeta(c, second)
	require.Empty(t, meta2.PromptCacheKey)
}

func BenchmarkBuildOpenAIRequestMeta(b *testing.B) {
	body := []byte(`{"model":"gpt-5","stream":true,"prompt_cache_key":"ses-1","previous_response_id":"resp_123","reasoning":{"effort":"medium"}}`)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchmarkOpenAIRequestMetaSink = buildOpenAIRequestMeta(body)
	}
}

func BenchmarkGetOpenAIRequestMetaCached(b *testing.B) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	body := []byte(`{"model":"gpt-5","stream":true,"prompt_cache_key":"ses-1","previous_response_id":"resp_123","reasoning":{"effort":"medium"}}`)
	benchmarkOpenAIRequestMetaSink = GetOpenAIRequestMeta(c, body)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchmarkOpenAIRequestMetaSink = GetOpenAIRequestMeta(c, nil)
	}
}
