package service

import (
	"context"
	"errors"
	"log/slog"
	"sync/atomic"
	"time"

	"golang.org/x/sync/singleflight"
)

// cachedVersionBounds 缓存 Claude Code 版本号上下限（进程内缓存，60s TTL）
type cachedVersionBounds struct {
	min       string
	max       string
	expiresAt int64
}

var versionBoundsCache atomic.Value
var versionBoundsSF singleflight.Group

const versionBoundsCacheTTL = 60 * time.Second
const versionBoundsErrorTTL = 5 * time.Second
const versionBoundsDBTimeout = 5 * time.Second

// cachedBackendMode Backend Mode cache (in-process, 60s TTL)
type cachedBackendMode struct {
	value     bool
	expiresAt int64
}

var backendModeCache atomic.Value
var backendModeSF singleflight.Group

const backendModeCacheTTL = 60 * time.Second
const backendModeErrorTTL = 5 * time.Second
const backendModeDBTimeout = 5 * time.Second

// cachedGatewayForwardingSettings 缓存网关转发行为设置（进程内缓存，60s TTL）
type cachedGatewayForwardingSettings struct {
	fingerprintUnification bool
	metadataPassthrough    bool
	cchSigning             bool
	expiresAt              int64
}

var gatewayForwardingCache atomic.Value
var gatewayForwardingSF singleflight.Group

const gatewayForwardingCacheTTL = 60 * time.Second
const gatewayForwardingErrorTTL = 5 * time.Second
const gatewayForwardingDBTimeout = 5 * time.Second

func storeVersionBoundsCache(min, max string, ttl time.Duration) {
	versionBoundsCache.Store(&cachedVersionBounds{
		min:       min,
		max:       max,
		expiresAt: time.Now().Add(ttl).UnixNano(),
	})
}

func storeBackendModeCache(enabled bool, ttl time.Duration) {
	backendModeCache.Store(&cachedBackendMode{
		value:     enabled,
		expiresAt: time.Now().Add(ttl).UnixNano(),
	})
}

func storeGatewayForwardingCache(fingerprintUnification, metadataPassthrough, cchSigning bool, ttl time.Duration) {
	gatewayForwardingCache.Store(&cachedGatewayForwardingSettings{
		fingerprintUnification: fingerprintUnification,
		metadataPassthrough:    metadataPassthrough,
		cchSigning:             cchSigning,
		expiresAt:              time.Now().Add(ttl).UnixNano(),
	})
}

func refreshSettingRuntimeCaches(settings *SystemSettings) {
	versionBoundsSF.Forget("version_bounds")
	storeVersionBoundsCache(settings.MinClaudeCodeVersion, settings.MaxClaudeCodeVersion, versionBoundsCacheTTL)

	backendModeSF.Forget("backend_mode")
	storeBackendModeCache(settings.BackendModeEnabled, backendModeCacheTTL)

	gatewayForwardingSF.Forget("gateway_forwarding")
	storeGatewayForwardingCache(settings.EnableFingerprintUnification, settings.EnableMetadataPassthrough, settings.EnableCCHSigning, gatewayForwardingCacheTTL)
}

// IsBackendModeEnabled checks if backend mode is enabled.
func (s *SettingService) IsBackendModeEnabled(ctx context.Context) bool {
	if cached, ok := backendModeCache.Load().(*cachedBackendMode); ok && cached != nil {
		if time.Now().UnixNano() < cached.expiresAt {
			return cached.value
		}
	}

	result, _, _ := backendModeSF.Do("backend_mode", func() (any, error) {
		if cached, ok := backendModeCache.Load().(*cachedBackendMode); ok && cached != nil {
			if time.Now().UnixNano() < cached.expiresAt {
				return cached.value, nil
			}
		}

		dbCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), backendModeDBTimeout)
		defer cancel()

		value, err := s.settingRepo.GetValue(dbCtx, SettingKeyBackendModeEnabled)
		if err != nil {
			if errors.Is(err, ErrSettingNotFound) {
				storeBackendModeCache(false, backendModeCacheTTL)
				return false, nil
			}

			slog.Warn("failed to get backend_mode_enabled setting", "error", err)
			storeBackendModeCache(false, backendModeErrorTTL)
			return false, nil
		}

		enabled := value == "true"
		storeBackendModeCache(enabled, backendModeCacheTTL)
		return enabled, nil
	})
	if val, ok := result.(bool); ok {
		return val
	}
	return false
}

