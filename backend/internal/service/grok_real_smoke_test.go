//go:build unit

package service

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/senran-N/sub2api/internal/pkg/apicompat"
	"github.com/senran-N/sub2api/internal/pkg/tlsfingerprint"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

type realHTTPUpstream struct {
	client *http.Client
}

func (u *realHTTPUpstream) Do(req *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
	return u.do(req)
}

func (u *realHTTPUpstream) DoWithTLS(req *http.Request, _ string, _ int64, _ int, _ *tlsfingerprint.Profile) (*http.Response, error) {
	return u.do(req)
}

func (u *realHTTPUpstream) do(req *http.Request) (*http.Response, error) {
	client := u.client
	if client == nil {
		client = &http.Client{Timeout: 120 * time.Second}
	}
	return client.Do(req)
}

func TestGrokRealSmoke_HandleMessages_WithSessionToken(t *testing.T) {
	account := requireRealGrokSessionAccount(t)

	gin.SetMode(gin.TestMode)

	body := []byte(`{"model":"grok-3","max_tokens":128,"stream":false,"messages":[{"role":"user","content":"Reply with exactly: smoke-ok"}]}`)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewReader(body))
	c.Request = c.Request.WithContext(WithGrokSessionTextRuntimeAllowed(context.Background()))
	c.Request.Header.Set("Content-Type", "application/json")

	repo := &mockAccountRepoForPlatform{accounts: []Account{account}}
	svc := NewGrokGatewayService(&GatewayService{
		accountRepo:  repo,
		httpUpstream: &realHTTPUpstream{},
		cfg:          testConfig(),
	}, nil)

	handled := svc.HandleMessages(c, nil, body)
	require.True(t, handled)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	var resp apicompat.AnthropicResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp), rec.Body.String())
	require.Equal(t, "message", resp.Type)
	require.Equal(t, "assistant", resp.Role)
	require.NotEmpty(t, resp.Content, rec.Body.String())
	require.Equal(t, "text", resp.Content[0].Type, rec.Body.String())

	text := strings.TrimSpace(resp.Content[0].Text)
	require.NotEmpty(t, text, rec.Body.String())
	t.Logf("messages response: %q", text)
}

func TestGrokRealSmoke_HandleImages_WithSessionToken(t *testing.T) {
	account := requireRealGrokSessionAccount(t)
	account.Extra = map[string]any{
		"grok": map[string]any{
			"tier": map[string]any{
				"normalized": "basic",
			},
		},
	}

	gin.SetMode(gin.TestMode)

	body := []byte(`{"model":"grok-2-image","prompt":"a tiny red square on a white background","n":1}`)
	c, rec := newGrokMediaTestContext(http.MethodPost, "/v1/images/generations", body)
	c.Request.Host = "gateway.example"
	c.Request.Header.Set("X-Forwarded-Proto", "https")

	repo := &mockAccountRepoForPlatform{accounts: []Account{account}}
	mediaAssets := &stubGrokMediaAssetRepository{}
	svc := NewGrokMediaService(&GatewayService{
		accountRepo:  repo,
		httpUpstream: &realHTTPUpstream{client: &http.Client{Timeout: 180 * time.Second}},
		cfg:          testConfig(),
	}, nil, mediaAssets)

	handled := svc.HandleImages(c, nil, body)
	require.True(t, handled)
	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected image status %d body=%s raw=%s", rec.Code, rec.Body.String(), debugGrokSessionImageStream(t, account, body))
	}

	url := strings.TrimSpace(gjson.Get(rec.Body.String(), "data.0.url").String())
	if url == "" {
		t.Fatalf("image url missing body=%s raw=%s", rec.Body.String(), debugGrokSessionImageStream(t, account, body))
	}
	require.Len(t, mediaAssets.upserts, 1, rec.Body.String())
	require.Contains(t, url, "/grok/media/assets/", rec.Body.String())

	upstreamURL := strings.TrimSpace(gjson.Get(rec.Body.String(), "data.0.upstream_url").String())
	require.NotEmpty(t, upstreamURL, rec.Body.String())
	t.Logf("image proxy url: %s", url)
}

