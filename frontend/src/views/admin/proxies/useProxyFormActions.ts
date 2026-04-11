import type { Ref } from 'vue'
import { adminAPI } from '@/api/admin'
import type { Proxy } from '@/types'
import { resolveRequestErrorMessage } from '@/utils/requestError'
import {
  buildCreateProxyRequest,
  buildUpdateProxyRequest,
  getProxyFormValidationError,
  hydrateProxyEditForm,
  type ProxyBatchParseState,
  type ProxyCreateForm,
  type ProxyEditForm,
  resetProxyBatchParseState,
  resetProxyCreateForm
} from './proxyForm'
import { parseBatchProxyInput } from './proxyUtils'

interface ProxyFormActionsOptions {
  showCreateModal: Ref<boolean>
  createMode: Ref<'standard' | 'batch'>
  createForm: ProxyCreateForm
  createPasswordVisible: Ref<boolean>
  batchInput: Ref<string>
  batchParseResult: ProxyBatchParseState
  showImportData: Ref<boolean>
  editingProxy: Ref<Proxy | null>
  editForm: ProxyEditForm
  showEditModal: Ref<boolean>
  editPasswordVisible: Ref<boolean>
  editPasswordDirty: Ref<boolean>
  submitting: Ref<boolean>
  loadProxies: () => void | Promise<void>
  t: (key: string, params?: Record<string, unknown>) => string
  showSuccess: (message: string) => void
  showError: (message: string) => void
  showInfo: (message: string) => void
}

export function useProxyFormActions(options: ProxyFormActionsOptions) {
  const closeCreateModal = () => {
    options.showCreateModal.value = false
    options.createMode.value = 'standard'
    resetProxyCreateForm(options.createForm)
    options.createPasswordVisible.value = false
    options.batchInput.value = ''
    resetProxyBatchParseState(options.batchParseResult)
  }

  const handleDataImported = () => {
    options.showImportData.value = false
    options.loadProxies()
  }

  const parseBatchInput = () => {
    Object.assign(options.batchParseResult, parseBatchProxyInput(options.batchInput.value))
  }

  const handleBatchCreate = async () => {
    if (options.batchParseResult.valid === 0) {
      return
    }

    options.submitting.value = true
    try {
      const result = await adminAPI.proxies.batchCreate(options.batchParseResult.proxies)
      const created = result.created || 0
      const skipped = result.skipped || 0

      if (created > 0) {
        options.showSuccess(options.t('admin.proxies.batchImportSuccess', { created, skipped }))
      } else {
        options.showInfo(options.t('admin.proxies.batchImportAllSkipped', { skipped }))
      }

      closeCreateModal()
      await options.loadProxies()
    } catch (error: unknown) {
      options.showError(
        resolveRequestErrorMessage(error, options.t('admin.proxies.failedToImport'))
      )
      console.error('Error batch creating proxies:', error)
    } finally {
      options.submitting.value = false
    }
  }

  const handleCreateProxy = async () => {
    const validationError = getProxyFormValidationError(options.createForm)
    if (validationError) {
      options.showError(options.t(validationError))
      return
    }

    options.submitting.value = true
    try {
      await adminAPI.proxies.create(buildCreateProxyRequest(options.createForm))
      options.showSuccess(options.t('admin.proxies.proxyCreated'))
      closeCreateModal()
      await options.loadProxies()
    } catch (error: unknown) {
      options.showError(
        resolveRequestErrorMessage(error, options.t('admin.proxies.failedToCreate'))
      )
      console.error('Error creating proxy:', error)
    } finally {
      options.submitting.value = false
    }
  }

  const handleEdit = (proxy: Proxy) => {
    options.editingProxy.value = proxy
    hydrateProxyEditForm(options.editForm, proxy)
    options.editPasswordVisible.value = false
    options.editPasswordDirty.value = false
    options.showEditModal.value = true
  }

  const closeEditModal = () => {
    options.showEditModal.value = false
    options.editingProxy.value = null
    options.editPasswordVisible.value = false
    options.editPasswordDirty.value = false
  }

  const handleUpdateProxy = async () => {
    if (!options.editingProxy.value) {
      return
    }

    const validationError = getProxyFormValidationError(options.editForm)
    if (validationError) {
      options.showError(options.t(validationError))
      return
    }

    options.submitting.value = true
    try {
      await adminAPI.proxies.update(
        options.editingProxy.value.id,
        buildUpdateProxyRequest(options.editForm, options.editPasswordDirty.value)
      )
      options.showSuccess(options.t('admin.proxies.proxyUpdated'))
      closeEditModal()
      await options.loadProxies()
    } catch (error: unknown) {
      options.showError(
        resolveRequestErrorMessage(error, options.t('admin.proxies.failedToUpdate'))
      )
      console.error('Error updating proxy:', error)
    } finally {
      options.submitting.value = false
    }
  }

  return {
    closeCreateModal,
    closeEditModal,
    handleBatchCreate,
    handleCreateProxy,
    handleDataImported,
    handleEdit,
    handleUpdateProxy,
    parseBatchInput
  }
}
