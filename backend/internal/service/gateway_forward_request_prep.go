package service

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/senran-N/sub2api/internal/pkg/logger"
	"github.com/senran-N/sub2api/internal/pkg/tlsfingerprint"
)

type forwardRequestPreparation struct {
	body                  []byte
	reqModel              string
	reqStream             bool
	originalModel         string
	shouldMimicClaudeCode bool
	token                 string
	tokenType             string
	proxyURL              string
	tlsProfile            *tlsfingerprint.Profile
}

func (s *GatewayService) prepareForwardRequest(
	ctx context.Context,
	c *gin.Context,
	account *Account,
	parsed *ParsedRequest,
) (*forwardRequestPreparation, error) {
	preparation := &forwardRequestPreparation{
		body:          parsed.Body,
		reqModel:      parsed.Model,
		reqStream:     parsed.Stream,
		originalModel: parsed.Model,
	}

	isClaudeCode := isClaudeCodeRequest(ctx, c, parsed)
	preparation.shouldMimicClaudeCode = account.IsOAuth() && !isClaudeCode
	if preparation.shouldMimicClaudeCode {
		preparation.body, preparation.reqModel = s.normalizeForwardOAuthRequestBody(ctx, c, account, parsed, preparation.body, preparation.reqModel)
	}

	preparation.body = enforceCacheControlLimit(preparation.body)
	preparation.body, preparation.reqModel = s.applyForwardModelMapping(account, preparation.originalModel, preparation.reqModel, preparation.body)
	preparation.body = StripEmptyTextBlocks(preparation.body)

	token, tokenType, err := s.GetAccessToken(ctx, account)
	if err != nil {
		return nil, err
	}
	preparation.token = token
	preparation.tokenType = tokenType
	preparation.proxyURL = s.resolveForwardProxyURL(account)
	preparation.tlsProfile = s.tlsFPProfileService.ResolveTLSProfile(account)
	return preparation, nil
}

func (s *GatewayService) normalizeForwardOAuthRequestBody(
	ctx context.Context,
	c *gin.Context,
	account *Account,
	parsed *ParsedRequest,
	body []byte,
	reqModel string,
) ([]byte, string) {
	isHaikuModel := strings.Contains(strings.ToLower(reqModel), "haiku")
	if !isHaikuModel && !systemIncludesClaudeCodePrompt(parsed.System) {
		body = injectClaudeCodePrompt(body, parsed.System)
	}

	if !isHaikuModel && s.identityService != nil && c != nil && c.Request != nil {
		if fp, err := s.identityService.GetOrCreateFingerprint(ctx, account.ID, c.Request.Header); err == nil && fp != nil {
			if attrHeader := buildAttributionHeaderText(body, fp.UserAgent); attrHeader != "" {
				body = injectAttributionHeaderBlock(body, attrHeader)
			}
		}
	}

	normalizeOpts := claudeOAuthNormalizeOptions{stripSystemCacheControl: true}
	if s.identityService != nil && c != nil && c.Request != nil {
		fp, err := s.identityService.GetOrCreateFingerprint(ctx, account.ID, c.Request.Header)
		if err == nil && fp != nil {
			_, mimicMPT := s.settingService.GetGatewayForwardingSettings(ctx)
			if !mimicMPT {
				if metadataUserID := s.buildOAuthMetadataUserID(parsed, account, fp); metadataUserID != "" {
					normalizeOpts.injectMetadata = true
					normalizeOpts.metadataUserID = metadataUserID
				}
			}
		}
	}

	return normalizeClaudeOAuthRequestBody(body, reqModel, normalizeOpts)
}

func (s *GatewayService) applyForwardModelMapping(
	account *Account,
	originalModel string,
	requestModel string,
	body []byte,
) ([]byte, string) {
	mappedModel := requestModel
	mappingSource := ""
	if resolvedModel, source := resolveAnthropicCompatForwardModel(account, requestModel); source != "" {
		mappedModel = resolvedModel
		mappingSource = source
	}
	if mappedModel == requestModel {
		return body, requestModel
	}

	logger.LegacyPrintf(
		"service.gateway",
		"Model mapping applied: %s -> %s (account: %s, source=%s)",
		originalModel,
		mappedModel,
		account.Name,
		mappingSource,
	)
	return s.replaceModelInBody(body, mappedModel), mappedModel
}

func (s *GatewayService) resolveForwardProxyURL(account *Account) string {
	if account == nil || account.ProxyID == nil || account.Proxy == nil {
		return ""
	}
	if account.IsCustomBaseURLEnabled() && account.GetCustomBaseURL() != "" {
		return ""
	}
	return account.Proxy.URL()
}
