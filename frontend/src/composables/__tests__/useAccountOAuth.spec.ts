import { beforeEach, describe, expect, it, vi } from 'vitest'

const { showError, generateAuthUrl, exchangeCode } = vi.hoisted(() => ({
  showError: vi.fn(),
  generateAuthUrl: vi.fn(),
  exchangeCode: vi.fn()
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
      exchangeCode
    }
  }
}))

import { useAccountOAuth } from '@/composables/useAccountOAuth'

function createDeferred<T>() {
  let resolve!: (value: T) => void
  const promise = new Promise<T>((res) => {
    resolve = res
  })

  return { promise, resolve }
}

describe('useAccountOAuth', () => {
  beforeEach(() => {
    showError.mockReset()
    generateAuthUrl.mockReset()
    exchangeCode.mockReset()
  })

  it('uses response detail when generating auth urls fails', async () => {
    const oauth = useAccountOAuth()
    generateAuthUrl.mockRejectedValueOnce({
      response: {
        data: {
          detail: 'account-auth-blocked'
        }
      }
    })

    await expect(oauth.generateAuthUrl('oauth')).resolves.toBe(false)

    expect(oauth.error.value).toBe('account-auth-blocked')
    expect(showError).toHaveBeenCalledWith('account-auth-blocked')
  })

  it('falls back to plain error messages when exchange fails', async () => {
    const oauth = useAccountOAuth()
    oauth.authCode.value = 'auth-code'
    oauth.sessionId.value = 'session-id'
    exchangeCode.mockRejectedValueOnce(new Error('network down'))

    await expect(oauth.exchangeAuthCode('oauth')).resolves.toBeNull()

    expect(oauth.error.value).toBe('network down')
    expect(showError).toHaveBeenCalledWith('network down')
  })

  it('ignores stale auth-url responses after reset', async () => {
    const oauth = useAccountOAuth()
    const deferred = createDeferred<{ auth_url: string; session_id: string }>()
    generateAuthUrl.mockReturnValueOnce(deferred.promise)

    const request = oauth.generateAuthUrl('oauth')
    oauth.resetState()

    deferred.resolve({
      auth_url: 'https://stale.example/auth',
      session_id: 'stale-session'
    })

    await expect(request).resolves.toBe(false)
    expect(oauth.authUrl.value).toBe('')
    expect(oauth.sessionId.value).toBe('')
    expect(oauth.loading.value).toBe(false)
    expect(showError).not.toHaveBeenCalled()
  })
})
