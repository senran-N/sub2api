package service

import (
	"context"
	"strconv"
	"strings"
	"time"
)

const (
	GrokMediaOutputFormatLocalURL    = "local_url"
	GrokMediaOutputFormatUpstreamURL = "upstream_url"
	GrokMediaOutputFormatMarkdown    = "markdown"
	GrokMediaOutputFormatBase64      = "base64"
	GrokMediaOutputFormatHTML        = "html"

	defaultGrokMediaCacheRetentionHours = 72
)

type GrokMediaSettings struct {
	ImageOutputFormat        string `json:"grok_image_output_format"`
	VideoOutputFormat        string `json:"grok_video_output_format"`
	MediaProxyEnabled        bool   `json:"grok_media_proxy_enabled"`
	MediaCacheRetentionHours int    `json:"grok_media_cache_retention_hours"`
}

func DefaultGrokMediaSettings() GrokMediaSettings {
	return GrokMediaSettings{
		ImageOutputFormat:        GrokMediaOutputFormatUpstreamURL,
		VideoOutputFormat:        GrokMediaOutputFormatUpstreamURL,
		MediaProxyEnabled:        true,
		MediaCacheRetentionHours: defaultGrokMediaCacheRetentionHours,
	}
}

func normalizeGrokImageOutputFormat(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case GrokMediaOutputFormatUpstreamURL:
		return GrokMediaOutputFormatUpstreamURL
	case GrokMediaOutputFormatMarkdown:
		return GrokMediaOutputFormatMarkdown
	case GrokMediaOutputFormatBase64:
		return GrokMediaOutputFormatBase64
	default:
		return DefaultGrokMediaSettings().ImageOutputFormat
	}
}

func normalizeGrokVideoOutputFormat(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case GrokMediaOutputFormatUpstreamURL:
		return GrokMediaOutputFormatUpstreamURL
	case GrokMediaOutputFormatHTML:
		return GrokMediaOutputFormatHTML
	default:
		return DefaultGrokMediaSettings().VideoOutputFormat
	}
}

func normalizeGrokMediaCacheRetentionHours(raw int) int {
	if raw <= 0 {
		return defaultGrokMediaCacheRetentionHours
	}
	return raw
}

func parseGrokMediaCacheRetentionHours(raw string) int {
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return defaultGrokMediaCacheRetentionHours
	}
	return normalizeGrokMediaCacheRetentionHours(value)
}

func (s *SettingService) GetGrokMediaSettings(ctx context.Context) GrokMediaSettings {
	defaults := DefaultGrokMediaSettings()
	if s == nil || s.settingRepo == nil {
		return defaults
	}

	values, err := s.settingRepo.GetMultiple(ctx, []string{
		SettingKeyGrokImageOutputFormat,
		SettingKeyGrokVideoOutputFormat,
		SettingKeyGrokMediaProxyEnabled,
		SettingKeyGrokMediaCacheRetentionHours,
	})
	if err != nil {
		return defaults
	}

	return GrokMediaSettings{
		ImageOutputFormat:        normalizeGrokImageOutputFormat(values[SettingKeyGrokImageOutputFormat]),
		VideoOutputFormat:        normalizeGrokVideoOutputFormat(values[SettingKeyGrokVideoOutputFormat]),
		MediaProxyEnabled:        !isFalseSettingValue(values[SettingKeyGrokMediaProxyEnabled]),
		MediaCacheRetentionHours: parseGrokMediaCacheRetentionHours(values[SettingKeyGrokMediaCacheRetentionHours]),
	}
}

func (s *SettingService) GetGrokMediaCacheRetention(ctx context.Context) time.Duration {
	settings := s.GetGrokMediaSettings(ctx)
	return time.Duration(settings.MediaCacheRetentionHours) * time.Hour
}
