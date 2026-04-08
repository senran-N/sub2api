package service

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/senran-N/sub2api/internal/config"
)

func BenchmarkOpenAIWSPoolAcquire(b *testing.B) {
	cfg := &config.Config{}
	cfg.Gateway.OpenAIWS.MaxConnsPerAccount = 8
	cfg.Gateway.OpenAIWS.MinIdlePerAccount = 1
	cfg.Gateway.OpenAIWS.MaxIdlePerAccount = 4
	cfg.Gateway.OpenAIWS.QueueLimitPerConn = 256
	cfg.Gateway.OpenAIWS.DialTimeoutSeconds = 1

	pool := newOpenAIWSConnPool(cfg)
	pool.setClientDialerForTest(&openAIWSCountingDialer{})

	account := &Account{ID: 1001, Platform: PlatformOpenAI, Type: AccountTypeAPIKey}
	req := openAIWSAcquireRequest{
		Account: account,
		WSURL:   "wss://example.com/v1/responses",
	}
	ctx := context.Background()

	lease, err := pool.Acquire(ctx, req)
	if err != nil {
		b.Fatalf("warm acquire failed: %v", err)
	}
	lease.Release()

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var (
				got        *openAIWSConnLease
				acquireErr error
			)
			for retry := 0; retry < 3; retry++ {
				got, acquireErr = pool.Acquire(ctx, req)
				if acquireErr == nil {
					break
				}
				if !errors.Is(acquireErr, errOpenAIWSConnClosed) {
					break
				}
			}
			if acquireErr != nil {
				b.Fatalf("acquire failed: %v", acquireErr)
			}
			got.Release()
		}
	})
}

func BenchmarkOpenAIWSPoolAcquire_ManyIdleConns(b *testing.B) {
	cfg := &config.Config{}
	cfg.Gateway.OpenAIWS.MaxConnsPerAccount = 8
	cfg.Gateway.OpenAIWS.MinIdlePerAccount = 0
	cfg.Gateway.OpenAIWS.MaxIdlePerAccount = 8
	cfg.Gateway.OpenAIWS.QueueLimitPerConn = 256

	pool := newOpenAIWSConnPool(cfg)
	account := &Account{ID: 1002, Platform: PlatformOpenAI, Type: AccountTypeAPIKey}
	ap := pool.getOrCreateAccountPool(account.ID)
	ap.mu.Lock()
	for i := 0; i < 8; i++ {
		conn := newOpenAIWSConn(pool.nextConnID(account.ID), account.ID, &openAIWSFakeConn{}, nil)
		conn.lastUsedNano.Store(int64(i + 1))
		ap.conns[conn.id] = conn
	}
	ap.mu.Unlock()

	req := openAIWSAcquireRequest{
		Account: account,
		WSURL:   "wss://example.com/v1/responses",
	}
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			lease, err := pool.Acquire(ctx, req)
			if err != nil {
				b.Fatalf("acquire failed: %v", err)
			}
			lease.Release()
		}
	})
}

func BenchmarkNormalizeOpenAIWSAcquireRequest_WithHeaders(b *testing.B) {
	account := &Account{ID: 1003, Platform: PlatformOpenAI, Type: AccountTypeAPIKey}
	req := openAIWSAcquireRequest{
		Account: account,
		WSURL:   "wss://example.com/v1/responses",
		Headers: http.Header{
			"User-Agent":      []string{"bench-agent/1.0"},
			"OpenAI-Beta":     []string{openAIWSBetaV2Value},
			"Session_ID":      []string{"session-bench"},
			"Conversation_ID": []string{"conversation-bench"},
		},
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		normalized := normalizeOpenAIWSAcquireRequest(req)
		benchmarkOpenAIWSStringSink = normalized.WSURL
		benchmarkOpenAIWSBoolSink = normalized.Headers != nil
	}
}

func BenchmarkCloneOpenAIWSAcquireRequest_WithHeaders(b *testing.B) {
	account := &Account{ID: 1003, Platform: PlatformOpenAI, Type: AccountTypeAPIKey}
	req := openAIWSAcquireRequest{
		Account: account,
		WSURL:   "wss://example.com/v1/responses",
		Headers: http.Header{
			"User-Agent":      []string{"bench-agent/1.0"},
			"OpenAI-Beta":     []string{openAIWSBetaV2Value},
			"Session_ID":      []string{"session-bench"},
			"Conversation_ID": []string{"conversation-bench"},
		},
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cloned := cloneOpenAIWSAcquireRequest(req)
		benchmarkOpenAIWSStringSink = cloned.WSURL
		benchmarkOpenAIWSBoolSink = cloned.Headers != nil
	}
}
