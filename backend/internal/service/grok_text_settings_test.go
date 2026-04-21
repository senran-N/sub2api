//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/senran-N/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestSettingServiceGetGrokTextSettings_UsesDefaults(t *testing.T) {
	svc := NewSettingService(&grokMediaSettingRepoStub{
		values: map[string]string{},
	}, &config.Config{})

	settings := svc.GetGrokTextSettings(context.Background())

	require.Equal(t, DefaultGrokTextSettings(), settings)
}

func TestSettingServiceGetGrokTextSettings_UsesConfiguredValues(t *testing.T) {
	svc := NewSettingService(&grokMediaSettingRepoStub{
		values: map[string]string{
			SettingKeyGrokThinkingSummary:   "true",
			SettingKeyGrokShowSearchSources: "true",
		},
	}, &config.Config{})

	settings := svc.GetGrokTextSettings(context.Background())

	require.True(t, settings.ThinkingSummary)
	require.True(t, settings.ShowSearchSources)
}
