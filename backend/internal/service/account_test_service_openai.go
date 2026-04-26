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

const defaultOpenAIImageTestPrompt = "Generate a clean square icon of a glass lighthouse at sunrise."

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

func (s *AccountTestService) testOpenAIImageAPIKey(c *gin.Context, ctx context.Context, account *Account, modelID, prompt string) error {
	authToken := account.GetOpenAIApiKey()
	if strings.TrimSpace(authToken) == "" {
		return s.sendErrorAndEnd(c, "No API key available")
	}

	baseURL := strings.TrimSpace(account.GetOpenAIBaseURL())
	if baseURL == "" {
		baseURL = CompatibleGatewayDefaultBaseURL(account.Platform)
	}
	normalizedBaseURL, err := s.validateUpstreamBaseURL(baseURL)
	if err != nil {
		return s.sendErrorAndEnd(c, fmt.Sprintf("Invalid base URL: %s", err.Error()))
	}
	target := newCompatiblePassthroughUpstreamTargetWithOptions(
		normalizedBaseURL,
		openAIImagesGenerationsEndpoint,
		account.GetCompatibleAuthMode(""),
		"",
		"",
	)

	s.prepareTestStream(c)
	s.sendEvent(c, TestEvent{Type: "test_start", Model: modelID})

	payload := map[string]any{
		"model":           modelID,
		"prompt":          prompt,
		"n":               1,
		"response_format": "b64_json",
	}
	payloadBytes, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target.URL, bytes.NewReader(payloadBytes))
	if err != nil {
		return s.sendErrorAndEnd(c, "Failed to create request")
	}
	req.Header.Set("Content-Type", "application/json")
	target.ApplyAuthHeader(req.Header, authToken)

	resp, err := s.httpUpstream.DoWithTLS(req, accountTestProxyURL(account), account.ID, account.Concurrency, s.tlsFPProfileService.ResolveTLSProfile(account))
	if err != nil {
		return s.sendErrorAndEnd(c, fmt.Sprintf("Request failed: %s", err.Error()))
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return s.sendErrorAndEnd(c, fmt.Sprintf("Failed to read response: %s", err.Error()))
	}
	if resp.StatusCode != http.StatusOK {
		return s.sendErrorAndEnd(c, fmt.Sprintf("API returned %d: %s", resp.StatusCode, string(body)))
	}

	var result struct {
		Data []struct {
			B64JSON       string `json:"b64_json"`
			URL           string `json:"url"`
			RevisedPrompt string `json:"revised_prompt"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return s.sendErrorAndEnd(c, fmt.Sprintf("Failed to parse response: %s", err.Error()))
	}
	if len(result.Data) == 0 {
		return s.sendErrorAndEnd(c, "No images returned from API")
	}
	for _, item := range result.Data {
		if strings.TrimSpace(item.RevisedPrompt) != "" {
			s.sendEvent(c, TestEvent{Type: "content", Text: item.RevisedPrompt})
		}
		switch {
		case strings.TrimSpace(item.B64JSON) != "":
			s.sendEvent(c, TestEvent{
				Type:     "image",
				ImageURL: "data:image/png;base64," + strings.TrimSpace(item.B64JSON),
				MimeType: "image/png",
			})
		case strings.TrimSpace(item.URL) != "":
			s.sendEvent(c, TestEvent{
				Type:     "image",
				ImageURL: strings.TrimSpace(item.URL),
				MimeType: "image/png",
			})
		}
	}
	s.sendEvent(c, TestEvent{Type: "test_complete", Success: true})
	return nil
}

func (s *AccountTestService) testOpenAIImageOAuth(c *gin.Context, ctx context.Context, account *Account, modelID, prompt string) error {
	authToken := account.GetOpenAIAccessToken()
	if strings.TrimSpace(authToken) == "" {
		return s.sendErrorAndEnd(c, "No access token available")
	}

	s.prepareTestStream(c)
	s.sendEvent(c, TestEvent{Type: "test_start", Model: modelID})

	parsed := &OpenAIImagesRequest{
		Endpoint: openAIImagesGenerationsEndpoint,
		Model:    strings.TrimSpace(modelID),
		Prompt:   prompt,
	}
	applyOpenAIImagesDefaults(parsed)
	responsesBody, err := buildOpenAIImagesResponsesRequest(parsed, parsed.Model)
	if err != nil {
		return s.sendErrorAndEnd(c, fmt.Sprintf("Failed to build image request: %s", err.Error()))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, chatgptCodexAPIURL, bytes.NewReader(responsesBody))
	if err != nil {
		return s.sendErrorAndEnd(c, "Failed to create request")
	}
	req.Host = "chatgpt.com"
	req.Header.Set("Authorization", "Bearer "+authToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("OpenAI-Beta", "responses=experimental")
	req.Header.Set("Originator", "opencode")
	if customUA := strings.TrimSpace(account.GetOpenAIUserAgent()); customUA != "" {
		req.Header.Set("User-Agent", customUA)
	} else {
		req.Header.Set("User-Agent", codexCLIUserAgent)
	}
	if chatgptAccountID := strings.TrimSpace(account.GetChatGPTAccountID()); chatgptAccountID != "" {
		req.Header.Set("chatgpt-account-id", chatgptAccountID)
	}

	resp, err := s.httpUpstream.DoWithTLS(req, accountTestProxyURL(account), account.ID, account.Concurrency, s.tlsFPProfileService.ResolveTLSProfile(account))
	if err != nil {
		return s.sendErrorAndEnd(c, fmt.Sprintf("Responses API request failed: %s", err.Error()))
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return s.sendErrorAndEnd(c, fmt.Sprintf("Failed to read image response: %s", err.Error()))
	}
	if resp.StatusCode >= 400 {
		message := strings.TrimSpace(extractUpstreamErrorMessage(body))
		if message == "" {
			message = fmt.Sprintf("Responses API returned %d", resp.StatusCode)
		}
		return s.sendErrorAndEnd(c, message)
	}

	results, _, _, _, _, err := collectOpenAIImagesFromResponsesBody(body)
	if err != nil {
		return s.sendErrorAndEnd(c, fmt.Sprintf("Failed to parse image response: %s", err.Error()))
	}
	if len(results) == 0 {
		return s.sendErrorAndEnd(c, "No images returned from responses API")
	}
	for _, item := range results {
		if strings.TrimSpace(item.RevisedPrompt) != "" {
			s.sendEvent(c, TestEvent{Type: "content", Text: item.RevisedPrompt})
		}
		mimeType := openAIImageOutputMIMEType(item.OutputFormat)
		s.sendEvent(c, TestEvent{
			Type:     "image",
			ImageURL: "data:" + mimeType + ";base64," + strings.TrimSpace(item.Result),
			MimeType: mimeType,
		})
	}
	s.sendEvent(c, TestEvent{Type: "test_complete", Success: true})
	return nil
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
