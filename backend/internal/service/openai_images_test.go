package service

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func newOpenAIImagesTestContext(method, path, contentType string, body []byte) *gin.Context {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	c.Request = req
	return c
}

func TestOpenAIGatewayServiceParseOpenAIImagesRequest_DefaultsMissingModelToGPTImage2(t *testing.T) {
	body := []byte(`{"prompt":"draw a cat"}`)
	c := newOpenAIImagesTestContext(http.MethodPost, "/v1/images/generations", "application/json", body)

	parsed, err := (&OpenAIGatewayService{}).ParseOpenAIImagesRequest(c, body)

	require.NoError(t, err)
	require.Equal(t, openAIImagesGenerationsEndpoint, parsed.Endpoint)
	require.Equal(t, "gpt-image-2", parsed.Model)
	require.False(t, parsed.ExplicitModel)
	require.Equal(t, OpenAIImagesCapabilityBasic, parsed.RequiredCapability)
}

func TestOpenAIGatewayServiceParseOpenAIImagesRequest_RejectsTextModelGPT55(t *testing.T) {
	body := []byte(`{"model":"gpt-5.5","prompt":"draw a cat"}`)
	c := newOpenAIImagesTestContext(http.MethodPost, "/v1/images/generations", "application/json", body)

	parsed, err := (&OpenAIGatewayService{}).ParseOpenAIImagesRequest(c, body)

	require.Nil(t, parsed)
	require.ErrorContains(t, err, `images endpoint requires an image model, got "gpt-5.5"`)
}

func TestOpenAIGatewayServiceParseOpenAIImagesRequest_MultipartEditPreservesGPTImage2(t *testing.T) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	require.NoError(t, writer.WriteField("model", "gpt-image-2"))
	require.NoError(t, writer.WriteField("prompt", "replace background"))
	part, err := writer.CreateFormFile("image", "source.png")
	require.NoError(t, err)
	_, err = part.Write([]byte("fake-image-bytes"))
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	c := newOpenAIImagesTestContext(http.MethodPost, "/v1/images/edits", writer.FormDataContentType(), body.Bytes())

	parsed, err := (&OpenAIGatewayService{}).ParseOpenAIImagesRequest(c, body.Bytes())

	require.NoError(t, err)
	require.Equal(t, openAIImagesEditsEndpoint, parsed.Endpoint)
	require.True(t, parsed.Multipart)
	require.Equal(t, "gpt-image-2", parsed.Model)
	require.Equal(t, "replace background", parsed.Prompt)
	require.Len(t, parsed.Uploads, 1)
	require.Equal(t, OpenAIImagesCapabilityNative, parsed.RequiredCapability)
}

func TestOpenAIGatewayServiceForwardImages_APIKeyUsesNativeImagesEndpointAndRewritesModel(t *testing.T) {
	body := []byte(`{"model":"gpt-image-2","prompt":"draw a cat","size":"1024x1024","n":2}`)
	c := newOpenAIImagesTestContext(http.MethodPost, "/v1/images/generations", "application/json", body)
	svc := &OpenAIGatewayService{}
	parsed, err := svc.ParseOpenAIImagesRequest(c, body)
	require.NoError(t, err)

	upstream := &httpUpstreamRecorder{
		resp: &http.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Type": []string{"application/json"},
				"X-Request-Id": []string{"req_img_native"},
			},
			Body: io.NopCloser(bytes.NewReader([]byte(`{
				"created":1710000000,
				"data":[{"b64_json":"YQ=="},{"b64_json":"Yg=="}],
				"usage":{"input_tokens":5,"output_tokens":9,"output_tokens_details":{"image_tokens":7}}
			}`))),
		},
	}
	svc.httpUpstream = upstream
	account := &Account{
		ID:       1,
		Name:     "openai-apikey",
		Platform: PlatformOpenAI,
		Type:     AccountTypeAPIKey,
		Credentials: map[string]any{
			"api_key": "sk-test",
		},
	}

	result, err := svc.ForwardImages(context.Background(), c, account, body, parsed, "gpt-image-2-2026-04-21")

	require.NoError(t, err)
	require.Equal(t, "https://api.openai.com/v1/images/generations", upstream.lastReq.URL.String())
	require.Equal(t, "Bearer sk-test", upstream.lastReq.Header.Get("Authorization"))
	require.Equal(t, "gpt-image-2-2026-04-21", gjson.GetBytes(upstream.lastBody, "model").String())
	require.Equal(t, "gpt-image-2-2026-04-21", result.Model)
	require.Equal(t, "gpt-image-2-2026-04-21", result.UpstreamModel)
	require.Equal(t, 2, result.ImageCount)
	require.Equal(t, "1K", result.ImageSize)
	require.Equal(t, "image", result.MediaType)
	require.Equal(t, 7, result.Usage.ImageOutputTokens)
}

