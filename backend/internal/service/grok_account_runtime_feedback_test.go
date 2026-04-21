//go:build unit

package service

import (
	"context"
	"net/http"
	"testing"

	"github.com/senran-N/sub2api/internal/pkg/grok"
	"github.com/stretchr/testify/require"
)

type grokRuntimeFeedbackRepoStub struct {
	mockAccountRepoForGemini
	extraUpdates  []map[string]any
	runtimeStates []map[string]any
}

func (r *grokRuntimeFeedbackRepoStub) UpdateExtra(_ context.Context, _ int64, updates map[string]any) error {
	r.extraUpdates = append(r.extraUpdates, cloneAnyMap(updates))
	return nil
}

func (r *grokRuntimeFeedbackRepoStub) UpdateGrokExtra(_ context.Context, _ int64, grokPatch map[string]any) error {
	r.extraUpdates = append(r.extraUpdates, map[string]any{"grok": cloneAnyMap(grokPatch)})
	return nil
}

func (r *grokRuntimeFeedbackRepoStub) UpdateGrokRuntimeState(_ context.Context, _ int64, runtimeState map[string]any) error {
	r.runtimeStates = append(r.runtimeStates, cloneAnyMap(runtimeState))
	return nil
}

func TestOpenAIGatewayService_PersistGrokRuntimeFeedbackSuccessEnrichesCapabilities(t *testing.T) {
	account := &Account{
		ID:       71,
		Platform: PlatformGrok,
		Type:     AccountTypeAPIKey,
		Extra: map[string]any{
			"grok": map[string]any{
				"tier": map[string]any{
					"normalized": "basic",
					"source":     "manual",
				},
				"quota_windows": map[string]any{
					grok.QuotaWindowAuto: map[string]any{
						"remaining": 9,
						"total":     20,
					},
				},
				"capabilities": map[string]any{
					"video": false,
				},
				"sync_state": map[string]any{
					"last_sync_at": "2026-04-19T00:00:00Z",
				},
			},
		},
	}
	repo := &grokRuntimeFeedbackRepoStub{}
	svc := &OpenAIGatewayService{accountRepo: repo}

	svc.PersistGrokRuntimeFeedback(context.Background(), GrokRuntimeFeedbackInput{
		Account:        account,
		RequestedModel: "grok-4.20-fast",
		UpstreamModel:  "grok-4.20-fast",
		Result: &OpenAIForwardResult{
			Model:         "grok-4.20-fast",
			UpstreamModel: "grok-4.20-fast",
		},
		StatusCode:     http.StatusOK,
		ProtocolFamily: grok.ProtocolFamilyChatCompletions,
	})

	require.Len(t, repo.extraUpdates, 1)
	require.Len(t, repo.runtimeStates, 1)

	grokExtra := grokExtraMap(repo.extraUpdates[0])
	require.Nil(t, getNestedGrokValue(grokExtra, "tier", "normalized"))
	require.Nil(t, getNestedGrokValue(grokExtra, "sync_state", "last_sync_at"))
	require.Equal(t, false, grokNestedMap(grokExtra["capabilities"])["video"])
	require.ElementsMatch(t, []string{"chat"}, grokParseStringSlice(getNestedGrokValue(grokExtra, "capabilities", "operations")))
	require.ElementsMatch(t, []string{"grok-4.20-fast"}, grokParseStringSlice(getNestedGrokValue(grokExtra, "capabilities", "models")))

	runtimeState := repo.runtimeStates[0]
	require.Equal(t, "success", runtimeState["last_outcome"])
	require.Equal(t, http.StatusOK, grokParseInt(runtimeState["last_request_status_code"]))
	require.Equal(t, "grok-4.20-fast", runtimeState["last_request_model"])
	require.Equal(t, "grok-4.20-fast", runtimeState["last_request_upstream_model"])
	require.Equal(t, string(grok.ProtocolFamilyChatCompletions), runtimeState["last_request_protocol_family"])
	require.Equal(t, string(grok.CapabilityChat), runtimeState["last_request_capability"])
	require.NotEmpty(t, runtimeState["last_request_at"])
	require.NotEmpty(t, runtimeState["last_use_at"])

	accountGrok := account.grokExtraMap()
	require.Equal(t, "basic", getNestedGrokValue(accountGrok, "tier", "normalized"))
	require.Equal(t, "grok-4.20-fast", getNestedGrokValue(accountGrok, "runtime_state", "last_request_upstream_model"))
}

func TestOpenAIGatewayService_PersistGrokRuntimeFeedbackKnownCapabilitySkipsExtraPatch(t *testing.T) {
	account := &Account{
		ID:       72,
		Platform: PlatformGrok,
		Type:     AccountTypeAPIKey,
		Extra: map[string]any{
			"grok": map[string]any{
				"capabilities": map[string]any{
					"models":     []any{"grok-3-fast"},
					"operations": []any{"chat"},
				},
			},
		},
	}
	repo := &grokRuntimeFeedbackRepoStub{}
	svc := &OpenAIGatewayService{accountRepo: repo}

	svc.PersistGrokRuntimeFeedback(context.Background(), GrokRuntimeFeedbackInput{
		Account:        account,
		RequestedModel: "grok-3-fast",
		UpstreamModel:  "grok-3-fast",
		Result: &OpenAIForwardResult{
			Model:         "grok-3-fast",
			UpstreamModel: "grok-3-fast",
		},
		StatusCode:     http.StatusOK,
		ProtocolFamily: grok.ProtocolFamilyChatCompletions,
	})

	require.Empty(t, repo.extraUpdates)
	require.Len(t, repo.runtimeStates, 1)
	require.Equal(t, "success", repo.runtimeStates[0]["last_outcome"])
}

