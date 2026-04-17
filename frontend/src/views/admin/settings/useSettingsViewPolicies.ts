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

function createLatestRequestTracker() {
  let sequence = 0

  return {
    next() {
      sequence += 1
      return sequence
    },
    isCurrent(requestSequence: number) {
      return requestSequence === sequence
    }
  }
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
  const adminApiKeyRequestTracker = createLatestRequestTracker()
  const overloadCooldownRequestTracker = createLatestRequestTracker()
  const streamTimeoutRequestTracker = createLatestRequestTracker()
  const rectifierRequestTracker = createLatestRequestTracker()
  const betaPolicyRequestTracker = createLatestRequestTracker()

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

  function normalizeBetaPolicyRules(rules: BetaPolicyRule[] | null | undefined): BetaPolicyRule[] {
    if (!Array.isArray(rules)) {
      return createDefaultBetaPolicyRules()
    }

    return rules.map((rule) => {
      const modelWhitelist = Array.isArray(rule.model_whitelist)
        ? rule.model_whitelist.map((pattern) => pattern.trim()).filter(Boolean)
        : []
      const hasWhitelist = modelWhitelist.length > 0
      const fallbackAction = hasWhitelist ? (rule.fallback_action ?? 'pass') : undefined

      return {
        beta_token: rule.beta_token,
        action: rule.action,
        scope: rule.scope,
        error_message: rule.action === 'block' ? rule.error_message?.trim() || undefined : undefined,
        model_whitelist: modelWhitelist,
        fallback_action: fallbackAction,
        fallback_error_message:
          hasWhitelist && fallbackAction === 'block'
            ? rule.fallback_error_message?.trim() || undefined
            : undefined
      }
    })
  }

  function confirmAction(message: string): boolean {
    return options.confirm(message)
  }

  async function loadAdminApiKey() {
    const requestSequence = adminApiKeyRequestTracker.next()
    adminApiKeyLoading.value = true
    try {
      const status = await adminAPI.settings.getAdminApiKey()
      if (!adminApiKeyRequestTracker.isCurrent(requestSequence)) {
        return
      }
      adminApiKeyExists.value = status.exists
      adminApiKeyMasked.value = status.masked_key
    } catch (error) {
      if (!adminApiKeyRequestTracker.isCurrent(requestSequence)) {
        return
      }
      console.error('Failed to load admin API key status:', error)
    } finally {
      if (adminApiKeyRequestTracker.isCurrent(requestSequence)) {
        adminApiKeyLoading.value = false
      }
    }
  }

  async function createAdminApiKey() {
    const requestSequence = adminApiKeyRequestTracker.next()
    adminApiKeyLoading.value = false
    adminApiKeyOperating.value = true
    try {
      const result = await adminAPI.settings.regenerateAdminApiKey()
      if (!adminApiKeyRequestTracker.isCurrent(requestSequence)) {
        return
      }
      newAdminApiKey.value = result.key
      adminApiKeyExists.value = true
      adminApiKeyMasked.value = maskSettingsApiKey(result.key)
      options.showSuccess(options.t('admin.settings.adminApiKey.keyGenerated'))
    } catch (error) {
      if (!adminApiKeyRequestTracker.isCurrent(requestSequence)) {
        return
      }
      options.showError(resolveRequestErrorMessage(error, options.t('common.unknownError')))
    } finally {
      if (adminApiKeyRequestTracker.isCurrent(requestSequence)) {
        adminApiKeyOperating.value = false
      }
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

    const requestSequence = adminApiKeyRequestTracker.next()
    adminApiKeyLoading.value = false
    adminApiKeyOperating.value = true
    try {
      await adminAPI.settings.deleteAdminApiKey()
      if (!adminApiKeyRequestTracker.isCurrent(requestSequence)) {
        return
      }
      adminApiKeyExists.value = false
      adminApiKeyMasked.value = ''
      newAdminApiKey.value = ''
      options.showSuccess(options.t('admin.settings.adminApiKey.keyDeleted'))
    } catch (error) {
      if (!adminApiKeyRequestTracker.isCurrent(requestSequence)) {
        return
      }
      options.showError(resolveRequestErrorMessage(error, options.t('common.unknownError')))
    } finally {
      if (adminApiKeyRequestTracker.isCurrent(requestSequence)) {
        adminApiKeyOperating.value = false
      }
    }
  }

  async function copyNewKey() {
    await options.copyToClipboard(
      newAdminApiKey.value,
      options.t('admin.settings.adminApiKey.keyCopied')
    )
  }

  async function loadOverloadCooldownSettings() {
    const requestSequence = overloadCooldownRequestTracker.next()
    overloadCooldownLoading.value = true
    try {
      const settings = await adminAPI.settings.getOverloadCooldownSettings()
      if (!overloadCooldownRequestTracker.isCurrent(requestSequence)) {
        return
      }
      Object.assign(overloadCooldownForm, settings)
    } catch (error) {
      if (!overloadCooldownRequestTracker.isCurrent(requestSequence)) {
        return
      }
      console.error('Failed to load overload cooldown settings:', error)
    } finally {
      if (overloadCooldownRequestTracker.isCurrent(requestSequence)) {
        overloadCooldownLoading.value = false
      }
    }
  }

  async function saveOverloadCooldownSettings() {
    const requestSequence = overloadCooldownRequestTracker.next()
    overloadCooldownLoading.value = false
    overloadCooldownSaving.value = true
    try {
      const settings = await adminAPI.settings.updateOverloadCooldownSettings({
        enabled: overloadCooldownForm.enabled,
        cooldown_minutes: overloadCooldownForm.cooldown_minutes
      })
      if (!overloadCooldownRequestTracker.isCurrent(requestSequence)) {
        return
      }
      Object.assign(overloadCooldownForm, settings)
      options.showSuccess(options.t('admin.settings.overloadCooldown.saved'))
    } catch (error) {
      if (!overloadCooldownRequestTracker.isCurrent(requestSequence)) {
        return
      }
      options.showError(
        `${options.t('admin.settings.overloadCooldown.saveFailed')}: ${resolveRequestErrorMessage(error, options.t('common.unknownError'))}`
      )
    } finally {
      if (overloadCooldownRequestTracker.isCurrent(requestSequence)) {
        overloadCooldownSaving.value = false
      }
    }
  }

  async function loadStreamTimeoutSettings() {
    const requestSequence = streamTimeoutRequestTracker.next()
    streamTimeoutLoading.value = true
    try {
      const settings = await adminAPI.settings.getStreamTimeoutSettings()
      if (!streamTimeoutRequestTracker.isCurrent(requestSequence)) {
        return
      }
      Object.assign(streamTimeoutForm, settings)
    } catch (error) {
      if (!streamTimeoutRequestTracker.isCurrent(requestSequence)) {
        return
      }
      console.error('Failed to load stream timeout settings:', error)
    } finally {
      if (streamTimeoutRequestTracker.isCurrent(requestSequence)) {
        streamTimeoutLoading.value = false
      }
    }
  }

  async function saveStreamTimeoutSettings() {
    const requestSequence = streamTimeoutRequestTracker.next()
    streamTimeoutLoading.value = false
    streamTimeoutSaving.value = true
    try {
      const settings = await adminAPI.settings.updateStreamTimeoutSettings({
        enabled: streamTimeoutForm.enabled,
        action: streamTimeoutForm.action,
        temp_unsched_minutes: streamTimeoutForm.temp_unsched_minutes,
        threshold_count: streamTimeoutForm.threshold_count,
        threshold_window_minutes: streamTimeoutForm.threshold_window_minutes
      })
      if (!streamTimeoutRequestTracker.isCurrent(requestSequence)) {
        return
      }
      Object.assign(streamTimeoutForm, settings)
      options.showSuccess(options.t('admin.settings.streamTimeout.saved'))
    } catch (error) {
      if (!streamTimeoutRequestTracker.isCurrent(requestSequence)) {
        return
      }
      options.showError(
        `${options.t('admin.settings.streamTimeout.saveFailed')}: ${resolveRequestErrorMessage(error, options.t('common.unknownError'))}`
      )
    } finally {
      if (streamTimeoutRequestTracker.isCurrent(requestSequence)) {
        streamTimeoutSaving.value = false
      }
    }
  }

  async function loadRectifierSettings() {
    const requestSequence = rectifierRequestTracker.next()
    rectifierLoading.value = true
    try {
      const settings = await adminAPI.settings.getRectifierSettings()
      if (!rectifierRequestTracker.isCurrent(requestSequence)) {
        return
      }
      Object.assign(rectifierForm, settings, {
        apikey_signature_patterns: sanitizeRectifierPatterns(settings.apikey_signature_patterns)
      })
    } catch (error) {
      if (!rectifierRequestTracker.isCurrent(requestSequence)) {
        return
      }
      console.error('Failed to load rectifier settings:', error)
    } finally {
      if (rectifierRequestTracker.isCurrent(requestSequence)) {
        rectifierLoading.value = false
      }
    }
  }

  async function saveRectifierSettings() {
    const requestSequence = rectifierRequestTracker.next()
    rectifierLoading.value = false
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
      if (!rectifierRequestTracker.isCurrent(requestSequence)) {
        return
      }
      Object.assign(rectifierForm, updated, {
        apikey_signature_patterns: sanitizeRectifierPatterns(updated.apikey_signature_patterns)
      })
      options.showSuccess(options.t('admin.settings.rectifier.saved'))
    } catch (error) {
      if (!rectifierRequestTracker.isCurrent(requestSequence)) {
        return
      }
      options.showError(
        `${options.t('admin.settings.rectifier.saveFailed')}: ${resolveRequestErrorMessage(error, options.t('common.unknownError'))}`
      )
    } finally {
      if (rectifierRequestTracker.isCurrent(requestSequence)) {
        rectifierSaving.value = false
      }
    }
  }

  async function loadBetaPolicySettings() {
    const requestSequence = betaPolicyRequestTracker.next()
    betaPolicyLoading.value = true
    try {
      const settings = await adminAPI.settings.getBetaPolicySettings()
      if (!betaPolicyRequestTracker.isCurrent(requestSequence)) {
        return
      }
      betaPolicyForm.rules = normalizeBetaPolicyRules(settings.rules)
    } catch (error) {
      if (!betaPolicyRequestTracker.isCurrent(requestSequence)) {
        return
      }
      console.error('Failed to load beta policy settings:', error)
    } finally {
      if (betaPolicyRequestTracker.isCurrent(requestSequence)) {
        betaPolicyLoading.value = false
      }
    }
  }

  async function saveBetaPolicySettings() {
    const requestSequence = betaPolicyRequestTracker.next()
    betaPolicyLoading.value = false
    betaPolicySaving.value = true
    try {
      const normalizedRules = normalizeBetaPolicyRules(betaPolicyForm.rules)
      const updated = await adminAPI.settings.updateBetaPolicySettings({
        rules: normalizedRules
      })
      if (!betaPolicyRequestTracker.isCurrent(requestSequence)) {
        return
      }
      betaPolicyForm.rules = updated.rules
      betaPolicyForm.rules = normalizeBetaPolicyRules(betaPolicyForm.rules)
      options.showSuccess(options.t('admin.settings.betaPolicy.saved'))
    } catch (error) {
      if (!betaPolicyRequestTracker.isCurrent(requestSequence)) {
        return
      }
      options.showError(
        `${options.t('admin.settings.betaPolicy.saveFailed')}: ${resolveRequestErrorMessage(error, options.t('common.unknownError'))}`
      )
    } finally {
      if (betaPolicyRequestTracker.isCurrent(requestSequence)) {
        betaPolicySaving.value = false
      }
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
