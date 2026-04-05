import { describe, expect, it } from 'vitest'
import {
  countMultilineEntries,
  extractNormalizedOAuthCallback,
  extractOAuthCallbackPayload,
  resolveOAuthKey,
  resolveOAuthKeyPrefix,
  resolveOAuthImportantNoticeKey,
  shouldExtractOAuthCallback
} from '../oauthAuthorizationFlowHelpers'

describe('oauthAuthorizationFlowHelpers', () => {
  it('resolves platform-specific oauth key prefixes', () => {
    expect(resolveOAuthKeyPrefix('openai')).toBe('admin.accounts.oauth.openai')
    expect(resolveOAuthKeyPrefix('sora')).toBe('admin.accounts.oauth.openai')
    expect(resolveOAuthKeyPrefix('gemini')).toBe('admin.accounts.oauth.gemini')
    expect(resolveOAuthKeyPrefix('antigravity')).toBe('admin.accounts.oauth.antigravity')
    expect(resolveOAuthKeyPrefix('anthropic')).toBe('admin.accounts.oauth')
    expect(resolveOAuthKey('openai', 'title')).toBe('admin.accounts.oauth.openai.title')
  })

  it('counts non-empty multiline entries', () => {
    expect(countMultilineEntries('a\n\n b \n')).toBe(2)
    expect(countMultilineEntries('')).toBe(0)
  })

  it('extracts oauth callback payload from url or query fragment text', () => {
    expect(
      extractOAuthCallbackPayload('http://localhost:8085/callback?code=abc123&state=xyz987')
    ).toEqual({
      authCode: 'abc123',
      oauthState: 'xyz987'
    })

    expect(
      extractOAuthCallbackPayload('callback?foo=1&code=token_only&state=state_only')
    ).toEqual({
      authCode: 'token_only',
      oauthState: 'state_only'
    })

    expect(extractOAuthCallbackPayload('plain-auth-code')).toEqual({
      authCode: null,
      oauthState: null
    })
  })

  it('resolves platform-specific oauth behavior flags', () => {
    expect(shouldExtractOAuthCallback('openai')).toBe(true)
    expect(shouldExtractOAuthCallback('sora')).toBe(true)
    expect(shouldExtractOAuthCallback('anthropic')).toBe(false)
    expect(resolveOAuthImportantNoticeKey('openai')).toBe('admin.accounts.oauth.openai.importantNotice')
    expect(resolveOAuthImportantNoticeKey('antigravity')).toBe('admin.accounts.oauth.antigravity.importantNotice')
    expect(resolveOAuthImportantNoticeKey('gemini')).toBeNull()
  })

  it('normalizes callback url input only for supported platforms', () => {
    expect(
      extractNormalizedOAuthCallback('openai', 'http://localhost:8085/callback?code=abc123&state=xyz987')
    ).toEqual({
      authCode: 'abc123',
      nextInputValue: 'abc123',
      oauthState: 'xyz987'
    })

    expect(extractNormalizedOAuthCallback('anthropic', 'callback?code=abc123&state=xyz987')).toEqual({
      authCode: null,
      nextInputValue: null,
      oauthState: null
    })
  })
})
