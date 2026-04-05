import { describe, expect, it } from 'vitest'
import {
  buildUserUsageApiKeyOptions,
  formatUserUsageCacheTokens,
  formatUserUsageDuration,
  formatUserUsageEndpoints,
  formatUserUsageTokens
} from '../userUsageView'

describe('userUsageView helpers', () => {
  it('formats usage numbers and endpoints', () => {
    expect(formatUserUsageDuration(250)).toBe('250ms')
    expect(formatUserUsageDuration(1250)).toBe('1.25s')
    expect(formatUserUsageTokens(2500)).toBe('2.50K')
    expect(formatUserUsageTokens(2_500_000)).toBe('2.50M')
    expect(formatUserUsageCacheTokens(1500)).toBe('1.5K')
    expect(formatUserUsageEndpoints('  /v1/chat/completions  ')).toBe('/v1/chat/completions')
    expect(formatUserUsageEndpoints('')).toBe('-')
  })

  it('builds api key options', () => {
    expect(buildUserUsageApiKeyOptions([{ id: 3, name: 'Demo' } as any], 'All')).toEqual([
      { value: null, label: 'All' },
      { value: 3, label: 'Demo' }
    ])
  })
})
