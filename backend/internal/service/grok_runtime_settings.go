package service

import (
	"context"
	"strconv"
	"strings"
	"time"
)

const (
	defaultGrokOfficialBaseURL                = "https://api.x.ai"
	defaultGrokSessionBaseURL                 = "https://grok.com"
	defaultGrokQuotaSyncIntervalSeconds       = int((15 * time.Minute) / time.Second)
	defaultGrokUsageSyncConcurrency           = 50
	defaultGrokCapabilityProbeIntervalSeconds = int((6 * time.Hour) / time.Second)
	defaultGrokCapabilityProbeConcurrency     = 10
	defaultGrokSessionValidityCheckSeconds    = int((30 * time.Minute) / time.Second)
	defaultGrokVideoTimeoutSeconds            = int((10 * time.Minute) / time.Second)
	minGrokRuntimeIntervalSeconds             = int(time.Minute / time.Second)
	maxGrokRuntimeIntervalSeconds             = int((7 * 24 * time.Hour) / time.Second)
	minGrokRuntimeConcurrency                 = 1
	maxGrokRuntimeConcurrency                 = 128
	minGrokRuntimeTimeoutSeconds              = 30
	maxGrokRuntimeTimeoutSeconds              = int(time.Hour / time.Second)
)

type GrokRuntimeSettings struct {
	OfficialBaseURL                string `json:"grok_official_base_url"`
	SessionBaseURL                 string `json:"grok_session_base_url"`
	QuotaSyncIntervalSeconds       int    `json:"grok_quota_sync_interval_seconds"`
	UsageSyncConcurrency           int    `json:"grok_usage_sync_concurrency"`
	CapabilityProbeIntervalSeconds int    `json:"grok_capability_probe_interval_seconds"`
	CapabilityProbeConcurrency     int    `json:"grok_capability_probe_concurrency"`
	SessionValidityCheckSeconds    int    `json:"grok_session_validity_check_interval"`
	VideoTimeoutSeconds            int    `json:"grok_video_timeout"`
}

func DefaultGrokRuntimeSettings() GrokRuntimeSettings {
	return GrokRuntimeSettings{
		OfficialBaseURL:                defaultGrokOfficialBaseURL,
		SessionBaseURL:                 defaultGrokSessionBaseURL,
		QuotaSyncIntervalSeconds:       defaultGrokQuotaSyncIntervalSeconds,
		UsageSyncConcurrency:           defaultGrokUsageSyncConcurrency,
		CapabilityProbeIntervalSeconds: defaultGrokCapabilityProbeIntervalSeconds,
		CapabilityProbeConcurrency:     defaultGrokCapabilityProbeConcurrency,
		SessionValidityCheckSeconds:    defaultGrokSessionValidityCheckSeconds,
		VideoTimeoutSeconds:            defaultGrokVideoTimeoutSeconds,
	}
}

func normalizeGrokRuntimeBaseURL(raw string, fallback string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		value = fallback
	}
	return strings.TrimRight(value, "/")
}

func normalizeGrokRuntimeIntervalSeconds(raw int, fallback int) int {
	if raw <= 0 {
		return fallback
	}
	if raw < minGrokRuntimeIntervalSeconds {
		return minGrokRuntimeIntervalSeconds
	}
	if raw > maxGrokRuntimeIntervalSeconds {
		return maxGrokRuntimeIntervalSeconds
	}
	return raw
}

func parseGrokRuntimeIntervalSeconds(raw string, fallback int) int {
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return fallback
	}
	return normalizeGrokRuntimeIntervalSeconds(value, fallback)
}

func normalizeGrokRuntimeConcurrency(raw int, fallback int) int {
	if raw <= 0 {
		return fallback
	}
	if raw < minGrokRuntimeConcurrency {
		return minGrokRuntimeConcurrency
	}
	if raw > maxGrokRuntimeConcurrency {
		return maxGrokRuntimeConcurrency
	}
	return raw
}

func parseGrokRuntimeConcurrency(raw string, fallback int) int {
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return fallback
	}
	return normalizeGrokRuntimeConcurrency(value, fallback)
}

func normalizeGrokRuntimeTimeoutSeconds(raw int, fallback int) int {
	if raw <= 0 {
		return fallback
	}
	if raw < minGrokRuntimeTimeoutSeconds {
		return minGrokRuntimeTimeoutSeconds
	}
	if raw > maxGrokRuntimeTimeoutSeconds {
		return maxGrokRuntimeTimeoutSeconds
	}
	return raw
}

