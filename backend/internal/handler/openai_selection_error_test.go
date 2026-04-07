package handler

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/senran-N/sub2api/internal/service"
)

func TestOpenAISelectionErrorResponseModelUnavailable(t *testing.T) {
	status, code, message := openAISelectionErrorResponse(
		fmt.Errorf("%w: %s", service.ErrOpenAIRequestedModelUnavailable, "gpt-does-not-exist"),
	)

	if status != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d", status)
	}
	if code != "invalid_request_error" {
		t.Fatalf("unexpected code: %s", code)
	}
	if message != "Requested model is not configured for any available OpenAI account: gpt-does-not-exist" {
		t.Fatalf("unexpected message: %s", message)
	}
}

func TestOpenAISelectionErrorResponseServiceUnavailableFallback(t *testing.T) {
	status, code, message := openAISelectionErrorResponse(fmt.Errorf("no available accounts"))
	if status != http.StatusServiceUnavailable {
		t.Fatalf("unexpected status: %d", status)
	}
	if code != "api_error" {
		t.Fatalf("unexpected code: %s", code)
	}
	if message != "Service temporarily unavailable" {
		t.Fatalf("unexpected message: %s", message)
	}
}
