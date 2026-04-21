package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/senran-N/sub2api/internal/pkg/apicompat"
	"github.com/senran-N/sub2api/internal/pkg/grok"
)

func (s *AccountTestService) testGrokAccountConnection(c *gin.Context, account *Account, modelID string, prompt string) error {
	if account == nil || NormalizeCompatibleGatewayPlatform(account.Platform) != PlatformGrok {
		platform := ""
		if account != nil {
			platform = account.Platform
		}
		return s.sendErrorAndEnd(c, fmt.Sprintf("Unsupported compatible platform: %s", platform))
	}

	runtimeSettings := s.currentGrokRuntimeSettings(c.Request.Context())
	target, err := resolveGrokTransportTargetWithSettings(account, s.accountTestBaseURLValidator(), runtimeSettings)
	if err != nil {
		return s.sendErrorAndEnd(c, err.Error())
	}

	switch target.Kind {
	case grokTransportKindCompatible:
		return s.testGrokCompatibleConnection(c, account, modelID, prompt, target)
	case grokTransportKindSession:
		return s.testGrokSessionConnection(c, account, modelID, prompt, target)
	default:
		return s.sendErrorAndEnd(c, fmt.Sprintf("Unsupported Grok transport kind: %s", target.Kind))
	}
}

func (s *AccountTestService) testGrokCompatibleConnection(
	c *gin.Context,
	account *Account,
	modelID string,
	prompt string,
	target grokTransportTarget,
) error {
	ctx := c.Request.Context()
	probeModels := grokCapabilityProbeModelCandidates(account, modelID)
	if len(probeModels) == 0 {
		probeModels = []string{grok.DefaultTestModel}
	}

	s.prepareTestStream(c)

	var (
		lastModel   string
		lastResp    *http.Response
		lastErr     error
		lastErrMsg  string
		authErrBody string
	)

	for _, candidateModel := range probeModels {
		testModelID := candidateModel
		if account.Type == AccountTypeAPIKey || account.Type == AccountTypeUpstream {
			testModelID = resolveOpenAIForwardModel(account, testModelID, "")
		}
		lastModel = testModelID

		payload := createCompatibleGatewayTestPayload(testModelID, false, prompt)
		payloadBytes, _ := json.Marshal(payload)
		s.sendEvent(c, TestEvent{Type: "test_start", Model: testModelID})

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, target.URL, bytes.NewReader(payloadBytes))
		if err != nil {
			lastResp = nil
			lastErr = fmt.Errorf("create request: %w", err)
			lastErrMsg = "Failed to create request"
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		target.Apply(req)

		resp, err := s.httpUpstream.DoWithTLS(req, accountTestProxyURL(account), account.ID, account.Concurrency, s.tlsFPProfileService.ResolveTLSProfile(account))
		if err != nil {
			lastResp = nil
			lastErr = fmt.Errorf("request failed: %w", err)
			lastErrMsg = fmt.Sprintf("Request failed: %s", err.Error())
			continue
		}

		if resp.StatusCode != http.StatusOK {
			body := grokReadProbeErrorBody(resp)
			_ = resp.Body.Close()
			lastResp = resp
			errSummary := grokSummarizeProbeHTTPError(resp, body)
			lastErrMsg = errSummary.Message
			lastErr = fmt.Errorf("%s", lastErrMsg)
			authErrBody = errSummary.AuthenticationMessage
			continue
		}

		s.persistCompatibleGatewayProbeState(ctx, account, testModelID, resp, nil, false)
		defer func() { _ = resp.Body.Close() }()
		return s.processCompatibleResponsesStream(c, resp.Body)
	}

	if lastModel == "" {
		lastModel = grok.DefaultTestModel
	}
	if lastErrMsg == "" && lastErr != nil {
		lastErrMsg = lastErr.Error()
	}
	if lastErrMsg == "" {
		lastErrMsg = "Grok probe failed"
	}
	s.persistCompatibleGatewayProbeState(ctx, account, lastModel, lastResp, lastErr, false)
	if s.accountRepo != nil && strings.TrimSpace(authErrBody) != "" {
		_ = s.accountRepo.SetError(ctx, account.ID, authErrBody)
	}
	return s.sendErrorAndEnd(c, lastErrMsg)
}

func (s *AccountTestService) accountTestBaseURLValidator() func(string) (string, error) {
	if s == nil || s.cfg == nil {
		return nil
	}
	return s.validateUpstreamBaseURL
}

