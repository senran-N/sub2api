import type { InstallRequest } from '@/api/setup'

export type SetupStepId = 'database' | 'redis' | 'admin' | 'complete'

export interface SetupStep {
  id: SetupStepId
  title: string
}

type Translate = (key: string) => string

interface SetupLocationLike {
  port: string
  protocol: string
}

interface SetupErrorLike {
  response?: {
    data?: {
      detail?: string
      message?: string
    }
  }
  message?: string
}

export interface PollSetupServiceReadyOptions {
  fetchStatus: () => Promise<unknown>
  sleep?: (ms: number) => Promise<void>
  initialDelayMs?: number
  intervalMs?: number
  maxAttempts?: number
}

export const SETUP_SERVICE_POLL_INITIAL_DELAY_MS = 3000
export const SETUP_SERVICE_POLL_INTERVAL_MS = 1000
export const SETUP_SERVICE_POLL_MAX_ATTEMPTS = 60
export const SETUP_SERVICE_REDIRECT_DELAY_MS = 1500

const sleep = (ms: number) => new Promise<void>((resolve) => window.setTimeout(resolve, ms))

export function buildSetupWizardSteps(t: Translate): SetupStep[] {
  return [
    { id: 'database', title: t('setup.database.title') },
    { id: 'redis', title: t('setup.redis.title') },
    { id: 'admin', title: t('setup.admin.title') },
    { id: 'complete', title: t('setup.ready.title') }
  ]
}

export function resolveSetupWizardPort(locationLike: SetupLocationLike): number {
  const parsedPort = Number.parseInt(locationLike.port, 10)
  if (Number.isFinite(parsedPort)) {
    return parsedPort
  }

  return locationLike.protocol === 'https:' ? 443 : 80
}

export function createSetupInstallRequest(locationLike: SetupLocationLike): InstallRequest {
  return {
    database: {
      host: 'localhost',
      port: 5432,
      user: 'postgres',
      password: '',
      dbname: 'sub2api',
      sslmode: 'disable'
    },
    redis: {
      host: 'localhost',
      port: 6379,
      password: '',
      db: 0,
      enable_tls: false
    },
    admin: {
      email: '',
      password: ''
    },
    server: {
      host: '0.0.0.0',
      port: resolveSetupWizardPort(locationLike),
      mode: 'release'
    }
  }
}

export function resolveSetupWizardErrorMessage(error: unknown, fallback: string): string {
  const setupError = error as SetupErrorLike | null

  return (
    setupError?.response?.data?.detail ||
    setupError?.response?.data?.message ||
    setupError?.message ||
    fallback
  )
}

export function isSetupServiceReady(payload: unknown): boolean {
  if (!payload || typeof payload !== 'object') {
    return false
  }

  const data = 'data' in payload ? (payload as { data?: unknown }).data : null
  if (!data || typeof data !== 'object' || !('needs_setup' in data)) {
    return false
  }

  return (data as { needs_setup?: boolean }).needs_setup === false
}

export async function pollSetupServiceReady({
  fetchStatus,
  sleep: wait = sleep,
  initialDelayMs = SETUP_SERVICE_POLL_INITIAL_DELAY_MS,
  intervalMs = SETUP_SERVICE_POLL_INTERVAL_MS,
  maxAttempts = SETUP_SERVICE_POLL_MAX_ATTEMPTS
}: PollSetupServiceReadyOptions): Promise<boolean> {
  await wait(initialDelayMs)

  for (let attempt = 0; attempt < maxAttempts; attempt += 1) {
    try {
      const payload = await fetchStatus()
      if (isSetupServiceReady(payload)) {
        return true
      }
    } catch {
      // Service may still be restarting; keep polling until the timeout is reached.
    }

    if (attempt < maxAttempts - 1) {
      await wait(intervalMs)
    }
  }

  return false
}
