import { beforeEach, describe, expect, it, vi } from 'vitest'
import { useBackupViewConfig } from '../backup/useBackupViewConfig'

const { getS3Config, updateS3Config, testS3Connection, getSchedule, updateSchedule } = vi.hoisted(
  () => ({
    getS3Config: vi.fn(),
    updateS3Config: vi.fn(),
    testS3Connection: vi.fn(),
    getSchedule: vi.fn(),
    updateSchedule: vi.fn()
  })
)

vi.mock('@/api', () => ({
  adminAPI: {
    backup: {
      getS3Config,
      updateS3Config,
      testS3Connection,
      getSchedule,
      updateSchedule
    }
  }
}))

function createDeferred<T>() {
  let resolve!: (value: T | PromiseLike<T>) => void
  let reject!: (reason?: unknown) => void
  const promise = new Promise<T>((resolvePromise, rejectPromise) => {
    resolve = resolvePromise
    reject = rejectPromise
  })

  return {
    promise,
    resolve,
    reject
  }
}

describe('useBackupViewConfig', () => {
  beforeEach(() => {
    getS3Config.mockReset()
    updateS3Config.mockReset()
    testS3Connection.mockReset()
    getSchedule.mockReset()
    updateSchedule.mockReset()

    getS3Config.mockResolvedValue({
      endpoint: 'https://example.com',
      region: 'auto',
      bucket: 'backups',
      access_key_id: 'AK',
      secret_access_key: '',
      prefix: 'backups/',
      force_path_style: true
    })
    testS3Connection.mockResolvedValue({ ok: true, message: 'ok' })
    getSchedule.mockResolvedValue({
      enabled: true,
      cron_expr: '0 1 * * *',
      retain_days: 7,
      retain_count: 5
    })
    updateS3Config.mockResolvedValue({})
    updateSchedule.mockResolvedValue({})
  })

  it('loads, tests, and saves S3 config', async () => {
    const showError = vi.fn()
    const showSuccess = vi.fn()
    const config = useBackupViewConfig({
      t: (key: string) => key,
      showError,
      showSuccess
    })

    await config.loadS3Config()
    expect(config.s3Form.value.endpoint).toBe('https://example.com')
    expect(config.s3SecretConfigured.value).toBe(true)

    await config.testS3()
    expect(showSuccess).toHaveBeenCalledWith('ok')

    await config.saveS3Config()
    expect(updateS3Config).toHaveBeenCalledWith(config.s3Form.value)
    expect(showSuccess).toHaveBeenCalledWith('admin.backup.s3.saved')
    expect(showError).not.toHaveBeenCalled()
  })

  it('loads and saves schedule config', async () => {
    const showError = vi.fn()
    const showSuccess = vi.fn()
    const config = useBackupViewConfig({
      t: (key: string) => key,
      showError,
      showSuccess
    })

    await config.loadSchedule()
    expect(config.scheduleForm.value).toEqual({
      enabled: true,
      cron_expr: '0 1 * * *',
      retain_days: 7,
      retain_count: 5
    })

    await config.saveSchedule()
    expect(updateSchedule).toHaveBeenCalledWith(config.scheduleForm.value)
    expect(showSuccess).toHaveBeenCalledWith('admin.backup.schedule.saved')
  })

  it('uses shared request error details when S3 load fails', async () => {
    const showError = vi.fn()
    const config = useBackupViewConfig({
      t: (key: string) => key,
      showError,
      showSuccess: vi.fn()
    })
    getS3Config.mockRejectedValueOnce({
      response: { data: { detail: 'backup-config-failed' } }
    })

    await config.loadS3Config()

    expect(showError).toHaveBeenCalledWith('backup-config-failed')
  })

  it('keeps the latest S3 config request authoritative across save reloads', async () => {
    const firstLoad = createDeferred<{
      endpoint: string
      region: string
      bucket: string
      access_key_id: string
      secret_access_key: string
      prefix: string
      force_path_style: boolean
    }>()
    const reloadAfterSave = createDeferred<{
      endpoint: string
      region: string
      bucket: string
      access_key_id: string
      secret_access_key: string
      prefix: string
      force_path_style: boolean
    }>()

    getS3Config.mockReset()
    getS3Config
      .mockImplementationOnce(() => firstLoad.promise)
      .mockImplementationOnce(() => reloadAfterSave.promise)

    const config = useBackupViewConfig({
      t: (key: string) => key,
      showError: vi.fn(),
      showSuccess: vi.fn()
    })

    const staleLoadPromise = config.loadS3Config()
    config.s3Form.value.endpoint = 'https://next.example.com'
    const savePromise = config.saveS3Config()

    firstLoad.resolve({
      endpoint: 'https://stale.example.com',
      region: 'auto',
      bucket: 'stale-bucket',
      access_key_id: 'STALE',
      secret_access_key: '',
      prefix: 'stale/',
      force_path_style: false
    })
    reloadAfterSave.resolve({
      endpoint: 'https://next.example.com',
      region: 'us-east-1',
      bucket: 'fresh-bucket',
      access_key_id: 'FRESH',
      secret_access_key: '',
      prefix: 'fresh/',
      force_path_style: true
    })

    await Promise.all([staleLoadPromise, savePromise])

    expect(config.s3Form.value).toEqual({
      endpoint: 'https://next.example.com',
      region: 'us-east-1',
      bucket: 'fresh-bucket',
      access_key_id: 'FRESH',
      secret_access_key: '',
      prefix: 'fresh/',
      force_path_style: true
    })
    expect(config.s3SecretConfigured.value).toBe(true)
  })

  it('keeps the latest schedule load authoritative', async () => {
    const firstLoad = createDeferred<{
      enabled: boolean
      cron_expr: string
      retain_days: number
      retain_count: number
    }>()
    const secondLoad = createDeferred<{
      enabled: boolean
      cron_expr: string
      retain_days: number
      retain_count: number
    }>()

    getSchedule.mockReset()
    getSchedule
      .mockImplementationOnce(() => firstLoad.promise)
      .mockImplementationOnce(() => secondLoad.promise)

    const config = useBackupViewConfig({
      t: (key: string) => key,
      showError: vi.fn(),
      showSuccess: vi.fn()
    })

    const firstPromise = config.loadSchedule()
    const secondPromise = config.loadSchedule()

    secondLoad.resolve({
      enabled: false,
      cron_expr: '0 3 * * *',
      retain_days: 30,
      retain_count: 3
    })
    firstLoad.resolve({
      enabled: true,
      cron_expr: '0 1 * * *',
      retain_days: 7,
      retain_count: 7
    })

    await Promise.all([firstPromise, secondPromise])

    expect(config.scheduleForm.value).toEqual({
      enabled: false,
      cron_expr: '0 3 * * *',
      retain_days: 30,
      retain_count: 3
    })
  })
})