// GetGatewayForwardingSettings returns cached gateway forwarding settings.
func (s *SettingService) GetGatewayForwardingSettings(ctx context.Context) (fingerprintUnification, metadataPassthrough, cchSigning bool) {
	if cached, ok := gatewayForwardingCache.Load().(*cachedGatewayForwardingSettings); ok && cached != nil {
		if time.Now().UnixNano() < cached.expiresAt {
			return cached.fingerprintUnification, cached.metadataPassthrough, cached.cchSigning
		}
	}

	type gatewayForwardingResult struct {
		fingerprintUnification bool
		metadataPassthrough    bool
		cchSigning             bool
	}

	val, _, _ := gatewayForwardingSF.Do("gateway_forwarding", func() (any, error) {
		if cached, ok := gatewayForwardingCache.Load().(*cachedGatewayForwardingSettings); ok && cached != nil {
			if time.Now().UnixNano() < cached.expiresAt {
				return gatewayForwardingResult{
					fingerprintUnification: cached.fingerprintUnification,
					metadataPassthrough:    cached.metadataPassthrough,
					cchSigning:             cached.cchSigning,
				}, nil
			}
		}

		dbCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), gatewayForwardingDBTimeout)
		defer cancel()

		values, err := s.settingRepo.GetMultiple(dbCtx, []string{
			SettingKeyEnableFingerprintUnification,
			SettingKeyEnableMetadataPassthrough,
			SettingKeyEnableCCHSigning,
		})
		if err != nil {
			slog.Warn("failed to get gateway forwarding settings", "error", err)
			storeGatewayForwardingCache(true, false, false, gatewayForwardingErrorTTL)
			return gatewayForwardingResult{
				fingerprintUnification: true,
				metadataPassthrough:    false,
				cchSigning:             false,
			}, nil
		}

		fingerprintUnification = true
		if value, ok := values[SettingKeyEnableFingerprintUnification]; ok && value != "" {
			fingerprintUnification = value == "true"
		}
		metadataPassthrough = values[SettingKeyEnableMetadataPassthrough] == "true"
		cchSigning = values[SettingKeyEnableCCHSigning] == "true"
		storeGatewayForwardingCache(fingerprintUnification, metadataPassthrough, cchSigning, gatewayForwardingCacheTTL)
		return gatewayForwardingResult{
			fingerprintUnification: fingerprintUnification,
			metadataPassthrough:    metadataPassthrough,
			cchSigning:             cchSigning,
		}, nil
	})
	if result, ok := val.(gatewayForwardingResult); ok {
		return result.fingerprintUnification, result.metadataPassthrough, result.cchSigning
	}
	return true, false, false
}

// GetClaudeCodeVersionBounds 获取 Claude Code 版本号上下限要求。
func (s *SettingService) GetClaudeCodeVersionBounds(ctx context.Context) (min, max string) {
	if cached, ok := versionBoundsCache.Load().(*cachedVersionBounds); ok && cached != nil {
		if time.Now().UnixNano() < cached.expiresAt {
			return cached.min, cached.max
		}
	}

	type versionBoundsResult struct {
		min string
		max string
	}

	result, err, _ := versionBoundsSF.Do("version_bounds", func() (any, error) {
		if cached, ok := versionBoundsCache.Load().(*cachedVersionBounds); ok && cached != nil {
			if time.Now().UnixNano() < cached.expiresAt {
				return versionBoundsResult{min: cached.min, max: cached.max}, nil
			}
		}

		dbCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), versionBoundsDBTimeout)
		defer cancel()

		values, err := s.settingRepo.GetMultiple(dbCtx, []string{
			SettingKeyMinClaudeCodeVersion,
			SettingKeyMaxClaudeCodeVersion,
		})
		if err != nil {
			slog.Warn("failed to get claude code version bounds setting, skipping version check", "error", err)
			storeVersionBoundsCache("", "", versionBoundsErrorTTL)
			return versionBoundsResult{}, nil
		}

		min = values[SettingKeyMinClaudeCodeVersion]
		max = values[SettingKeyMaxClaudeCodeVersion]
		storeVersionBoundsCache(min, max, versionBoundsCacheTTL)
		return versionBoundsResult{min: min, max: max}, nil
	})
	if err != nil {
		return "", ""
	}
	if bounds, ok := result.(versionBoundsResult); ok {
		return bounds.min, bounds.max
	}
	return "", ""
}