func TestGrokRealSmoke_HandleVideos_WithSessionToken(t *testing.T) {
	account := requireRealGrokSessionAccount(t)
	account.Extra = map[string]any{
		"grok": map[string]any{
			"tier": map[string]any{
				"normalized": "super",
			},
		},
	}

	gin.SetMode(gin.TestMode)

	body := []byte(`{"model":"grok-imagine-video","prompt":"a tiny red ball bouncing on a white background","seconds":6}`)
	videoJobs := &stubGrokVideoJobRepository{}
	mediaAssets := &stubGrokMediaAssetRepository{}
	repo := &mockAccountRepoForPlatform{
		accounts: []Account{account},
		accountsByID: map[int64]*Account{
			account.ID: accountPtr(account),
		},
	}
	svc := NewGrokMediaService(&GatewayService{
		accountRepo:  repo,
		httpUpstream: &realHTTPUpstream{client: &http.Client{Timeout: 360 * time.Second}},
		cfg:          testConfig(),
	}, videoJobs, mediaAssets)
	svc.sessionRuntime.spawn = func(fn func()) { fn() }

	createCtx, createRec := newGrokMediaTestContext(http.MethodPost, "/v1/videos", body)
	createCtx.Request.Host = "gateway.example"
	createCtx.Request.Header.Set("X-Forwarded-Proto", "https")

	handled := svc.HandleVideos(createCtx, nil, body)
	require.True(t, handled)
	require.Equal(t, http.StatusOK, createRec.Code, createRec.Body.String())

	jobID := strings.TrimSpace(gjson.Get(createRec.Body.String(), "id").String())
	require.NotEmpty(t, jobID, createRec.Body.String())

	record, ok := videoJobs.records[jobID]
	require.True(t, ok, createRec.Body.String())
	if record.NormalizedStatus != "completed" {
		require.Equal(t, "rate_limit_error", record.ErrorCode, createRec.Body.String())
		require.Len(t, repo.runtimeStates, 1, createRec.Body.String())
		runtimeState := repo.runtimeStates[0]
		require.Equal(t, "error", runtimeState["last_outcome"])
		require.Equal(t, http.StatusTooManyRequests, grokParseInt(runtimeState["last_request_status_code"]))
		require.Equal(t, string(grokRuntimeErrorClassRateLimited), runtimeState["last_fail_class"])
		require.NotEmpty(t, runtimeState["selection_cooldown_until"])
		t.Fatalf("video job not completed status=%s upstream=%s error=%s/%s body=%s", record.NormalizedStatus, record.UpstreamStatus, record.ErrorCode, record.ErrorMessage, createRec.Body.String())
	}
	require.NotEmpty(t, strings.TrimSpace(record.OutputAssetID), createRec.Body.String())

	statusCtx, statusRec := newGrokMediaTestContext(http.MethodGet, "/v1/videos/"+jobID, nil)
	statusHandled := svc.HandleVideos(statusCtx, nil, nil)
	require.True(t, statusHandled)
	require.Equal(t, http.StatusOK, statusRec.Code, statusRec.Body.String())
	require.Equal(t, "completed", strings.TrimSpace(gjson.Get(statusRec.Body.String(), "status").String()), statusRec.Body.String())

	contentCtx, contentRec := newGrokMediaTestContext(http.MethodGet, "/v1/videos/"+jobID+"/content", nil)
	contentCtx.Request.Host = "gateway.example"
	contentCtx.Request.Header.Set("X-Forwarded-Proto", "https")

	contentHandled := svc.HandleVideos(contentCtx, nil, nil)
	require.True(t, contentHandled)
	require.Equal(t, http.StatusOK, contentRec.Code, contentRec.Body.String())

	contentURL := strings.TrimSpace(gjson.Get(contentRec.Body.String(), "content_url").String())
	upstreamContentURL := strings.TrimSpace(gjson.Get(contentRec.Body.String(), "upstream_content_url").String())
	require.NotEmpty(t, contentURL, contentRec.Body.String())
	require.NotEmpty(t, upstreamContentURL, contentRec.Body.String())
	require.Contains(t, contentURL, "/grok/media/assets/", contentRec.Body.String())
	t.Logf("video job=%s proxy=%s", jobID, contentURL)
}

func requireRealGrokSessionAccount(t *testing.T) Account {
	t.Helper()

	rawToken := strings.TrimSpace(os.Getenv("GROK_SSO_TOKEN"))
	if rawToken == "" {
		t.Skip("GROK_SSO_TOKEN is not set")
	}

	account := Account{
		ID:          900001,
		Name:        "grok-real-smoke",
		Platform:    PlatformGrok,
		Type:        AccountTypeSession,
		Status:      StatusActive,
		Schedulable: true,
		Credentials: map[string]any{
			"session_token": rawToken,
		},
	}

	if cfCookies := strings.TrimSpace(os.Getenv("GROK_CF_COOKIES")); cfCookies != "" {
		account.Credentials["cf_cookies"] = cfCookies
	}
	if cfClearance := strings.TrimSpace(os.Getenv("GROK_CF_CLEARANCE")); cfClearance != "" {
		account.Credentials["cf_clearance"] = cfClearance
	}
	if userAgent := strings.TrimSpace(os.Getenv("GROK_SESSION_USER_AGENT")); userAgent != "" {
		account.Credentials["user_agent"] = userAgent
	}
	if acceptLang := strings.TrimSpace(os.Getenv("GROK_SESSION_ACCEPT_LANGUAGE")); acceptLang != "" {
		account.Credentials["accept_language"] = acceptLang
	}

	return account
}

func debugGrokSessionImageStream(t *testing.T, account Account, requestBody []byte) string {
	t.Helper()

	req, err := parseGrokSessionImageGenerationRequest(requestBody, "grok-2-image")
	require.NoError(t, err)

	payload, err := buildGrokSessionTextPayload(grokSessionTextRequest{
		ModelID: req.Model,
		ModeID:  resolveGrokSessionMediaModeID(req.Model),
		Message: "Drawing: " + strings.TrimSpace(req.Prompt),
	})
	require.NoError(t, err)

	payload["disableTextFollowUps"] = true
	payload["enableImageGeneration"] = true
	payload["enableImageStreaming"] = true
	payload["imageGenerationCount"] = req.N
	payload["toolOverrides"] = map[string]any{"imageGen": true}
	payload["returnImageBytes"] = false
	payload["returnRawGrokInXaiRequest"] = false

	payloadBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	target, err := resolveGrokTransportTargetWithSettings(&account, nil, DefaultGrokRuntimeSettings())
	require.NoError(t, err)

	httpReq, err := newGrokSessionJSONRequest(
		context.Background(),
		http.MethodPost,
		target,
		payloadBytes,
		"application/json, text/event-stream, text/plain, */*",
	)
	require.NoError(t, err)

	resp, err := (&realHTTPUpstream{client: &http.Client{Timeout: 180 * time.Second}}).DoWithTLS(httpReq, "", account.ID, account.Concurrency, nil)
	if err != nil {
		return "request_error=" + err.Error()
	}
	defer func() { _ = resp.Body.Close() }()

	raw, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return "read_error=" + readErr.Error()
	}
	return "status=" + resp.Status + " body=" + string(raw)
}
