package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSettingRuntimeReadersNilSafeDefaults(t *testing.T) {
	var settings *SettingService

	fp, metadata, cch := settings.GatewayRuntimeReader().Forwarding(context.Background())
	require.True(t, fp)
	require.False(t, metadata)
	require.False(t, cch)

	require.True(t, settings.GatewayRuntimeReader().ChannelMonitor(context.Background()).Enabled)
	require.False(t, settings.GatewayRuntimeReader().AvailableChannels(context.Background()).Enabled)
	require.Equal(t, DefaultGrokRuntimeSettings(), settings.GrokRuntimeReader().Settings(context.Background()))

	webSearch, err := settings.WebSearchRuntimeReader().Config(context.Background())
	require.NoError(t, err)
	require.NotNil(t, webSearch)

	authDefaults, err := settings.AuthRuntimeReader().SourceDefaults(context.Background())
	require.NoError(t, err)
	require.NotNil(t, authDefaults)
}
