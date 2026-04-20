//go:build unit

package service

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/senran-N/sub2api/internal/pkg/grok"
	"github.com/stretchr/testify/require"
)

func TestExtractGrokSessionImageCardURLs_UsesCompletedCardAttachmentImage(t *testing.T) {
	raw := []byte(`{"result":{"response":{"cardAttachment":{"jsonData":"{\"id\":\"card_1\",\"image_chunk\":{\"imageUrl\":\"users/demo/generated/part/image.jpg\",\"progress\":50}}"},"modelResponse":{"cardAttachmentsJson":["{\"id\":\"card_1\",\"image_chunk\":{\"imageUrl\":\"users/demo/generated/image.jpg\",\"progress\":100,\"moderated\":false}}"]}}}}`)

	urls := extractGrokSessionImageCardURLs(raw)

	require.Equal(t, []string{"https://assets.grok.com/users/demo/generated/image.jpg"}, urls)
}

func TestExtractGrokSessionImageCardURLs_SkipsModeratedImage(t *testing.T) {
	raw := []byte(`{"result":{"response":{"cardAttachment":{"jsonData":"{\"id\":\"card_1\",\"image_chunk\":{\"imageUrl\":\"users/demo/generated/image.jpg\",\"progress\":100,\"moderated\":true}}"}}}}`)

	urls := extractGrokSessionImageCardURLs(raw)

	require.Empty(t, urls)
}

func TestParseGrokSessionImageEditRequest_JSONCollectsMultipleImageReferences(t *testing.T) {
	body := []byte(`{"model":"grok-imagine-image-edit","prompt":"replace background","image":["https://media.example/a.png","https://media.example/b.png"]}`)

	req, err := parseGrokSessionImageEditRequest(nil, body, "grok-imagine-image-edit")

	require.NoError(t, err)
	require.Len(t, req.InputImages, 2)
	require.Equal(t, "https://media.example/a.png", req.InputImages[0].Source)
	require.Equal(t, "https://media.example/b.png", req.InputImages[1].Source)
	require.Equal(t, 1, req.N)
}

func TestParseGrokSessionImageGenerationRequest_NormalizesResponseFormat(t *testing.T) {
	req, err := parseGrokSessionImageGenerationRequest(
		[]byte(`{"model":"grok-2-image","prompt":"cat","response_format":"B64_JSON"}`),
		"grok-2-image",
	)

	require.NoError(t, err)
	require.Equal(t, grokOpenAIImageResponseFormatB64JSON, req.ResponseFormat)
}

func TestParseGrokSessionImageEditRequest_MultipartReadsResponseFormat(t *testing.T) {
	body := "--boundary\r\n" +
		"Content-Disposition: form-data; name=\"model\"\r\n\r\n" +
		"grok-imagine-image-edit\r\n" +
		"--boundary\r\n" +
		"Content-Disposition: form-data; name=\"prompt\"\r\n\r\n" +
		"replace sky\r\n" +
		"--boundary\r\n" +
		"Content-Disposition: form-data; name=\"response_format\"\r\n\r\n" +
		"b64_json\r\n" +
		"--boundary\r\n" +
		"Content-Disposition: form-data; name=\"image\"; filename=\"a.png\"\r\n" +
		"Content-Type: image/png\r\n\r\n" +
		"PNGDATA\r\n" +
		"--boundary--\r\n"

	req, err := parseMultipartGrokSessionImageEditRequest([]byte(body), "boundary", "grok-imagine-image-edit")

	require.NoError(t, err)
	require.Equal(t, grokOpenAIImageResponseFormatB64JSON, req.ResponseFormat)
	require.Len(t, req.InputImages, 1)
}

func TestCollectGrokSessionImageURLsWithProgress_ImageEditWaitsForFinalIndexedImages(t *testing.T) {
	resp := &http.Response{
		Body: io.NopCloser(strings.NewReader(strings.Join([]string{
			`{"result":{"response":{"streamingImageGenerationResponse":{"imageIndex":1,"progress":25,"imageUrl":"https://media.example/partial-b.png"}}}}`,
			`{"result":{"response":{"streamingImageGenerationResponse":{"imageIndex":0,"progress":40,"imageUrl":"https://media.example/partial-a.png"}}}}`,
			`{"result":{"response":{"streamingImageGenerationResponse":{"imageIndex":1,"progress":100,"imageUrl":"https://media.example/final-b.png"}}}}`,
			`{"result":{"response":{"streamingImageGenerationResponse":{"imageIndex":0,"progress":100,"imageUrl":"https://media.example/final-a.png"}}}}`,
		}, "\n"))),
	}

	urls, err := collectGrokSessionImageURLsWithProgress(
		resp,
		grokTransportTarget{},
		"/v1/images/edits",
		2,
		nil,
	)

	require.NoError(t, err)
	require.Equal(t, []string{
		"https://media.example/final-a.png",
		"https://media.example/final-b.png",
	}, urls)
}

func TestCollectGrokSessionImageURLsWithProgress_AbsolutizesModelResponseGeneratedURLs(t *testing.T) {
	resp := &http.Response{
		Body: io.NopCloser(strings.NewReader(`{"result":{"response":{"modelResponse":{"generatedImageUrls":["users/demo/generated/image.png"]}}}}`)),
	}

	urls, err := collectGrokSessionImageURLsWithProgress(
		resp,
		grokTransportTarget{},
		"/v1/images/generations",
		1,
		nil,
	)

	require.NoError(t, err)
	require.Equal(t, []string{"https://assets.grok.com/users/demo/generated/image.png"}, urls)
}

func TestGrokSessionMediaRuntimePersistSessionMediaRuntimeFeedback_RateLimitedVideoSetsCooldown(t *testing.T) {
	account := &Account{
		ID:       88,
		Platform: PlatformGrok,
		Type:     AccountTypeSession,
		Extra: map[string]any{
			"grok": map[string]any{
				"tier": map[string]any{
					"normalized": "super",
				},
			},
		},
	}
	repo := &grokRuntimeFeedbackRepoStub{}
	runtime := NewGrokSessionMediaRuntime(&GatewayService{accountRepo: repo}, nil, nil)

	runtime.persistSessionMediaRuntimeFeedback(
		context.Background(),
		account,
		"grok-imagine-video",
		"/v1/videos",
		&grokSessionMediaUpstreamError{
			statusCode: http.StatusTooManyRequests,
			code:       "rate_limit_error",
			message:    "Too many requests",
		},
	)

	require.Empty(t, repo.extraUpdates)
	require.Len(t, repo.runtimeStates, 1)
	runtimeState := repo.runtimeStates[0]
	require.Equal(t, "error", runtimeState["last_outcome"])
	require.Equal(t, http.StatusTooManyRequests, grokParseInt(runtimeState["last_request_status_code"]))
	require.Equal(t, http.StatusTooManyRequests, grokParseInt(runtimeState["last_fail_status_code"]))
	require.Equal(t, string(grokRuntimeErrorClassRateLimited), runtimeState["last_fail_class"])
	require.Equal(t, string(grokRuntimePenaltyScopeAccount), runtimeState["last_fail_scope"])
	require.Equal(t, true, runtimeState["last_fail_retryable"])
	require.Equal(t, string(grok.ProtocolFamilyMediaJob), runtimeState["last_request_protocol_family"])
	require.Equal(t, string(grok.CapabilityVideo), runtimeState["last_request_capability"])
	require.NotEmpty(t, runtimeState["selection_cooldown_until"])
}
