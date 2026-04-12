import { computed, reactive, ref, watch } from 'vue'
import { adminAPI } from '@/api/admin'
import type { AdminGroup } from '@/types'
import { resolveRequestErrorMessage } from '@/utils/requestError'
import {
  applyGroupFormPlatformRules,
  applyGroupFormSubscriptionTypeRules,
  buildCreateGroupPayload,
  buildUpdateGroupPayload,
  createDefaultCreateGroupForm,
  createDefaultEditGroupForm,
  hydrateEditGroupForm,
  resetCreateGroupForm,
  resetEditGroupForm,
  toggleModelScope,
  type CreateGroupForm,
  type EditGroupForm
} from './groupsForm'
import { useGroupRoutingRules } from './useGroupRoutingRules'

interface GroupsViewManagementOptions {
  t: (key: string, params?: Record<string, unknown>) => string
  showError: (message: string) => void
  showSuccess: (message: string) => void
  loadGroups: () => Promise<void>
  isCurrentOnboardingStep: (selector: string) => boolean
  advanceOnboarding: (delay?: number) => void
}

export function useGroupsViewManagement(options: GroupsViewManagementOptions) {
  const showCreateModal = ref(false)
  const showEditModal = ref(false)
  const showDeleteDialog = ref(false)
  const showRateMultipliersModal = ref(false)
  const submitting = ref(false)
  const editingGroup = ref<AdminGroup | null>(null)
  const deletingGroup = ref<AdminGroup | null>(null)
  const rateMultipliersGroup = ref<AdminGroup | null>(null)

  const createForm = reactive<CreateGroupForm>(createDefaultCreateGroupForm())
  const editForm = reactive<EditGroupForm>(createDefaultEditGroupForm())

  const {
    rules: createModelRoutingRules,
    accountSearchKeyword: createAccountSearchKeyword,
    accountSearchResults: createAccountSearchResults,
    showAccountDropdown: createShowAccountDropdown,
    getRuleRenderKey: getCreateRuleRenderKey,
    getRuleSearchKey: getCreateRuleSearchKey,
    searchAccountsByRule: searchCreateAccountsByRule,
    selectAccount: selectCreateAccount,
    removeSelectedAccount: removeCreateSelectedAccount,
    onAccountSearchFocus: onCreateAccountSearchFocus,
    addRoutingRule: addCreateRoutingRule,
    removeRoutingRule: removeCreateRoutingRule,
    hideAllDropdowns: hideCreateAccountDropdowns,
    reset: resetCreateRoutingRules
  } = useGroupRoutingRules('create', () => createForm.platform)

  const {
    rules: editModelRoutingRules,
    accountSearchKeyword: editAccountSearchKeyword,
    accountSearchResults: editAccountSearchResults,
    showAccountDropdown: editShowAccountDropdown,
    getRuleRenderKey: getEditRuleRenderKey,
    getRuleSearchKey: getEditRuleSearchKey,
    searchAccountsByRule: searchEditAccountsByRule,
    selectAccount: selectEditAccount,
    removeSelectedAccount: removeEditSelectedAccount,
    onAccountSearchFocus: onEditAccountSearchFocus,
    addRoutingRule: addEditRoutingRule,
    removeRoutingRule: removeEditRoutingRule,
    hideAllDropdowns: hideEditAccountDropdowns,
    loadRulesFromApi: loadEditRoutingRulesFromApi,
    reset: resetEditRoutingRules
  } = useGroupRoutingRules('edit', () => editForm.platform)

  const deleteConfirmMessage = computed(() => {
    if (!deletingGroup.value) {
      return ''
    }
    if (deletingGroup.value.subscription_type === 'subscription') {
      return options.t('admin.groups.deleteConfirmSubscription', {
        name: deletingGroup.value.name
      })
    }
    return options.t('admin.groups.deleteConfirm', { name: deletingGroup.value.name })
  })

  const toggleCreateScope = (scope: string) => {
    toggleModelScope(createForm.supported_model_scopes, scope)
  }

  const toggleEditScope = (scope: string) => {
    toggleModelScope(editForm.supported_model_scopes, scope)
  }

  const closeCreateModal = () => {
    showCreateModal.value = false
    resetCreateRoutingRules()
    resetCreateGroupForm(createForm)
  }

  const closeEditModal = () => {
    resetEditRoutingRules()
    showEditModal.value = false
    editingGroup.value = null
    resetEditGroupForm(editForm)
  }

  const handleCreateGroup = async () => {
    if (!createForm.name.trim()) {
      options.showError(options.t('admin.groups.nameRequired'))
      return
    }

    submitting.value = true
    try {
      await adminAPI.groups.create(
        buildCreateGroupPayload(createForm, createModelRoutingRules.value)
      )
      options.showSuccess(options.t('admin.groups.groupCreated'))
      closeCreateModal()
      await options.loadGroups()
      if (options.isCurrentOnboardingStep('[data-tour="group-form-submit"]')) {
        options.advanceOnboarding(500)
      }
    } catch (error) {
      options.showError(
        resolveRequestErrorMessage(error, options.t('admin.groups.failedToCreate'))
      )
      console.error('Error creating group:', error)
    } finally {
      submitting.value = false
    }
  }

  const handleEdit = async (group: AdminGroup) => {
    editingGroup.value = group
    hydrateEditGroupForm(editForm, group)
    applyGroupFormSubscriptionTypeRules(editForm)
    applyGroupFormPlatformRules(editForm)
    await loadEditRoutingRulesFromApi(group.model_routing)
    showEditModal.value = true
  }

  const handleUpdateGroup = async () => {
    if (!editingGroup.value) {
      return
    }
    if (!editForm.name.trim()) {
      options.showError(options.t('admin.groups.nameRequired'))
      return
    }

    submitting.value = true
    try {
      await adminAPI.groups.update(
        editingGroup.value.id,
        buildUpdateGroupPayload(editForm, editModelRoutingRules.value)
      )
      options.showSuccess(options.t('admin.groups.groupUpdated'))
      closeEditModal()
      await options.loadGroups()
    } catch (error) {
      options.showError(
        resolveRequestErrorMessage(error, options.t('admin.groups.failedToUpdate'))
      )
      console.error('Error updating group:', error)
    } finally {
      submitting.value = false
    }
  }

  const handleRateMultipliers = (group: AdminGroup) => {
    rateMultipliersGroup.value = group
    showRateMultipliersModal.value = true
  }

  const handleDelete = (group: AdminGroup) => {
    deletingGroup.value = group
    showDeleteDialog.value = true
  }

  const confirmDelete = async () => {
    if (!deletingGroup.value) {
      return
    }

    try {
      await adminAPI.groups.delete(deletingGroup.value.id)
      options.showSuccess(options.t('admin.groups.groupDeleted'))
      showDeleteDialog.value = false
      deletingGroup.value = null
      await options.loadGroups()
    } catch (error) {
      options.showError(
        resolveRequestErrorMessage(error, options.t('admin.groups.failedToDelete'))
      )
      console.error('Error deleting group:', error)
    }
  }

  const handleClickOutside = (event: MouseEvent) => {
    const target = event.target as HTMLElement
    if (!target.closest('.account-search-container')) {
      hideCreateAccountDropdowns()
      hideEditAccountDropdowns()
    }
  }

  watch(
    () => createForm.subscription_type,
    () => {
      applyGroupFormSubscriptionTypeRules(createForm)
    }
  )

  watch(
    () => createForm.platform,
    () => {
      applyGroupFormPlatformRules(createForm)
    }
  )

  watch(
    () => editForm.subscription_type,
    () => {
      applyGroupFormSubscriptionTypeRules(editForm)
    }
  )

  watch(
    () => editForm.platform,
    () => {
      applyGroupFormPlatformRules(editForm)
    }
  )

  return {
    showCreateModal,
    showEditModal,
    showDeleteDialog,
    showRateMultipliersModal,
    submitting,
    editingGroup,
    deletingGroup,
    rateMultipliersGroup,
    createForm,
    editForm,
    createModelRoutingRules,
    createAccountSearchKeyword,
    createAccountSearchResults,
    createShowAccountDropdown,
    getCreateRuleRenderKey,
    getCreateRuleSearchKey,
    searchCreateAccountsByRule,
    selectCreateAccount,
    removeCreateSelectedAccount,
    onCreateAccountSearchFocus,
    addCreateRoutingRule,
    removeCreateRoutingRule,
    editModelRoutingRules,
    editAccountSearchKeyword,
    editAccountSearchResults,
    editShowAccountDropdown,
    getEditRuleRenderKey,
    getEditRuleSearchKey,
    searchEditAccountsByRule,
    selectEditAccount,
    removeEditSelectedAccount,
    onEditAccountSearchFocus,
    addEditRoutingRule,
    removeEditRoutingRule,
    deleteConfirmMessage,
    toggleCreateScope,
    toggleEditScope,
    closeCreateModal,
    handleCreateGroup,
    handleEdit,
    closeEditModal,
    handleUpdateGroup,
    handleRateMultipliers,
    handleDelete,
    confirmDelete,
    handleClickOutside,
    hideCreateAccountDropdowns,
    hideEditAccountDropdowns
  }
}
