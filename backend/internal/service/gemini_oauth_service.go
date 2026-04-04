package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/senran-N/sub2api/internal/config"
	"github.com/senran-N/sub2api/internal/domain"
	"github.com/senran-N/sub2api/internal/pkg/geminicli"
)

type GeminiOAuthService struct {
	sessionStore *geminicli.SessionStore
	proxyRepo    ProxyRepository
	oauthClient  GeminiOAuthClient
	codeAssist   GeminiCliCodeAssistClient
	driveClient  geminicli.DriveClient
	cfg          *config.Config
}

type GeminiOAuthCapabilities = domain.GeminiOAuthCapabilities

func NewGeminiOAuthService(
	proxyRepo ProxyRepository,
	oauthClient GeminiOAuthClient,
	codeAssist GeminiCliCodeAssistClient,
	driveClient geminicli.DriveClient,
	cfg *config.Config,
) *GeminiOAuthService {
	return &GeminiOAuthService{
		sessionStore: geminicli.NewSessionStore(),
		proxyRepo:    proxyRepo,
		oauthClient:  oauthClient,
		codeAssist:   codeAssist,
		driveClient:  driveClient,
		cfg:          cfg,
	}
}

func (s *GeminiOAuthService) GetOAuthConfig() *GeminiOAuthCapabilities {
	clientID := strings.TrimSpace(s.cfg.Gemini.OAuth.ClientID)
	clientSecret := strings.TrimSpace(s.cfg.Gemini.OAuth.ClientSecret)
	enabled := clientID != "" && clientSecret != "" && clientID != geminicli.GeminiCLIOAuthClientID

	return &GeminiOAuthCapabilities{
		AIStudioOAuthEnabled: enabled,
		RequiredRedirectURIs: []string{geminicli.AIStudioOAuthRedirectURI},
	}
}

type GeminiAuthURLResult = domain.GeminiAuthURLResult

func (s *GeminiOAuthService) GenerateAuthURL(ctx context.Context, proxyID *int64, redirectURI, projectID, oauthType, tierID string) (*GeminiAuthURLResult, error) {
	state, err := geminicli.GenerateState()
	if err != nil {
		return nil, fmt.Errorf("failed to generate state: %w", err)
	}
	codeVerifier, err := geminicli.GenerateCodeVerifier()
	if err != nil {
		return nil, fmt.Errorf("failed to generate code verifier: %w", err)
	}
	codeChallenge := geminicli.GenerateCodeChallenge(codeVerifier)
	sessionID, err := geminicli.GenerateSessionID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session ID: %w", err)
	}

	proxyURL := s.resolveProxyURL(ctx, proxyID, "")

	oauthCfg := s.configuredOAuthConfig()
	if oauthType == "code_assist" || oauthType == "google_one" {
		oauthCfg.ClientID = ""
		oauthCfg.ClientSecret = ""
	}

	session := &geminicli.OAuthSession{
		State:        state,
		CodeVerifier: codeVerifier,
		ProxyURL:     proxyURL,
		RedirectURI:  redirectURI,
		ProjectID:    strings.TrimSpace(projectID),
		TierID:       canonicalGeminiTierIDForOAuthType(oauthType, tierID),
		OAuthType:    oauthType,
		CreatedAt:    time.Now(),
	}
	s.sessionStore.Set(sessionID, session)

	effectiveCfg, err := geminicli.EffectiveOAuthConfig(oauthCfg, oauthType)
	if err != nil {
		return nil, err
	}

	isBuiltinClient := effectiveCfg.ClientID == geminicli.GeminiCLIOAuthClientID
	if oauthType == "ai_studio" && isBuiltinClient {
		return nil, fmt.Errorf("AI Studio OAuth requires a custom OAuth Client (GEMINI_OAUTH_CLIENT_ID / GEMINI_OAUTH_CLIENT_SECRET). If you don't want to configure an OAuth client, please use an AI Studio API Key account instead")
	}

	if isBuiltinClient {
		redirectURI = geminicli.GeminiCLIRedirectURI
	} else {
		redirectURI = geminicli.AIStudioOAuthRedirectURI
	}
	session.RedirectURI = redirectURI
	s.sessionStore.Set(sessionID, session)

	authURL, err := geminicli.BuildAuthorizationURL(effectiveCfg, state, codeChallenge, redirectURI, session.ProjectID, oauthType)
	if err != nil {
		return nil, err
	}

	return &GeminiAuthURLResult{
		AuthURL:   authURL,
		SessionID: sessionID,
		State:     state,
	}, nil
}

type GeminiExchangeCodeInput = domain.GeminiExchangeCodeInput
type GeminiTokenInfo = domain.GeminiTokenInfo

func (s *GeminiOAuthService) Stop() {
	s.sessionStore.Stop()
}
