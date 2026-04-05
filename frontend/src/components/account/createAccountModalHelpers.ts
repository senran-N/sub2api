import type { AccountPlatform, CreateAccountRequest } from '@/types'
import { isOpenAIWSModeEnabled, type OpenAIWSMode } from '@/utils/openaiWsMode'

export type CreateAccountCategory = 'oauth-based' | 'apikey' | 'bedrock'
export type AntigravityAccountType = 'oauth' | 'upstream'
export type SoraAccountType = 'oauth' | 'apikey'
export type GeminiOAuthType = 'code_assist' | 'google_one' | 'ai_studio'

interface ResolveCreateAccountOAuthFlowOptions {
  accountCategory: CreateAccountCategory
  antigravityAccountType: AntigravityAccountType
  platform: AccountPlatform
}

interface ResolveGeminiSelectedTierOptions {
  accountCategory: CreateAccountCategory
  geminiOAuthType: GeminiOAuthType
  geminiTierAIStudio: string
  geminiTierGcp: string
  geminiTierGoogleOne: string
  platform: AccountPlatform
}

interface BuildCreateOpenAIExtraOptions {
  accountCategory: CreateAccountCategory
  base?: Record<string, unknown>
  codexCLIOnlyEnabled: boolean
  openaiAPIKeyResponsesWebSocketV2Mode: OpenAIWSMode
  openaiOAuthResponsesWebSocketV2Mode: OpenAIWSMode
  openaiPassthroughEnabled: boolean
  platform: AccountPlatform
}

interface BuildCreateAnthropicExtraOptions {
  accountCategory: CreateAccountCategory
  anthropicPassthroughEnabled: boolean
  base?: Record<string, unknown>
  platform: AccountPlatform
}

interface BuildCreateAnthropicQuotaControlExtraOptions {
  baseExtra?: Record<string, unknown>
  baseRpm: number | null
  cacheTTLOverrideEnabled: boolean
  cacheTTLOverrideTarget: string
  customBaseUrl: string
  customBaseUrlEnabled: boolean
  rpmLimitEnabled: boolean
  rpmStickyBuffer: number | null
  rpmStrategy: 'tiered' | 'sticky_exempt'
  sessionIdMaskingEnabled: boolean
  sessionIdleTimeout: number | null
  sessionLimitEnabled: boolean
  tlsFingerprintEnabled: boolean
  tlsFingerprintProfileId: number | null
  userMsgQueueMode: string
  windowCostEnabled: boolean
  windowCostLimit: number | null
  windowCostStickyReserve: number | null
  maxSessions: number | null
}

interface BuildCreateAccountSharedPayloadOptions {
  autoPauseOnExpired: boolean
  concurrency: number
  expiresAt: number | null
  groupIds: number[]
  loadFactor: number | null
  notes: string
  priority: number
  proxyId: number | null
  rateMultiplier: number
}

interface BuildCreateOAuthAccountPayloadOptions {
  common: Omit<
    CreateAccountRequest,
    'credentials' | 'extra' | 'name' | 'platform' | 'type'
  >
  credentials: Record<string, unknown>
  extra?: Record<string, unknown>
  name: string
  platform: AccountPlatform
  type: CreateAccountRequest['type']
}

interface ResolveBatchCreateOutcomeOptions {
  failedCount: number
  successCount: number
  t: (key: string, values?: Record<string, unknown>) => string
}

export function resolveCreateAccountOAuthFlow(
  options: ResolveCreateAccountOAuthFlowOptions
) {
  if (options.platform === 'antigravity' && options.antigravityAccountType === 'upstream') {
    return false
  }
  if (options.platform === 'anthropic' && options.accountCategory === 'bedrock') {
    return false
  }
  return options.accountCategory === 'oauth-based'
}

export function resolveCreateAccountGeminiSelectedTier(
  options: ResolveGeminiSelectedTierOptions
) {
  if (options.platform !== 'gemini') return ''
  if (options.accountCategory === 'apikey') return options.geminiTierAIStudio

  switch (options.geminiOAuthType) {
    case 'google_one':
      return options.geminiTierGoogleOne
    case 'code_assist':
      return options.geminiTierGcp
    default:
      return options.geminiTierAIStudio
  }
}

