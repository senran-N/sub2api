//go:build unit

package service

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/senran-N/sub2api/internal/pkg/apicompat"
	"github.com/stretchr/testify/require"
)

type grokCompatibleExecutorCall struct {
	accountID int64
	body      string
}

type stubGrokCompatibleTextExecutor struct {
	responsesCalls       []grokCompatibleExecutorCall
	chatCompletionsCalls []grokCompatibleExecutorCall
	messagesCalls        []grokCompatibleExecutorCall
}

func (s *stubGrokCompatibleTextExecutor) ForwardResponses(_ context.Context, c *gin.Context, account *Account, body []byte, _ string) (*OpenAIForwardResult, error) {
	s.responsesCalls = append(s.responsesCalls, grokCompatibleExecutorCall{accountID: account.ID, body: string(body)})
	c.JSON(http.StatusOK, gin.H{"ok": true})
	return nil, nil
}

func (s *stubGrokCompatibleTextExecutor) ForwardChatCompletions(_ context.Context, c *gin.Context, account *Account, body []byte, _ string, _ string) (*OpenAIForwardResult, error) {
	s.chatCompletionsCalls = append(s.chatCompletionsCalls, grokCompatibleExecutorCall{accountID: account.ID, body: string(body)})
	c.JSON(http.StatusOK, gin.H{"ok": true})
	return nil, nil
}

func (s *stubGrokCompatibleTextExecutor) ForwardMessages(_ context.Context, c *gin.Context, account *Account, body []byte, _ string, _ string) (*OpenAIForwardResult, error) {
	s.messagesCalls = append(s.messagesCalls, grokCompatibleExecutorCall{accountID: account.ID, body: string(body)})
	c.JSON(http.StatusOK, gin.H{"ok": true})
	return nil, nil
}

func TestGrokGatewayServiceHandleResponses_UsesCompatibleAccount(t *testing.T) {
	gin.SetMode(gin.TestMode)

	body := []byte(`{"model":"grok-3","input":"hello"}`)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", bytes.NewReader(body))
	c.Request = c.Request.WithContext(WithGrokSessionTextRuntimeAllowed(context.Background()))

	repo := &mockAccountRepoForPlatform{
		accounts: []Account{
			{
				ID:          1,
				Name:        "grok-compatible",
				Platform:    PlatformGrok,
				Type:        AccountTypeAPIKey,
				Status:      StatusActive,
				Schedulable: true,
			},
		},
	}
	executor := &stubGrokCompatibleTextExecutor{}
	svc := NewGrokGatewayServiceWithCompatibleExecutor(&GatewayService{
		accountRepo: repo,
		cfg:         testConfig(),
	}, executor)

	handled := svc.HandleResponses(c, nil, body)
	require.True(t, handled)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Len(t, executor.responsesCalls, 1)
	require.Equal(t, int64(1), executor.responsesCalls[0].accountID)
	require.JSONEq(t, string(body), executor.responsesCalls[0].body)
}

