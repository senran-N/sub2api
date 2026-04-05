import { describe, expect, it, vi } from 'vitest'
import {
  createSetupInstallRequest,
  isSetupServiceReady,
  pollSetupServiceReady,
  resolveSetupWizardErrorMessage,
  resolveSetupWizardPort
} from '../setupWizardView'

describe('setupWizardView', () => {
  it('resolves the server port from browser location', () => {
    expect(resolveSetupWizardPort({ port: '7788', protocol: 'http:' })).toBe(7788)
    expect(resolveSetupWizardPort({ port: '', protocol: 'https:' })).toBe(443)
    expect(resolveSetupWizardPort({ port: '', protocol: 'http:' })).toBe(80)
  })

  it('creates the default install request with the detected port', () => {
    expect(
      createSetupInstallRequest({ port: '9527', protocol: 'http:' })
    ).toMatchObject({
      database: {
        host: 'localhost',
        port: 5432,
        user: 'postgres',
        dbname: 'sub2api',
        sslmode: 'disable'
      },
      redis: {
        host: 'localhost',
        port: 6379,
        db: 0,
        enable_tls: false
      },
      server: {
        host: '0.0.0.0',
        port: 9527,
        mode: 'release'
      }
    })
  })

  it('prefers setup API detail fields when building error messages', () => {
    expect(
      resolveSetupWizardErrorMessage(
        {
          response: {
            data: {
              detail: 'detail message',
              message: 'response message'
            }
          },
          message: 'generic message'
        },
        'fallback'
      )
    ).toBe('detail message')

    expect(resolveSetupWizardErrorMessage({ message: 'generic message' }, 'fallback')).toBe(
      'generic message'
    )
    expect(resolveSetupWizardErrorMessage(null, 'fallback')).toBe('fallback')
  })

  it('detects when the setup service has restarted into normal mode', () => {
    expect(isSetupServiceReady({ data: { needs_setup: false } })).toBe(true)
    expect(isSetupServiceReady({ data: { needs_setup: true } })).toBe(false)
    expect(isSetupServiceReady({})).toBe(false)
  })

  it('polls until the setup service is ready', async () => {
    const fetchStatus = vi
      .fn<() => Promise<unknown>>()
      .mockResolvedValueOnce({ data: { needs_setup: true } })
      .mockResolvedValueOnce({ data: { needs_setup: false } })
    const sleep = vi.fn<(_: number) => Promise<void>>().mockResolvedValue(undefined)

    await expect(
      pollSetupServiceReady({
        fetchStatus,
        sleep,
        initialDelayMs: 0,
        intervalMs: 0,
        maxAttempts: 3
      })
    ).resolves.toBe(true)

    expect(fetchStatus).toHaveBeenCalledTimes(2)
    expect(sleep).toHaveBeenCalledTimes(2)
  })

  it('stops polling after the configured attempt limit', async () => {
    const fetchStatus = vi.fn<() => Promise<unknown>>().mockResolvedValue({ data: { needs_setup: true } })

    await expect(
      pollSetupServiceReady({
        fetchStatus,
        sleep: async () => undefined,
        initialDelayMs: 0,
        intervalMs: 0,
        maxAttempts: 2
      })
    ).resolves.toBe(false)

    expect(fetchStatus).toHaveBeenCalledTimes(2)
  })
})
