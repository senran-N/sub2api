import { beforeEach, describe, expect, it, vi } from 'vitest'

const { showError, generateAuthUrl, exchangeCode, refreshAntigravityToken } = vi.hoisted(
  () => ({
    showError: vi.fn(),
    generateAuthUrl: vi.fn(),
    exchangeCode: vi.fn(),
    refreshAntigravityToken: vi.fn()
  })
)

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key
  })
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError
  })
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    antigravity: {
      generateAuthUrl,
      exchangeCode,
      refreshAntigravityToken
    }
  }
}))

import { useAntigravityOAuth } from '@/composables/useAntigravityOAuth'

describe('useAntigravityOAuth', () => {
  beforeEach(() => {
    showError.mockReset()
    generateAuthUrl.mockReset()
    exchangeCode.mockReset()
    refreshAntigravityToken.mockReset()
  })

  it('uses backend detail when generating auth urls fails', async () => {
    const oauth = useAntigravityOAuth()
    generateAuthUrl.mockRejectedValueOnce({
      response: {
        data: {
          detail: 'antigravity-auth-blocked'
        }
      }
    })

    await expect(oauth.generateAuthUrl(null)).resolves.toBe(false)

    expect(oauth.error.value).toBe('antigravity-auth-blocked')
    expect(showError).toHaveBeenCalledWith('antigravity-auth-blocked')
  })

  it('falls back to plain error messages when refresh token validation fails', async () => {
    const oauth = useAntigravityOAuth()
    refreshAntigravityToken.mockRejectedValueOnce(new Error('network down'))

    await expect(oauth.validateRefreshToken('rt-token')).resolves.toBeNull()

    expect(oauth.error.value).toBe('network down')
    expect(showError).not.toHaveBeenCalled()
  })
})
