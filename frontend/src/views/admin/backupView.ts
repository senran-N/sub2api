import type { BackupRecord, BackupS3Config, BackupScheduleConfig } from '@/api/admin/backup'

export const BACKUP_MAX_POLL_COUNT = 900

export function createDefaultBackupS3Config(): BackupS3Config {
  return {
    endpoint: '',
    region: 'auto',
    bucket: '',
    access_key_id: '',
    secret_access_key: '',
    prefix: 'backups/',
    force_path_style: false
  }
}

export function createDefaultBackupScheduleConfig(): BackupScheduleConfig {
  return {
    enabled: false,
    cron_expr: '0 2 * * *',
    retain_days: 14,
    retain_count: 10
  }
}

export function getBackupStatusClass(status: string): string {
  switch (status) {
    case 'completed':
      return 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300'
    case 'running':
      return 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300'
    case 'failed':
      return 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300'
    default:
      return 'bg-gray-100 text-gray-700 dark:bg-dark-800 dark:text-gray-300'
  }
}

export function formatBackupSize(bytes: number): string {
  if (!bytes || bytes <= 0) {
    return '-'
  }
  if (bytes < 1024) {
    return `${bytes} B`
  }
  if (bytes < 1024 * 1024) {
    return `${(bytes / 1024).toFixed(1)} KB`
  }
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

export function formatBackupDate(value?: string): string {
  if (!value) {
    return '-'
  }
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return value
  }
  return date.toLocaleString()
}

export function buildBackupR2ConfigRows(
  t: (key: string, params?: Record<string, unknown>) => string
) {
  return [
    { field: t('admin.backup.s3.endpoint'), value: 'https://<account_id>.r2.cloudflarestorage.com' },
    { field: t('admin.backup.s3.region'), value: 'auto' },
    { field: t('admin.backup.s3.bucket'), value: t('admin.backup.r2Guide.step4.bucketValue') },
    { field: t('admin.backup.s3.prefix'), value: 'backups/' },
    { field: 'Access Key ID', value: t('admin.backup.r2Guide.step4.fromStep2') },
    { field: 'Secret Access Key', value: t('admin.backup.r2Guide.step4.fromStep2') },
    { field: t('admin.backup.s3.forcePathStyle'), value: t('admin.backup.r2Guide.step4.unchecked') }
  ]
}

export function findRunningBackup(records: BackupRecord[]): BackupRecord | undefined {
  return records.find((record) => record.status === 'running')
}

export function findRestoringBackup(records: BackupRecord[]): BackupRecord | undefined {
  return records.find((record) => record.restore_status === 'running')
}