func TestGrokGatewayServiceHandleResponses_UsesScoredGrokSelector(t *testing.T) {
	gin.SetMode(gin.TestMode)

	now := time.Now().UTC()
	body := []byte(`{"model":"grok-3","input":"hello"}`)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", bytes.NewReader(body))
	c.Request = c.Request.WithContext(WithGrokSessionTextRuntimeAllowed(context.Background()))

	repo := &mockAccountRepoForPlatform{
		accounts: []Account{
			{
				ID:          101,
				Name:        "grok-heavy-stale",
				Platform:    PlatformGrok,
				Type:        AccountTypeAPIKey,
				Status:      StatusActive,
				Schedulable: true,
				Priority:    1,
				Extra: map[string]any{
					"grok": map[string]any{
						"tier": map[string]any{
							"normalized": "heavy",
						},
						"sync_state": map[string]any{
							"last_sync_at": now.Add(-10 * time.Hour).Format(time.RFC3339),
						},
						"quota_windows": map[string]any{
							"auto": map[string]any{
								"remaining": 100,
								"total":     150,
							},
						},
					},
				},
			},
			{
				ID:          102,
				Name:        "grok-basic-fresh",
				Platform:    PlatformGrok,
				Type:        AccountTypeAPIKey,
				Status:      StatusActive,
				Schedulable: true,
				Priority:    1,
				Extra: map[string]any{
					"grok": map[string]any{
						"tier": map[string]any{
							"normalized": "basic",
						},
						"capabilities": map[string]any{
							"models": []any{"grok-3"},
						},
						"sync_state": map[string]any{
							"last_sync_at": now.Add(-5 * time.Minute).Format(time.RFC3339),
						},
						"quota_windows": map[string]any{
							"auto": map[string]any{
								"remaining": 18,
								"total":     20,
							},
						},
					},
				},
			},
		},
	}
	executor := &stubGrokCompatibleTextExecutor{}
	svc := NewGrokGatewayServiceWithCompatibleExecutor(&GatewayService{
		accountRepo: repo,
		cfg:         testConfig(),
	}, executor)

	handled := svc.HandleResponses(c, nil, body)
	require.True(t, handled)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Len(t, executor.responsesCalls, 1)
	require.Equal(t, int64(102), executor.responsesCalls[0].accountID)
}

func TestGrokGatewayServiceHandleChatCompletions_UsesCompatibleAccount(t *testing.T) {
	gin.SetMode(gin.TestMode)

	body := []byte(`{"model":"grok-3","messages":[{"role":"user","content":"hello"}]}`)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(body))
	c.Request = c.Request.WithContext(WithGrokSessionTextRuntimeAllowed(context.Background()))

	repo := &mockAccountRepoForPlatform{
		accounts: []Account{
			{
				ID:          11,
				Name:        "grok-compatible-chat",
				Platform:    PlatformGrok,
				Type:        AccountTypeAPIKey,
				Status:      StatusActive,
				Schedulable: true,
			},
		},
	}
	executor := &stubGrokCompatibleTextExecutor{}
	svc := NewGrokGatewayServiceWithCompatibleExecutor(&GatewayService{
		accountRepo: repo,
		cfg:         testConfig(),
	}, executor)

	handled := svc.HandleChatCompletions(c, nil, body)
	require.True(t, handled)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Len(t, executor.chatCompletionsCalls, 1)
	require.Equal(t, int64(11), executor.chatCompletionsCalls[0].accountID)
	require.JSONEq(t, string(body), executor.chatCompletionsCalls[0].body)
}

func TestGrokGatewayServiceHandleMessages_UsesCompatibleAccount(t *testing.T) {
	gin.SetMode(gin.TestMode)

	body := []byte(`{"model":"grok-3","max_tokens":256,"messages":[{"role":"user","content":"hello"}]}`)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewReader(body))
	c.Request = c.Request.WithContext(WithGrokSessionTextRuntimeAllowed(context.Background()))

	repo := &mockAccountRepoForPlatform{
		accounts: []Account{
			{
				ID:          12,
				Name:        "grok-compatible-messages",
				Platform:    PlatformGrok,
				Type:        AccountTypeAPIKey,
				Status:      StatusActive,
				Schedulable: true,
			},
		},
	}
	executor := &stubGrokCompatibleTextExecutor{}
	svc := NewGrokGatewayServiceWithCompatibleExecutor(&GatewayService{
		accountRepo: repo,
		cfg:         testConfig(),
	}, executor)

	handled := svc.HandleMessages(c, nil, body)
	require.True(t, handled)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Len(t, executor.messagesCalls, 1)
	require.Equal(t, int64(12), executor.messagesCalls[0].accountID)
	require.JSONEq(t, string(body), executor.messagesCalls[0].body)
}

