import type { Account } from '@/types'

const GROK_QUOTA_WINDOW_ORDER = ['auto', 'fast', 'expert', 'heavy'] as const
const GROK_CAPABILITY_ORDER = ['chat', 'image', 'image_edit', 'video', 'voice'] as const
const GROK_TIERS = new Set(['basic', 'heavy', 'super'])

type UnknownRecord = Record<string, unknown>

export interface GrokRuntimeTierInfo {
  normalized: string
  raw: string | null
  source: string | null
  confidence: number | null
}

export interface GrokRuntimeCapabilityInfo {
  operations: string[]
  models: string[]
}

export interface GrokRuntimeQuotaWindowInfo {
  name: string
  remaining: number
  total: number
  windowSeconds: number | null
  source: string | null
  resetAt: string | null
  hasSignal: boolean
}

export interface GrokRuntimeSyncInfo {
  lastSyncAt: string | null
  lastProbeAt: string | null
  lastProbeOkAt: string | null
  lastProbeErrorAt: string | null
  lastProbeError: string | null
  lastProbeStatusCode: number | null
}

export interface GrokRuntimeFeedbackInfo {
  lastRequestAt: string | null
  lastRequestCapability: string | null
  lastRequestModel: string | null
  lastRequestUpstreamModel: string | null
  lastFailAt: string | null
  lastFailReason: string | null
  lastFailStatusCode: number | null
  lastFailClass: string | null
  lastFailScope: string | null
  cooldownUntil: string | null
  cooldownModel: string | null
}

export interface GrokAccountRuntimeInfo {
  authMode: string | null
  authFingerprint: string | null
  tier: GrokRuntimeTierInfo
  capabilities: GrokRuntimeCapabilityInfo
  quotaWindows: GrokRuntimeQuotaWindowInfo[]
  sync: GrokRuntimeSyncInfo
  runtime: GrokRuntimeFeedbackInfo
  hasState: boolean
}

function asRecord(value: unknown): UnknownRecord | null {
  if (!value || typeof value !== 'object' || Array.isArray(value)) {
    return null
  }
  return value as UnknownRecord
}

function asString(value: unknown): string | null {
  if (typeof value !== 'string') {
    return null
  }
  const normalized = value.trim()
  return normalized === '' ? null : normalized
}

function asNumber(value: unknown): number | null {
  if (typeof value === 'number' && Number.isFinite(value)) {
    return value
  }
  if (typeof value === 'string') {
    const normalized = value.trim()
    if (normalized === '') {
      return null
    }
    const parsed = Number(normalized)
    return Number.isFinite(parsed) ? parsed : null
  }
  return null
}

function asStringArray(value: unknown): string[] {
  if (!Array.isArray(value)) {
    return []
  }

  return value
    .map((item) => asString(item))
    .filter((item): item is string => Boolean(item))
}

function inferTierFromQuotaWindows(quotaWindows: UnknownRecord | null): string {
  const autoWindow = quotaWindows ? asRecord(quotaWindows.auto) : null
  const total = asNumber(autoWindow?.total)
  switch (total) {
    case 20:
      return 'basic'
    case 50:
      return 'super'
    case 150:
      return 'heavy'
    default:
      return 'unknown'
  }
}

function parseTier(grok: UnknownRecord, quotaWindows: UnknownRecord | null): GrokRuntimeTierInfo {
  const tier = asRecord(grok.tier)
  const normalized = asString(tier?.normalized)?.toLowerCase()

  return {
    normalized: normalized && GROK_TIERS.has(normalized) ? normalized : inferTierFromQuotaWindows(quotaWindows),
    raw: asString(tier?.raw) ?? asString(grok.raw_tier) ?? asString(grok.tier_raw),
    source: asString(tier?.source) ?? asString(grok.tier_source),
    confidence: asNumber(tier?.confidence) ?? asNumber(grok.tier_confidence)
  }
}

function parseCapabilities(grok: UnknownRecord): GrokRuntimeCapabilityInfo {
  const capabilities = asRecord(grok.capabilities)
  if (!capabilities) {
    return { operations: [], models: [] }
  }

  const enabled = new Set<string>()
  for (const operation of asStringArray(capabilities.operations)) {
    enabled.add(operation.toLowerCase())
  }

  for (const operation of GROK_CAPABILITY_ORDER) {
    if (capabilities[operation] === true) {
      enabled.add(operation)
    }
  }

  const operations = Array.from(enabled).sort((left, right) => {
    const leftIndex = GROK_CAPABILITY_ORDER.indexOf(left as (typeof GROK_CAPABILITY_ORDER)[number])
    const rightIndex = GROK_CAPABILITY_ORDER.indexOf(right as (typeof GROK_CAPABILITY_ORDER)[number])
    if (leftIndex === -1 && rightIndex === -1) return left.localeCompare(right)
    if (leftIndex === -1) return 1
    if (rightIndex === -1) return -1
    return leftIndex - rightIndex
  })

  const models = Array.from(new Set(asStringArray(capabilities.models))).sort((left, right) =>
    left.localeCompare(right)
  )

  return { operations, models }
}

