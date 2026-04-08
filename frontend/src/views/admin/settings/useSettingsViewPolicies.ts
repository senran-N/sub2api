import { computed, reactive, ref } from 'vue'
import type { BetaPolicyRule } from '@/api/admin/settings'
import { adminAPI } from '@/api'
import { resolveRequestErrorMessage } from '@/utils/requestError'
import {
  createDefaultBetaPolicyRules,
  createDefaultOverloadCooldownSettings,
  createDefaultRectifierSettings,
  createDefaultStreamTimeoutSettings,
} from './settingsForm'
import {
  getSettingsBetaDisplayName,
  maskSettingsApiKey,
  sanitizeRectifierPatterns
} from './settingsPolicies'

interface SettingsViewPoliciesOptions {
  t: (key: string, params?: Record<string, unknown>) => string
  showError: (message: string) => void
  showSuccess: (message: string) => void
  confirm: (message?: string) => boolean
  copyToClipboard: (text: string, successMessage?: string) => Promise<boolean>
}

export function useSettingsViewPolicies(options: SettingsViewPoliciesOptions) {
  const adminApiKeyLoading = ref(true)
  const adminApiKeyExists = ref(false)
  const adminApiKeyMasked = ref('')
  const adminApiKeyOperating = ref(false)
  const newAdminApiKey = ref('')

  const overloadCooldownLoading = ref(true)
  const overloadCooldownSaving = ref(false)
  const overloadCooldownForm = reactive(createDefaultOverloadCooldownSettings())

  const streamTimeoutLoading = ref(true)
  const streamTimeoutSaving = ref(false)
  const streamTimeoutForm = reactive(createDefaultStreamTimeoutSettings())

  const rectifierLoading = ref(true)
  const rectifierSaving = ref(false)
  const rectifierForm = reactive(createDefaultRectifierSettings())

  const betaPolicyLoading = ref(true)
  const betaPolicySaving = ref(false)
  const betaPolicyForm = reactive({
    rules: createDefaultBetaPolicyRules() as BetaPolicyRule[]
  })

  const betaPolicyActionOptions = computed(() => [
    { value: 'pass', label: options.t('admin.settings.betaPolicy.actionPass') },
    { value: 'filter', label: options.t('admin.settings.betaPolicy.actionFilter') },
    { value: 'block', label: options.t('admin.settings.betaPolicy.actionBlock') }
  ])

  const betaPolicyScopeOptions = computed(() => [
    { value: 'all', label: options.t('admin.settings.betaPolicy.scopeAll') },
    { value: 'oauth', label: options.t('admin.settings.betaPolicy.scopeOAuth') },
    { value: 'apikey', label: options.t('admin.settings.betaPolicy.scopeAPIKey') },
    { value: 'bedrock', label: options.t('admin.settings.betaPolicy.scopeBedrock') }
  ])

  function getBetaDisplayName(token: string): string {
    return getSettingsBetaDisplayName(token)
  }

  function confirmAction(message: string): boolean {
    return options.confirm(message)
  }

  async function loadAdminApiKey() {
    adminApiKeyLoading.value = true
    try {
      const status = await adminAPI.settings.getAdminApiKey()
      adminApiKeyExists.value = status.exists
      adminApiKeyMasked.value = status.masked_key
    } catch (error) {
      console.error('Failed to load admin API key status:', error)
    } finally {
      adminApiKeyLoading.value = false
    }
  }

  async function createAdminApiKey() {
    adminApiKeyOperating.value = true
    try {
      const result = await adminAPI.settings.regenerateAdminApiKey()
      newAdminApiKey.value = result.key
      adminApiKeyExists.value = true
      adminApiKeyMasked.value = maskSettingsApiKey(result.key)
      options.showSuccess(options.t('admin.settings.adminApiKey.keyGenerated'))
    } catch (error) {
      options.showError(resolveRequestErrorMessage(error, options.t('common.unknownError')))
    } finally {
      adminApiKeyOperating.value = false
    }
  }

  async function regenerateAdminApiKey() {
    if (!confirmAction(options.t('admin.settings.adminApiKey.regenerateConfirm'))) {
      return
    }
    await createAdminApiKey()
  }

  async function deleteAdminApiKey() {
    if (!confirmAction(options.t('admin.settings.adminApiKey.deleteConfirm'))) {
      return
    }

    adminApiKeyOperating.value = true
    try {
      await adminAPI.settings.deleteAdminApiKey()
      adminApiKeyExists.value = false
      adminApiKeyMasked.value = ''
      newAdminApiKey.value = ''
      options.showSuccess(options.t('admin.settings.adminApiKey.keyDeleted'))
    } catch (error) {
      options.showError(resolveRequestErrorMessage(error, options.t('common.unknownError')))
    } finally {
      adminApiKeyOperating.value = false
    }
  }

  async function copyNewKey() {
    await options.copyToClipboard(
      newAdminApiKey.value,
      options.t('admin.settings.adminApiKey.keyCopied')
    )
  }

  async function loadOverloadCooldownSettings() {
    overloadCooldownLoading.value = true
    try {
      Object.assign(overloadCooldownForm, await adminAPI.settings.getOverloadCooldownSettings())
    } catch (error) {
      console.error('Failed to load overload cooldown settings:', error)
    } finally {
      overloadCooldownLoading.value = false
    }
  }

  async function saveOverloadCooldownSettings() {
    overloadCooldownSaving.value = true
    try {
      Object.assign(
        overloadCooldownForm,
        await adminAPI.settings.updateOverloadCooldownSettings({
          enabled: overloadCooldownForm.enabled,
          cooldown_minutes: overloadCooldownForm.cooldown_minutes
        })
      )
      options.showSuccess(options.t('admin.settings.overloadCooldown.saved'))
    } catch (error) {
      options.showError(
        `${options.t('admin.settings.overloadCooldown.saveFailed')}: ${resolveRequestErrorMessage(error, options.t('common.unknownError'))}`
      )
    } finally {
      overloadCooldownSaving.value = false
    }
  }

  async function loadStreamTimeoutSettings() {
    streamTimeoutLoading.value = true
    try {
      Object.assign(streamTimeoutForm, await adminAPI.settings.getStreamTimeoutSettings())
    } catch (error) {
      console.error('Failed to load stream timeout settings:', error)
    } finally {
      streamTimeoutLoading.value = false
    }
  }

  async function saveStreamTimeoutSettings() {
    streamTimeoutSaving.value = true
    try {
      Object.assign(
        streamTimeoutForm,
        await adminAPI.settings.updateStreamTimeoutSettings({
          enabled: streamTimeoutForm.enabled,
          action: streamTimeoutForm.action,
          temp_unsched_minutes: streamTimeoutForm.temp_unsched_minutes,
          threshold_count: streamTimeoutForm.threshold_count,
          threshold_window_minutes: streamTimeoutForm.threshold_window_minutes
        })
      )
      options.showSuccess(options.t('admin.settings.streamTimeout.saved'))
    } catch (error) {
      options.showError(
        `${options.t('admin.settings.streamTimeout.saveFailed')}: ${resolveRequestErrorMessage(error, options.t('common.unknownError'))}`
      )
    } finally {
      streamTimeoutSaving.value = false
    }
  }

  async function loadRectifierSettings() {
    rectifierLoading.value = true
    try {
      const settings = await adminAPI.settings.getRectifierSettings()
      Object.assign(rectifierForm, settings, {
        apikey_signature_patterns: sanitizeRectifierPatterns(settings.apikey_signature_patterns)
      })
    } catch (error) {
      console.error('Failed to load rectifier settings:', error)
    } finally {
      rectifierLoading.value = false
    }
  }

  async function saveRectifierSettings() {
    rectifierSaving.value = true
    try {
      const updated = await adminAPI.settings.updateRectifierSettings({
        enabled: rectifierForm.enabled,
        thinking_signature_enabled: rectifierForm.thinking_signature_enabled,
        thinking_budget_enabled: rectifierForm.thinking_budget_enabled,
        apikey_signature_enabled: rectifierForm.apikey_signature_enabled,
        apikey_signature_patterns: sanitizeRectifierPatterns(
          rectifierForm.apikey_signature_patterns
        )
      })
      Object.assign(rectifierForm, updated, {
        apikey_signature_patterns: sanitizeRectifierPatterns(updated.apikey_signature_patterns)
      })
      options.showSuccess(options.t('admin.settings.rectifier.saved'))
    } catch (error) {
      options.showError(
        `${options.t('admin.settings.rectifier.saveFailed')}: ${resolveRequestErrorMessage(error, options.t('common.unknownError'))}`
      )
    } finally {
      rectifierSaving.value = false
    }
  }

  async function loadBetaPolicySettings() {
    betaPolicyLoading.value = true
    try {
      betaPolicyForm.rules = (await adminAPI.settings.getBetaPolicySettings()).rules
    } catch (error) {
      console.error('Failed to load beta policy settings:', error)
    } finally {
      betaPolicyLoading.value = false
    }
  }

  async function saveBetaPolicySettings() {
    betaPolicySaving.value = true
    try {
      betaPolicyForm.rules = (
        await adminAPI.settings.updateBetaPolicySettings({
          rules: betaPolicyForm.rules
        })
      ).rules
      options.showSuccess(options.t('admin.settings.betaPolicy.saved'))
    } catch (error) {
      options.showError(
        `${options.t('admin.settings.betaPolicy.saveFailed')}: ${resolveRequestErrorMessage(error, options.t('common.unknownError'))}`
      )
    } finally {
      betaPolicySaving.value = false
    }
  }

  return {
    adminApiKeyLoading,
    adminApiKeyExists,
    adminApiKeyMasked,
    adminApiKeyOperating,
    newAdminApiKey,
    overloadCooldownLoading,
    overloadCooldownSaving,
    overloadCooldownForm,
    streamTimeoutLoading,
    streamTimeoutSaving,
    streamTimeoutForm,
    rectifierLoading,
    rectifierSaving,
    rectifierForm,
    betaPolicyLoading,
    betaPolicySaving,
    betaPolicyForm,
    betaPolicyActionOptions,
    betaPolicyScopeOptions,
    getBetaDisplayName,
    loadAdminApiKey,
    createAdminApiKey,
    regenerateAdminApiKey,
    deleteAdminApiKey,
    copyNewKey,
    loadOverloadCooldownSettings,
    saveOverloadCooldownSettings,
    loadStreamTimeoutSettings,
    saveStreamTimeoutSettings,
    loadRectifierSettings,
    saveRectifierSettings,
    loadBetaPolicySettings,
    saveBetaPolicySettings
  }
}
