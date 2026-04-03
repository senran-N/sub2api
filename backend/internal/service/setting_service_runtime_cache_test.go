//go:build unit

package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/senran-N/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type runtimeCacheRepoStub struct {
	getValueFn    func(ctx context.Context, key string) (string, error)
	getMultipleFn func(ctx context.Context, keys []string) (map[string]string, error)
	getValueCalls int
	getMultiCalls int
	updates       map[string]string
}

func (s *runtimeCacheRepoStub) Get(ctx context.Context, key string) (*Setting, error) {
	panic("unexpected Get call")
}

func (s *runtimeCacheRepoStub) GetValue(ctx context.Context, key string) (string, error) {
	s.getValueCalls++
	if s.getValueFn == nil {
		panic("unexpected GetValue call")
	}
	return s.getValueFn(ctx, key)
}

func (s *runtimeCacheRepoStub) Set(ctx context.Context, key, value string) error {
	panic("unexpected Set call")
}

func (s *runtimeCacheRepoStub) GetMultiple(ctx context.Context, keys []string) (map[string]string, error) {
	s.getMultiCalls++
	if s.getMultipleFn == nil {
		panic("unexpected GetMultiple call")
	}
	return s.getMultipleFn(ctx, keys)
}

func (s *runtimeCacheRepoStub) SetMultiple(ctx context.Context, settings map[string]string) error {
	s.updates = make(map[string]string, len(settings))
	for key, value := range settings {
		s.updates[key] = value
	}
	return nil
}

func (s *runtimeCacheRepoStub) GetAll(ctx context.Context) (map[string]string, error) {
	panic("unexpected GetAll call")
}

func (s *runtimeCacheRepoStub) Delete(ctx context.Context, key string) error {
	panic("unexpected Delete call")
}

func resetRuntimeSettingCaches(t *testing.T) {
	t.Helper()

	versionBoundsCache.Store((*cachedVersionBounds)(nil))
	backendModeCache.Store((*cachedBackendMode)(nil))
	gatewayForwardingCache.Store((*cachedGatewayForwardingSettings)(nil))
	t.Cleanup(func() {
		versionBoundsCache.Store((*cachedVersionBounds)(nil))
		backendModeCache.Store((*cachedBackendMode)(nil))
		gatewayForwardingCache.Store((*cachedGatewayForwardingSettings)(nil))
	})
}

func TestGetGatewayForwardingSettings_FailOpenOnRepoError(t *testing.T) {
	resetRuntimeSettingCaches(t)

	repo := &runtimeCacheRepoStub{
		getMultipleFn: func(ctx context.Context, keys []string) (map[string]string, error) {
			require.ElementsMatch(t, []string{
				SettingKeyEnableFingerprintUnification,
				SettingKeyEnableMetadataPassthrough,
			}, keys)
			return nil, errors.New("db down")
		},
	}
	svc := NewSettingService(repo, &config.Config{})

	fingerprintUnification, metadataPassthrough := svc.GetGatewayForwardingSettings(context.Background())
	require.True(t, fingerprintUnification)
	require.False(t, metadataPassthrough)
	require.Equal(t, 1, repo.getMultiCalls)
}

func TestGetGatewayForwardingSettings_CachesResult(t *testing.T) {
	resetRuntimeSettingCaches(t)

	repo := &runtimeCacheRepoStub{
		getMultipleFn: func(ctx context.Context, keys []string) (map[string]string, error) {
			return map[string]string{
				SettingKeyEnableFingerprintUnification: "false",
				SettingKeyEnableMetadataPassthrough:    "true",
			}, nil
		},
	}
	svc := NewSettingService(repo, &config.Config{})

	fingerprintUnification, metadataPassthrough := svc.GetGatewayForwardingSettings(context.Background())
	require.False(t, fingerprintUnification)
	require.True(t, metadataPassthrough)

	fingerprintUnification, metadataPassthrough = svc.GetGatewayForwardingSettings(context.Background())
	require.False(t, fingerprintUnification)
	require.True(t, metadataPassthrough)
	require.Equal(t, 1, repo.getMultiCalls)
}

func TestGetClaudeCodeVersionBounds_CachesResult(t *testing.T) {
	resetRuntimeSettingCaches(t)

	repo := &runtimeCacheRepoStub{
		getMultipleFn: func(ctx context.Context, keys []string) (map[string]string, error) {
			require.ElementsMatch(t, []string{
				SettingKeyMinClaudeCodeVersion,
				SettingKeyMaxClaudeCodeVersion,
			}, keys)
			return map[string]string{
				SettingKeyMinClaudeCodeVersion: "1.2.3",
				SettingKeyMaxClaudeCodeVersion: "9.9.9",
			}, nil
		},
	}
	svc := NewSettingService(repo, &config.Config{})

	min, max := svc.GetClaudeCodeVersionBounds(context.Background())
	require.Equal(t, "1.2.3", min)
	require.Equal(t, "9.9.9", max)

	min, max = svc.GetClaudeCodeVersionBounds(context.Background())
	require.Equal(t, "1.2.3", min)
	require.Equal(t, "9.9.9", max)
	require.Equal(t, 1, repo.getMultiCalls)
}

func TestUpdateSettings_InvalidatesGatewayForwardingAndVersionBoundsCaches(t *testing.T) {
	resetRuntimeSettingCaches(t)

	storeGatewayForwardingCache(true, false, gatewayForwardingCacheTTL)
	storeVersionBoundsCache("0.0.1", "0.0.2", versionBoundsCacheTTL)

	repo := &runtimeCacheRepoStub{
		getValueFn: func(ctx context.Context, key string) (string, error) {
			return "", ErrSettingNotFound
		},
		getMultipleFn: func(ctx context.Context, keys []string) (map[string]string, error) {
			return map[string]string{}, nil
		},
	}
	svc := NewSettingService(repo, &config.Config{})

	err := svc.UpdateSettings(context.Background(), &SystemSettings{
		EnableFingerprintUnification: false,
		EnableMetadataPassthrough:    true,
		MinClaudeCodeVersion:         "2.0.0",
		MaxClaudeCodeVersion:         "3.0.0",
	})
	require.NoError(t, err)

	fingerprintUnification, metadataPassthrough := svc.GetGatewayForwardingSettings(context.Background())
	require.False(t, fingerprintUnification)
	require.True(t, metadataPassthrough)

	min, max := svc.GetClaudeCodeVersionBounds(context.Background())
	require.Equal(t, "2.0.0", min)
	require.Equal(t, "3.0.0", max)
	require.Zero(t, repo.getMultiCalls)
}

func TestStoreBackendModeCache_ReplacesExistingValue(t *testing.T) {
	resetRuntimeSettingCaches(t)

	storeBackendModeCache(true, time.Minute)
	storeBackendModeCache(false, time.Minute)

	cached, ok := backendModeCache.Load().(*cachedBackendMode)
	require.True(t, ok)
	require.NotNil(t, cached)
	require.False(t, cached.value)
}
