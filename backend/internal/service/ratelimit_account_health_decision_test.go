package service

import (
	"net/http"
	"testing"
)

func TestClassifyAccountHealthDecision(t *testing.T) {
	svc := &RateLimitService{}

	tests := []struct {
		name       string
		account    *Account
		statusCode int
		body       []byte
		action     AccountHealthAction
		kind       string
	}{
		{
			name:       "openai token revoked",
			account:    &Account{ID: 1, Platform: PlatformOpenAI, Type: AccountTypeOAuth},
			statusCode: http.StatusUnauthorized,
			body:       []byte(`{"error":{"code":"token_revoked","message":"revoked"}}`),
			action:     AccountHealthActionSetError,
			kind:       "token_revoked",
		},
		{
			name:       "oauth 401 forces refresh cooldown",
			account:    &Account{ID: 2, Platform: PlatformAnthropic, Type: AccountTypeOAuth},
			statusCode: http.StatusUnauthorized,
			body:       []byte(`{"error":{"message":"expired"}}`),
			action:     AccountHealthActionOAuth401,
			kind:       "oauth_401",
		},
		{
			name:       "429 records rate limit",
			account:    &Account{ID: 3, Platform: PlatformOpenAI, Type: AccountTypeOAuth},
			statusCode: http.StatusTooManyRequests,
			body:       []byte(`{"error":{"message":"rate limited"}}`),
			action:     AccountHealthActionRateLimit,
			kind:       "rate_limited",
		},
		{
			name:       "antigravity validation 403 disables",
			account:    &Account{ID: 4, Platform: PlatformAntigravity, Type: AccountTypeOAuth},
			statusCode: http.StatusForbidden,
			body:       []byte(`{"error":{"message":"validation_required","details":"verify your account"}}`),
			action:     AccountHealthActionSetError,
			kind:       "validation_required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.classifyAccountHealthDecision(tt.account, tt.statusCode, tt.body, false)
			if got.Action != tt.action {
				t.Fatalf("action=%q, want %q", got.Action, tt.action)
			}
			if got.FailureKind != tt.kind {
				t.Fatalf("failure kind=%q, want %q", got.FailureKind, tt.kind)
			}
		})
	}
}
