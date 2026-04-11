import { describe, expect, it } from 'vitest'
import {
  getUsageCacheOverrideBadgeText,
  getUsageCacheOverrideLabelKey,
  getUsageTotalTokens,
  hasUsageCacheCreationBreakdown
} from '../usageTokens'

describe('usageTokens helpers', () => {
  it('keeps cache ttl override badge text aligned with billed tier', () => {
    expect(
      getUsageCacheOverrideBadgeText({
        cache_creation_5m_tokens: 0,
        cache_creation_1h_tokens: 300
      })
    ).toBe('R-1h')
    expect(
      getUsageCacheOverrideLabelKey({
        cache_creation_5m_tokens: 0,
        cache_creation_1h_tokens: 300
      })
    ).toBe('usage.cacheTtlOverridden1h')
    expect(
      getUsageCacheOverrideBadgeText({
        cache_creation_5m_tokens: 200,
        cache_creation_1h_tokens: 0
      })
    ).toBe('R-5m')
    expect(
      getUsageCacheOverrideLabelKey({
        cache_creation_5m_tokens: 200,
        cache_creation_1h_tokens: 0
      })
    ).toBe('usage.cacheTtlOverridden5m')
  })

  it('derives cache breakdown and total token counts', () => {
    expect(
      hasUsageCacheCreationBreakdown({
        cache_creation_5m_tokens: 0,
        cache_creation_1h_tokens: 0
      })
    ).toBe(false)
    expect(
      hasUsageCacheCreationBreakdown({
        cache_creation_5m_tokens: 200,
        cache_creation_1h_tokens: 0
      })
    ).toBe(true)
    expect(
      getUsageTotalTokens({
        input_tokens: 4,
        output_tokens: 5,
        cache_creation_tokens: 6,
        cache_read_tokens: 7
      })
    ).toBe(22)
    expect(getUsageTotalTokens(null)).toBe(0)
  })
})
