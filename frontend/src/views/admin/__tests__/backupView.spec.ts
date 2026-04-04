import { describe, expect, it } from 'vitest'
import type { BackupRecord } from '@/api/admin/backup'
import {
  buildBackupR2ConfigRows,
  createDefaultBackupS3Config,
  createDefaultBackupScheduleConfig,
  findRestoringBackup,
  findRunningBackup,
  formatBackupDate,
  formatBackupSize,
  getBackupStatusClass
} from '../backupView'

function createRecord(overrides: Partial<BackupRecord> = {}): BackupRecord {
  return {
    id: 'b_1',
    status: 'completed',
    backup_type: 'full',
    file_name: 'backup.tar.gz',
    s3_key: 'backups/backup.tar.gz',
    size_bytes: 1024,
    triggered_by: 'manual',
    started_at: '2026-04-04T00:00:00Z',
    ...overrides
  }
}

describe('backupView helpers', () => {
  it('creates default config forms', () => {
    expect(createDefaultBackupS3Config()).toEqual({
      endpoint: '',
      region: 'auto',
      bucket: '',
      access_key_id: '',
      secret_access_key: '',
      prefix: 'backups/',
      force_path_style: false
    })
    expect(createDefaultBackupScheduleConfig()).toEqual({
      enabled: false,
      cron_expr: '0 2 * * *',
      retain_days: 14,
      retain_count: 10
    })
  })

  it('formats backup status, size, date, and R2 guide rows', () => {
    expect(getBackupStatusClass('completed')).toContain('bg-green-100')
    expect(getBackupStatusClass('running')).toContain('bg-blue-100')
    expect(getBackupStatusClass('failed')).toContain('bg-red-100')
    expect(getBackupStatusClass('pending')).toContain('bg-gray-100')

    expect(formatBackupSize(0)).toBe('-')
    expect(formatBackupSize(100)).toBe('100 B')
    expect(formatBackupSize(2048)).toBe('2.0 KB')
    expect(formatBackupSize(2 * 1024 * 1024)).toBe('2.0 MB')

    expect(formatBackupDate()).toBe('-')
    expect(formatBackupDate('invalid')).toBe('invalid')
    expect(buildBackupR2ConfigRows((key: string) => key)).toHaveLength(7)
  })

  it('finds running and restoring records', () => {
    const records = [
      createRecord({ id: 'a', status: 'completed' }),
      createRecord({ id: 'b', status: 'running' }),
      createRecord({ id: 'c', restore_status: 'running' })
    ]

    expect(findRunningBackup(records)?.id).toBe('b')
    expect(findRestoringBackup(records)?.id).toBe('c')
  })
})
