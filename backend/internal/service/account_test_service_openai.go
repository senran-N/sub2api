package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/senran-N/sub2api/internal/pkg/grok"
	"github.com/senran-N/sub2api/internal/pkg/openai"
)

func compatibleGatewayDefaultTestModel(platform string) string {
	switch NormalizeCompatibleGatewayPlatform(platform) {
	case PlatformGrok:
		return grok.DefaultTestModel
	default:
		return openai.DefaultTestModel
	}
}

func compatibleGatewayUnsupportedSessionTestError(platform string) string {
	_ = platform
	return "Unsupported account type: session"
}

// testCompatibleGatewayAPIKeyConnection tests a shared compatible-gateway API-key/upstream account.
func (s *AccountTestService) testCompatibleGatewayAPIKeyConnection(c *gin.Context, account *Account, modelID string, prompt string) error {
	ctx := c.Request.Context()

	testModelID := modelID
	if testModelID == "" {
		testModelID = compatibleGatewayDefaultTestModel(account.Platform)
	}
	if account.Type == AccountTypeAPIKey || account.Type == AccountTypeUpstream {
		testModelID = resolveOpenAIForwardModel(account, testModelID, "")
	}

	var (
		authToken      string
		apiURL         string
		isOAuth        bool
		upstreamTarget compatibleResponsesUpstreamTarget
	)
	if account.IsOAuth() {
		if !account.IsOpenAI() {
			return s.sendErrorAndEnd(c, fmt.Sprintf("Unsupported account type: %s", account.Type))
		}
		isOAuth = true
		authToken = account.GetOpenAIAccessToken()
		if authToken == "" {
			return s.sendErrorAndEnd(c, "No access token available")
		}
		apiURL = chatgptCodexAPIURL
	} else if account.Type == AccountTypeSession {
		return s.sendErrorAndEnd(c, compatibleGatewayUnsupportedSessionTestError(account.Platform))
	} else if account.Type == AccountTypeAPIKey || account.Type == AccountTypeUpstream {
		authToken = account.GetOpenAIApiKey()
		if authToken == "" {
			return s.sendErrorAndEnd(c, "No API key available")
		}

		baseURL := strings.TrimSpace(account.GetOpenAIBaseURL())
		if baseURL == "" {
			baseURL = CompatibleGatewayDefaultBaseURL(account.Platform)
		}
		if baseURL == "" {
			return s.sendErrorAndEnd(c, fmt.Sprintf("Unsupported compatible platform: %s", account.Platform))
		}
		normalizedBaseURL, err := s.validateUpstreamBaseURL(baseURL)
		if err != nil {
			return s.sendErrorAndEnd(c, fmt.Sprintf("Invalid base URL: %s", err.Error()))
		}
		upstreamTarget = newCompatibleResponsesUpstreamTargetWithOptions(
			normalizedBaseURL,
			account.GetCompatibleAuthMode(""),
			account.GetCompatibleEndpointOverride("responses"),
		)
		apiURL = upstreamTarget.URL
	} else {
		return s.sendErrorAndEnd(c, fmt.Sprintf("Unsupported account type: %s", account.Type))
	}

	s.prepareTestStream(c)
	payload := createCompatibleGatewayTestPayload(testModelID, isOAuth, prompt)
	payloadBytes, _ := json.Marshal(payload)
	s.sendEvent(c, TestEvent{Type: "test_start", Model: testModelID})

	var req *http.Request
	var err error
	if isOAuth {
		req, err = newOpenAICodexOAuthResponsesRequest(ctx, apiURL, authToken, payloadBytes, "text/event-stream", account)
	} else {
		req, err = http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(payloadBytes))
		if err != nil {
			return s.sendErrorAndEnd(c, "Failed to create request")
		}
		req.Header.Set("Content-Type", "application/json")
		upstreamTarget.ApplyAuthHeader(req.Header, authToken)
	}
	if err != nil {
		return s.sendErrorAndEnd(c, "Failed to create request")
	}

	resp, err := s.httpUpstream.DoWithTLS(req, accountTestProxyURL(account), account.ID, account.Concurrency, s.tlsFPProfileService.ResolveTLSProfile(account))
	if err != nil {
		s.persistCompatibleGatewayProbeState(ctx, account, testModelID, nil, fmt.Errorf("request failed: %w", err), isOAuth)
		return s.sendErrorAndEnd(c, fmt.Sprintf("Request failed: %s", err.Error()))
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode == http.StatusTooManyRequests {
			s.reconcileOpenAI429State(ctx, account, resp.Header, body)
		}
		errMsg := fmt.Sprintf("API returned %d: %s", resp.StatusCode, string(body))
		s.persistCompatibleGatewayProbeState(ctx, account, testModelID, resp, fmt.Errorf("%s", errMsg), isOAuth)
		if s.accountRepo != nil && (resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden) {
			authErrMsg := fmt.Sprintf("Authentication failed (%d): %s", resp.StatusCode, string(body))
			_ = s.accountRepo.SetError(ctx, account.ID, authErrMsg)
		}
		return s.sendErrorAndEnd(c, errMsg)
	}
	s.persistCompatibleGatewayProbeState(ctx, account, testModelID, resp, nil, isOAuth)
	return s.processCompatibleResponsesStream(c, resp.Body)
}

func (s *AccountTestService) reconcileOpenAI429State(ctx context.Context, account *Account, headers http.Header, body []byte) {
	if s == nil || s.accountRepo == nil || account == nil || !account.IsOpenAI() {
		return
	}

	var resetAt *time.Time
	if calculated := (&RateLimitService{}).calculateOpenAI429ResetTime(headers); calculated != nil {
		resetAt = calculated
	} else if unixTs := parseOpenAIRateLimitResetTime(body); unixTs != nil {
		t := time.Unix(*unixTs, 0)
		resetAt = &t
	}
	if resetAt == nil {
		return
	}

	if err := s.accountRepo.SetRateLimited(ctx, account.ID, *resetAt); err != nil {
		return
	}
	now := time.Now()
	account.RateLimitedAt = &now
	account.RateLimitResetAt = resetAt

	if account.Status == StatusError {
		if err := s.accountRepo.ClearError(ctx, account.ID); err != nil {
			return
		}
		account.Status = StatusActive
		account.ErrorMessage = ""
	}
}

// createCompatibleGatewayTestPayload creates a test payload for compatible Responses APIs.
func createCompatibleGatewayTestPayload(modelID string, isOAuth bool, prompt string) map[string]any {
	testPrompt := strings.TrimSpace(prompt)
	if testPrompt == "" {
		testPrompt = defaultTestPrompt
	}
	payload := map[string]any{
		"model": modelID,
		"input": []map[string]any{
			{
				"role": "user",
				"content": []map[string]any{
					{
						"type": "input_text",
						"text": testPrompt,
					},
				},
			},
		},
		"stream": true,
	}
	if isOAuth {
		payload["store"] = false
	}
	payload["instructions"] = openai.DefaultInstructions
	return payload
}