func (s *AccountTestService) testGrokSessionConnection(
	c *gin.Context,
	account *Account,
	modelID string,
	prompt string,
	target grokTransportTarget,
) error {
	ctx := c.Request.Context()
	s.prepareTestStream(c)

	probeModels := grokCapabilityProbeModelCandidates(account, modelID)
	if len(probeModels) == 0 {
		probeModels = []string{grok.DefaultTestModel}
	}

	var (
		lastModel   string
		lastResp    *http.Response
		lastErr     error
		lastErrMsg  string
		authErrBody string
	)

	for _, testModelID := range probeModels {
		lastModel = testModelID
		payload, err := createGrokSessionTestPayload(testModelID, prompt)
		if err != nil {
			lastResp = nil
			lastErr = err
			lastErrMsg = err.Error()
			continue
		}
		payloadBytes, _ := json.Marshal(payload)
		s.sendEvent(c, TestEvent{Type: "test_start", Model: testModelID})

		req, err := newGrokSessionJSONRequest(
			ctx,
			http.MethodPost,
			target,
			payloadBytes,
			grokSessionProbeAcceptHeader,
		)
		if err != nil {
			lastResp = nil
			lastErr = fmt.Errorf("create request: %w", err)
			lastErrMsg = "Failed to create request"
			continue
		}

		resp, err := s.httpUpstream.DoWithTLS(req, accountTestProxyURL(account), account.ID, account.Concurrency, s.tlsFPProfileService.ResolveTLSProfile(account))
		if err != nil {
			lastResp = nil
			lastErr = fmt.Errorf("request failed: %w", err)
			lastErrMsg = fmt.Sprintf("Request failed: %s", err.Error())
			continue
		}

		if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
			body := grokReadProbeErrorBody(resp)
			_ = resp.Body.Close()
			lastResp = resp
			errSummary := grokSummarizeProbeHTTPError(resp, body)
			lastErrMsg = errSummary.Message
			lastErr = fmt.Errorf("%s", lastErrMsg)
			authErrBody = errSummary.AuthenticationMessage
			continue
		}

		s.persistCompatibleGatewayProbeState(ctx, account, testModelID, resp, nil, false)
		defer func() { _ = resp.Body.Close() }()
		return s.processGrokSessionTestStream(c, resp.Body, testModelID)
	}

	if lastModel == "" {
		lastModel = grok.DefaultTestModel
	}
	if lastErrMsg == "" && lastErr != nil {
		lastErrMsg = lastErr.Error()
	}
	if lastErrMsg == "" {
		lastErrMsg = "Grok session probe failed"
	}
	s.persistCompatibleGatewayProbeState(ctx, account, lastModel, lastResp, lastErr, false)
	if s.accountRepo != nil && strings.TrimSpace(authErrBody) != "" {
		_ = s.accountRepo.SetError(ctx, account.ID, authErrBody)
	}
	return s.sendErrorAndEnd(c, lastErrMsg)
}

func createGrokSessionTestPayload(modelID string, prompt string) (map[string]any, error) {
	testPrompt := strings.TrimSpace(prompt)
	if testPrompt == "" {
		testPrompt = defaultTestPrompt
	}

	payload, err := buildGrokSessionTextPayload(grokSessionTextRequest{
		ModelID: modelID,
		Message: testPrompt,
	})
	if err != nil {
		return nil, err
	}
	return payload, nil
}

func (s *AccountTestService) processGrokSessionTestStream(c *gin.Context, body io.Reader, model string) error {
	sawContent := false

	finalResponse, _, err := collectGrokSessionResponses(body, model, nil, &grokSessionResponsesCallbacks{
		onEvent: func(event apicompat.ResponsesStreamEvent) error {
			if event.Type == "response.output_text.delta" && event.Delta != "" {
				sawContent = true
				s.sendEvent(c, TestEvent{Type: "content", Text: event.Delta})
			}
			return nil
		},
		onStreamFailure: func(message string, _ apicompat.ResponsesStreamEvent) error {
			return s.sendErrorAndEnd(c, message)
		},
	})
	if errors.Is(err, errGrokSessionTextStreamHandled) {
		return nil
	}
	if err != nil {
		return err
	}

	if !sawContent {
		text := strings.TrimSpace(extractResponsesOutputText(finalResponse))
		if text == "" {
			text = "(empty response)"
		}
		s.sendEvent(c, TestEvent{Type: "content", Text: text})
	}

	s.sendEvent(c, TestEvent{Type: "test_complete", Success: true})
	return nil
}

func extractResponsesOutputText(resp *apicompat.ResponsesResponse) string {
	if resp == nil || len(resp.Output) == 0 {
		return ""
	}

	var builder strings.Builder
	for _, output := range resp.Output {
		switch output.Type {
		case "message":
			for _, content := range output.Content {
				if content.Type != "output_text" {
					continue
				}
				builder.WriteString(content.Text)
			}
		case "reasoning":
			for _, summary := range output.Summary {
				if summary.Type != "summary_text" {
					continue
				}
				builder.WriteString(summary.Text)
			}
		}
	}
	return builder.String()
}
