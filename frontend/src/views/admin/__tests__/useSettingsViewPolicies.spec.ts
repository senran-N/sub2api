import { beforeEach, describe, expect, it, vi } from 'vitest'
import { useSettingsViewPolicies } from '../settings/useSettingsViewPolicies'

const {
  getAdminApiKey,
  regenerateAdminApiKey,
  deleteAdminApiKey,
  getOverloadCooldownSettings,
  updateOverloadCooldownSettings,
  getStreamTimeoutSettings,
  updateStreamTimeoutSettings,
  getRectifierSettings,
  updateRectifierSettings,
  getBetaPolicySettings,
  updateBetaPolicySettings
} = vi.hoisted(() => ({
  getAdminApiKey: vi.fn(),
  regenerateAdminApiKey: vi.fn(),
  deleteAdminApiKey: vi.fn(),
  getOverloadCooldownSettings: vi.fn(),
  updateOverloadCooldownSettings: vi.fn(),
  getStreamTimeoutSettings: vi.fn(),
  updateStreamTimeoutSettings: vi.fn(),
  getRectifierSettings: vi.fn(),
  updateRectifierSettings: vi.fn(),
  getBetaPolicySettings: vi.fn(),
  updateBetaPolicySettings: vi.fn()
}))

vi.mock('@/api', () => ({
  adminAPI: {
    settings: {
      getAdminApiKey,
      regenerateAdminApiKey,
      deleteAdminApiKey,
      getOverloadCooldownSettings,
      updateOverloadCooldownSettings,
      getStreamTimeoutSettings,
      updateStreamTimeoutSettings,
      getRectifierSettings,
      updateRectifierSettings,
      getBetaPolicySettings,
      updateBetaPolicySettings
    }
  }
}))

describe('useSettingsViewPolicies', () => {
  beforeEach(() => {
    getAdminApiKey.mockReset()
    regenerateAdminApiKey.mockReset()
    deleteAdminApiKey.mockReset()
    getOverloadCooldownSettings.mockReset()
    updateOverloadCooldownSettings.mockReset()
    getStreamTimeoutSettings.mockReset()
    updateStreamTimeoutSettings.mockReset()
    getRectifierSettings.mockReset()
    updateRectifierSettings.mockReset()
    getBetaPolicySettings.mockReset()
    updateBetaPolicySettings.mockReset()

    getAdminApiKey.mockResolvedValue({ exists: true, masked_key: 'abc...1234' })
    regenerateAdminApiKey.mockResolvedValue({ key: 'abcdefghijklmnopqrstuvwxyz' })
    deleteAdminApiKey.mockResolvedValue({ message: 'deleted' })
    getOverloadCooldownSettings.mockResolvedValue({ enabled: true, cooldown_minutes: 15 })
    updateOverloadCooldownSettings.mockImplementation(async (payload) => payload)
    getStreamTimeoutSettings.mockResolvedValue({
      enabled: true,
      action: 'error',
      temp_unsched_minutes: 5,
      threshold_count: 7,
      threshold_window_minutes: 30
    })
    updateStreamTimeoutSettings.mockImplementation(async (payload) => payload)
    getRectifierSettings.mockResolvedValue({
      enabled: true,
      thinking_signature_enabled: true,
      thinking_budget_enabled: false,
      apikey_signature_enabled: true,
      apikey_signature_patterns: null
    })
    updateRectifierSettings.mockImplementation(async (payload) => payload)
    getBetaPolicySettings.mockResolvedValue({
      rules: [
        {
          beta_token: 'fast-mode-2026-02-01',
          action: 'filter',
          scope: 'oauth',
          error_message: 'blocked'
        }
      ]
    })
    updateBetaPolicySettings.mockImplementation(async (payload) => payload)
  })

  it('loads, creates, copies, and deletes admin api keys', async () => {
    const showError = vi.fn()
    const showSuccess = vi.fn()
    const copyToClipboard = vi.fn().mockResolvedValue(true)
    const state = useSettingsViewPolicies({
      t: (key: string) => key,
      showError,
      showSuccess,
      confirm: vi.fn(() => true),
      copyToClipboard
    })

    await state.loadAdminApiKey()
    expect(state.adminApiKeyExists.value).toBe(true)
    expect(state.adminApiKeyMasked.value).toBe('abc...1234')

    await state.createAdminApiKey()
    expect(state.newAdminApiKey.value).toBe('abcdefghijklmnopqrstuvwxyz')
    expect(state.adminApiKeyMasked.value).toBe('abcdefghij...wxyz')

    await state.copyNewKey()
    expect(copyToClipboard).toHaveBeenCalledWith(
      'abcdefghijklmnopqrstuvwxyz',
      'admin.settings.adminApiKey.keyCopied'
    )

    await state.deleteAdminApiKey()
    expect(deleteAdminApiKey).toHaveBeenCalledTimes(1)
    expect(state.adminApiKeyExists.value).toBe(false)
    expect(showError).not.toHaveBeenCalled()
  })

  it('loads and saves overload, stream timeout, rectifier, and beta policy settings', async () => {
    const showSuccess = vi.fn()
    const state = useSettingsViewPolicies({
      t: (key: string) => key,
      showError: vi.fn(),
      showSuccess,
      confirm: vi.fn(() => true),
      copyToClipboard: vi.fn().mockResolvedValue(true)
    })

    await state.loadOverloadCooldownSettings()
    expect(state.overloadCooldownForm.cooldown_minutes).toBe(15)
    state.overloadCooldownForm.cooldown_minutes = 20
    await state.saveOverloadCooldownSettings()
    expect(updateOverloadCooldownSettings).toHaveBeenCalledWith({
      enabled: true,
      cooldown_minutes: 20
    })

    await state.loadStreamTimeoutSettings()
    expect(state.streamTimeoutForm.action).toBe('error')
    state.streamTimeoutForm.threshold_count = 9
    await state.saveStreamTimeoutSettings()
    expect(updateStreamTimeoutSettings).toHaveBeenCalledWith(
      expect.objectContaining({
        threshold_count: 9
      })
    )

    await state.loadRectifierSettings()
    expect(state.rectifierForm.apikey_signature_patterns).toEqual([])
    state.rectifierForm.apikey_signature_patterns = [' alpha ', '', 'beta']
    await state.saveRectifierSettings()
    expect(updateRectifierSettings).toHaveBeenCalledWith(
      expect.objectContaining({
        apikey_signature_patterns: ['alpha', 'beta']
      })
    )

    await state.loadBetaPolicySettings()
    expect(state.getBetaDisplayName('fast-mode-2026-02-01')).toBe('Fast Mode')
    expect(state.betaPolicyActionOptions.value).toHaveLength(3)
    expect(state.betaPolicyScopeOptions.value).toHaveLength(4)
    await state.saveBetaPolicySettings()
    expect(updateBetaPolicySettings).toHaveBeenCalledWith({
      rules: [
        {
          beta_token: 'fast-mode-2026-02-01',
          action: 'filter',
          scope: 'oauth',
          error_message: 'blocked'
        }
      ]
    })
    expect(showSuccess).toHaveBeenCalledWith('admin.settings.betaPolicy.saved')
  })
})
