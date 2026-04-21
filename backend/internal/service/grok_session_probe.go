package service

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/senran-N/sub2api/internal/pkg/grok"
)

const grokSessionProbeBodyLimit = 8192

type GrokSessionProbeResult struct {
	ModelID            string
	StatusCode         int
	Body               string
	TargetURL          string
	HasSSO             bool
	HasSSORW           bool
	HasCFClearance     bool
	HasExtraCFCookies  bool
	ResolvedUserAgent  string
	ResolvedAcceptLang string
}

func ProbeGrokSessionConnection(
	ctx context.Context,
	upstream HTTPUpstream,
	account *Account,
	modelID string,
	prompt string,
) (*GrokSessionProbeResult, error) {
	return ProbeGrokSessionConnectionWithSettings(
		ctx,
		upstream,
		account,
		modelID,
		prompt,
		DefaultGrokRuntimeSettings(),
	)
}

func ProbeGrokSessionConnectionWithSettings(
	ctx context.Context,
	upstream HTTPUpstream,
	account *Account,
	modelID string,
	prompt string,
	settings GrokRuntimeSettings,
) (*GrokSessionProbeResult, error) {
	if upstream == nil {
		return nil, errors.New("http upstream is nil")
	}
	if account == nil {
		return nil, errors.New("account is nil")
	}
	if NormalizeCompatibleGatewayPlatform(account.Platform) != PlatformGrok || account.Type != AccountTypeSession {
		return nil, errors.New("account is not a grok session account")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	target, err := resolveGrokTransportTargetWithSettings(account, nil, settings)
	if err != nil {
		return nil, err
	}
	if target.Kind != grokTransportKindSession {
		return nil, errors.New("resolved transport is not grok session")
	}

	testModelID := strings.TrimSpace(modelID)
	if testModelID == "" {
		testModelID = grok.DefaultTestModel
	}

	payload, err := createGrokSessionTestPayload(testModelID, prompt)
	if err != nil {
		return nil, err
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := newGrokSessionJSONRequest(
		ctx,
		http.MethodPost,
		target,
		payloadBytes,
		grokSessionProbeAcceptHeader,
	)
	if err != nil {
		return nil, err
	}

	resp, err := upstream.DoWithTLS(
		req,
		accountTestProxyURL(account),
		account.ID,
		account.Concurrency,
		resolveGrokTLSProfile(account, nil),
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, grokSessionProbeBodyLimit))
	if err != nil {
		return nil, err
	}

	bodyText := strings.TrimSpace(string(body))
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		bodyText = grokSummarizeProbeHTTPError(resp, body).Message
	}

	return &GrokSessionProbeResult{
		ModelID:            testModelID,
		StatusCode:         resp.StatusCode,
		Body:               bodyText,
		TargetURL:          target.URL,
		HasSSO:             strings.Contains(target.CookieHeader, "sso="),
		HasSSORW:           strings.Contains(target.CookieHeader, "sso-rw="),
		HasCFClearance:     strings.Contains(target.CookieHeader, "cf_clearance="),
		HasExtraCFCookies:  strings.TrimSpace(account.GetGrokSessionCFCookies()) != "",
		ResolvedUserAgent:  grokSessionResolvedUserAgent(target),
		ResolvedAcceptLang: grokSessionResolvedAcceptLanguage(target),
	}, nil
}
