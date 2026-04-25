package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type runtimeSessionCacheStub struct {
	sessions map[int64]map[string]int64
}

func newRuntimeSessionCacheStub() *runtimeSessionCacheStub {
	return &runtimeSessionCacheStub{sessions: make(map[int64]map[string]int64)}
}

func (c *runtimeSessionCacheStub) GetSessionAccountID(_ context.Context, groupID int64, sessionHash string) (int64, error) {
	if c == nil || c.sessions == nil {
		return 0, nil
	}
	return c.sessions[groupID][sessionHash], nil
}

func (c *runtimeSessionCacheStub) SetSessionAccountID(_ context.Context, groupID int64, sessionHash string, accountID int64, _ time.Duration) error {
	if c.sessions == nil {
		c.sessions = make(map[int64]map[string]int64)
	}
	if c.sessions[groupID] == nil {
		c.sessions[groupID] = make(map[string]int64)
	}
	c.sessions[groupID][sessionHash] = accountID
	return nil
}

func (c *runtimeSessionCacheStub) RefreshSessionTTL(context.Context, int64, string, time.Duration) error {
	return nil
}

func (c *runtimeSessionCacheStub) DeleteSessionAccountID(_ context.Context, groupID int64, sessionHash string) error {
	if c != nil && c.sessions != nil && c.sessions[groupID] != nil {
		delete(c.sessions[groupID], sessionHash)
	}
	return nil
}

func TestPrepareRuntimeSessionParsesContextAndHash(t *testing.T) {
	svc := &GatewayService{}

	result := svc.PrepareRuntimeSession(context.Background(), RuntimeSessionPrepareRequest{
		Body:          []byte(`{"model":"claude-sonnet","stream":true,"messages":[{"role":"user","content":"hello"}]}`),
		ParseProtocol: string(GatewayProtocolChatCompletions),
		Model:         "fallback-model",
		Stream:        false,
		ClientIP:      "203.0.113.10",
		UserAgent:     "Codex CLI 2.1.78",
		APIKeyID:      42,
	})

	require.NotNil(t, result.Parsed)
	require.Equal(t, "claude-sonnet", result.Parsed.Model)
	require.True(t, result.Parsed.Stream)
	require.NotEmpty(t, result.SessionHash)
	require.Equal(t, result.SessionHash, result.SessionKey)
	require.Zero(t, result.BoundAccountID)
	require.Equal(t, "203.0.113.10", result.Parsed.SessionContext.ClientIP)
	require.Equal(t, "Codex CLI 2.1.78", result.Parsed.SessionContext.UserAgent)
	require.Equal(t, int64(42), result.Parsed.SessionContext.APIKeyID)
}

func TestPrepareRuntimeSessionFallsBackWhenParsingFails(t *testing.T) {
	svc := &GatewayService{}

	result := svc.PrepareRuntimeSession(context.Background(), RuntimeSessionPrepareRequest{
		Body:          []byte(`{"model":"claude-sonnet","stream":"yes"}`),
		ParseProtocol: string(GatewayProtocolResponses),
		Model:         "claude-sonnet",
		Stream:        true,
		APIKeyID:      9,
	})

	require.NotNil(t, result.Parsed)
	require.Equal(t, "claude-sonnet", result.Parsed.Model)
	require.True(t, result.Parsed.Stream)
	require.Equal(t, []byte(`{"model":"claude-sonnet","stream":"yes"}`), result.Parsed.Body)
}

func TestPrepareRuntimeSessionUsesProvidedParsedRequest(t *testing.T) {
	svc := &GatewayService{}
	parsed := &ParsedRequest{
		Model:    "already-parsed",
		Stream:   true,
		Messages: []any{map[string]any{"role": "user", "content": "hello"}},
	}
	body := []byte(`{"model":123}`)

	result := svc.PrepareRuntimeSession(context.Background(), RuntimeSessionPrepareRequest{
		Parsed:        parsed,
		Body:          body,
		ClientIP:      "198.51.100.9",
		UserAgent:     "codex_cli_rs/0.1.0",
		APIKeyID:      11,
		ParseProtocol: string(GatewayProtocolMessages),
	})

	require.Same(t, parsed, result.Parsed)
	require.Equal(t, body, result.Parsed.Body)
	require.NotEmpty(t, result.SessionHash)
	require.Equal(t, "198.51.100.9", result.Parsed.SessionContext.ClientIP)
	require.Equal(t, int64(11), result.Parsed.SessionContext.APIKeyID)
}

func TestPrepareRuntimeSessionUsesProvidedSessionHash(t *testing.T) {
	groupID := int64(5)
	cache := newRuntimeSessionCacheStub()
	require.NoError(t, cache.SetSessionAccountID(context.Background(), groupID, "gemini:cli-session", 123, time.Minute))
	svc := &GatewayService{cache: cache}

	result := svc.PrepareRuntimeSession(context.Background(), RuntimeSessionPrepareRequest{
		Body:             []byte(`{"contents":[{"parts":[{"text":"ignored for hash"}]}]}`),
		ParseProtocol:    "gemini",
		SessionHash:      " cli-session ",
		Model:            "gemini-2.5-pro",
		GroupID:          &groupID,
		SessionKeyPrefix: "gemini:",
	})

	require.Equal(t, "cli-session", result.SessionHash)
	require.Equal(t, "gemini:cli-session", result.SessionKey)
	require.Equal(t, int64(123), result.BoundAccountID)
}

func TestPrefetchRuntimeStickySessionBindsContextMetadata(t *testing.T) {
	groupID := int64(7)
	cache := newRuntimeSessionCacheStub()
	require.NoError(t, cache.SetSessionAccountID(context.Background(), groupID, "session-1", 99, time.Minute))
	svc := &GatewayService{cache: cache}

	result := svc.PrefetchRuntimeStickySession(context.Background(), RuntimeStickySessionPrefetchRequest{
		GroupID:    &groupID,
		SessionKey: "session-1",
	})

	require.Equal(t, int64(99), result.AccountID)
	require.Equal(t, groupID, result.GroupID)
	accountID, ok := PrefetchedStickyAccountIDFromContext(result.Context)
	require.True(t, ok)
	require.Equal(t, int64(99), accountID)
	prefetchedGroupID, ok := PrefetchedStickyGroupIDFromContext(result.Context)
	require.True(t, ok)
	require.Equal(t, groupID, prefetchedGroupID)
}

func TestPrepareRuntimeSessionUsesSessionKeyPrefixForStickyLookup(t *testing.T) {
	groupID := int64(3)
	cache := newRuntimeSessionCacheStub()
	svc := &GatewayService{cache: cache}
	body := []byte(`{"model":"claude-sonnet","metadata":{"user_id":"{\"device_id\":\"dev\",\"session_id\":\"123e4567-e89b-12d3-a456-426614174000\"}"}}`)

	require.NoError(t, cache.SetSessionAccountID(context.Background(), groupID, "gemini:123e4567-e89b-12d3-a456-426614174000", 77, time.Minute))

	result := svc.PrepareRuntimeSession(context.Background(), RuntimeSessionPrepareRequest{
		Body:             body,
		ParseProtocol:    string(GatewayProtocolResponses),
		Model:            "claude-sonnet",
		GroupID:          &groupID,
		SessionKeyPrefix: "gemini:",
	})

	require.Equal(t, "123e4567-e89b-12d3-a456-426614174000", result.SessionHash)
	require.Equal(t, "gemini:123e4567-e89b-12d3-a456-426614174000", result.SessionKey)
	require.Equal(t, int64(77), result.BoundAccountID)
}
