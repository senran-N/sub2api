import type { Ref } from 'vue'
import { adminAPI } from '@/api/admin'
import type { Account, ClaudeModel, SelectOption } from '@/types'
import { resolveRequestErrorMessage } from '@/utils/requestError'
import { normalizeBulkSchedulableResult } from './accountsList'

interface AccountsViewActionsOptions {
  showEdit: Ref<boolean>
  showTempUnsched: Ref<boolean>
  showDeleteDialog: Ref<boolean>
  showReAuth: Ref<boolean>
  showTest: Ref<boolean>
  showStats: Ref<boolean>
  showSchedulePanel: Ref<boolean>
  edAcc: Ref<Account | null>
  tempUnschedAcc: Ref<Account | null>
  deletingAcc: Ref<Account | null>
  reAuthAcc: Ref<Account | null>
  testingAcc: Ref<Account | null>
  statsAcc: Ref<Account | null>
  scheduleAcc: Ref<Account | null>
  scheduleModelOptions: Ref<SelectOption[]>
  togglingSchedulable: Ref<number | null>
  getSelectedIds: () => number[]
  confirmAction: () => boolean
  clearSelection: () => void
  setSelectedIds: (ids: number[]) => void
  load: () => void | Promise<void>
  reload: () => void | Promise<void>
  patchAccountInList: (updatedAccount: Account) => void
  updateSchedulableInList: (accountIds: number[], schedulable: boolean) => void
  enterAutoRefreshSilentWindow: () => void
  t: (key: string, params?: Record<string, unknown>) => string
  showSuccess: (message: string) => void
  showError: (message: string) => void
}

