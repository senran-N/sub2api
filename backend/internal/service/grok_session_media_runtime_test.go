//go:build unit

package service

import (
	"context"
	"net/http"
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
