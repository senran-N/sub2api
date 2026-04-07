package handler

import (
	"net/http"

	"github.com/senran-N/sub2api/internal/service"
)

func openAISelectionErrorResponse(err error) (int, string, string) {
	if service.IsOpenAIRequestedModelUnavailableError(err) {
		model := service.ExtractOpenAIRequestedModelUnavailable(err)
		if model == "" {
			return http.StatusBadRequest, "invalid_request_error", "Requested model is not configured for any available OpenAI account"
		}
		return http.StatusBadRequest, "invalid_request_error", "Requested model is not configured for any available OpenAI account: " + model
	}
	return http.StatusServiceUnavailable, "api_error", "Service temporarily unavailable"
}
