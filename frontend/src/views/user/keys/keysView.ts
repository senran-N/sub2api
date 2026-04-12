import type { ApiKey, PublicSettings } from '@/types'
import type { BatchApiKeyUsageStats } from '@/api/usage'

export interface ApiKeyRateLimitWindow {
  key: '5h' | '1d' | '7d'
  label: string
  usage: number
  limit: number
  resetAt: string | null
}

export type CcsClientType = 'claude' | 'gemini'

export function maskUserApiKey(key: string): string {
  if (key.length <= 12) {
    return key
  }

  return `${key.slice(0, 8)}...${key.slice(-4)}`
}

export function hasApiKeyIpRestrictions(row: ApiKey): boolean {
  return (row.ip_whitelist?.length ?? 0) > 0 || (row.ip_blacklist?.length ?? 0) > 0
}

export function getApiKeyQuotaProgressWidth(row: ApiKey): string {
  if (row.quota <= 0) {
    return '0%'
  }

  return `${Math.min((row.quota_used / row.quota) * 100, 100)}%`
}

export function getApiKeyQuotaTextTone(row: ApiKey): string {
  if (row.quota_used >= row.quota) {
    return 'theme-text-danger'
  }
  if (row.quota_used >= row.quota * 0.8) {
    return 'theme-text-warning'
  }

  return 'theme-text-default'
}

export function getApiKeyQuotaBarTone(row: ApiKey): string {
  if (row.quota_used >= row.quota) {
    return 'theme-progress-fill--danger'
  }
  if (row.quota_used >= row.quota * 0.8) {
    return 'theme-progress-fill--warning'
  }

  return 'theme-progress-fill--accent'
}

export function getApiKeyRateLimitTextTone(
  usage: number,
  limit: number
): string {
  if (usage >= limit) {
    return 'theme-text-danger'
  }
  if (usage >= limit * 0.8) {
    return 'theme-text-warning'
  }

  return 'theme-text-default'
}

export function getApiKeyRateLimitBarTone(
  usage: number,
  limit: number
): string {
  if (usage >= limit) {
    return 'theme-progress-fill--danger'
  }
  if (usage >= limit * 0.8) {
    return 'theme-progress-fill--warning'
  }

  return 'theme-progress-fill--success'
}

export function getApiKeyRateLimitProgressWidth(
  usage: number,
  limit: number
): string {
  if (limit <= 0) {
    return '0%'
  }

  return `${Math.min((usage / limit) * 100, 100)}%`
}

export function hasApiKeyRateLimitUsage(row: ApiKey): boolean {
  return row.usage_5h > 0 || row.usage_1d > 0 || row.usage_7d > 0
}

export function getApiKeyRateLimitWindows(row: ApiKey): ApiKeyRateLimitWindow[] {
  const windows: ApiKeyRateLimitWindow[] = []

  if (row.rate_limit_5h > 0) {
    windows.push({
      key: '5h',
      label: '5h',
      usage: row.usage_5h,
      limit: row.rate_limit_5h,
      resetAt: row.reset_5h_at
    })
  }

  if (row.rate_limit_1d > 0) {
    windows.push({
      key: '1d',
      label: '1d',
      usage: row.usage_1d,
      limit: row.rate_limit_1d,
      resetAt: row.reset_1d_at
    })
  }

  if (row.rate_limit_7d > 0) {
    windows.push({
      key: '7d',
      label: '7d',
      usage: row.usage_7d,
      limit: row.rate_limit_7d,
      resetAt: row.reset_7d_at
    })
  }

  return windows
}

export function getApiKeyStatusBadgeClass(status: ApiKey['status']): string {
  if (status === 'active') {
    return 'badge-success'
  }
  if (status === 'quota_exhausted') {
    return 'badge-warning'
  }
  if (status === 'expired') {
    return 'badge-danger'
  }

  return 'badge-gray'
}

export function getApiKeyUsageSummary(
  stats: BatchApiKeyUsageStats | undefined
): { todayCost: string; totalCost: string } {
  return {
    todayCost: (stats?.today_actual_cost ?? 0).toFixed(4),
    totalCost: (stats?.total_actual_cost ?? 0).toFixed(4)
  }
}

export function getApiKeyExpirationTextClass(
  expiresAt: string | null | undefined,
  now: Date = new Date()
): string {
  if (!expiresAt) {
    return 'text-sm theme-text-subtle'
  }

  return new Date(expiresAt) < now
    ? 'text-sm theme-text-danger'
    : 'text-sm theme-text-muted'
}

type Translate = (key: string) => string

export function formatApiKeyResetTime(
  resetAt: string | null,
  now: Date,
  t: Translate
): string {
  if (!resetAt) return ''

  const diff = new Date(resetAt).getTime() - now.getTime()
  if (diff <= 0) return t('keys.resetNow')

  const days = Math.floor(diff / 86400000)
  const hours = Math.floor((diff % 86400000) / 3600000)
  const mins = Math.floor((diff % 3600000) / 60000)

  if (days > 0) return `${days}d ${hours}h`
  if (hours > 0) return `${hours}h ${mins}m`
  return `${mins}m`
}

interface CcsImportTarget {
  app: string
  endpoint: string
}

function resolveCcsImportTarget(
  platform: string,
  clientType: CcsClientType,
  baseUrl: string
): CcsImportTarget {
  if (platform === 'antigravity') {
    return {
      app: clientType === 'gemini' ? 'gemini' : 'claude',
      endpoint: `${baseUrl}/antigravity`
    }
  }

  if (platform === 'openai') {
    return { app: 'codex', endpoint: baseUrl }
  }

  if (platform === 'gemini') {
    return { app: 'gemini', endpoint: baseUrl }
  }

  return { app: 'claude', endpoint: baseUrl }
}

export function buildCcsImportDeeplink(
  row: ApiKey,
  publicSettings: Pick<PublicSettings, 'api_base_url' | 'site_name'> | null | undefined,
  clientType: CcsClientType,
  fallbackBaseUrl: string
): string {
  const baseUrl = publicSettings?.api_base_url || fallbackBaseUrl
  const platform = row.group?.platform || 'anthropic'
  const { app, endpoint } = resolveCcsImportTarget(platform, clientType, baseUrl)
  const providerName = (publicSettings?.site_name || 'sub2api').trim() || 'sub2api'
  const usageEndpoint = new URL('/v1/usage', baseUrl).toString()
  const usageScript = `({
    request: {
      url: "${usageEndpoint}",
      method: "GET",
      headers: { "Authorization": "Bearer {{apiKey}}" }
    },
    extractor: function(response) {
      const remaining = response?.remaining ?? response?.quota?.remaining ?? response?.balance;
      const unit = response?.unit ?? response?.quota?.unit ?? "USD";
      return {
        isValid: response?.is_active ?? response?.isValid ?? true,
        remaining,
        unit
      };
    }
  })`

  const params = new URLSearchParams({
    resource: 'provider',
    app,
    name: providerName,
    homepage: baseUrl,
    endpoint,
    apiKey: row.key,
    configFormat: 'json',
    usageEnabled: 'true',
    usageScript: btoa(usageScript),
    usageAutoInterval: '30'
  })

  return `ccswitch://v1/import?${params.toString()}`
}
