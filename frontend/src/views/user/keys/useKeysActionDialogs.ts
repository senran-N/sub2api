import { computed, onMounted, onUnmounted, ref, type Ref } from 'vue'
import type { ApiKey, PublicSettings } from '@/types'
import { resolveRequestErrorMessage } from '@/utils/requestError'
import {
  applyUserKeyExpirationPreset,
  buildDefaultUserKeyFormData,
  buildEditUserKeyFormData,
  buildUserKeyExpirationPayload,
  buildUserKeyRateLimitPayload,
  parseUserKeyIpList,
  resolveUserKeyQuotaValue,
  type UserKeyFormData
} from './keysForm'
import {
  buildCcsImportDeeplink,
  formatApiKeyResetTime,
  type CcsClientType
} from './keysView'

interface KeysActionDialogsOptions {
  t: (key: string, params?: Record<string, unknown>) => string
  showError: (message: string) => void
  showSuccess: (message: string) => void
  apiKeys: Ref<ApiKey[]>
  publicSettings: Ref<PublicSettings | null | undefined>
  keysAPI: {
    create: (
      name: string,
      groupId: number,
      customKey: string | undefined,
      ipWhitelist: string[],
      ipBlacklist: string[],
      quota: number,
      expiresInDays: number | undefined,
      rateLimitData: { rate_limit_5h: number; rate_limit_1d: number; rate_limit_7d: number }
    ) => Promise<unknown>
    update: (id: number, payload: Record<string, unknown>) => Promise<unknown>
    delete: (id: number) => Promise<unknown>
    toggleStatus: (id: number, status: 'active' | 'inactive') => Promise<unknown>
  }
  loadApiKeys: () => Promise<void>
  isOnboardingSubmitStep: () => boolean
  advanceOnboardingStep: (delayMs: number) => void
}

