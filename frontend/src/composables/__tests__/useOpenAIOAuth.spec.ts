import { beforeEach, describe, expect, it, vi } from 'vitest'

const {
  showError,
  generateAuthUrl,
  exchangeCode,
  refreshOpenAIToken
} = vi.hoisted(() => ({
  showError: vi.fn(),
  generateAuthUrl: vi.fn(),
  exchangeCode: vi.fn(),
  refreshOpenAIToken: vi.fn()
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError
  })
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    accounts: {
      generateAuthUrl,
      exchangeCode,
      refreshOpenAIToken
    }
  }
}))

import { useOpenAIOAuth } from '@/composables/useOpenAIOAuth'

function createDeferred<T>() {
  let resolve!: (value: T) => void
  const promise = new Promise<T>((res) => {
    resolve = res
  })

  return { promise, resolve }
}

describe('useOpenAIOAuth', () => {
  beforeEach(() => {
    showError.mockReset()
    generateAuthUrl.mockReset()
    exchangeCode.mockReset()
    refreshOpenAIToken.mockReset()
  })

  it('should keep client_id when token response contains it', () => {
    const oauth = useOpenAIOAuth({ platform: 'openai' })
    const creds = oauth.buildCredentials({
      access_token: 'at',
      refresh_token: 'rt',
      client_id: 'app_sora_client',
      expires_at: 1700000000
    })

    expect(creds.client_id).toBe('app_sora_client')
    expect(creds.access_token).toBe('at')
    expect(creds.refresh_token).toBe('rt')
  })

  it('should keep legacy behavior when client_id is missing', () => {
    const oauth = useOpenAIOAuth({ platform: 'openai' })
    const creds = oauth.buildCredentials({
      access_token: 'at',
      refresh_token: 'rt',
      expires_at: 1700000000
    })

    expect(Object.prototype.hasOwnProperty.call(creds, 'client_id')).toBe(false)
    expect(creds.access_token).toBe('at')
    expect(creds.refresh_token).toBe('rt')
  })

  it('prefers backend detail when generating auth url fails', async () => {
    const oauth = useOpenAIOAuth()
    generateAuthUrl.mockRejectedValueOnce({
      response: {
        data: {
          detail: 'openai-auth-blocked'
        }
      }
    })

    await expect(oauth.generateAuthUrl()).resolves.toBe(false)

    expect(oauth.error.value).toBe('openai-auth-blocked')
    expect(showError).toHaveBeenCalledWith('openai-auth-blocked')
  })

  it('falls back to plain error messages when validating refresh tokens', async () => {
    const oauth = useOpenAIOAuth()
    refreshOpenAIToken.mockRejectedValueOnce(new Error('network down'))

    await expect(oauth.validateRefreshToken('rt-token')).resolves.toBeNull()

    expect(oauth.error.value).toBe('network down')
    expect(showError).toHaveBeenCalledWith('network down')
  })

  it('ignores stale auth-url responses after reset', async () => {
    const oauth = useOpenAIOAuth()
    const deferred = createDeferred<{ auth_url: string; session_id: string }>()
    generateAuthUrl.mockReturnValueOnce(deferred.promise)

    const request = oauth.generateAuthUrl()
    oauth.resetState()

    deferred.resolve({
      auth_url: 'https://stale.example/auth?state=old-state',
      session_id: 'stale-session'
    })

    await expect(request).resolves.toBe(false)
    expect(oauth.authUrl.value).toBe('')
    expect(oauth.sessionId.value).toBe('')
    expect(oauth.oauthState.value).toBe('')
    expect(oauth.loading.value).toBe(false)
    expect(showError).not.toHaveBeenCalled()
  })
})
