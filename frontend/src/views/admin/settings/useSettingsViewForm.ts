import { computed, reactive, ref } from 'vue'
import { adminAPI } from '@/api'
import type { AdminGroup } from '@/types'
import { resolveRequestErrorMessage } from '@/utils/requestError'
import {
  isRegistrationEmailSuffixDomainValid,
  normalizeRegistrationEmailSuffixDomain,
  parseRegistrationEmailSuffixWhitelistInput
} from '@/utils/registrationEmailPolicy'
import {
  addCustomEndpoint,
  addCustomMenuItem,
  addNextDefaultSubscription,
  buildSendTestEmailRequest,
  buildSmtpTestConnectionRequest,
  buildSettingsUpdatePayload,
  createDefaultSettingsForm,
  getSettingsLinuxdoRedirectUrlSuggestion,
  hydrateSettingsForm,
  moveCustomMenuItem,
  removeCustomEndpoint,
  removeCustomMenuItem,
  removeDefaultSubscription as removeDefaultSubscriptionItem
} from './settingsForm'

interface DefaultSubscriptionGroupOption {
  value: number
  label: string
  description: string | null
  platform: AdminGroup['platform']
  subscriptionType: AdminGroup['subscription_type']
  rate: number
  [key: string]: unknown
}

interface SettingsViewFormOptions {
  t: (key: string, params?: Record<string, unknown>) => string
  showError: (message: string) => void
  showSuccess: (message: string) => void
  refreshPublicSettings: (force?: boolean) => Promise<unknown>
  refreshAdminSettings: (force?: boolean) => Promise<unknown>
  copyToClipboard: (text: string, successMessage?: string) => Promise<boolean>
  location?: Pick<Location, 'origin' | 'protocol' | 'host'>
}

