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
	"github.com/senran-N/sub2api/internal/pkg/openai"
)

// testOpenAIAccountConnection tests an OpenAI account's connection
func (s *AccountTestService) testOpenAIAccountConnection(c *gin.Context, account *Account, modelID string, prompt string) error {
	ctx := c.Request.Context()

	testModelID := modelID
	if testModelID == "" {
		testModelID = openai.DefaultTestModel
	}
	if account.Type == AccountTypeAPIKey || account.Type == AccountTypeUpstream {
		testModelID = resolveOpenAIForwardModel(account, testModelID, "")
	}

	var (
		authToken        string
		apiURL           string
		isOAuth          bool
		chatgptAccountID string
	)
	if account.IsOAuth() {
		isOAuth = true
		authToken = account.GetOpenAIAccessToken()
		if authToken == "" {
			return s.sendErrorAndEnd(c, "No access token available")
		}
		apiURL = chatgptCodexAPIURL
		chatgptAccountID = account.GetChatGPTAccountID()
	} else if account.Type == AccountTypeAPIKey || account.Type == AccountTypeUpstream {
		authToken = account.GetOpenAIApiKey()
		if authToken == "" {
			return s.sendErrorAndEnd(c, "No API key available")
		}

		baseURL := account.GetOpenAIBaseURL()
		if baseURL == "" {
			baseURL = "https://api.openai.com"
		}
		normalizedBaseURL, err := s.validateUpstreamBaseURL(baseURL)
		if err != nil {
			return s.sendErrorAndEnd(c, fmt.Sprintf("Invalid base URL: %s", err.Error()))
		}
		apiURL = newOpenAIResponsesUpstreamTargetWithOptions(
			normalizedBaseURL,
			account.GetCompatibleAuthMode(""),
			account.GetCompatibleEndpointOverride("responses"),
		).URL
	} else {
		return s.sendErrorAndEnd(c, fmt.Sprintf("Unsupported account type: %s", account.Type))
	}

	s.prepareTestStream(c)
	payload := createOpenAITestPayload(testModelID, isOAuth, prompt)
	payloadBytes, _ := json.Marshal(payload)
	s.sendEvent(c, TestEvent{Type: "test_start", Model: testModelID})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(payloadBytes))
	if err != nil {
		return s.sendErrorAndEnd(c, "Failed to create request")
	}
	req.Header.Set("Content-Type", "application/json")
	if isOAuth {
		req.Header.Set("Authorization", "Bearer "+authToken)
	} else {
		newOpenAIResponsesUpstreamTargetWithOptions(
			account.GetOpenAIBaseURL(),
			account.GetCompatibleAuthMode(""),
			account.GetCompatibleEndpointOverride("responses"),
		).ApplyAuthHeader(req.Header, authToken)
	}
	if isOAuth {
		req.Host = "chatgpt.com"
		req.Header.Set("accept", "text/event-stream")
		if chatgptAccountID != "" {
			req.Header.Set("chatgpt-account-id", chatgptAccountID)
		}
	}

	resp, err := s.httpUpstream.DoWithTLS(req, accountTestProxyURL(account), account.ID, account.Concurrency, s.tlsFPProfileService.ResolveTLSProfile(account))
	if err != nil {
		return s.sendErrorAndEnd(c, fmt.Sprintf("Request failed: %s", err.Error()))
	}
	defer func() { _ = resp.Body.Close() }()

	s.handleOpenAIOAuthProbeState(ctx, account, resp, isOAuth)
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		errMsg := fmt.Sprintf("API returned %d: %s", resp.StatusCode, string(body))
		if isOAuth && s.accountRepo != nil {
			if resetAt := (&RateLimitService{}).calculateOpenAI429ResetTime(resp.Header); resetAt != nil {
				_ = s.accountRepo.SetRateLimited(ctx, account.ID, *resetAt)
				account.RateLimitResetAt = resetAt
			}
		}
		if s.accountRepo != nil && (resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden) {
			authErrMsg := fmt.Sprintf("Authentication failed (%d): %s", resp.StatusCode, string(body))
			_ = s.accountRepo.SetError(ctx, account.ID, authErrMsg)
		}
		return s.sendErrorAndEnd(c, errMsg)
	}
	return s.processOpenAIStream(c, resp.Body)
}

func (s *AccountTestService) handleOpenAIOAuthProbeState(ctx context.Context, account *Account, resp *http.Response, isOAuth bool) {
	if !isOAuth || s.accountRepo == nil || resp == nil || account == nil {
		return
	}
	if updates, err := extractOpenAICodexProbeUpdates(resp); err == nil && len(updates) > 0 {
		_ = s.accountRepo.UpdateExtra(ctx, account.ID, updates)
		mergeAccountExtra(account, updates)
	}
	if snapshot := ParseCodexRateLimitHeaders(resp.Header); snapshot != nil {
		if resetAt := codexRateLimitResetAtFromSnapshot(snapshot, time.Now()); resetAt != nil {
			_ = s.accountRepo.SetRateLimited(ctx, account.ID, *resetAt)
			account.RateLimitResetAt = resetAt
		}
	}
}

// createOpenAITestPayload creates a test payload for OpenAI Responses API
func createOpenAITestPayload(modelID string, isOAuth bool, prompt string) map[string]any {
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
