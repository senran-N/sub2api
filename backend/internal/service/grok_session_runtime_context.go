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

func AllowsGrokSessionTextRuntime(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	allowed, _ := ctx.Value(ctxkey.GrokSessionTextRuntimeAllowed).(bool)
	return allowed
}
