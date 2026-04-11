import type { ApiKey, UsageLog } from '@/types'
import {
  getUsageCacheOverrideBadgeText,
  getUsageCacheOverrideLabelKey,
  getUsageTotalTokens,
  hasUsageCacheCreationBreakdown
} from '@/utils/usageTokens'

export function formatUserUsageDuration(ms: number): string {
  if (ms < 1000) {
    return `${ms.toFixed(0)}ms`
  }

  return `${(ms / 1000).toFixed(2)}s`
}

export function formatUserUsageTokens(value: number): string {
  if (value >= 1_000_000_000) {
    return `${(value / 1_000_000_000).toFixed(2)}B`
  }
  if (value >= 1_000_000) {
    return `${(value / 1_000_000).toFixed(2)}M`
  }
  if (value >= 1_000) {
    return `${(value / 1_000).toFixed(2)}K`
  }

  return value.toLocaleString()
}

export function formatUserUsageCacheTokens(value: number): string {
  if (value >= 1_000_000) {
    return `${(value / 1_000_000).toFixed(1)}M`
  }
  if (value >= 1_000) {
    return `${(value / 1_000).toFixed(1)}K`
  }

  return value.toLocaleString()
}

export function formatUserUsageEndpoints(inboundEndpoint?: string | null): string {
  const inbound = inboundEndpoint?.trim()
  return inbound || '-'
}

export function buildUserUsageApiKeyOptions(
  apiKeys: ApiKey[],
  allLabel: string
): Array<{ value: number | null; label: string }> {
  return [
    { value: null, label: allLabel },
    ...apiKeys.map((key) => ({
      value: key.id,
      label: key.name
    }))
  ]
}

export function hasUserUsageCacheCreationBreakdown(log: UsageLog): boolean {
  return hasUsageCacheCreationBreakdown(log)
}

export function getUserUsageCacheOverrideBadgeText(log: UsageLog): string {
  return getUsageCacheOverrideBadgeText(log)
}

export function getUserUsageCacheOverrideLabelKey(log: UsageLog): string {
  return getUsageCacheOverrideLabelKey(log)
}

export function getUserUsageTotalTokens(log: UsageLog | null): number {
  return getUsageTotalTokens(log)
}