func TestOpenAIGatewayServiceForwardImages_OAuthUsesResponsesImageTool(t *testing.T) {
	body := []byte(`{"model":"gpt-image-2","prompt":"draw a cat","size":"1024x1024","quality":"high"}`)
	c := newOpenAIImagesTestContext(http.MethodPost, "/v1/images/generations", "application/json", body)
	svc := &OpenAIGatewayService{}
	parsed, err := svc.ParseOpenAIImagesRequest(c, body)
	require.NoError(t, err)

	upstream := &httpUpstreamRecorder{
		resp: &http.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Type": []string{"text/event-stream"},
				"X-Request-Id": []string{"req_img_oauth"},
			},
			Body: io.NopCloser(bytes.NewReader([]byte(
				"data: {\"type\":\"response.completed\",\"response\":{\"created_at\":1710000000,\"usage\":{\"input_tokens\":11,\"output_tokens\":22,\"input_tokens_details\":{\"cached_tokens\":3},\"output_tokens_details\":{\"image_tokens\":7}},\"tool_usage\":{\"image_gen\":{\"images\":1}},\"output\":[{\"type\":\"image_generation_call\",\"result\":\"aGVsbG8=\",\"revised_prompt\":\"draw a cat\",\"output_format\":\"png\",\"quality\":\"high\",\"size\":\"1024x1024\"}]}}\n\n" +
					"data: [DONE]\n\n",
			))),
		},
	}
	svc.httpUpstream = upstream
	account := &Account{
		ID:       2,
		Name:     "openai-oauth",
		Platform: PlatformOpenAI,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"access_token":       "token-123",
			"chatgpt_account_id": "acct-123",
		},
	}

	result, err := svc.ForwardImages(context.Background(), c, account, body, parsed, "")

	require.NoError(t, err)
	require.Equal(t, chatgptCodexURL, upstream.lastReq.URL.String())
	require.Equal(t, "application/json", upstream.lastReq.Header.Get("Content-Type"))
	require.Equal(t, "text/event-stream", upstream.lastReq.Header.Get("Accept"))
	require.Equal(t, "acct-123", upstream.lastReq.Header.Get("chatgpt-account-id"))
	require.Equal(t, openAIImagesResponsesMainModel, gjson.GetBytes(upstream.lastBody, "model").String())
	require.Equal(t, "image_generation", gjson.GetBytes(upstream.lastBody, "tools.0.type").String())
	require.Equal(t, "generate", gjson.GetBytes(upstream.lastBody, "tools.0.action").String())
	require.Equal(t, "gpt-image-2", gjson.GetBytes(upstream.lastBody, "tools.0.model").String())
	require.Equal(t, "1024x1024", gjson.GetBytes(upstream.lastBody, "tools.0.size").String())
	require.Equal(t, "high", gjson.GetBytes(upstream.lastBody, "tools.0.quality").String())
	require.Equal(t, 1, result.ImageCount)
	require.Equal(t, 11, result.Usage.InputTokens)
	require.Equal(t, 7, result.Usage.ImageOutputTokens)
}

func TestAccountSupportsOpenAIImageCapability_DistinguishesNativeAndOAuthFallback(t *testing.T) {
	tests := []struct {
		name       string
		account    *Account
		wantBasic  bool
		wantNative bool
	}{
		{name: "oauth fallback", account: &Account{Platform: PlatformOpenAI, Type: AccountTypeOAuth}, wantBasic: true},
		{name: "api key native", account: &Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey}, wantBasic: true, wantNative: true},
		{name: "upstream native", account: &Account{Platform: PlatformOpenAI, Type: AccountTypeUpstream}, wantBasic: true, wantNative: true},
		{name: "grok is not openai images", account: &Account{Platform: PlatformGrok, Type: AccountTypeAPIKey}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.wantBasic, tt.account.SupportsOpenAIImageCapability(OpenAIImagesCapabilityBasic))
			require.Equal(t, tt.wantNative, tt.account.SupportsOpenAIImageCapability(OpenAIImagesCapabilityNative))
		})
	}
}
