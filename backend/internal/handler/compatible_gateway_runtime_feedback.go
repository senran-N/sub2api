package handler

import (
	"context"

	"github.com/senran-N/sub2api/internal/service"
)

type compatibleGatewayRuntimeFeedbackInput struct {
	Account        *service.Account
	RequestedModel string
	UpstreamModel  string
	Result         *service.OpenAIForwardResult
	StatusCode     int
	ProtocolFamily service.CompatibleGatewayProtocolFamily
	Endpoint       string
	Err            error
}

func persistCompatibleGatewayRuntimeFeedback(
	ctx context.Context,
	gatewayService *service.OpenAIGatewayService,
	input compatibleGatewayRuntimeFeedbackInput,
) {
	if gatewayService == nil || input.Account == nil {
		return
	}

	gatewayService.PersistCompatibleGatewayRuntimeFeedback(ctx, service.CompatibleGatewayRuntimeFeedbackInput{
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

func (h *OpenAIGatewayHandler) persistCompatibleGatewayRuntimeFeedback(ctx context.Context, input compatibleGatewayRuntimeFeedbackInput) {
	if h == nil {
		return
	}
	persistCompatibleGatewayRuntimeFeedback(ctx, h.gatewayService, input)
}

func (h *CompatibleGatewayTextHandler) persistCompatibleGatewayRuntimeFeedback(ctx context.Context, input compatibleGatewayRuntimeFeedbackInput) {
	if h == nil {
		return
	}
	persistCompatibleGatewayRuntimeFeedback(ctx, h.gatewayService, input)
}