func TestGrokGatewayServiceHandleResponses_UsesSessionAccount(t *testing.T) {
	gin.SetMode(gin.TestMode)

	body := []byte(`{"model":"grok-3","input":"hello"}`)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", bytes.NewReader(body))
	c.Request = c.Request.WithContext(WithGrokSessionTextRuntimeAllowed(context.Background()))

	upstream := &queuedHTTPUpstream{
		responses: []*http.Response{
			newJSONResponse(http.StatusOK, strings.Join([]string{
				`{"result":{"response":{"token":"answer","messageTag":"final"}}}`,
				`{"result":{"response":{"finalMetadata":{"stop_reason":"end_turn"}}}}`,
			}, "\n")),
		},
	}
	repo := &mockAccountRepoForPlatform{
		accounts: []Account{
			{
				ID:          2,
				Name:        "grok-session",
				Platform:    PlatformGrok,
				Type:        AccountTypeSession,
				Status:      StatusActive,
				Schedulable: true,
				Credentials: map[string]any{
					"session_token": "session-cookie",
				},
			},
		},
	}
	svc := NewGrokGatewayService(&GatewayService{
		accountRepo:  repo,
		httpUpstream: upstream,
		cfg:          testConfig(),
	}, nil)

	handled := svc.HandleResponses(c, nil, body)
	require.True(t, handled)
	require.Len(t, upstream.requests, 1)
	require.Equal(t, requireGrokSessionCookieHeader(t, "session-cookie"), upstream.requests[0].Header.Get("Cookie"))
	require.Contains(t, upstream.requests[0].URL.String(), "/rest/app-chat/conversations/new")

	var payload map[string]any
	require.NoError(t, json.NewDecoder(upstream.requests[0].Body).Decode(&payload))
	require.Equal(t, "hello", payload["message"])

	var response apicompat.ResponsesResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
	require.Equal(t, "completed", response.Status)
	require.Len(t, response.Output, 1)
	require.Equal(t, "answer", response.Output[0].Content[0].Text)
}

func TestGrokGatewayServiceHandleChatCompletions_UsesSessionAccount(t *testing.T) {
	gin.SetMode(gin.TestMode)

	body := []byte(`{"model":"grok-3","messages":[{"role":"user","content":"hello"}]}`)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(body))
	c.Request = c.Request.WithContext(WithGrokSessionTextRuntimeAllowed(context.Background()))

	upstream := &queuedHTTPUpstream{
		responses: []*http.Response{
			newJSONResponse(http.StatusOK, strings.Join([]string{
				`{"result":{"response":{"token":"answer","messageTag":"final"}}}`,
				`{"result":{"response":{"finalMetadata":{"stop_reason":"end_turn"}}}}`,
			}, "\n")),
		},
	}
	repo := &mockAccountRepoForPlatform{
		accounts: []Account{
			{
				ID:          3,
				Name:        "grok-session-chat",
				Platform:    PlatformGrok,
				Type:        AccountTypeSession,
				Status:      StatusActive,
				Schedulable: true,
				Credentials: map[string]any{
					"session_token": "session-cookie",
				},
			},
		},
	}
	svc := NewGrokGatewayService(&GatewayService{
		accountRepo:  repo,
		httpUpstream: upstream,
		cfg:          testConfig(),
	}, nil)

	handled := svc.HandleChatCompletions(c, nil, body)
	require.True(t, handled)
	require.Len(t, upstream.requests, 1)
	require.Equal(t, requireGrokSessionCookieHeader(t, "session-cookie"), upstream.requests[0].Header.Get("Cookie"))

	var response apicompat.ChatCompletionsResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
	require.Len(t, response.Choices, 1)

	var content string
	require.NoError(t, json.Unmarshal(response.Choices[0].Message.Content, &content))
	require.Equal(t, "answer", content)
	require.Equal(t, "stop", response.Choices[0].FinishReason)
}

