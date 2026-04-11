import { reactive, ref } from 'vue'
import { adminAPI } from '@/api/admin'
import type { PromoCode, PromoCodeUsage } from '@/types'
import { resolveRequestErrorMessage } from '@/utils/requestError'
import {
  buildCreatePromoCodeRequest,
  buildPromoRegisterLink,
  buildUpdatePromoCodeRequest,
  createDefaultPromoCodeCreateForm,
  createDefaultPromoCodeEditForm,
  hydratePromoCodeEditForm,
  resetPromoCodeCreateForm,
  resetPromoCodeEditForm
} from './promoCodeForm'

interface PromoCodesViewActionsOptions {
  origin: string
  t: (key: string, params?: Record<string, unknown>) => string
  showSuccess: (message: string) => void
  showError: (message: string) => void
  copyToClipboard: (text: string, successMessage?: string) => Promise<boolean>
  reloadCodes: () => Promise<void>
}

export function usePromoCodesViewActions(options: PromoCodesViewActionsOptions) {
  const creating = ref(false)
  const updating = ref(false)
  const showCreateDialog = ref(false)
  const showEditDialog = ref(false)
  const showDeleteDialog = ref(false)
  const showUsagesDialog = ref(false)

  const editingCode = ref<PromoCode | null>(null)
  const deletingCode = ref<PromoCode | null>(null)
  const currentViewingCode = ref<PromoCode | null>(null)

  const usages = ref<PromoCodeUsage[]>([])
  const usagesLoading = ref(false)
  const usagesPage = ref(1)
  const usagesPageSize = ref(20)
  const usagesTotal = ref(0)

  const createForm = reactive(createDefaultPromoCodeCreateForm())
  const editForm = reactive(createDefaultPromoCodeEditForm())

  const resetCreateForm = () => {
    resetPromoCodeCreateForm(createForm)
  }

  const closeCreateDialog = () => {
    showCreateDialog.value = false
    resetCreateForm()
  }

  const handleCreate = async () => {
    creating.value = true
    try {
      await adminAPI.promo.create(buildCreatePromoCodeRequest(createForm))
      options.showSuccess(options.t('admin.promo.codeCreated'))
      closeCreateDialog()
      await options.reloadCodes()
    } catch (error: unknown) {
      options.showError(
        resolveRequestErrorMessage(error, options.t('admin.promo.failedToCreate'))
      )
    } finally {
      creating.value = false
    }
  }

  const handleEdit = (code: PromoCode) => {
    editingCode.value = code
    hydratePromoCodeEditForm(editForm, code)
    showEditDialog.value = true
  }

  const closeEditDialog = () => {
    showEditDialog.value = false
    editingCode.value = null
    resetPromoCodeEditForm(editForm)
  }

  const handleUpdate = async () => {
    if (!editingCode.value) {
      return
    }

    updating.value = true
    try {
      await adminAPI.promo.update(editingCode.value.id, buildUpdatePromoCodeRequest(editForm))
      options.showSuccess(options.t('admin.promo.codeUpdated'))
      closeEditDialog()
      await options.reloadCodes()
    } catch (error: unknown) {
      options.showError(
        resolveRequestErrorMessage(error, options.t('admin.promo.failedToUpdate'))
      )
    } finally {
      updating.value = false
    }
  }

  const copyRegisterLink = async (code: PromoCode) => {
    const registerLink = buildPromoRegisterLink(options.origin, code.code)
    await options.copyToClipboard(registerLink, options.t('admin.promo.registerLinkCopied'))
  }

  const handleDelete = (code: PromoCode) => {
    deletingCode.value = code
    showDeleteDialog.value = true
  }

  const confirmDelete = async () => {
    if (!deletingCode.value) {
      return
    }

    try {
      await adminAPI.promo.delete(deletingCode.value.id)
      options.showSuccess(options.t('admin.promo.codeDeleted'))
      showDeleteDialog.value = false
      deletingCode.value = null
      await options.reloadCodes()
    } catch (error: unknown) {
      options.showError(
        resolveRequestErrorMessage(error, options.t('admin.promo.failedToDelete'))
      )
    }
  }

  const handleViewUsages = async (code: PromoCode) => {
    currentViewingCode.value = code
    showUsagesDialog.value = true
    usagesPage.value = 1
    await loadUsages()
  }

  const closeUsagesDialog = () => {
    showUsagesDialog.value = false
    currentViewingCode.value = null
    usages.value = []
    usagesTotal.value = 0
    usagesPage.value = 1
    usagesPageSize.value = 20
  }

  const loadUsages = async () => {
    if (!currentViewingCode.value) {
      return
    }

    usagesLoading.value = true
    usages.value = []

    try {
      const response = await adminAPI.promo.getUsages(
        currentViewingCode.value.id,
        usagesPage.value,
        usagesPageSize.value
      )
      usages.value = response.items
      usagesTotal.value = response.total
      usagesPage.value = response.page
      usagesPageSize.value = response.page_size
    } catch (error: unknown) {
      options.showError(
        resolveRequestErrorMessage(error, options.t('admin.promo.failedToLoadUsages'))
      )
    } finally {
      usagesLoading.value = false
    }
  }

  const handleUsagesPageChange = (page: number) => {
    usagesPage.value = page
    void loadUsages()
  }

  const handleUsagesPageSizeChange = (pageSize: number) => {
    usagesPageSize.value = pageSize
    usagesPage.value = 1
    void loadUsages()
  }

  return {
    creating,
    updating,
    showCreateDialog,
    showEditDialog,
    showDeleteDialog,
    showUsagesDialog,
    createForm,
    editForm,
    usages,
    usagesLoading,
    usagesPage,
    usagesPageSize,
    usagesTotal,
    handleCreate,
    closeCreateDialog,
    handleEdit,
    closeEditDialog,
    handleUpdate,
    copyRegisterLink,
    handleDelete,
    confirmDelete,
    handleViewUsages,
    closeUsagesDialog,
    loadUsages,
    handleUsagesPageChange,
    handleUsagesPageSizeChange
  }
}
