package service

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/senran-N/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type compatibleModelsAccountRepoStub struct {
	AccountRepository
	accounts []Account
}

func (r compatibleModelsAccountRepoStub) ListSchedulable(ctx context.Context) ([]Account, error) {
	return append([]Account(nil), r.accounts...), nil
}

func (r compatibleModelsAccountRepoStub) ListSchedulableByGroupID(ctx context.Context, groupID int64) ([]Account, error) {
	return append([]Account(nil), r.accounts...), nil
}

func TestCompatibleModelsEndpointURLPreservesBasePath(t *testing.T) {
	account := &Account{
		Platform: PlatformAnthropic,
		Type:     AccountTypeAPIKey,
		Credentials: map[string]any{
			"base_url": "https://proxy.example.com/anthropic",
		},
	}

	require.Equal(
		t,
		"https://proxy.example.com/anthropic/v1/models",
		compatibleModelsEndpointURL(account, account.GetCompatibleBaseURL()),
	)
}

func TestParseCompatibleUpstreamModelsSupportsCommonShapes(t *testing.T) {
	tests := []struct {
		name string
		body string
		want []CompatibleUpstreamModel
	}{
		{
			name: "data envelope",
			body: `{"data":[{"id":"gpt-4.1","object":"model","owned_by":"openai"}]}`,
			want: []CompatibleUpstreamModel{{
				ID:          "gpt-4.1",
				Object:      "model",
				Type:        "model",
				DisplayName: "gpt-4.1",
				OwnedBy:     "openai",
			}},
		},
		{
			name: "models envelope",
			body: `{"models":[{"name":"claude-sonnet-4-5","display_name":"Claude Sonnet 4.5","created_at":"2025-09-29T00:00:00Z"}]}`,
			want: []CompatibleUpstreamModel{{
				ID:          "claude-sonnet-4-5",
				Object:      "model",
				Type:        "model",
				DisplayName: "Claude Sonnet 4.5",
				CreatedAt:   "2025-09-29T00:00:00Z",
			}},
		},
		{
			name: "bare array",
			body: `[{"id":"models/gpt-5","created":1735689600}]`,
			want: []CompatibleUpstreamModel{{
				ID:          "gpt-5",
				Object:      "model",
				Type:        "model",
				DisplayName: "gpt-5",
				Created:     1735689600,
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseCompatibleUpstreamModels([]byte(tt.body))
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestParseCompatibleUpstreamModelsRejectsUnsupportedShape(t *testing.T) {
	_, err := parseCompatibleUpstreamModels([]byte(`{"object":"list"}`))
	require.Error(t, err)
}

func TestResolveCompatibleEndpointURL_RelativeOverridePreservesBasePathAndQuery(t *testing.T) {
	require.Equal(
		t,
		"https://proxy.example.com/anthropic/custom/messages?beta=true",
		resolveCompatibleEndpointURL("https://proxy.example.com/anthropic", "/v1/messages", "custom/messages?beta=true"),
	)
}

func TestCompatibleUpstreamModelsServiceDiscoverGroupModelsDeduplicates(t *testing.T) {
	groupID := int64(1)
	repo := compatibleModelsAccountRepoStub{
		accounts: []Account{
			{
				ID:        1,
				Platform:  PlatformOpenAI,
				Type:      AccountTypeAPIKey,
				UpdatedAt: time.Unix(100, 0),
				Credentials: map[string]any{
					"base_url": "https://proxy.example.com/openai-a",
					"api_key":  "sk-a",
				},
			},
			{
				ID:        2,
				Platform:  PlatformOpenAI,
				Type:      AccountTypeUpstream,
				UpdatedAt: time.Unix(200, 0),
				Credentials: map[string]any{
					"base_url": "https://proxy.example.com/openai-b",
					"api_key":  "sk-b",
				},
			},
		},
	}
	upstream := &queuedHTTPUpstream{
		responses: []*http.Response{
			newJSONResponse(http.StatusOK, `{"data":[{"id":"gpt-5"},{"id":"gpt-5-mini"}]}`),
			newJSONResponse(http.StatusOK, `{"data":[{"id":"gpt-5-mini"},{"id":"gpt-5-nano"}]}`),
		},
	}
	svc := NewCompatibleUpstreamModelsService(repo, upstream, &config.Config{}, nil)

	models, err := svc.DiscoverGroupModels(context.Background(), &groupID, PlatformOpenAI)
	require.NoError(t, err)
	require.Equal(t, []CompatibleUpstreamModel{
		{ID: "gpt-5", Object: "model", Type: "model", DisplayName: "gpt-5"},
		{ID: "gpt-5-mini", Object: "model", Type: "model", DisplayName: "gpt-5-mini"},
		{ID: "gpt-5-nano", Object: "model", Type: "model", DisplayName: "gpt-5-nano"},
	}, models)

	require.Len(t, upstream.requests, 2)
	require.Equal(t, "https://proxy.example.com/openai-a/v1/models", upstream.requests[0].URL.String())
	require.Equal(t, "Bearer sk-a", upstream.requests[0].Header.Get("authorization"))
	require.Equal(t, "https://proxy.example.com/openai-b/v1/models", upstream.requests[1].URL.String())
	require.Equal(t, "Bearer sk-b", upstream.requests[1].Header.Get("authorization"))
}
