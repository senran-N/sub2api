package service

import (
	"context"
	"time"
)

const cacheLockReleaseTimeout = 5 * time.Second

func newCacheLockReleaseContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), cacheLockReleaseTimeout)
}
