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

function createDeferred<T>() {
  let resolve!: (value: T | PromiseLike<T>) => void
  let reject!: (reason?: unknown) => void
  const promise = new Promise<T>((resolvePromise, rejectPromise) => {
    resolve = resolvePromise
    reject = rejectPromise
  })

  return {
    promise,
    resolve,
    reject
  }
}

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
          error_message: 'blocked',
          model_whitelist: [' claude-opus-* ', '', 'claude-opus-4-1'],
          fallback_action: 'block',
          fallback_error_message: '  fallback blocked  '
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

  it('uses the provided confirm handler for regenerate and delete actions', async () => {
    const confirm = vi.fn(() => true)
    const state = useSettingsViewPolicies({
      t: (key: string) => key,
      showError: vi.fn(),
      showSuccess: vi.fn(),
      confirm,
      copyToClipboard: vi.fn().mockResolvedValue(true)
    })

    await state.regenerateAdminApiKey()
    expect(confirm).toHaveBeenNthCalledWith(
      1,
      'admin.settings.adminApiKey.regenerateConfirm'
    )

    await state.deleteAdminApiKey()
    expect(confirm).toHaveBeenNthCalledWith(
      2,
      'admin.settings.adminApiKey.deleteConfirm'
    )
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
    expect(state.betaPolicyForm.rules[0].model_whitelist).toEqual([
      'claude-opus-*',
      'claude-opus-4-1'
    ])
    expect(state.betaPolicyForm.rules[0].fallback_error_message).toBe('fallback blocked')
    await state.saveBetaPolicySettings()
    expect(updateBetaPolicySettings).toHaveBeenCalledWith({
      rules: [
        {
          beta_token: 'fast-mode-2026-02-01',
          action: 'filter',
          scope: 'oauth',
          error_message: undefined,
          model_whitelist: ['claude-opus-*', 'claude-opus-4-1'],
          fallback_action: 'block',
          fallback_error_message: 'fallback blocked'
        }
      ]
    })
    expect(showSuccess).toHaveBeenCalledWith('admin.settings.betaPolicy.saved')
  })

  it('keeps admin api key state bound to the latest action', async () => {
    const firstLoad = createDeferred<{ exists: boolean; masked_key: string }>()
    const createRequest = createDeferred<{ key: string }>()

    getAdminApiKey.mockReset().mockReturnValueOnce(firstLoad.promise)
    regenerateAdminApiKey.mockReset().mockReturnValueOnce(createRequest.promise)

    const showSuccess = vi.fn()
    const state = useSettingsViewPolicies({
      t: (key: string) => key,
      showError: vi.fn(),
      showSuccess,
      confirm: vi.fn(() => true),
      copyToClipboard: vi.fn().mockResolvedValue(true)
    })

    const loadPromise = state.loadAdminApiKey()
    const createPromise = state.createAdminApiKey()

    createRequest.resolve({ key: 'abcdefghijklmnopqrstuvwxyz' })
    await createPromise

    firstLoad.resolve({ exists: true, masked_key: 'stale-mask' })
    await loadPromise

    expect(state.newAdminApiKey.value).toBe('abcdefghijklmnopqrstuvwxyz')
    expect(state.adminApiKeyMasked.value).toBe('abcdefghij...wxyz')
    expect(showSuccess).toHaveBeenCalledWith('admin.settings.adminApiKey.keyGenerated')
    expect(state.adminApiKeyOperating.value).toBe(false)
  })

  it('does not let a stale overload settings load overwrite a newer save', async () => {
    const firstLoad = createDeferred<{ enabled: boolean; cooldown_minutes: number }>()
    const saveRequest = createDeferred<{ enabled: boolean; cooldown_minutes: number }>()

    getOverloadCooldownSettings.mockReset().mockReturnValueOnce(firstLoad.promise)
    updateOverloadCooldownSettings.mockReset().mockReturnValueOnce(saveRequest.promise)

    const showSuccess = vi.fn()
    const state = useSettingsViewPolicies({
      t: (key: string) => key,
      showError: vi.fn(),
      showSuccess,
      confirm: vi.fn(() => true),
      copyToClipboard: vi.fn().mockResolvedValue(true)
    })

    const loadPromise = state.loadOverloadCooldownSettings()
    state.overloadCooldownForm.cooldown_minutes = 42
    const savePromise = state.saveOverloadCooldownSettings()

    saveRequest.resolve({ enabled: true, cooldown_minutes: 42 })
    await savePromise

    firstLoad.resolve({ enabled: true, cooldown_minutes: 15 })
    await loadPromise

    expect(state.overloadCooldownForm.cooldown_minutes).toBe(42)
    expect(showSuccess).toHaveBeenCalledWith('admin.settings.overloadCooldown.saved')
    expect(state.overloadCooldownLoading.value).toBe(false)
    expect(state.overloadCooldownSaving.value).toBe(false)
  })
})