func parseGrokRuntimeTimeoutSeconds(raw string, fallback int) int {
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return fallback
	}
	return normalizeGrokRuntimeTimeoutSeconds(value, fallback)
}

func (s GrokRuntimeSettings) QuotaSyncInterval() time.Duration {
	return time.Duration(
		normalizeGrokRuntimeIntervalSeconds(s.QuotaSyncIntervalSeconds, defaultGrokQuotaSyncIntervalSeconds),
	) * time.Second
}

func (s GrokRuntimeSettings) CapabilityProbeInterval() time.Duration {
	return time.Duration(
		normalizeGrokRuntimeIntervalSeconds(s.CapabilityProbeIntervalSeconds, defaultGrokCapabilityProbeIntervalSeconds),
	) * time.Second
}

func (s GrokRuntimeSettings) UsageSyncWorkers() int {
	return normalizeGrokRuntimeConcurrency(s.UsageSyncConcurrency, defaultGrokUsageSyncConcurrency)
}

func (s GrokRuntimeSettings) CapabilityProbeWorkers() int {
	return normalizeGrokRuntimeConcurrency(s.CapabilityProbeConcurrency, defaultGrokCapabilityProbeConcurrency)
}

func (s GrokRuntimeSettings) SessionValidityCheckInterval() time.Duration {
	return time.Duration(
		normalizeGrokRuntimeIntervalSeconds(s.SessionValidityCheckSeconds, defaultGrokSessionValidityCheckSeconds),
	) * time.Second
}

func (s GrokRuntimeSettings) VideoTimeout() time.Duration {
	return time.Duration(
		normalizeGrokRuntimeTimeoutSeconds(s.VideoTimeoutSeconds, defaultGrokVideoTimeoutSeconds),
	) * time.Second
}

func (s *SettingService) GetGrokRuntimeSettings(ctx context.Context) GrokRuntimeSettings {
	defaults := DefaultGrokRuntimeSettings()
	if s == nil || s.settingRepo == nil {
		return defaults
	}

	values, err := s.settingRepo.GetMultiple(ctx, []string{
		SettingKeyGrokOfficialBaseURL,
		SettingKeyGrokSessionBaseURL,
		SettingKeyGrokQuotaSyncIntervalSeconds,
		SettingKeyGrokUsageSyncConcurrency,
		SettingKeyGrokCapabilityProbeIntervalSeconds,
		SettingKeyGrokCapabilityProbeConcurrency,
		SettingKeyGrokSessionValidityCheckInterval,
		SettingKeyGrokVideoTimeout,
	})
	if err != nil {
		return defaults
	}

	return GrokRuntimeSettings{
		OfficialBaseURL: normalizeGrokRuntimeBaseURL(
			values[SettingKeyGrokOfficialBaseURL],
			defaults.OfficialBaseURL,
		),
		SessionBaseURL: normalizeGrokRuntimeBaseURL(
			values[SettingKeyGrokSessionBaseURL],
			defaults.SessionBaseURL,
		),
		QuotaSyncIntervalSeconds: parseGrokRuntimeIntervalSeconds(
			values[SettingKeyGrokQuotaSyncIntervalSeconds],
			defaults.QuotaSyncIntervalSeconds,
		),
		UsageSyncConcurrency: parseGrokRuntimeConcurrency(
			values[SettingKeyGrokUsageSyncConcurrency],
			defaults.UsageSyncConcurrency,
		),
		CapabilityProbeIntervalSeconds: parseGrokRuntimeIntervalSeconds(
			values[SettingKeyGrokCapabilityProbeIntervalSeconds],
			defaults.CapabilityProbeIntervalSeconds,
		),
		CapabilityProbeConcurrency: parseGrokRuntimeConcurrency(
			values[SettingKeyGrokCapabilityProbeConcurrency],
			defaults.CapabilityProbeConcurrency,
		),
		SessionValidityCheckSeconds: parseGrokRuntimeIntervalSeconds(
			values[SettingKeyGrokSessionValidityCheckInterval],
			defaults.SessionValidityCheckSeconds,
		),
		VideoTimeoutSeconds: parseGrokRuntimeTimeoutSeconds(
			values[SettingKeyGrokVideoTimeout],
			defaults.VideoTimeoutSeconds,
		),
	}
}
