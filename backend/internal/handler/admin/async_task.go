package admin

import (
	"context"
	"log/slog"
	"time"
)

func runDetachedAdminTask(taskName string, timeout time.Duration, fn func(context.Context), attrs ...any) <-chan struct{} {
	done := make(chan struct{})

	go func() {
		defer close(done)
		defer func() {
			if recovered := recover(); recovered != nil {
				logArgs := []any{"task", taskName, "recover", recovered}
				logArgs = append(logArgs, attrs...)
				slog.Error("admin_async_task_panic", logArgs...)
			}
		}()

		ctx := context.Background()
		cancel := func() {}
		if timeout > 0 {
			ctx, cancel = context.WithTimeout(ctx, timeout)
		}
		defer cancel()

		fn(ctx)
	}()

	return done
}
