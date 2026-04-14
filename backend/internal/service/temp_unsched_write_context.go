package service

import (
	"context"
	"time"
)

const tempUnschedWriteTimeout = 3 * time.Second

func newTempUnschedWriteContext(parent context.Context) (context.Context, context.CancelFunc) {
	base := context.Background()
	if parent != nil {
		base = context.WithoutCancel(parent)
	}
	return context.WithTimeout(base, tempUnschedWriteTimeout)
}
