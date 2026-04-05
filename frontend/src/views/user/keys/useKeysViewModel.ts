import { computed, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { keysAPI } from '@/api'
import type { ApiKey } from '@/types'
import { useAppStore } from '@/stores/app'
import { useOnboardingStore } from '@/stores/onboarding'
import { filterUserKeyGroupOptions } from './keysForm'
import { useKeysActionDialogs } from './useKeysActionDialogs'
import { useKeysOverlayState } from './useKeysOverlayState'
import { useKeysViewData } from './useKeysViewData'

export function useKeysViewModel() {
  const { t } = useI18n()
  const appStore = useAppStore()
  const onboardingStore = useOnboardingStore()
  const publicSettings = computed(() => appStore.cachedPublicSettings)

  const dataState = useKeysViewData({
    t,
    showError: appStore.showError,
    fetchPublicSettings: () => appStore.fetchPublicSettings(),
    publicSettings
  })

  const overlayState = useKeysOverlayState({
    apiKeys: dataState.apiKeys,
    groupOptions: dataState.groupOptions,
    filterGroupOptions: filterUserKeyGroupOptions,
    copiedMessage: t('keys.copied')
  })

  const actionState = useKeysActionDialogs({
    t,
    showError: appStore.showError,
    showSuccess: appStore.showSuccess,
    apiKeys: dataState.apiKeys,
    publicSettings,
    keysAPI,
    loadApiKeys: dataState.loadApiKeys,
    isOnboardingSubmitStep: () =>
      onboardingStore.isCurrentStep('[data-tour="key-form-submit"]'),
    advanceOnboardingStep: (delayMs) => onboardingStore.nextStep(delayMs)
  })

  const changeGroup = async (key: ApiKey, newGroupId: number | null) => {
    overlayState.closeGroupSelector()
    await actionState.changeGroup(key, newGroupId)
  }

  onMounted(() => {
    void Promise.all([
      dataState.loadApiKeys(),
      dataState.loadGroups(),
      dataState.loadUserGroupRates(),
      dataState.loadPublicSettings()
    ])
  })

  onUnmounted(() => {
    dataState.dispose()
  })

  return {
    columns: dataState.columns,
    apiKeys: dataState.apiKeys,
    loading: dataState.loading,
    submitting: actionState.submitting,
    usageStats: dataState.usageStats,
    userGroupRates: dataState.userGroupRates,
    pagination: dataState.pagination,
    filterSearch: dataState.filterSearch,
    filterStatus: dataState.filterStatus,
    filterGroupId: dataState.filterGroupId,
    showCreateModal: actionState.showCreateModal,
    showEditModal: actionState.showEditModal,
    showDeleteDialog: actionState.showDeleteDialog,
    showResetQuotaDialog: actionState.showResetQuotaDialog,
    showResetRateLimitDialog: actionState.showResetRateLimitDialog,
    showUseKeyModal: actionState.showUseKeyModal,
    showCcsClientSelect: actionState.showCcsClientSelect,
    selectedKey: actionState.selectedKey,
    copiedKeyId: overlayState.copiedKeyId,
    groupSelectorKeyId: overlayState.groupSelectorKeyId,
    publicSettings: dataState.publicSettings,
    dropdownPosition: overlayState.dropdownPosition,
    moreMenuKeyId: overlayState.moreMenuKeyId,
    moreMenuPosition: overlayState.moreMenuPosition,
    moreMenuRow: overlayState.moreMenuRow,
    selectedKeyForGroup: overlayState.selectedKeyForGroup,
    formData: actionState.formData,
    customKeyError: actionState.customKeyError,
    statusOptions: actionState.statusOptions,
    groupFilterOptions: dataState.groupFilterOptions,
    statusFilterOptions: dataState.statusFilterOptions,
    groupOptions: dataState.groupOptions,
    groupSearchQuery: overlayState.groupSearchQuery,
    filteredGroupOptions: overlayState.filteredGroupOptions,
    loadApiKeys: dataState.loadApiKeys,
    onFilterChange: dataState.onFilterChange,
    onGroupFilterChange: dataState.onGroupFilterChange,
    onStatusFilterChange: dataState.onStatusFilterChange,
    copyToClipboard: overlayState.copyToClipboard,
    setMoreMenuRef: overlayState.setMoreMenuRef,
    toggleMoreMenu: overlayState.toggleMoreMenu,
    closeMoreMenu: overlayState.closeMoreMenu,
    setGroupButtonRef: overlayState.setGroupButtonRef,
    openGroupSelector: overlayState.openGroupSelector,
    closeGroupSelector: overlayState.closeGroupSelector,
    openUseKeyModal: actionState.openUseKeyModal,
    closeUseKeyModal: actionState.closeUseKeyModal,
    handlePageChange: dataState.handlePageChange,
    handlePageSizeChange: dataState.handlePageSizeChange,
    editKey: actionState.editKey,
    toggleKeyStatus: actionState.toggleKeyStatus,
    changeGroup,
    confirmDelete: actionState.confirmDelete,
    handleSubmit: actionState.handleSubmit,
    handleDelete: actionState.handleDelete,
    closeModals: actionState.closeModals,
    confirmResetQuota: actionState.confirmResetQuota,
    setExpirationDays: actionState.setExpirationDays,
    resetQuotaUsed: actionState.resetQuotaUsed,
    confirmResetRateLimit: actionState.confirmResetRateLimit,
    confirmResetRateLimitFromTable: actionState.confirmResetRateLimitFromTable,
    resetRateLimitUsage: actionState.resetRateLimitUsage,
    importToCcswitch: actionState.importToCcswitch,
    handleCcsClientSelect: actionState.handleCcsClientSelect,
    closeCcsClientSelect: actionState.closeCcsClientSelect,
    formatResetTime: actionState.formatResetTime
  }
}