function parseQuotaWindows(grok: UnknownRecord): GrokRuntimeQuotaWindowInfo[] {
  const quotaWindows = asRecord(grok.quota_windows)
  if (!quotaWindows) {
    return []
  }

  const orderedNames = [
    ...GROK_QUOTA_WINDOW_ORDER,
    ...Object.keys(quotaWindows).filter((name) => !GROK_QUOTA_WINDOW_ORDER.includes(name as never)).sort()
  ]

  return orderedNames
    .map((name) => {
      const value = asRecord(quotaWindows[name])
      if (!value) {
        return null
      }

      const remaining = asNumber(value.remaining)
      const total = asNumber(value.total)
      const windowSeconds = asNumber(value.window_seconds)
      const source = asString(value.source)
      const resetAt = asString(value.reset_at)

      return {
        name,
        remaining: remaining ?? 0,
        total: total ?? 0,
        windowSeconds,
        source,
        resetAt,
        hasSignal:
          remaining !== null ||
          total !== null ||
          windowSeconds !== null ||
          source !== null ||
          resetAt !== null
      } satisfies GrokRuntimeQuotaWindowInfo
    })
    .filter((window): window is GrokRuntimeQuotaWindowInfo => Boolean(window))
}

function parseSyncInfo(grok: UnknownRecord): GrokRuntimeSyncInfo {
  const sync = asRecord(grok.sync_state)
  return {
    lastSyncAt: asString(sync?.last_sync_at),
    lastProbeAt: asString(sync?.last_probe_at),
    lastProbeOkAt: asString(sync?.last_probe_ok_at),
    lastProbeErrorAt: asString(sync?.last_probe_error_at),
    lastProbeError: asString(sync?.last_probe_error),
    lastProbeStatusCode: asNumber(sync?.last_probe_status_code)
  }
}

function parseRuntimeInfo(grok: UnknownRecord): GrokRuntimeFeedbackInfo {
  const runtime = asRecord(grok.runtime_state)
  return {
    lastRequestAt: asString(runtime?.last_request_at),
    lastRequestCapability: asString(runtime?.last_request_capability),
    lastRequestModel: asString(runtime?.last_request_model),
    lastRequestUpstreamModel: asString(runtime?.last_request_upstream_model),
    lastFailAt: asString(runtime?.last_fail_at),
    lastFailReason: asString(runtime?.last_fail_reason),
    lastFailStatusCode: asNumber(runtime?.last_fail_status_code),
    lastFailClass: asString(runtime?.last_fail_class),
    lastFailScope: asString(runtime?.last_fail_scope),
    cooldownUntil: asString(runtime?.selection_cooldown_until),
    cooldownModel: asString(runtime?.selection_cooldown_model)
  }
}

function hasRuntimeState(runtime: GrokAccountRuntimeInfo): boolean {
  return Boolean(
    runtime.authMode ||
      runtime.authFingerprint ||
      runtime.tier.normalized !== 'unknown' ||
      runtime.tier.raw ||
      runtime.capabilities.operations.length > 0 ||
      runtime.capabilities.models.length > 0 ||
      runtime.quotaWindows.some((window) => window.hasSignal) ||
      runtime.sync.lastSyncAt ||
      runtime.sync.lastProbeAt ||
      runtime.sync.lastProbeOkAt ||
      runtime.sync.lastProbeErrorAt ||
      runtime.sync.lastProbeError ||
      runtime.runtime.lastRequestAt ||
      runtime.runtime.lastFailAt ||
      runtime.runtime.lastFailReason ||
      runtime.runtime.cooldownUntil
  )
}

export function getGrokAccountRuntime(
  account: Pick<Account, 'platform' | 'extra'> | null | undefined
): GrokAccountRuntimeInfo | null {
  if (!account || account.platform !== 'grok') {
    return null
  }

  const extra = asRecord(account.extra)
  const grok = asRecord(extra?.grok) ?? {}
  const quotaWindows = asRecord(grok.quota_windows)

  const runtime: GrokAccountRuntimeInfo = {
    authMode: asString(grok.auth_mode),
    authFingerprint: asString(grok.auth_fingerprint),
    tier: parseTier(grok, quotaWindows),
    capabilities: parseCapabilities(grok),
    quotaWindows: parseQuotaWindows(grok),
    sync: parseSyncInfo(grok),
    runtime: parseRuntimeInfo(grok),
    hasState: false
  }

  runtime.hasState = hasRuntimeState(runtime)
  return runtime
}
