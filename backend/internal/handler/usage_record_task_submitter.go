package handler

import (
	"context"

	"github.com/senran-N/sub2api/internal/pkg/logger"
	"github.com/senran-N/sub2api/internal/service"
	"go.uber.org/zap"
)

func submitUsageRecordTaskWithPool(
	parent context.Context,
	workerPool *service.UsageRecordWorkerPool,
	component string,
	panicEvent string,
	task service.UsageRecordTask,
) {
	if task == nil {
		return
	}
	if workerPool != nil {
		workerPool.Submit(task)
		return
	}

	ctx, cancel := newDetachedTimeoutContext(parent, usageRecordFallbackTaskTimeout)
	defer cancel()
	defer func() {
		if recovered := recover(); recovered != nil {
			logger.L().With(
				zap.String("component", component),
				zap.Any("panic", recovered),
			).Error(panicEvent)
		}
	}()
	task(ctx)
}
