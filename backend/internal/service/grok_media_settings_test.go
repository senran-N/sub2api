//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/senran-N/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestSettingServiceGetGrokMediaSettings_UsesDefaultsForInvalidValues(t *testing.T) {
	svc := NewSettingService(&grokMediaSettingRepoStub{
		values: map[string]string{
			SettingKeyGrokImageOutputFormat:        "invalid",
			SettingKeyGrokVideoOutputFormat:        "bad",
			SettingKeyGrokMediaProxyEnabled:        "false",
			SettingKeyGrokMediaCacheRetentionHours: "-1",
		},
	}, &config.Config{})

	settings := svc.GetGrokMediaSettings(context.Background())

	require.Equal(t, GrokMediaOutputFormatLocalURL, settings.ImageOutputFormat)
	require.Equal(t, GrokMediaOutputFormatLocalURL, settings.VideoOutputFormat)
	require.False(t, settings.MediaProxyEnabled)
	require.Equal(t, defaultGrokMediaCacheRetentionHours, settings.MediaCacheRetentionHours)
	require.Equal(t, time.Duration(defaultGrokMediaCacheRetentionHours)*time.Hour, svc.GetGrokMediaCacheRetention(context.Background()))
}