export function buildCreateAntigravityExtra(options: {
  allowOverages: boolean
  mixedScheduling: boolean
}) {
  const extra: Record<string, unknown> = {}
  if (options.mixedScheduling) extra.mixed_scheduling = true
  if (options.allowOverages) extra.allow_overages = true
  return Object.keys(extra).length > 0 ? extra : undefined
}

export function buildCreateOpenAIExtra(
  options: BuildCreateOpenAIExtraOptions
): Record<string, unknown> | undefined {
  if (options.platform !== 'openai') {
    return options.base
  }

  const extra: Record<string, unknown> = { ...(options.base || {}) }

  if (options.accountCategory === 'oauth-based') {
    extra.openai_oauth_responses_websockets_v2_mode =
      options.openaiOAuthResponsesWebSocketV2Mode
    extra.openai_oauth_responses_websockets_v2_enabled = isOpenAIWSModeEnabled(
      options.openaiOAuthResponsesWebSocketV2Mode
    )
  } else if (options.accountCategory === 'apikey') {
    extra.openai_apikey_responses_websockets_v2_mode =
      options.openaiAPIKeyResponsesWebSocketV2Mode
    extra.openai_apikey_responses_websockets_v2_enabled = isOpenAIWSModeEnabled(
      options.openaiAPIKeyResponsesWebSocketV2Mode
    )
  }

  delete extra.responses_websockets_v2_enabled
  delete extra.openai_ws_enabled

  if (options.openaiPassthroughEnabled) {
    extra.openai_passthrough = true
  } else {
    delete extra.openai_passthrough
    delete extra.openai_oauth_passthrough
  }

  if (options.accountCategory === 'oauth-based' && options.codexCLIOnlyEnabled) {
    extra.codex_cli_only = true
  } else {
    delete extra.codex_cli_only
  }

  return Object.keys(extra).length > 0 ? extra : undefined
}

export function buildCreateAnthropicExtra(
  options: BuildCreateAnthropicExtraOptions
): Record<string, unknown> | undefined {
  if (options.platform !== 'anthropic' || options.accountCategory !== 'apikey') {
    return options.base
  }

  const extra: Record<string, unknown> = { ...(options.base || {}) }
  if (options.anthropicPassthroughEnabled) {
    extra.anthropic_passthrough = true
  } else {
    delete extra.anthropic_passthrough
  }

  return Object.keys(extra).length > 0 ? extra : undefined
}

export function buildCreateAnthropicQuotaControlExtra(
  options: BuildCreateAnthropicQuotaControlExtraOptions
) {
  const extra: Record<string, unknown> = { ...(options.baseExtra || {}) }

  if (options.windowCostEnabled && options.windowCostLimit != null && options.windowCostLimit > 0) {
    extra.window_cost_limit = options.windowCostLimit
    extra.window_cost_sticky_reserve = options.windowCostStickyReserve ?? 10
  }

  if (options.sessionLimitEnabled && options.maxSessions != null && options.maxSessions > 0) {
    extra.max_sessions = options.maxSessions
    extra.session_idle_timeout_minutes = options.sessionIdleTimeout ?? 5
  }

  if (options.rpmLimitEnabled) {
    const DEFAULT_BASE_RPM = 15
    extra.base_rpm =
      options.baseRpm != null && options.baseRpm > 0 ? options.baseRpm : DEFAULT_BASE_RPM
    extra.rpm_strategy = options.rpmStrategy
    if (options.rpmStickyBuffer != null && options.rpmStickyBuffer > 0) {
      extra.rpm_sticky_buffer = options.rpmStickyBuffer
    }
  }

  if (options.userMsgQueueMode) {
    extra.user_msg_queue_mode = options.userMsgQueueMode
  }

  if (options.tlsFingerprintEnabled) {
    extra.enable_tls_fingerprint = true
    if (options.tlsFingerprintProfileId) {
      extra.tls_fingerprint_profile_id = options.tlsFingerprintProfileId
    }
  }

  if (options.sessionIdMaskingEnabled) {
    extra.session_id_masking_enabled = true
  }

  if (options.cacheTTLOverrideEnabled) {
    extra.cache_ttl_override_enabled = true
    extra.cache_ttl_override_target = options.cacheTTLOverrideTarget
  }

  if (options.customBaseUrlEnabled && options.customBaseUrl.trim()) {
    extra.custom_base_url_enabled = true
    extra.custom_base_url = options.customBaseUrl.trim()
  }

  return extra
}

