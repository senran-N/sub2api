import { describe, expect, it, vi } from 'vitest'
import {
  consumeValidationFailureMessage,
  countMultilineEntries,
  extractNormalizedOAuthCallback,
  extractOAuthCallbackPayload,
  parseBatchTokenInput,
  resolveAnthropicExchangeEndpoint,
  resolveOAuthExchangeState,
  resolveOAuthKey,
  resolveOAuthKeyPrefix,
  resolveOAuthImportantNoticeKey,
  runBatchCreateFlow,
  runOAuthExchangeFlow,
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
    expect(parseBatchTokenInput('a\n\n b \n')).toEqual(['a', 'b'])
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

  it('resolves exchange state and anthropic endpoints', () => {
    const onMissingState = vi.fn()

    expect(
      resolveOAuthExchangeState({
        fallbackState: 'fallback-state',
        inputState: '',
        onMissingState,
        authFailedMessage: 'auth failed'
      })
    ).toBe('fallback-state')

    expect(
      resolveOAuthExchangeState({
        fallbackState: '',
        inputState: '',
        onMissingState,
        authFailedMessage: 'auth failed'
      })
    ).toBeNull()
    expect(onMissingState).toHaveBeenCalledWith('auth failed')

    expect(resolveAnthropicExchangeEndpoint('oauth', 'code')).toBe('/admin/accounts/exchange-code')
    expect(resolveAnthropicExchangeEndpoint('setup-token', 'cookie')).toBe('/admin/accounts/setup-token-cookie-auth')
  })

  it('consumes validation failures and controls oauth loading state', async () => {
    const errorRef = { value: 'validation failed' }
    expect(consumeValidationFailureMessage(errorRef)).toBe('validation failed')
    expect(errorRef.value).toBe('')

    const stateRefs = {
      loading: { value: false },
      error: { value: '' }
    }
    const showError = vi.fn()

    await runOAuthExchangeFlow(
      stateRefs,
      async () => {
        throw new Error('exchange failed')
      },
      (error) => (error instanceof Error ? error.message : 'unknown'),
      showError
    )

    expect(stateRefs.loading.value).toBe(false)
    expect(stateRefs.error.value).toBe('exchange failed')
    expect(showError).toHaveBeenCalledWith('exchange failed')
  })

  it('runs batch create flow with aggregated results', async () => {
    const loadingRef = { value: false }
    const errorRef = { value: '' }
    const onComplete = vi.fn()

    await runBatchCreateFlow({
      rawInput: 'first\nsecond',
      emptyInputMessage: 'missing',
      loadingRef,
      errorRef,
      onComplete,
      processEntry: async (entry) => (entry === 'second' ? 'bad token' : null),
      resolveUnexpectedError: () => 'unexpected'
    })

    expect(loadingRef.value).toBe(false)
    expect(errorRef.value).toBe('')
    expect(onComplete).toHaveBeenCalledWith({
      failedCount: 1,
      successCount: 1,
      errors: ['#2: bad token']
    })
  })
})
