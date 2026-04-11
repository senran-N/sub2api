//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type redeemLockReleaseCacheStub struct {
	releaseCalled bool
	releaseCtxErr error
}

func (s *redeemLockReleaseCacheStub) GetRedeemAttemptCount(context.Context, int64) (int, error) {
	return 0, nil
}

func (s *redeemLockReleaseCacheStub) IncrementRedeemAttemptCount(context.Context, int64) error {
	return nil
}

func (s *redeemLockReleaseCacheStub) AcquireRedeemLock(context.Context, string, time.Duration) (bool, error) {
	return true, nil
}

func (s *redeemLockReleaseCacheStub) ReleaseRedeemLock(ctx context.Context, _ string) error {
	s.releaseCalled = true
	s.releaseCtxErr = ctx.Err()
	return nil
}

func TestRedeemService_ReleaseRedeemLock_UsesDetachedContext(t *testing.T) {
	cache := &redeemLockReleaseCacheStub{}
	svc := &RedeemService{cache: cache}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	svc.releaseRedeemLock(ctx, "TEST-CODE")

	require.True(t, cache.releaseCalled)
	require.NoError(t, cache.releaseCtxErr)
}
