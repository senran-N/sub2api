import { beforeEach, describe, expect, it, vi } from 'vitest'

const { showError, generateAuthUrl, exchangeCode, getCapabilities } = vi.hoisted(() => ({
  showError: vi.fn(),
  generateAuthUrl: vi.fn(),
  exchangeCode: vi.fn(),
  getCapabilities: vi.fn()
}))

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
    gemini: {
      generateAuthUrl,
      exchangeCode,
      getCapabilities
    }
  }
}))

import { useGeminiOAuth } from '@/composables/useGeminiOAuth'

describe('useGeminiOAuth', () => {
  beforeEach(() => {
    showError.mockReset()
    generateAuthUrl.mockReset()
    exchangeCode.mockReset()
    getCapabilities.mockReset()
  })

  it('uses backend detail when generating auth urls fails', async () => {
    const oauth = useGeminiOAuth()
    generateAuthUrl.mockRejectedValueOnce({
      response: {
        data: {
          detail: 'gemini-auth-blocked'
        }
      }
    })

    await expect(oauth.generateAuthUrl(null)).resolves.toBe(false)

    expect(oauth.error.value).toBe('gemini-auth-blocked')
    expect(showError).toHaveBeenCalledWith('gemini-auth-blocked')
  })

  it('maps missing project_id exchange failures to the dedicated translation key', async () => {
    const oauth = useGeminiOAuth()
    exchangeCode.mockRejectedValueOnce({
      response: {
        data: {
          detail: 'missing project_id for oauth exchange'
        }
      }
    })

    await expect(
      oauth.exchangeAuthCode({
        code: 'code',
        sessionId: 'session',
        state: 'state'
      })
    ).resolves.toBeNull()

    expect(oauth.error.value).toBe('admin.accounts.oauth.gemini.missingProjectId')
    expect(showError).toHaveBeenCalledWith('admin.accounts.oauth.gemini.missingProjectId')
  })
})
