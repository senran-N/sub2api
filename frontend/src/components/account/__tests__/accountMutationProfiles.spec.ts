import { describe, expect, it } from 'vitest'
import {
  accountMutationProfileHasSection,
  resolveAccountMutationProfile
} from '../accountMutationProfiles'

describe('accountMutationProfiles', () => {
  it('marks compatible key accounts as compatible credentials with quota limits', () => {
    const profile = resolveAccountMutationProfile('openai', 'apikey')

    expect(accountMutationProfileHasSection(profile, 'compatible-credentials')).toBe(true)
    expect(accountMutationProfileHasSection(profile, 'quota-limits')).toBe(true)
    expect(accountMutationProfileHasSection(profile, 'openai-runtime')).toBe(true)
  })

  it('keeps Grok session on its provider-owned session section', () => {
    const profile = resolveAccountMutationProfile('grok', 'session')

    expect(accountMutationProfileHasSection(profile, 'grok-session')).toBe(true)
    expect(accountMutationProfileHasSection(profile, 'compatible-credentials')).toBe(false)
    expect(accountMutationProfileHasSection(profile, 'temp-unsched')).toBe(true)
  })

  it('keeps Antigravity upstream out of generic compatible and quota sections', () => {
    const profile = resolveAccountMutationProfile('antigravity', 'upstream')

    expect(accountMutationProfileHasSection(profile, 'antigravity-extra')).toBe(true)
    expect(accountMutationProfileHasSection(profile, 'model-restriction')).toBe(true)
    expect(accountMutationProfileHasSection(profile, 'warmup')).toBe(true)
    expect(accountMutationProfileHasSection(profile, 'compatible-credentials')).toBe(false)
    expect(accountMutationProfileHasSection(profile, 'quota-limits')).toBe(false)
  })
})
