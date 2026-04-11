//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type userMsgQueueReleaseCacheStub struct {
	releaseCalled bool
	releaseCtxErr error
}

func (s *userMsgQueueReleaseCacheStub) AcquireLock(context.Context, int64, string, int) (bool, error) {
	return true, nil
}

func (s *userMsgQueueReleaseCacheStub) ReleaseLock(ctx context.Context, _ int64, _ string) (bool, error) {
	s.releaseCalled = true
	s.releaseCtxErr = ctx.Err()
	return true, nil
}

func (s *userMsgQueueReleaseCacheStub) GetLastCompletedMs(context.Context, int64) (int64, error) {
	return 0, nil
}

func (s *userMsgQueueReleaseCacheStub) GetCurrentTimeMs(context.Context) (int64, error) {
	return 0, nil
}

func (s *userMsgQueueReleaseCacheStub) ForceReleaseLock(context.Context, int64) error {
	return nil
}

func (s *userMsgQueueReleaseCacheStub) ScanLockKeys(context.Context, int) ([]int64, error) {
	return nil, nil
}

func TestUserMessageQueueService_Release_UsesDetachedContext(t *testing.T) {
	cache := &userMsgQueueReleaseCacheStub{}
	svc := &UserMessageQueueService{cache: cache}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := svc.Release(ctx, 123, "req-1")

	require.NoError(t, err)
	require.True(t, cache.releaseCalled)
	require.NoError(t, cache.releaseCtxErr)
}
