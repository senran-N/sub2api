package admin

import (
	"context"
	"time"
)

const systemLockReleaseTimeout = 2 * time.Second

func newDetachedTimeoutContext(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	base := context.Background()
	if parent != nil {
		base = context.WithoutCancel(parent)
	}
	return context.WithTimeout(base, timeout)
}
