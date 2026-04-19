package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/senran-N/sub2api/internal/pkg/grok"
)

const grokSessionProbeUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36"

func (s *AccountTestService) testGrokAccountConnection(c *gin.Context, account *Account, modelID string, prompt string) error {
	if account == nil || NormalizeCompatibleGatewayPlatform(account.Platform) != PlatformGrok {
		platform := ""
		if account != nil {
			platform = account.Platform
		}
		return s.sendErrorAndEnd(c, fmt.Sprintf("Unsupported compatible platform: %s", platform))
	}

	target, err := resolveGrokTransportTarget(account, s.accountTestBaseURLValidator())
	if err != nil {
		return s.sendErrorAndEnd(c, err.Error())
	}

	switch target.Kind {
	case grokTransportKindCompatible:
		return s.testCompatibleGatewayAPIKeyConnection(c, account, modelID, prompt)
	case grokTransportKindSession:
		return s.testGrokSessionConnection(c, account, modelID, prompt, target)
	default:
		return s.sendErrorAndEnd(c, fmt.Sprintf("Unsupported Grok transport kind: %s", target.Kind))
	}
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

	testModelID := strings.TrimSpace(modelID)
	if testModelID == "" {
		testModelID = grok.DefaultTestModel
	}

	s.prepareTestStream(c)
	payload, err := createGrokSessionTestPayload(testModelID, prompt)
	if err != nil {
		return s.sendErrorAndEnd(c, err.Error())
	}
	payloadBytes, _ := json.Marshal(payload)
	s.sendEvent(c, TestEvent{Type: "test_start", Model: testModelID})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target.URL, bytes.NewReader(payloadBytes))
	if err != nil {
		return s.sendErrorAndEnd(c, "Failed to create request")
	}
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", grokWebBaseURL)
	req.Header.Set("Referer", grokWebBaseURL+"/")
	req.Header.Set("User-Agent", grokSessionProbeUserAgent)
	target.Apply(req)

	resp, err := s.httpUpstream.DoWithTLS(req, accountTestProxyURL(account), account.ID, account.Concurrency, s.tlsFPProfileService.ResolveTLSProfile(account))
	if err != nil {
		s.persistCompatibleGatewayProbeState(ctx, account, testModelID, nil, fmt.Errorf("request failed: %w", err), false)
		return s.sendErrorAndEnd(c, fmt.Sprintf("Request failed: %s", err.Error()))
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(resp.Body)
		errMsg := fmt.Sprintf("API returned %d: %s", resp.StatusCode, string(body))
		s.persistCompatibleGatewayProbeState(ctx, account, testModelID, resp, fmt.Errorf("%s", errMsg), false)
		if s.accountRepo != nil && (resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden) {
			authErrMsg := fmt.Sprintf("Authentication failed (%d): %s", resp.StatusCode, string(body))
			_ = s.accountRepo.SetError(ctx, account.ID, authErrMsg)
		}
		return s.sendErrorAndEnd(c, errMsg)
	}

	s.persistCompatibleGatewayProbeState(ctx, account, testModelID, resp, nil, false)
	s.sendEvent(c, TestEvent{Type: "test_complete", Success: true})
	return nil
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

	payload["disableSearch"] = true
	payload["sendFinalMetadata"] = false
	return payload, nil
}
