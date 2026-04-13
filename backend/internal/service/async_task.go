package service

import (
	"context"
	"log/slog"
)

func runDetachedTask(taskName string, fn func(context.Context) error, attrs ...any) <-chan struct{} {
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

		if err := fn(context.Background()); err != nil {
			logArgs := []any{"task", taskName, "error", err}
			logArgs = append(logArgs, attrs...)
			slog.Error("service_async_task_failed", logArgs...)
		}
	}()

	return done
}
