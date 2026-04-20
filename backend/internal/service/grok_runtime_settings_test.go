//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type grokRuntimeSettingRepoStub struct {
	values map[string]string
	err    error
}

func (s *grokRuntimeSettingRepoStub) Get(ctx context.Context, key string) (*Setting, error) {
	panic("unexpected Get call")
}

func (s *grokRuntimeSettingRepoStub) GetValue(ctx context.Context, key string) (string, error) {
	panic("unexpected GetValue call")
}

func (s *grokRuntimeSettingRepoStub) Set(ctx context.Context, key, value string) error {
	panic("unexpected Set call")
}

func (s *grokRuntimeSettingRepoStub) GetMultiple(ctx context.Context, keys []string) (map[string]string, error) {
	if s.err != nil {
		return nil, s.err
	}
	result := make(map[string]string, len(keys))
	for _, key := range keys {
		result[key] = s.values[key]
	}
	return result, nil
}

func (s *grokRuntimeSettingRepoStub) SetMultiple(ctx context.Context, settings map[string]string) error {
	panic("unexpected SetMultiple call")
}

func (s *grokRuntimeSettingRepoStub) GetAll(ctx context.Context) (map[string]string, error) {
	panic("unexpected GetAll call")
}

func (s *grokRuntimeSettingRepoStub) Delete(ctx context.Context, key string) error {
	panic("unexpected Delete call")
}

func TestSettingServiceGetGrokRuntimeSettings_UsesDefaultsForInvalidValues(t *testing.T) {
	svc := NewSettingService(&grokRuntimeSettingRepoStub{
		values: map[string]string{
			SettingKeyGrokQuotaSyncIntervalSeconds:       "abc",
			SettingKeyGrokCapabilityProbeIntervalSeconds: "0",
			SettingKeyGrokSessionValidityCheckInterval:   "invalid",
			SettingKeyGrokVideoTimeout:                   "0",
		},
	}, nil)

	settings := svc.GetGrokRuntimeSettings(context.Background())
	require.Equal(t, defaultGrokQuotaSyncIntervalSeconds, settings.QuotaSyncIntervalSeconds)
	require.Equal(t, defaultGrokCapabilityProbeIntervalSeconds, settings.CapabilityProbeIntervalSeconds)
	require.Equal(t, defaultGrokSessionValidityCheckSeconds, settings.SessionValidityCheckSeconds)
	require.Equal(t, defaultGrokVideoTimeoutSeconds, settings.VideoTimeoutSeconds)
	require.Equal(t, 15*time.Minute, settings.QuotaSyncInterval())
	require.Equal(t, 6*time.Hour, settings.CapabilityProbeInterval())
	require.Equal(t, 30*time.Minute, settings.SessionValidityCheckInterval())
	require.Equal(t, 10*time.Minute, settings.VideoTimeout())
}

func TestSettingServiceGetGrokRuntimeSettings_ClampsConfiguredValues(t *testing.T) {
	svc := NewSettingService(&grokRuntimeSettingRepoStub{
		values: map[string]string{
			SettingKeyGrokQuotaSyncIntervalSeconds:       "30",
			SettingKeyGrokCapabilityProbeIntervalSeconds: "99999999",
			SettingKeyGrokSessionValidityCheckInterval:   "10",
			SettingKeyGrokVideoTimeout:                   "99999999",
		},
	}, nil)

	settings := svc.GetGrokRuntimeSettings(context.Background())
	require.Equal(t, minGrokRuntimeIntervalSeconds, settings.QuotaSyncIntervalSeconds)
	require.Equal(t, maxGrokRuntimeIntervalSeconds, settings.CapabilityProbeIntervalSeconds)
	require.Equal(t, minGrokRuntimeIntervalSeconds, settings.SessionValidityCheckSeconds)
	require.Equal(t, maxGrokRuntimeTimeoutSeconds, settings.VideoTimeoutSeconds)
}