export function useKeysActionDialogs(options: KeysActionDialogsOptions) {
  const submitting = ref(false)
  const now = ref(new Date())

  const showCreateModal = ref(false)
  const showEditModal = ref(false)
  const showDeleteDialog = ref(false)
  const showResetQuotaDialog = ref(false)
  const showResetRateLimitDialog = ref(false)
  const showUseKeyModal = ref(false)
  const showCcsClientSelect = ref(false)
  const pendingCcsRow = ref<ApiKey | null>(null)
  const selectedKey = ref<ApiKey | null>(null)
  const formData = ref<UserKeyFormData>(buildDefaultUserKeyFormData())

  const customKeyError = computed(() => {
    if (!formData.value.use_custom_key || !formData.value.custom_key) {
      return ''
    }

    if (formData.value.custom_key.length < 16) {
      return options.t('keys.customKeyTooShort')
    }

    if (!/^[a-zA-Z0-9_-]+$/.test(formData.value.custom_key)) {
      return options.t('keys.customKeyInvalidChars')
    }

    return ''
  })

  const statusOptions = computed(() => [
    { value: 'active', label: options.t('common.active') },
    { value: 'inactive', label: options.t('common.inactive') }
  ])

  function openUseKeyModal(key: ApiKey) {
    selectedKey.value = key
    showUseKeyModal.value = true
  }

  function closeUseKeyModal() {
    showUseKeyModal.value = false
    selectedKey.value = null
  }

  function editKey(key: ApiKey) {
    selectedKey.value = key
    formData.value = buildEditUserKeyFormData(key)
    showEditModal.value = true
  }

  async function toggleKeyStatus(key: ApiKey) {
    const newStatus = key.status === 'active' ? 'inactive' : 'active'

    try {
      await options.keysAPI.toggleStatus(key.id, newStatus)
      options.showSuccess(
        newStatus === 'active'
          ? options.t('keys.keyEnabledSuccess')
          : options.t('keys.keyDisabledSuccess')
      )
      await options.loadApiKeys()
    } catch {
      options.showError(options.t('keys.failedToUpdateStatus'))
    }
  }

  async function changeGroup(key: ApiKey, newGroupId: number | null) {
    if (key.group_id === newGroupId) return

    try {
      await options.keysAPI.update(key.id, { group_id: newGroupId })
      options.showSuccess(options.t('keys.groupChangedSuccess'))
      await options.loadApiKeys()
    } catch {
      options.showError(options.t('keys.failedToChangeGroup'))
    }
  }

  function confirmDelete(key: ApiKey) {
    selectedKey.value = key
    showDeleteDialog.value = true
  }

  async function handleSubmit() {
    if (formData.value.group_id === null) {
      options.showError(options.t('keys.groupRequired'))
      return
    }

    if (!showEditModal.value && formData.value.use_custom_key) {
      if (!formData.value.custom_key) {
        options.showError(options.t('keys.customKeyRequired'))
        return
      }
      if (customKeyError.value) {
        options.showError(customKeyError.value)
        return
      }
    }

    const ipWhitelist = formData.value.enable_ip_restriction
      ? parseUserKeyIpList(formData.value.ip_whitelist)
      : []
    const ipBlacklist = formData.value.enable_ip_restriction
      ? parseUserKeyIpList(formData.value.ip_blacklist)
      : []
    const quota = resolveUserKeyQuotaValue(formData.value.quota)
    const { expiresInDays, expiresAt } = buildUserKeyExpirationPayload(
      formData.value,
      showEditModal.value
    )
    const rateLimitData = buildUserKeyRateLimitPayload(formData.value)

    submitting.value = true
    try {
      if (showEditModal.value && selectedKey.value) {
        await options.keysAPI.update(selectedKey.value.id, {
          name: formData.value.name,
          group_id: formData.value.group_id,
          status: formData.value.status,
          ip_whitelist: ipWhitelist,
          ip_blacklist: ipBlacklist,
          quota,
          expires_at: expiresAt,
          rate_limit_5h: rateLimitData.rate_limit_5h,
          rate_limit_1d: rateLimitData.rate_limit_1d,
          rate_limit_7d: rateLimitData.rate_limit_7d
        })
        options.showSuccess(options.t('keys.keyUpdatedSuccess'))
      } else {
        const customKey = formData.value.use_custom_key ? formData.value.custom_key : undefined
        await options.keysAPI.create(
          formData.value.name,
          formData.value.group_id,
          customKey,
          ipWhitelist,
          ipBlacklist,
          quota,
          expiresInDays,
          rateLimitData
        )
        options.showSuccess(options.t('keys.keyCreatedSuccess'))

        if (options.isOnboardingSubmitStep()) {
          options.advanceOnboardingStep(500)
        }
      }

      closeModals()
      await options.loadApiKeys()
    } catch (error: unknown) {
      options.showError(resolveRequestErrorMessage(error, options.t('keys.failedToSave')))
    } finally {
      submitting.value = false
    }
  }

  async function handleDelete() {
    if (!selectedKey.value) return

    try {
      await options.keysAPI.delete(selectedKey.value.id)
      options.showSuccess(options.t('keys.keyDeletedSuccess'))
      showDeleteDialog.value = false
      await options.loadApiKeys()
    } catch (error: unknown) {
      options.showError(resolveRequestErrorMessage(error, options.t('keys.failedToDelete')))
    }
  }

  function closeModals() {
    showCreateModal.value = false
    showEditModal.value = false
    selectedKey.value = null
    formData.value = buildDefaultUserKeyFormData()
  }

  function confirmResetQuota() {
    showResetQuotaDialog.value = true
  }

  function setExpirationDays(days: number) {
    formData.value = applyUserKeyExpirationPreset(formData.value, days)
  }

  async function resetQuotaUsed() {
    if (!selectedKey.value) return

    showResetQuotaDialog.value = false
    try {
      await options.keysAPI.update(selectedKey.value.id, { reset_quota: true })
      options.showSuccess(options.t('keys.quotaResetSuccess'))
      if (selectedKey.value) {
        selectedKey.value.quota_used = 0
      }
    } catch (error: unknown) {
      options.showError(resolveRequestErrorMessage(error, options.t('keys.failedToResetQuota')))
    }
  }

  function confirmResetRateLimit() {
    showResetRateLimitDialog.value = true
  }

  function confirmResetRateLimitFromTable(row: ApiKey) {
    selectedKey.value = row
    showResetRateLimitDialog.value = true
  }

  async function resetRateLimitUsage() {
    if (!selectedKey.value) return

    showResetRateLimitDialog.value = false
    try {
      await options.keysAPI.update(selectedKey.value.id, { reset_rate_limit_usage: true })
      options.showSuccess(options.t('keys.rateLimitResetSuccess'))
      await options.loadApiKeys()
      const refreshedKey = options.apiKeys.value.find((key) => key.id === selectedKey.value!.id)
      if (refreshedKey) {
        selectedKey.value = refreshedKey
      }
    } catch (error: unknown) {
      options.showError(
        resolveRequestErrorMessage(error, options.t('keys.failedToResetRateLimit'))
      )
    }
  }

  function executeCcsImport(row: ApiKey, clientType: CcsClientType) {
    const deeplink = buildCcsImportDeeplink(
      row,
      options.publicSettings.value,
      clientType,
      window.location.origin
    )

    try {
      window.open(deeplink, '_self')
      setTimeout(() => {
        if (document.hasFocus()) {
          options.showError(options.t('keys.ccSwitchNotInstalled'))
        }
      }, 100)
    } catch {
      options.showError(options.t('keys.ccSwitchNotInstalled'))
    }
  }

  function importToCcswitch(row: ApiKey) {
    const platform = row.group?.platform || 'anthropic'
    if (platform === 'antigravity') {
      pendingCcsRow.value = row
      showCcsClientSelect.value = true
      return
    }

    executeCcsImport(row, platform === 'gemini' ? 'gemini' : 'claude')
  }

  function handleCcsClientSelect(clientType: CcsClientType) {
    if (pendingCcsRow.value) {
      executeCcsImport(pendingCcsRow.value, clientType)
    }
    closeCcsClientSelect()
  }

  function closeCcsClientSelect() {
    showCcsClientSelect.value = false
    pendingCcsRow.value = null
  }

  function formatResetTime(resetAt: string | null): string {
    return formatApiKeyResetTime(resetAt, now.value, options.t)
  }

  let resetTimer: ReturnType<typeof setInterval> | null = null

  onMounted(() => {
    resetTimer = setInterval(() => {
      now.value = new Date()
    }, 60000)
  })

  onUnmounted(() => {
    if (resetTimer) {
      clearInterval(resetTimer)
    }
  })

  return {
    submitting,
    showCreateModal,
    showEditModal,
    showDeleteDialog,
    showResetQuotaDialog,
    showResetRateLimitDialog,
    showUseKeyModal,
    showCcsClientSelect,
    selectedKey,
    formData,
    customKeyError,
    statusOptions,
    openUseKeyModal,
    closeUseKeyModal,
    editKey,
    toggleKeyStatus,
    changeGroup,
    confirmDelete,
    handleSubmit,
    handleDelete,
    closeModals,
    confirmResetQuota,
    setExpirationDays,
    resetQuotaUsed,
    confirmResetRateLimit,
    confirmResetRateLimitFromTable,
    resetRateLimitUsage,
    importToCcswitch,
    handleCcsClientSelect,
    closeCcsClientSelect,
    formatResetTime
  }
}
