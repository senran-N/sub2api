package service

import (
	"context"

	"github.com/senran-N/sub2api/internal/pkg/grok"
)

// CompatibleGatewayProtocolFamily exposes neutral names for shared compatible
// runtime feedback so OpenAI/Grok sibling handlers do not need Grok-owned
// protocol constants at the call site.
type CompatibleGatewayProtocolFamily = grok.ProtocolFamily

const (
	CompatibleGatewayProtocolFamilyChatCompletions CompatibleGatewayProtocolFamily = grok.ProtocolFamilyChatCompletions
	CompatibleGatewayProtocolFamilyResponses       CompatibleGatewayProtocolFamily = grok.ProtocolFamilyResponses
	CompatibleGatewayProtocolFamilyMessages        CompatibleGatewayProtocolFamily = grok.ProtocolFamilyMessages
)

type CompatibleGatewayRuntimeFeedbackInput struct {
	Account        *Account
	RequestedModel string
	UpstreamModel  string
	Result         *OpenAIForwardResult
	StatusCode     int
	ProtocolFamily CompatibleGatewayProtocolFamily
	Endpoint       string
	Err            error
}

func (s *OpenAIGatewayService) PersistCompatibleGatewayRuntimeFeedback(ctx context.Context, input CompatibleGatewayRuntimeFeedbackInput) {
	if s == nil || input.Account == nil {
		return
	}

	switch NormalizeCompatibleGatewayPlatform(input.Account.Platform) {
	case PlatformGrok:
		s.PersistGrokRuntimeFeedback(ctx, GrokRuntimeFeedbackInput{
			Account:        input.Account,
			RequestedModel: input.RequestedModel,
			UpstreamModel:  input.UpstreamModel,
			Result:         input.Result,
			StatusCode:     input.StatusCode,
			ProtocolFamily: input.ProtocolFamily,
			Endpoint:       input.Endpoint,
			Err:            input.Err,
		})
	}
}
