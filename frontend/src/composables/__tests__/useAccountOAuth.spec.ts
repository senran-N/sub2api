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
})
