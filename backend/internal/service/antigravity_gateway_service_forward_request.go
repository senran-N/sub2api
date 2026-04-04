package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/senran-N/sub2api/internal/pkg/antigravity"
	"github.com/senran-N/sub2api/internal/pkg/logger"
)

const antigravityUpstreamStreamAction = "streamGenerateContent"

var errAntigravityTokenProviderNotConfigured = errors.New("antigravity token provider not configured")

type antigravityForwardTransportContext struct {
	accessToken string
	projectID   string
	proxyURL    string
}

func (s *AntigravityGatewayService) resolveForwardTransportContext(ctx context.Context, account *Account) (*antigravityForwardTransportContext, error) {
	if s.tokenProvider == nil {
		return nil, errAntigravityTokenProviderNotConfigured
	}

	accessToken, err := s.tokenProvider.GetAccessToken(ctx, account)
	if err != nil {
		return nil, err
	}

	return &antigravityForwardTransportContext{
		accessToken: accessToken,
		projectID:   strings.TrimSpace(account.GetCredential("project_id")),
		proxyURL:    antigravityProxyURL(account),
	}, nil
}

func antigravityProxyURL(account *Account) string {
	if account == nil || account.ProxyID == nil || account.Proxy == nil {
		return ""
	}
	return account.Proxy.URL()
}

func (s *AntigravityGatewayService) resolveClaudeForwardModels(account *Account, claudeReq *antigravity.ClaudeRequest) (originalModel, mappedModel, billingModel string) {
	originalModel = claudeReq.Model
	mappedModel = s.getMappedModel(account, claudeReq.Model)
	if mappedModel == "" {
		return originalModel, "", ""
	}

	thinkingEnabled := claudeReq.Thinking != nil && (claudeReq.Thinking.Type == "enabled" || claudeReq.Thinking.Type == "adaptive")
	mappedModel = applyThinkingModelSuffix(mappedModel, thinkingEnabled)
	return originalModel, mappedModel, mappedModel
}

func (s *AntigravityGatewayService) buildClaudeUpstreamBody(ctx context.Context, claudeReq *antigravity.ClaudeRequest, projectID, mappedModel string) ([]byte, antigravity.TransformOptions, error) {
	transformOpts := s.getClaudeTransformOptions(ctx)
	transformOpts.EnableIdentityPatch = true

	geminiBody, err := antigravity.TransformClaudeToGeminiWithOptions(claudeReq, projectID, mappedModel, transformOpts)
	if err != nil {
		return nil, antigravity.TransformOptions{}, err
	}

	return geminiBody, transformOpts, nil
}

func (s *AntigravityGatewayService) resolveGeminiForwardModels(account *Account, originalModel string) (mappedModel, billingModel string) {
	mappedModel = s.getMappedModel(account, originalModel)
	if mappedModel == "" {
		return "", ""
	}
	return mappedModel, mappedModel
}

func (s *AntigravityGatewayService) buildGeminiUpstreamBody(account *Account, body []byte, projectID, mappedModel string) (injectedBody, wrappedBody []byte, err error) {
	injectedBody, err = injectIdentityPatchToGeminiRequest(body)
	if err != nil {
		return nil, nil, err
	}

	if cleanedBody, cleanErr := cleanGeminiRequest(injectedBody); cleanErr == nil {
		injectedBody = cleanedBody
		logger.LegacyPrintf("service.antigravity_gateway", "[Antigravity] Cleaned request schema in forwarded request for account %s", account.Name)
	} else {
		logger.LegacyPrintf("service.antigravity_gateway", "[Antigravity] Failed to clean schema: %v", cleanErr)
	}

	wrappedBody, err = s.wrapV1InternalRequest(projectID, mappedModel, injectedBody)
	if err != nil {
		return nil, nil, err
	}

	return injectedBody, wrappedBody, nil
}

func (s *AntigravityGatewayService) newRetryLoopParams(
	ctx context.Context,
	prefix string,
	account *Account,
	transport *antigravityForwardTransportContext,
	action string,
	body []byte,
	c *gin.Context,
	requestedModel string,
	isStickySession bool,
	handleError func(ctx context.Context, prefix string, account *Account, statusCode int, headers http.Header, body []byte, requestedModel string, groupID int64, sessionHash string, isStickySession bool) *handleModelRateLimitResult,
) antigravityRetryLoopParams {
	return antigravityRetryLoopParams{
		ctx:             ctx,
		prefix:          prefix,
		account:         account,
		proxyURL:        transport.proxyURL,
		accessToken:     transport.accessToken,
		action:          action,
		body:            body,
		c:               c,
		httpUpstream:    s.httpUpstream,
		settingService:  s.settingService,
		accountRepo:     s.accountRepo,
		handleError:     handleError,
		requestedModel:  requestedModel,
		isStickySession: isStickySession,
		groupID:         0,
		sessionHash:     "",
	}
}

func parseClaudeForwardRequest(body []byte) (antigravity.ClaudeRequest, error) {
	var claudeReq antigravity.ClaudeRequest
	if err := json.Unmarshal(body, &claudeReq); err != nil {
		return antigravity.ClaudeRequest{}, fmt.Errorf("parse claude request: %w", err)
	}
	return claudeReq, nil
}
