import { ref } from 'vue'
import { adminAPI } from '@/api'
import { resolveRequestErrorMessage } from '@/utils/requestError'
import {
  createDefaultBackupS3Config,
  createDefaultBackupScheduleConfig
} from './backupView'

interface BackupViewConfigOptions {
  t: (key: string, params?: Record<string, unknown>) => string
  showError: (message: string) => void
  showSuccess: (message: string) => void
}

export function useBackupViewConfig(options: BackupViewConfigOptions) {
  const s3Form = ref(createDefaultBackupS3Config())
  const s3SecretConfigured = ref(false)
  const savingS3 = ref(false)
  const testingS3 = ref(false)

  const scheduleForm = ref(createDefaultBackupScheduleConfig())
  const savingSchedule = ref(false)

  let s3RequestSequence = 0
  let scheduleRequestSequence = 0

  const loadS3Config = async (requestSequence = ++s3RequestSequence) => {
    try {
      const config = await adminAPI.backup.getS3Config()
      if (requestSequence !== s3RequestSequence) {
        return false
      }

      s3Form.value = {
        endpoint: config.endpoint || '',
        region: config.region || 'auto',
        bucket: config.bucket || '',
        access_key_id: config.access_key_id || '',
        secret_access_key: '',
        prefix: config.prefix || 'backups/',
        force_path_style: config.force_path_style
      }
      s3SecretConfigured.value = Boolean(config.access_key_id)
      return true
    } catch (error) {
      if (requestSequence !== s3RequestSequence) {
        return false
      }

      options.showError(resolveRequestErrorMessage(error, options.t('errors.networkError')))
      return false
    }
  }

  const saveS3Config = async () => {
    const requestSequence = ++s3RequestSequence
    savingS3.value = true
    try {
      await adminAPI.backup.updateS3Config(s3Form.value)
      if (requestSequence !== s3RequestSequence) {
        return
      }

      options.showSuccess(options.t('admin.backup.s3.saved'))
      await loadS3Config(requestSequence)
    } catch (error) {
      if (requestSequence !== s3RequestSequence) {
        return
      }

      options.showError(resolveRequestErrorMessage(error, options.t('errors.networkError')))
    } finally {
      if (requestSequence === s3RequestSequence) {
        savingS3.value = false
      }
    }
  }

  const testS3 = async () => {
    testingS3.value = true
    try {
      const result = await adminAPI.backup.testS3Connection(s3Form.value)
      if (result.ok) {
        options.showSuccess(result.message || options.t('admin.backup.s3.testSuccess'))
      } else {
        options.showError(result.message || options.t('admin.backup.s3.testFailed'))
      }
    } catch (error) {
      options.showError(resolveRequestErrorMessage(error, options.t('errors.networkError')))
    } finally {
      testingS3.value = false
    }
  }

  const loadSchedule = async (requestSequence = ++scheduleRequestSequence) => {
    try {
      const config = await adminAPI.backup.getSchedule()
      if (requestSequence !== scheduleRequestSequence) {
        return false
      }

      scheduleForm.value = {
        enabled: config.enabled,
        cron_expr: config.cron_expr || '0 2 * * *',
        retain_days: config.retain_days || 14,
        retain_count: config.retain_count || 10
      }
      return true
    } catch (error) {
      if (requestSequence !== scheduleRequestSequence) {
        return false
      }

      options.showError(resolveRequestErrorMessage(error, options.t('errors.networkError')))
      return false
    }
  }

  const saveSchedule = async () => {
    const requestSequence = ++scheduleRequestSequence
    savingSchedule.value = true
    try {
      await adminAPI.backup.updateSchedule(scheduleForm.value)
      if (requestSequence !== scheduleRequestSequence) {
        return
      }

      options.showSuccess(options.t('admin.backup.schedule.saved'))
    } catch (error) {
      if (requestSequence !== scheduleRequestSequence) {
        return
      }

      options.showError(resolveRequestErrorMessage(error, options.t('errors.networkError')))
    } finally {
      if (requestSequence === scheduleRequestSequence) {
        savingSchedule.value = false
      }
    }
  }

  const dispose = () => {
    s3RequestSequence += 1
    scheduleRequestSequence += 1
    savingS3.value = false
    savingSchedule.value = false
    testingS3.value = false
  }

  return {
    s3Form,
    s3SecretConfigured,
    savingS3,
    testingS3,
    scheduleForm,
    savingSchedule,
    loadS3Config,
    saveS3Config,
    testS3,
    loadSchedule,
    saveSchedule,
    dispose
  }
}
