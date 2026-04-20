package service

import (
	"context"

	"github.com/senran-N/sub2api/internal/pkg/ctxkey"
)

func WithGrokSessionTextRuntimeAllowed(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, ctxkey.GrokSessionTextRuntimeAllowed, true)
}

func WithGrokSessionMediaRuntimeAllowed(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, ctxkey.GrokSessionMediaRuntimeAllowed, true)
}

func AllowsGrokSessionTextRuntime(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	allowed, _ := ctx.Value(ctxkey.GrokSessionTextRuntimeAllowed).(bool)
	return allowed
}

func AllowsGrokSessionMediaRuntime(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	allowed, _ := ctx.Value(ctxkey.GrokSessionMediaRuntimeAllowed).(bool)
	return allowed
}

func AllowsGrokSessionProviderRuntime(ctx context.Context) bool {
	return AllowsGrokSessionTextRuntime(ctx) || AllowsGrokSessionMediaRuntime(ctx)
}
