import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { BackupRecord } from '@/api/admin/backup'
import { useBackupViewOperations } from '../backup/useBackupViewOperations'

const {
  listBackups,
  createBackupRequest,
  getBackup,
  deleteBackup,
  getDownloadURL,
  restoreBackupRequest
} = vi.hoisted(() => ({
  listBackups: vi.fn(),
  createBackupRequest: vi.fn(),
  getBackup: vi.fn(),
  deleteBackup: vi.fn(),
  getDownloadURL: vi.fn(),
  restoreBackupRequest: vi.fn()
}))

vi.mock('@/api', () => ({
  adminAPI: {
    backup: {
      listBackups,
      createBackup: createBackupRequest,
      getBackup,
      deleteBackup,
      getDownloadURL,
      restoreBackup: restoreBackupRequest
    }
  }
}))

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

describe('useBackupViewOperations', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    listBackups.mockReset()
    createBackupRequest.mockReset()
    getBackup.mockReset()
    deleteBackup.mockReset()
    getDownloadURL.mockReset()
    restoreBackupRequest.mockReset()

    listBackups.mockResolvedValue({
      items: [createRecord()]
    })
    createBackupRequest.mockResolvedValue(createRecord({ id: 'b_2', status: 'running' }))
    getBackup.mockResolvedValue(createRecord({ id: 'b_2', status: 'completed' }))
    deleteBackup.mockResolvedValue(undefined)
    getDownloadURL.mockResolvedValue({ url: 'https://example.com/backup' })
    restoreBackupRequest.mockResolvedValue(
      createRecord({ id: 'b_1', restore_status: 'running' })
    )
  })

  it('loads backups, creates backups, and polls them to completion', async () => {
    const showSuccess = vi.fn()
    const showError = vi.fn()
    const showWarning = vi.fn()
    const operations = useBackupViewOperations({
      t: (key: string) => key,
      showSuccess,
      showError,
      showWarning,
      confirm: vi.fn(() => true),
      prompt: vi.fn(() => 'secret'),
      openUrl: vi.fn()
    })

    await operations.initialize()
    expect(listBackups).toHaveBeenCalledTimes(1)
    expect(operations.backups.value).toHaveLength(1)

    await operations.createBackup()
    expect(createBackupRequest).toHaveBeenCalledWith({ expire_days: 14 })
    await vi.advanceTimersByTimeAsync(2000)
    expect(getBackup).toHaveBeenCalledWith('b_2')
    expect(showSuccess).toHaveBeenCalledWith('admin.backup.operations.backupCreated')
    expect(showError).not.toHaveBeenCalled()
    expect(showWarning).not.toHaveBeenCalled()
  })

  it('downloads, restores, and deletes backups', async () => {
    const confirm = vi.fn(() => true)
    const prompt = vi.fn(() => 'secret')
    const openUrl = vi.fn()
    const showSuccess = vi.fn()
    const operations = useBackupViewOperations({
      t: (key: string) => key,
      showSuccess,
      showError: vi.fn(),
      showWarning: vi.fn(),
      confirm,
      prompt,
      openUrl
    })

    await operations.initialize()
    await operations.downloadBackup('b_1')
    expect(openUrl).toHaveBeenCalledWith('https://example.com/backup')

    await operations.restoreBackup('b_1')
    expect(confirm).toHaveBeenCalledWith('admin.backup.actions.restoreConfirm')
    expect(prompt).toHaveBeenCalledWith('admin.backup.actions.restorePasswordPrompt')
    expect(restoreBackupRequest).toHaveBeenCalledWith('b_1', 'secret')

    await operations.removeBackup('b_1')
    expect(deleteBackup).toHaveBeenCalledWith('b_1')
    expect(showSuccess).toHaveBeenCalledWith('admin.backup.actions.deleted')
  })

  it('routes conflict and detail errors through shared request helpers', async () => {
    const showError = vi.fn()
    const showWarning = vi.fn()
    const operations = useBackupViewOperations({
      t: (key: string) => key,
      showSuccess: vi.fn(),
      showError,
      showWarning,
      confirm: vi.fn(() => true),
      prompt: vi.fn(() => 'secret'),
      openUrl: vi.fn()
    })

    createBackupRequest.mockRejectedValueOnce({ response: { status: 409 } })
    await operations.createBackup()
    expect(showWarning).toHaveBeenCalledWith('admin.backup.operations.alreadyInProgress')

    getDownloadURL.mockRejectedValueOnce({
      response: { data: { detail: 'download-failed' } }
    })
    await operations.downloadBackup('b_1')
    expect(showError).toHaveBeenCalledWith('download-failed')
  })
})
