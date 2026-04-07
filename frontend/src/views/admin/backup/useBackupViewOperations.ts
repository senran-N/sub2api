import { ref } from 'vue'
import { adminAPI } from '@/api'
import type { BackupRecord } from '@/api/admin/backup'
import { hasResponseStatus, resolveRequestErrorMessage } from '@/utils/requestError'
import {
  BACKUP_MAX_POLL_COUNT,
  findRestoringBackup,
  findRunningBackup
} from './backupView'

interface BackupViewOperationsOptions {
  t: (key: string, params?: Record<string, unknown>) => string
  showSuccess: (message: string) => void
  showError: (message: string) => void
  showWarning: (message: string) => void
  confirm: (message: string) => boolean
  prompt: (message: string) => string | null
  openUrl: (url: string) => void
}

export function useBackupViewOperations(options: BackupViewOperationsOptions) {
  const backups = ref<BackupRecord[]>([])
  const loadingBackups = ref(false)
  const creatingBackup = ref(false)
  const restoringId = ref('')
  const manualExpireDays = ref(14)

  const pollingTimer = ref<ReturnType<typeof setInterval> | null>(null)
  const restoringPollingTimer = ref<ReturnType<typeof setInterval> | null>(null)

  const updateRecordInList = (updated: BackupRecord) => {
    const index = backups.value.findIndex((record) => record.id === updated.id)
    if (index >= 0) {
      backups.value[index] = updated
    }
  }

  const stopPolling = () => {
    if (pollingTimer.value) {
      clearInterval(pollingTimer.value)
      pollingTimer.value = null
    }
  }

  const stopRestorePolling = () => {
    if (restoringPollingTimer.value) {
      clearInterval(restoringPollingTimer.value)
      restoringPollingTimer.value = null
    }
  }

  const loadBackups = async () => {
    loadingBackups.value = true
    try {
      const result = await adminAPI.backup.listBackups()
      backups.value = result.items || []
    } catch (error) {
      options.showError(resolveRequestErrorMessage(error, options.t('errors.networkError')))
    } finally {
      loadingBackups.value = false
    }
  }

  const startPolling = (backupId: string) => {
    stopPolling()
    let count = 0
    pollingTimer.value = setInterval(async () => {
      if (count++ >= BACKUP_MAX_POLL_COUNT) {
        stopPolling()
        creatingBackup.value = false
        options.showWarning(options.t('admin.backup.operations.backupRunning'))
        return
      }
      try {
        const record = await adminAPI.backup.getBackup(backupId)
        updateRecordInList(record)
        if (record.status === 'completed' || record.status === 'failed') {
          stopPolling()
          creatingBackup.value = false
          if (record.status === 'completed') {
            options.showSuccess(options.t('admin.backup.operations.backupCreated'))
          } else {
            options.showError(record.error_message || options.t('admin.backup.operations.backupFailed'))
          }
          await loadBackups()
        }
      } catch {
        // Polling failures should not stop the running task.
      }
    }, 2000)
  }

  const startRestorePolling = (backupId: string) => {
    stopRestorePolling()
    let count = 0
    restoringPollingTimer.value = setInterval(async () => {
      if (count++ >= BACKUP_MAX_POLL_COUNT) {
        stopRestorePolling()
        restoringId.value = ''
        options.showWarning(options.t('admin.backup.operations.restoreRunning'))
        return
      }
      try {
        const record = await adminAPI.backup.getBackup(backupId)
        updateRecordInList(record)
        if (record.restore_status === 'completed' || record.restore_status === 'failed') {
          stopRestorePolling()
          restoringId.value = ''
          if (record.restore_status === 'completed') {
            options.showSuccess(options.t('admin.backup.actions.restoreSuccess'))
          } else {
            options.showError(record.restore_error || options.t('admin.backup.operations.restoreFailed'))
          }
          await loadBackups()
        }
      } catch {
        // Polling failures should not stop the running task.
      }
    }, 2000)
  }

  const resumeActiveOperations = () => {
    const runningBackup = findRunningBackup(backups.value)
    if (runningBackup) {
      creatingBackup.value = true
      startPolling(runningBackup.id)
    }

    const restoringBackup = findRestoringBackup(backups.value)
    if (restoringBackup) {
      restoringId.value = restoringBackup.id
      startRestorePolling(restoringBackup.id)
    }
  }

  const handleVisibilityChange = () => {
    if (document.hidden) {
      stopPolling()
      stopRestorePolling()
      return
    }

    void loadBackups().then(() => {
      resumeActiveOperations()
    })
  }

  const createBackup = async () => {
    creatingBackup.value = true
    try {
      const record = await adminAPI.backup.createBackup({ expire_days: manualExpireDays.value })
      backups.value.unshift(record)
      startPolling(record.id)
    } catch (error: any) {
      if (hasResponseStatus(error, 409)) {
        options.showWarning(options.t('admin.backup.operations.alreadyInProgress'))
      } else {
        options.showError(resolveRequestErrorMessage(error, options.t('errors.networkError')))
      }
      creatingBackup.value = false
    }
  }

  const downloadBackup = async (id: string) => {
    try {
      const result = await adminAPI.backup.getDownloadURL(id)
      options.openUrl(result.url)
    } catch (error) {
      options.showError(resolveRequestErrorMessage(error, options.t('errors.networkError')))
    }
  }

  const restoreBackup = async (id: string) => {
    if (!options.confirm(options.t('admin.backup.actions.restoreConfirm'))) {
      return
    }
    const password = options.prompt(options.t('admin.backup.actions.restorePasswordPrompt'))
    if (!password) {
      return
    }

    restoringId.value = id
    try {
      const record = await adminAPI.backup.restoreBackup(id, password)
      updateRecordInList(record)
      startRestorePolling(id)
    } catch (error) {
      if (hasResponseStatus(error, 409)) {
        options.showWarning(options.t('admin.backup.operations.restoreRunning'))
      } else {
        options.showError(resolveRequestErrorMessage(error, options.t('errors.networkError')))
      }
      restoringId.value = ''
    }
  }

  const removeBackup = async (id: string) => {
    if (!options.confirm(options.t('admin.backup.actions.deleteConfirm'))) {
      return
    }

    try {
      await adminAPI.backup.deleteBackup(id)
      options.showSuccess(options.t('admin.backup.actions.deleted'))
      await loadBackups()
    } catch (error) {
      options.showError(resolveRequestErrorMessage(error, options.t('errors.networkError')))
    }
  }

  const initialize = async () => {
    document.addEventListener('visibilitychange', handleVisibilityChange)
    await loadBackups()
    resumeActiveOperations()
  }

  const dispose = () => {
    stopPolling()
    stopRestorePolling()
    document.removeEventListener('visibilitychange', handleVisibilityChange)
  }

  return {
    backups,
    loadingBackups,
    creatingBackup,
    restoringId,
    manualExpireDays,
    loadBackups,
    createBackup,
    downloadBackup,
    restoreBackup,
    removeBackup,
    initialize,
    dispose
  }
}