export function useAccountsViewActions(options: AccountsViewActionsOptions) {
  const handleEdit = (account: Account) => {
    options.edAcc.value = account
    options.showEdit.value = true
  }

  const handleBulkDelete = async () => {
    if (!options.confirmAction()) {
      return
    }

    try {
      await Promise.all(options.getSelectedIds().map((id) => adminAPI.accounts.delete(id)))
      options.clearSelection()
      await options.reload()
    } catch (error) {
      console.error('Failed to bulk delete accounts:', error)
    }
  }

  const handleBulkResetStatus = async () => {
    if (!options.confirmAction()) {
      return
    }

    try {
      const result = await adminAPI.accounts.batchClearError(options.getSelectedIds())
      if (result.failed > 0) {
        options.showError(
          options.t('admin.accounts.bulkActions.partialSuccess', {
            success: result.success,
            failed: result.failed
          })
        )
      } else {
        options.showSuccess(
          options.t('admin.accounts.bulkActions.resetStatusSuccess', { count: result.success })
        )
        options.clearSelection()
      }

      await options.reload()
    } catch (error) {
      console.error('Failed to bulk reset status:', error)
      options.showError(
        resolveRequestErrorMessage(error, options.t('common.error'))
      )
    }
  }

  const handleBulkRefreshToken = async () => {
    if (!options.confirmAction()) {
      return
    }

    try {
      const selectedIds = options.getSelectedIds()
      const result = await adminAPI.accounts.batchRefresh(selectedIds)
      if (result.failed > 0) {
        const failedIDs = Array.isArray(result.errors)
          ? result.errors
              .map((entry) => entry.account_id)
              .filter((accountID) => Number.isInteger(accountID) && accountID > 0)
          : []
        if (failedIDs.length > 0) {
          options.setSelectedIds(failedIDs)
        } else {
          options.setSelectedIds(selectedIds)
        }
        options.showError(
          options.t('admin.accounts.bulkActions.partialSuccess', {
            success: result.success,
            failed: result.failed
          })
        )
      } else {
        options.showSuccess(
          options.t('admin.accounts.bulkActions.refreshTokenSuccess', { count: result.success })
        )
        options.clearSelection()
      }

      await options.reload()
    } catch (error) {
      console.error('Failed to bulk refresh token:', error)
      options.showError(
        resolveRequestErrorMessage(error, options.t('common.error'))
      )
    }
  }

  const handleBulkToggleSchedulable = async (schedulable: boolean) => {
    const accountIds = [...options.getSelectedIds()]

    try {
      const result = await adminAPI.accounts.bulkUpdate(accountIds, { schedulable })
      const { successIds, failedIds, successCount, failedCount, hasIds, hasCounts } =
        normalizeBulkSchedulableResult(result, accountIds)

      if (!hasIds && !hasCounts) {
        options.showError(options.t('admin.accounts.bulkSchedulableResultUnknown'))
        options.setSelectedIds(accountIds)
        await options.load()
        return
      }

      if (successIds.length > 0) {
        options.updateSchedulableInList(successIds, schedulable)
      }

      if (successCount > 0 && failedCount === 0) {
        options.showSuccess(
          schedulable
            ? options.t('admin.accounts.bulkSchedulableEnabled', { count: successCount })
            : options.t('admin.accounts.bulkSchedulableDisabled', { count: successCount })
        )
      }

      if (failedCount > 0) {
        options.showError(
          hasCounts || hasIds
            ? options.t('admin.accounts.bulkSchedulablePartial', {
                success: successCount,
                failed: failedCount
              })
            : options.t('admin.accounts.bulkSchedulableResultUnknown')
        )
        options.setSelectedIds(failedIds.length > 0 ? failedIds : accountIds)
        return
      }

      if (hasIds) {
        options.clearSelection()
      } else {
        options.setSelectedIds(accountIds)
      }
    } catch (error) {
      console.error('Failed to bulk toggle schedulable:', error)
      options.showError(options.t('common.error'))
    }
  }

  const handleAccountUpdated = (updatedAccount: Account) => {
    options.patchAccountInList(updatedAccount)
    options.enterAutoRefreshSilentWindow()
  }

  const closeTestModal = () => {
    options.showTest.value = false
    options.testingAcc.value = null
  }

  const closeStatsModal = () => {
    options.showStats.value = false
    options.statsAcc.value = null
  }

  const closeReAuthModal = () => {
    options.showReAuth.value = false
    options.reAuthAcc.value = null
  }

  const handleTest = (account: Account) => {
    options.testingAcc.value = account
    options.showTest.value = true
  }

  const handleViewStats = (account: Account) => {
    options.statsAcc.value = account
    options.showStats.value = true
  }

  const handleSchedule = async (account: Account) => {
    options.scheduleAcc.value = account
    options.scheduleModelOptions.value = []
    options.showSchedulePanel.value = true

    try {
      const models = await adminAPI.accounts.getAvailableModels(account.id)
      options.scheduleModelOptions.value = models.map((model: ClaudeModel) => ({
        value: model.id,
        label: model.display_name || model.id
      }))
    } catch {
      options.scheduleModelOptions.value = []
    }
  }

  const closeSchedulePanel = () => {
    options.showSchedulePanel.value = false
    options.scheduleAcc.value = null
    options.scheduleModelOptions.value = []
  }

  const handleReAuth = (account: Account) => {
    options.reAuthAcc.value = account
    options.showReAuth.value = true
  }

  const runSingleAccountUpdate = (
    request: Promise<Account>,
    optionsForMessage?: {
      successMessage?: string
      errorMessage?: string
      mapErrorMessage?: (error: unknown) => string
    }
  ) =>
    request
      .then((updatedAccount) => {
        options.patchAccountInList(updatedAccount)
        options.enterAutoRefreshSilentWindow()

        if (optionsForMessage?.successMessage) {
          options.showSuccess(optionsForMessage.successMessage)
        }
      })
      .catch((error: any) => {
        const errorMessage = optionsForMessage?.mapErrorMessage
          ? optionsForMessage.mapErrorMessage(error)
          : optionsForMessage?.errorMessage

        console.error('Failed to update account:', error)
        if (errorMessage) {
          options.showError(errorMessage)
        }
      })

  const handleRefresh = async (account: Account) => {
    try {
      const updated = await adminAPI.accounts.refreshCredentials(account.id)
      options.patchAccountInList(updated)
      options.enterAutoRefreshSilentWindow()
    } catch (error) {
      console.error('Failed to refresh credentials:', error)
    }
  }

  const handleRecoverState = async (account: Account) => {
    await runSingleAccountUpdate(adminAPI.accounts.recoverState(account.id), {
      successMessage: options.t('admin.accounts.recoverStateSuccess'),
      errorMessage: options.t('admin.accounts.recoverStateFailed'),
      mapErrorMessage: (error: unknown) =>
        resolveRequestErrorMessage(error, options.t('admin.accounts.recoverStateFailed'))
    })
  }

  const handleResetQuota = async (account: Account) => {
    try {
      const updated = await adminAPI.accounts.resetAccountQuota(account.id)
      options.patchAccountInList(updated)
      options.enterAutoRefreshSilentWindow()
      options.showSuccess(options.t('common.success'))
    } catch (error) {
      console.error('Failed to reset quota:', error)
    }
  }

  const handleSetPrivacy = async (account: Account) => {
    await runSingleAccountUpdate(adminAPI.accounts.setPrivacy(account.id), {
      successMessage: options.t('common.success'),
      mapErrorMessage: (error: any) =>
        error?.response?.data?.message || options.t('admin.accounts.privacyFailed')
    })
  }

  const handleDelete = (account: Account) => {
    options.deletingAcc.value = account
    options.showDeleteDialog.value = true
  }

  const confirmDelete = async () => {
    if (!options.deletingAcc.value) {
      return
    }

    try {
      await adminAPI.accounts.delete(options.deletingAcc.value.id)
      options.showDeleteDialog.value = false
      options.deletingAcc.value = null
      await options.reload()
    } catch (error) {
      console.error('Failed to delete account:', error)
    }
  }

  const handleToggleSchedulable = async (account: Account) => {
    const nextSchedulable = !account.schedulable
    options.togglingSchedulable.value = account.id

    try {
      const updated = await adminAPI.accounts.setSchedulable(account.id, nextSchedulable)
      options.updateSchedulableInList([account.id], updated?.schedulable ?? nextSchedulable)
      options.enterAutoRefreshSilentWindow()
    } catch (error) {
      console.error('Failed to toggle schedulable:', error)
      options.showError(options.t('admin.accounts.failedToToggleSchedulable'))
    } finally {
      options.togglingSchedulable.value = null
    }
  }

  const handleShowTempUnsched = (account: Account) => {
    options.tempUnschedAcc.value = account
    options.showTempUnsched.value = true
  }

  const handleTempUnschedReset = async (updatedAccount: Account) => {
    options.showTempUnsched.value = false
    options.tempUnschedAcc.value = null
    options.patchAccountInList(updatedAccount)
    options.enterAutoRefreshSilentWindow()
  }

  return {
    showEdit: options.showEdit,
    showTempUnsched: options.showTempUnsched,
    showDeleteDialog: options.showDeleteDialog,
    showReAuth: options.showReAuth,
    showTest: options.showTest,
    showStats: options.showStats,
    showSchedulePanel: options.showSchedulePanel,
    edAcc: options.edAcc,
    tempUnschedAcc: options.tempUnschedAcc,
    deletingAcc: options.deletingAcc,
    reAuthAcc: options.reAuthAcc,
    testingAcc: options.testingAcc,
    statsAcc: options.statsAcc,
    scheduleAcc: options.scheduleAcc,
    scheduleModelOptions: options.scheduleModelOptions,
    togglingSchedulable: options.togglingSchedulable,
    handleEdit,
    handleBulkDelete,
    handleBulkResetStatus,
    handleBulkRefreshToken,
    handleBulkToggleSchedulable,
    handleAccountUpdated,
    closeTestModal,
    closeStatsModal,
    closeReAuthModal,
    handleTest,
    handleViewStats,
    handleSchedule,
    closeSchedulePanel,
    handleReAuth,
    handleRefresh,
    handleRecoverState,
    handleResetQuota,
    handleSetPrivacy,
    handleDelete,
    confirmDelete,
    handleToggleSchedulable,
    handleShowTempUnsched,
    handleTempUnschedReset
  }
}