export function buildCreateAccountSharedPayload(
  options: BuildCreateAccountSharedPayloadOptions
): Omit<CreateAccountRequest, 'credentials' | 'extra' | 'name' | 'platform' | 'type'> {
  return {
    proxy_id: options.proxyId,
    concurrency: options.concurrency,
    load_factor: options.loadFactor ?? undefined,
    priority: options.priority,
    rate_multiplier: options.rateMultiplier,
    group_ids: options.groupIds,
    expires_at: options.expiresAt,
    auto_pause_on_expired: options.autoPauseOnExpired,
    notes: options.notes
  }
}

export function buildCreateOAuthAccountPayload(
  options: BuildCreateOAuthAccountPayloadOptions
): CreateAccountRequest {
  return {
    ...options.common,
    name: options.name,
    platform: options.platform,
    type: options.type,
    credentials: options.credentials,
    extra: options.extra
  }
}

export function buildCreateBatchAccountName(
  baseName: string,
  index: number,
  total: number,
  fallbackBaseName?: string,
  suffix?: string
) {
  const resolvedBaseName = baseName || fallbackBaseName || ''
  const indexedName = total > 1 ? `${resolvedBaseName} #${index + 1}` : resolvedBaseName
  return suffix ? `${indexedName} ${suffix}` : indexedName
}

export function buildCreateSoraOAuthCredentials(
  credentials: Record<string, unknown>,
  sessionToken?: string
) {
  const result: Record<string, unknown> = {
    access_token: credentials.access_token,
    refresh_token: credentials.refresh_token,
    client_id: credentials.client_id,
    expires_at: credentials.expires_at
  }

  if (sessionToken) {
    result.session_token = sessionToken
  }

  return result
}

export function resolveBatchCreateOutcome(options: ResolveBatchCreateOutcomeOptions) {
  if (options.successCount > 0 && options.failedCount === 0) {
    return {
      type: 'success' as const,
      message:
        options.successCount > 1
          ? options.t('admin.accounts.oauth.batchSuccess', { count: options.successCount })
          : options.t('admin.accounts.accountCreated'),
      shouldClose: true,
      shouldEmitCreated: true
    }
  }

  if (options.successCount > 0) {
    return {
      type: 'warning' as const,
      message: options.t('admin.accounts.oauth.batchPartialSuccess', {
        success: options.successCount,
        failed: options.failedCount
      }),
      shouldClose: false,
      shouldEmitCreated: true
    }
  }

  return {
    type: 'error' as const,
    message: options.t('admin.accounts.oauth.batchFailed'),
    shouldClose: false,
    shouldEmitCreated: false
  }
}

export function buildCreateSoraExtra(
  base?: Record<string, unknown>,
  linkedOpenAIAccountId?: string | number
): Record<string, unknown> | undefined {
  const extra: Record<string, unknown> = { ...(base || {}) }

  if (linkedOpenAIAccountId !== undefined && linkedOpenAIAccountId !== null) {
    const id = String(linkedOpenAIAccountId).trim()
    if (id) {
      extra.linked_openai_account_id = id
    }
  }

  delete extra.openai_passthrough
  delete extra.openai_oauth_passthrough
  delete extra.codex_cli_only
  delete extra.openai_oauth_responses_websockets_v2_mode
  delete extra.openai_apikey_responses_websockets_v2_mode
  delete extra.openai_oauth_responses_websockets_v2_enabled
  delete extra.openai_apikey_responses_websockets_v2_enabled
  delete extra.responses_websockets_v2_enabled
  delete extra.openai_ws_enabled

  return Object.keys(extra).length > 0 ? extra : undefined
}