func TestGrokGatewayServiceHandleChatCompletions_HydratesSnapshotSelectedSessionAccount(t *testing.T) {
	gin.SetMode(gin.TestMode)

	body := []byte(`{"model":"grok-3","messages":[{"role":"user","content":"hello"}]}`)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/grok/v1/chat/completions", bytes.NewReader(body))
	c.Request = c.Request.WithContext(WithGrokSessionTextRuntimeAllowed(context.Background()))

	snapshotAccount := &Account{
		ID:          303,
		Name:        "grok-session-snapshot",
		Platform:    PlatformGrok,
		Type:        AccountTypeSession,
		Status:      StatusActive,
		Schedulable: true,
	}
	fullAccount := &Account{
		ID:          snapshotAccount.ID,
		Name:        snapshotAccount.Name,
		Platform:    snapshotAccount.Platform,
		Type:        snapshotAccount.Type,
		Status:      snapshotAccount.Status,
		Schedulable: snapshotAccount.Schedulable,
		Credentials: map[string]any{
			"session_token": "session-cookie",
		},
	}
	snapshotCache := &openAISnapshotCacheStub{
		snapshotAccounts: []*Account{snapshotAccount},
		accountsByID:     map[int64]*Account{fullAccount.ID: fullAccount},
	}
	upstream := &queuedHTTPUpstream{
		responses: []*http.Response{
			newJSONResponse(http.StatusOK, strings.Join([]string{
				`{"result":{"response":{"token":"answer","messageTag":"final"}}}`,
				`{"result":{"response":{"finalMetadata":{"stop_reason":"end_turn"}}}}`,
			}, "\n")),
		},
	}
	repo := &mockAccountRepoForPlatform{
		accounts: []Account{*fullAccount},
		accountsByID: map[int64]*Account{
			fullAccount.ID: fullAccount,
		},
	}
	svc := NewGrokGatewayService(&GatewayService{
		accountRepo:        repo,
		httpUpstream:       upstream,
		cfg:                testConfig(),
		schedulerSnapshot:  &SchedulerSnapshotService{cache: snapshotCache},
	}, nil)

	handled := svc.HandleChatCompletions(c, nil, body)
	require.True(t, handled)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Len(t, upstream.requests, 1)
	require.Equal(t, requireGrokSessionCookieHeader(t, "session-cookie"), upstream.requests[0].Header.Get("Cookie"))
	require.Equal(t, 0, repo.getByIDCalls)

	var response apicompat.ChatCompletionsResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
	require.Len(t, response.Choices, 1)
}

func TestGrokGatewayServiceHandleMessages_UsesSessionAccount(t *testing.T) {
	gin.SetMode(gin.TestMode)

	body := []byte(`{"model":"grok-3","max_tokens":256,"messages":[{"role":"user","content":"hello"}]}`)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewReader(body))
	c.Request = c.Request.WithContext(WithGrokSessionTextRuntimeAllowed(context.Background()))

	upstream := &queuedHTTPUpstream{
		responses: []*http.Response{
			newJSONResponse(http.StatusOK, strings.Join([]string{
				`{"result":{"response":{"token":"answer","messageTag":"final"}}}`,
				`{"result":{"response":{"finalMetadata":{"stop_reason":"end_turn"}}}}`,
			}, "\n")),
		},
	}
	repo := &mockAccountRepoForPlatform{
		accounts: []Account{
			{
				ID:          4,
				Name:        "grok-session-messages",
				Platform:    PlatformGrok,
				Type:        AccountTypeSession,
				Status:      StatusActive,
				Schedulable: true,
				Credentials: map[string]any{
					"session_token": "session-cookie",
				},
			},
		},
	}
	svc := NewGrokGatewayService(&GatewayService{
		accountRepo:  repo,
		httpUpstream: upstream,
		cfg:          testConfig(),
	}, nil)

	handled := svc.HandleMessages(c, nil, body)
	require.True(t, handled)
	require.Len(t, upstream.requests, 1)
	require.Equal(t, requireGrokSessionCookieHeader(t, "session-cookie"), upstream.requests[0].Header.Get("Cookie"))

	var response apicompat.AnthropicResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
	require.Equal(t, "message", response.Type)
	require.Equal(t, "assistant", response.Role)
	require.Len(t, response.Content, 1)
	require.Equal(t, "text", response.Content[0].Type)
	require.Equal(t, "answer", response.Content[0].Text)
	require.Equal(t, "end_turn", response.StopReason)
}