const REGISTRATION_EMAIL_SUFFIX_SEPARATOR_KEYS = new Set([
  ' ',
  ',',
  '，',
  'Enter',
  'Tab'
])
export function useSettingsViewForm(options: SettingsViewFormOptions) {
  const loading = ref(true)
  const loadFailed = ref(false)
  const saving = ref(false)
  const testingSmtp = ref(false)
  const sendingTestEmail = ref(false)
  const smtpPasswordManuallyEdited = ref(false)
  const testEmailAddress = ref('')
  const registrationEmailSuffixWhitelistTags = ref<string[]>([])
  const registrationEmailSuffixWhitelistDraft = ref('')
  const subscriptionGroups = ref<AdminGroup[]>([])
  const form = reactive(createDefaultSettingsForm())

  const defaultSubscriptionGroupOptions = computed<DefaultSubscriptionGroupOption[]>(() =>
    subscriptionGroups.value.map((group) => ({
      value: group.id,
      label: group.name,
      description: group.description,
      platform: group.platform,
      subscriptionType: group.subscription_type,
      rate: group.rate_multiplier
    }))
  )

  const linuxdoRedirectUrlSuggestion = computed(() =>
    getSettingsLinuxdoRedirectUrlSuggestion(
      options.location ?? (typeof window === 'undefined' ? undefined : window.location)
    )
  )

  function removeRegistrationEmailSuffixWhitelistTag(suffix: string) {
    registrationEmailSuffixWhitelistTags.value = registrationEmailSuffixWhitelistTags.value.filter(
      (item) => item !== suffix
    )
  }

  function addRegistrationEmailSuffixWhitelistTag(raw: string) {
    const suffix = normalizeRegistrationEmailSuffixDomain(raw)
    if (
      !isRegistrationEmailSuffixDomainValid(suffix) ||
      registrationEmailSuffixWhitelistTags.value.includes(suffix)
    ) {
      return
    }

    registrationEmailSuffixWhitelistTags.value = [
      ...registrationEmailSuffixWhitelistTags.value,
      suffix
    ]
  }

  function commitRegistrationEmailSuffixWhitelistDraft() {
    if (!registrationEmailSuffixWhitelistDraft.value) {
      return
    }

    addRegistrationEmailSuffixWhitelistTag(registrationEmailSuffixWhitelistDraft.value)
    registrationEmailSuffixWhitelistDraft.value = ''
  }

  function handleRegistrationEmailSuffixWhitelistDraftInput() {
    registrationEmailSuffixWhitelistDraft.value = normalizeRegistrationEmailSuffixDomain(
      registrationEmailSuffixWhitelistDraft.value
    )
  }

  function handleRegistrationEmailSuffixWhitelistDraftKeydown(event: KeyboardEvent) {
    if (event.isComposing) {
      return
    }

    if (REGISTRATION_EMAIL_SUFFIX_SEPARATOR_KEYS.has(event.key)) {
      event.preventDefault()
      commitRegistrationEmailSuffixWhitelistDraft()
      return
    }

    if (
      event.key === 'Backspace' &&
      !registrationEmailSuffixWhitelistDraft.value &&
      registrationEmailSuffixWhitelistTags.value.length > 0
    ) {
      registrationEmailSuffixWhitelistTags.value.pop()
    }
  }

  function handleRegistrationEmailSuffixWhitelistPaste(event: ClipboardEvent) {
    const text = event.clipboardData?.getData('text') || ''
    if (!text.trim()) {
      return
    }

    event.preventDefault()
    const tokens = parseRegistrationEmailSuffixWhitelistInput(text)
    for (const token of tokens) {
      addRegistrationEmailSuffixWhitelistTag(token)
    }
  }

  async function setAndCopyLinuxdoRedirectUrl() {
    const url = linuxdoRedirectUrlSuggestion.value
    if (!url) {
      return
    }

    form.linuxdo_connect_redirect_url = url
    await options.copyToClipboard(url, options.t('admin.settings.linuxdo.redirectUrlSetAndCopied'))
  }

  function addMenuItem() {
    addCustomMenuItem(form.custom_menu_items)
  }

  function removeMenuItem(index: number) {
    removeCustomMenuItem(form.custom_menu_items, index)
  }

  function moveMenuItem(index: number, direction: -1 | 1) {
    moveCustomMenuItem(form.custom_menu_items, index, direction)
  }

  function addEndpoint() {
    addCustomEndpoint(form.custom_endpoints)
  }

  function removeEndpoint(index: number) {
    removeCustomEndpoint(form.custom_endpoints, index)
  }

  async function loadSettings() {
    loading.value = true
    loadFailed.value = false

    try {
      const settings = await adminAPI.settings.getSettings()
      registrationEmailSuffixWhitelistTags.value = hydrateSettingsForm(form, settings)
      registrationEmailSuffixWhitelistDraft.value = ''
      smtpPasswordManuallyEdited.value = false
    } catch (error) {
      loadFailed.value = true
      options.showError(
        `${options.t('admin.settings.failedToLoad')}: ${resolveRequestErrorMessage(error, options.t('common.unknownError'))}`
      )
    } finally {
      loading.value = false
    }
  }

  async function loadSubscriptionGroups() {
    try {
      const groups = await adminAPI.groups.getAll()
      subscriptionGroups.value = groups.filter(
        (group) => group.subscription_type === 'subscription' && group.status === 'active'
      )
    } catch (error) {
      console.error('Failed to load subscription groups:', error)
      subscriptionGroups.value = []
    }
  }

  function addDefaultSubscription() {
    addNextDefaultSubscription(form.default_subscriptions, subscriptionGroups.value)
  }

  function removeDefaultSubscription(index: number) {
    removeDefaultSubscriptionItem(form.default_subscriptions, index)
  }

  async function saveSettings() {
    saving.value = true

    try {
      const payloadResult = buildSettingsUpdatePayload(
        form,
        registrationEmailSuffixWhitelistTags.value
      )

      if (!payloadResult.ok) {
        if (payloadResult.error.code === 'duplicate_default_subscription') {
          options.showError(
            options.t('admin.settings.defaults.defaultSubscriptionsDuplicate', {
              groupId: payloadResult.error.groupId
            })
          )
          return
        }

        if (payloadResult.error.code === 'purchase_url_required') {
          options.showError(
            `${options.t('admin.settings.purchase.url')}: URL is required when purchase is enabled`
          )
          return
        }

        options.showError(
          `${options.t('admin.settings.purchase.url')}: must be an absolute http(s) URL (e.g. https://example.com)`
        )
        return
      }

      const updated = await adminAPI.settings.updateSettings(payloadResult.payload)
      registrationEmailSuffixWhitelistTags.value = hydrateSettingsForm(form, updated)
      registrationEmailSuffixWhitelistDraft.value = ''
      smtpPasswordManuallyEdited.value = false
      await options.refreshPublicSettings(true)
      await options.refreshAdminSettings(true)
      options.showSuccess(options.t('admin.settings.settingsSaved'))
    } catch (error) {
      options.showError(
        `${options.t('admin.settings.failedToSave')}: ${resolveRequestErrorMessage(error, options.t('common.unknownError'))}`
      )
    } finally {
      saving.value = false
    }
  }

  async function testSmtpConnection() {
    testingSmtp.value = true

    try {
      const result = await adminAPI.settings.testSmtpConnection(
        buildSmtpTestConnectionRequest(form, smtpPasswordManuallyEdited.value)
      )
      options.showSuccess(result.message || options.t('admin.settings.smtpConnectionSuccess'))
    } catch (error) {
      options.showError(
        `${options.t('admin.settings.failedToTestSmtp')}: ${resolveRequestErrorMessage(error, options.t('common.unknownError'))}`
      )
    } finally {
      testingSmtp.value = false
    }
  }

  async function sendTestEmail() {
    if (!testEmailAddress.value) {
      options.showError(options.t('admin.settings.testEmail.enterRecipientHint'))
      return
    }

    sendingTestEmail.value = true

    try {
      const result = await adminAPI.settings.sendTestEmail(
        buildSendTestEmailRequest(
          form,
          testEmailAddress.value,
          smtpPasswordManuallyEdited.value
        )
      )
      options.showSuccess(result.message || options.t('admin.settings.testEmailSent'))
    } catch (error) {
      options.showError(
        `${options.t('admin.settings.failedToSendTestEmail')}: ${resolveRequestErrorMessage(error, options.t('common.unknownError'))}`
      )
    } finally {
      sendingTestEmail.value = false
    }
  }

  return {
    loading,
    loadFailed,
    saving,
    testingSmtp,
    sendingTestEmail,
    smtpPasswordManuallyEdited,
    testEmailAddress,
    registrationEmailSuffixWhitelistTags,
    registrationEmailSuffixWhitelistDraft,
    subscriptionGroups,
    form,
    defaultSubscriptionGroupOptions,
    linuxdoRedirectUrlSuggestion,
    removeRegistrationEmailSuffixWhitelistTag,
    addRegistrationEmailSuffixWhitelistTag,
    commitRegistrationEmailSuffixWhitelistDraft,
    handleRegistrationEmailSuffixWhitelistDraftInput,
    handleRegistrationEmailSuffixWhitelistDraftKeydown,
    handleRegistrationEmailSuffixWhitelistPaste,
    setAndCopyLinuxdoRedirectUrl,
    addMenuItem,
    removeMenuItem,
    moveMenuItem,
    addEndpoint,
    removeEndpoint,
    loadSettings,
    loadSubscriptionGroups,
    addDefaultSubscription,
    removeDefaultSubscription,
    saveSettings,
    testSmtpConnection,
    sendTestEmail
  }
}
