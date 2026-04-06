import type { ApiResponse, PublicSettings } from '@/types'
import { getLocale } from '@/i18n'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api/v1'

export interface SetupStatus {
  needs_setup: boolean
  step: string
}

function resolveErrorMessage(response: Response, payload: unknown): string {
  if (payload && typeof payload === 'object') {
    const data = payload as Partial<ApiResponse<unknown>> & { detail?: string }
    if (typeof data.message === 'string' && data.message.trim()) {
      return data.message
    }
    if (typeof data.detail === 'string' && data.detail.trim()) {
      return data.detail
    }
  }

  return `Request failed with status ${response.status}`
}

async function readJsonResponse<T>(url: string): Promise<T> {
  const response = await fetch(url, {
    headers: {
      'Accept-Language': getLocale()
    }
  })

  const payload = await response.json().catch(() => null)

  if (!response.ok) {
    throw new Error(resolveErrorMessage(response, payload))
  }

  if (payload && typeof payload === 'object' && 'code' in payload) {
    const apiPayload = payload as ApiResponse<T>
    if (apiPayload.code !== 0) {
      throw new Error(apiPayload.message || 'Unknown error')
    }
    return apiPayload.data
  }

  return payload as T
}

export function fetchPublicSettings(): Promise<PublicSettings> {
  return readJsonResponse<PublicSettings>(`${API_BASE_URL}/settings/public`)
}

export function fetchSetupStatus(): Promise<SetupStatus> {
  return readJsonResponse<SetupStatus>('/setup/status')
}
