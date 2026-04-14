package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	"github.com/imroc/req/v3"
	"github.com/senran-N/sub2api/internal/pkg/openai"
	"github.com/stretchr/testify/require"
)

type openaiOAuthClientRefreshStub struct {
	refreshCalls  int32
	tokenResponse *openai.TokenResponse
}

func (s *openaiOAuthClientRefreshStub) ExchangeCode(ctx context.Context, code, codeVerifier, redirectURI, proxyURL, clientID string) (*openai.TokenResponse, error) {
	return nil, errors.New("not implemented")
}

func (s *openaiOAuthClientRefreshStub) RefreshToken(ctx context.Context, refreshToken, proxyURL string) (*openai.TokenResponse, error) {
	atomic.AddInt32(&s.refreshCalls, 1)
	if s.tokenResponse != nil {
		return s.tokenResponse, nil
	}
	return nil, errors.New("not implemented")
}

func (s *openaiOAuthClientRefreshStub) RefreshTokenWithClientID(ctx context.Context, refreshToken, proxyURL string, clientID string) (*openai.TokenResponse, error) {
	atomic.AddInt32(&s.refreshCalls, 1)
	if s.tokenResponse != nil {
		return s.tokenResponse, nil
	}
	return nil, errors.New("not implemented")
}

func mustEncodeOpenAIJWT(t *testing.T, claims map[string]any) string {
	t.Helper()
	payload, err := json.Marshal(claims)
	require.NoError(t, err)
	return base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none","typ":"JWT"}`)) +
		"." + base64.RawURLEncoding.EncodeToString(payload) +
		".signature"
}

func TestOpenAIOAuthService_RefreshAccountToken_NoRefreshTokenUsesExistingAccessToken(t *testing.T) {
	client := &openaiOAuthClientRefreshStub{}
	svc := NewOpenAIOAuthService(nil, client)

	expiresAt := time.Now().Add(30 * time.Minute).UTC().Format(time.RFC3339)
	account := &Account{
		ID:       77,
		Platform: PlatformOpenAI,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"access_token": "existing-access-token",
			"expires_at":   expiresAt,
			"client_id":    "client-id-1",
		},
	}

	info, err := svc.RefreshAccountToken(context.Background(), account)
	require.NoError(t, err)
	require.NotNil(t, info)
	require.Equal(t, "existing-access-token", info.AccessToken)
	require.Equal(t, "client-id-1", info.ClientID)
	require.Zero(t, atomic.LoadInt32(&client.refreshCalls), "existing access token should be reused without calling refresh")
}

func TestOpenAIOAuthService_RefreshTokenWithClientID_RefreshesPlanTypeFromBackendAPI(t *testing.T) {
	orgID := "org-live"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/backend-api/accounts/check/v4-2023-04-27", r.URL.Path)
		require.Equal(t, "Bearer access-token-1", r.Header.Get("Authorization"))
		require.Equal(t, codexCLIUserAgent, r.Header.Get("User-Agent"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"accounts":{"org-live":{"account":{"plan_type":"team","is_default":true}}}}`))
	}))
	defer server.Close()

	target, err := url.Parse(server.URL)
	require.NoError(t, err)

	client := &openaiOAuthClientRefreshStub{
		tokenResponse: &openai.TokenResponse{
			AccessToken:  "access-token-1",
			RefreshToken: "refresh-token-1",
			ExpiresIn:    3600,
			IDToken: mustEncodeOpenAIJWT(t, map[string]any{
				"exp": time.Now().Add(time.Hour).Unix(),
				"https://api.openai.com/auth": map[string]any{
					"chatgpt_plan_type": "plus",
					"organizations": []map[string]any{
						{"id": orgID, "is_default": true},
					},
				},
			}),
		},
	}

	svc := NewOpenAIOAuthService(nil, client)
	svc.SetPrivacyClientFactory(func(proxyURL string) (*req.Client, error) {
		reqClient := req.C()
		reqClient.GetTransport().WrapRoundTripFunc(func(rt http.RoundTripper) req.HttpRoundTripFunc {
			return func(r *http.Request) (*http.Response, error) {
				if r.URL.String() == chatGPTAccountsCheckURL {
					r.URL.Scheme = target.Scheme
					r.URL.Host = target.Host
					r.Host = target.Host
				}
				return rt.RoundTrip(r)
			}
		})
		return reqClient, nil
	})

	info, err := svc.RefreshTokenWithClientID(context.Background(), "refresh-token-1", "", "client-id-2")
	require.NoError(t, err)
	require.NotNil(t, info)
	require.Equal(t, "team", info.PlanType)
	require.Equal(t, orgID, info.OrganizationID)
	require.Equal(t, "client-id-2", info.ClientID)
	require.Equal(t, int32(1), atomic.LoadInt32(&client.refreshCalls))
}
