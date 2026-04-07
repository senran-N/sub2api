package service

import (
	"context"
	"log/slog"
)

func runDetachedTask(taskName string, fn func(context.Context), attrs ...any) <-chan struct{} {
	done := make(chan struct{})

	go func() {
		defer close(done)
		defer func() {
			if recovered := recover(); recovered != nil {
				logArgs := []any{"task", taskName, "recover", recovered}
				logArgs = append(logArgs, attrs...)
				slog.Error("service_async_task_panic", logArgs...)
			}
		}()

		fn(context.Background())
	}()

	return done
}
