package service

import "context"

// SettingService keeps the historical public methods for compatibility. These
// small readers give runtime code a narrower dependency surface when a caller
// only needs gateway, provider, auth, ops, or websearch settings.

type GatewayRuntimeSettingsReader struct {
	settings *SettingService
}

func (s *SettingService) GatewayRuntimeReader() GatewayRuntimeSettingsReader {
	return GatewayRuntimeSettingsReader{settings: s}
}

func (r GatewayRuntimeSettingsReader) Forwarding(ctx context.Context) (fingerprintUnification, metadataPassthrough, cchSigning bool) {
	if r.settings == nil {
		return true, false, false
	}
	return r.settings.GetGatewayForwardingSettings(ctx)
}

func (r GatewayRuntimeSettingsReader) ChannelMonitor(ctx context.Context) ChannelMonitorRuntime {
	if r.settings == nil {
		return ChannelMonitorRuntime{Enabled: true, DefaultIntervalSeconds: channelMonitorIntervalFallback}
	}
	return r.settings.GetChannelMonitorRuntime(ctx)
}

func (r GatewayRuntimeSettingsReader) AvailableChannels(ctx context.Context) AvailableChannelsRuntime {
	if r.settings == nil {
		return AvailableChannelsRuntime{}
	}
	return r.settings.GetAvailableChannelsRuntime(ctx)
}

type GrokRuntimeSettingsReader struct {
	settings *SettingService
}

func (s *SettingService) GrokRuntimeReader() GrokRuntimeSettingsReader {
	return GrokRuntimeSettingsReader{settings: s}
}

func (r GrokRuntimeSettingsReader) Settings(ctx context.Context) GrokRuntimeSettings {
	if r.settings == nil {
		return DefaultGrokRuntimeSettings()
	}
	return r.settings.GetGrokRuntimeSettings(ctx)
}

type WebSearchRuntimeSettingsReader struct {
	settings *SettingService
}

func (s *SettingService) WebSearchRuntimeReader() WebSearchRuntimeSettingsReader {
	return WebSearchRuntimeSettingsReader{settings: s}
}

func (r WebSearchRuntimeSettingsReader) Config(ctx context.Context) (*WebSearchEmulationConfig, error) {
	if r.settings == nil {
		return &WebSearchEmulationConfig{}, nil
	}
	return r.settings.GetWebSearchEmulationConfig(ctx)
}

type AuthRuntimeSettingsReader struct {
	settings *SettingService
}

func (s *SettingService) AuthRuntimeReader() AuthRuntimeSettingsReader {
	return AuthRuntimeSettingsReader{settings: s}
}

func (r AuthRuntimeSettingsReader) SourceDefaults(ctx context.Context) (*AuthSourceDefaultSettings, error) {
	if r.settings == nil {
		return &AuthSourceDefaultSettings{}, nil
	}
	return r.settings.GetAuthSourceDefaultSettings(ctx)
}

type OpsRuntimeSettingsReader struct {
	settings *SettingService
}

func (s *SettingService) OpsRuntimeReader() OpsRuntimeSettingsReader {
	return OpsRuntimeSettingsReader{settings: s}
}

func (r OpsRuntimeSettingsReader) ChannelMonitor(ctx context.Context) ChannelMonitorRuntime {
	return GatewayRuntimeSettingsReader(r).ChannelMonitor(ctx)
}

func (r OpsRuntimeSettingsReader) AvailableChannels(ctx context.Context) AvailableChannelsRuntime {
	return GatewayRuntimeSettingsReader(r).AvailableChannels(ctx)
}
