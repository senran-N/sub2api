//go:build unit

package service

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOpenAIGatewayService_PersistCompatibleGatewayRuntimeFeedbackDelegatesToGrokRuntimeState(t *testing.T) {
	account := &Account{
		ID:       91,
		Platform: PlatformGrok,
		Type:     AccountTypeAPIKey,
	}
	repo := &grokRuntimeFeedbackRepoStub{}
	svc := &OpenAIGatewayService{accountRepo: repo}

	svc.PersistCompatibleGatewayRuntimeFeedback(context.Background(), CompatibleGatewayRuntimeFeedbackInput{
		Account:        account,
		RequestedModel: "grok-4.20-fast",
		UpstreamModel:  "grok-4.20-fast",
		Result: &OpenAIForwardResult{
			Model:         "grok-4.20-fast",
			UpstreamModel: "grok-4.20-fast",
		},
		StatusCode:     http.StatusOK,
		ProtocolFamily: CompatibleGatewayProtocolFamilyResponses,
	})

	require.Len(t, repo.extraUpdates, 1)
	require.Len(t, repo.runtimeStates, 1)
	require.Equal(t, "grok-3-fast", repo.runtimeStates[0]["last_request_upstream_model"])
	require.Equal(t, "responses", repo.runtimeStates[0]["last_request_protocol_family"])
	accountGrok := account.grokExtraMap()
	require.ElementsMatch(t, []string{"chat"}, grokParseStringSlice(getNestedGrokValue(accountGrok, "capabilities", "operations")))
}

func TestOpenAIGatewayService_PersistCompatibleGatewayRuntimeFeedbackOpenAINoop(t *testing.T) {
	account := &Account{
		ID:       92,
		Platform: PlatformOpenAI,
		Type:     AccountTypeAPIKey,
	}
	repo := &grokRuntimeFeedbackRepoStub{}
	svc := &OpenAIGatewayService{accountRepo: repo}

	svc.PersistCompatibleGatewayRuntimeFeedback(context.Background(), CompatibleGatewayRuntimeFeedbackInput{
		Account:        account,
		RequestedModel: "gpt-5.1",
		UpstreamModel:  "gpt-5.1",
		StatusCode:     http.StatusBadRequest,
		ProtocolFamily: CompatibleGatewayProtocolFamilyResponses,
		Err:            errors.New("boom"),
	})

	require.Empty(t, repo.extraUpdates)
	require.Empty(t, repo.runtimeStates)
	require.Nil(t, account.Extra)
}
