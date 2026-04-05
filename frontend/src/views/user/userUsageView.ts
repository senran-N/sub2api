import type { ApiKey, UsageLog } from '@/types'

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
  return log.cache_creation_5m_tokens > 0 || log.cache_creation_1h_tokens > 0
}

export function getUserUsageCacheOverrideBadgeText(log: UsageLog): string {
  return `R-${log.cache_creation_1h_tokens > 0 ? '5m' : '1H'}`
}

export function getUserUsageCacheOverrideLabelKey(log: UsageLog): string {
  return log.cache_creation_1h_tokens > 0
    ? 'usage.cacheTtlOverridden1h'
    : 'usage.cacheTtlOverridden5m'
}

export function getUserUsageTotalTokens(log: UsageLog | null): number {
  if (log == null) {
    return 0
  }

  return (
    log.input_tokens +
    log.output_tokens +
    log.cache_creation_tokens +
    log.cache_read_tokens
  )
}
