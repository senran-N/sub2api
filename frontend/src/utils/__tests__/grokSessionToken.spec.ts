import { describe, expect, it } from 'vitest'
import { findInvalidGrokSessionBatchImportLine, normalizeGrokSessionToken } from '../grokSessionToken'

describe('grokSessionToken', () => {
  it('normalizes a raw session token into a canonical cookie header', () => {
    const rawToken = 'groksessiontoken1234567890abcd'

    expect(normalizeGrokSessionToken(rawToken)).toBe(
      `sso=${rawToken}; sso-rw=${rawToken}`
    )
  })

  it('normalizes a cookie header and preserves supported extra cookies', () => {
    expect(
      normalizeGrokSessionToken(
        'Cookie: sso=abcdefghijklmnopqrstuvwxyz123456; x-anonuserid=anon-1'
      )
    ).toBe(
      'sso=abcdefghijklmnopqrstuvwxyz123456; sso-rw=abcdefghijklmnopqrstuvwxyz123456; x-anonuserid=anon-1'
    )
  })

  it('rejects short or malformed session tokens', () => {
    expect(normalizeGrokSessionToken('abc')).toBeNull()
    expect(normalizeGrokSessionToken('sso=short')).toBeNull()
  })

  it('reports the first invalid non-empty batch line', () => {
    expect(
      findInvalidGrokSessionBatchImportLine(
        'sso=abcdefghijklmnopqrstuvwxyz123456\nabc\nsso=mnopqrstuvwxyzabcdef123456'
      )
    ).toBe(2)
  })
})