func TestOpenAIGatewayService_PersistGrokRuntimeFeedbackFailoverOnlyUpdatesRuntimeState(t *testing.T) {
	account := &Account{
		ID:       73,
		Platform: PlatformGrok,
		Type:     AccountTypeSession,
		Extra: map[string]any{
			"grok": map[string]any{
				"tier": map[string]any{
					"normalized": "super",
				},
				"quota_windows": map[string]any{
					grok.QuotaWindowAuto: map[string]any{
						"remaining": 3,
						"total":     20,
					},
				},
				"capabilities": map[string]any{
					"models":     []any{"grok-3"},
					"operations": []any{"chat"},
				},
			},
		},
	}
	repo := &grokRuntimeFeedbackRepoStub{}
	svc := &OpenAIGatewayService{accountRepo: repo}

	svc.PersistGrokRuntimeFeedback(context.Background(), GrokRuntimeFeedbackInput{
		Account:        account,
		RequestedModel: "grok-imagine-video",
		UpstreamModel:  "grok-imagine-video",
		StatusCode:     http.StatusForbidden,
		ProtocolFamily: grok.ProtocolFamilyMediaJob,
		Err: &UpstreamFailoverError{
			StatusCode:   http.StatusForbidden,
			ResponseBody: []byte(`{"error":{"message":"video tier required"}}`),
		},
	})

	require.Len(t, repo.extraUpdates, 1)
	require.Len(t, repo.runtimeStates, 1)

	grokExtra := grokExtraMap(repo.extraUpdates[0])
	require.Equal(t, false, grokNestedMap(grokExtra["capabilities"])["video"])
	require.ElementsMatch(t, []string{"chat"}, grokParseStringSlice(getNestedGrokValue(grokExtra, "capabilities", "operations")))
	require.ElementsMatch(t, []string{"grok-4.20-auto"}, grokParseStringSlice(getNestedGrokValue(grokExtra, "capabilities", "models")))

	runtimeState := repo.runtimeStates[0]
	require.Equal(t, "failover", runtimeState["last_outcome"])
	require.Equal(t, http.StatusForbidden, grokParseInt(runtimeState["last_request_status_code"]))
	require.Equal(t, http.StatusForbidden, grokParseInt(runtimeState["last_fail_status_code"]))
	require.Equal(t, "video tier required", runtimeState["last_fail_reason"])
	require.Equal(t, string(grokRuntimeErrorClassModelUnsupported), runtimeState["last_fail_class"])
	require.Equal(t, string(grokRuntimePenaltyScopeModel), runtimeState["last_fail_scope"])
	require.Equal(t, false, runtimeState["last_fail_retryable"])
	require.Equal(t, string(grok.ProtocolFamilyMediaJob), runtimeState["last_request_protocol_family"])
	require.Equal(t, string(grok.CapabilityVideo), runtimeState["last_request_capability"])
	require.NotEmpty(t, runtimeState["last_fail_at"])
	require.NotEmpty(t, runtimeState["last_failover_at"])
	require.Equal(t, "grok-imagine-video", runtimeState["selection_cooldown_model"])
	require.NotEmpty(t, runtimeState["selection_cooldown_until"])

	accountGrok := account.grokExtraMap()
	require.Equal(t, "super", getNestedGrokValue(accountGrok, "tier", "normalized"))
	require.ElementsMatch(t, []string{"grok-4.20-auto"}, grokParseStringSlice(getNestedGrokValue(accountGrok, "capabilities", "models")))
	require.Equal(t, "video tier required", getNestedGrokValue(accountGrok, "runtime_state", "last_fail_reason"))
	require.Equal(t, "model_unsupported", getNestedGrokValue(accountGrok, "runtime_state", "last_fail_class"))
}

func TestOpenAIGatewayService_PersistGrokRuntimeFeedbackOpenAINoop(t *testing.T) {
	account := &Account{
		ID:       74,
		Platform: PlatformOpenAI,
		Type:     AccountTypeAPIKey,
	}
	repo := &grokRuntimeFeedbackRepoStub{}
	svc := &OpenAIGatewayService{accountRepo: repo}

	svc.PersistGrokRuntimeFeedback(context.Background(), GrokRuntimeFeedbackInput{
		Account:        account,
		RequestedModel: "gpt-5.4",
		Result:         &OpenAIForwardResult{Model: "gpt-5.4"},
		StatusCode:     http.StatusOK,
	})

	require.Empty(t, repo.extraUpdates)
	require.Empty(t, repo.runtimeStates)
}
