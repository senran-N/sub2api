import type { UsageLog } from '@/types'

type UsageCacheBreakdownLike = Pick<
  UsageLog,
  'cache_creation_5m_tokens' | 'cache_creation_1h_tokens'
>

type UsageTokenTotalsLike = Pick<
  UsageLog,
  'input_tokens' | 'output_tokens' | 'cache_creation_tokens' | 'cache_read_tokens'
>

export function hasUsageCacheCreationBreakdown(
  log: UsageCacheBreakdownLike
): boolean {
  return log.cache_creation_5m_tokens > 0 || log.cache_creation_1h_tokens > 0
}

export function getUsageCacheOverrideTier(
  log: UsageCacheBreakdownLike
): '5m' | '1h' {
  return log.cache_creation_1h_tokens > 0 ? '1h' : '5m'
}

export function getUsageCacheOverrideBadgeText(
  log: UsageCacheBreakdownLike
): string {
  return `R-${getUsageCacheOverrideTier(log)}`
}

export function getUsageCacheOverrideLabelKey(
  log: UsageCacheBreakdownLike
): 'usage.cacheTtlOverridden5m' | 'usage.cacheTtlOverridden1h' {
  return getUsageCacheOverrideTier(log) === '1h'
    ? 'usage.cacheTtlOverridden1h'
    : 'usage.cacheTtlOverridden5m'
}

export function getUsageTotalTokens(log: UsageTokenTotalsLike | null): number {
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
