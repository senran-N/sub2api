import type { AccountPlatform, CreateAccountRequest } from '@/types'
import { isOpenAIWSModeEnabled, type OpenAIWSMode } from '@/utils/openaiWsMode'
import {
  applyInterceptWarmup,
  assignBuiltModelMapping,
  getDefaultBaseURL,
  normalizePoolModeRetryCount,
  replaceAntigravityModelMapping,
  type ModelMapping,
  type TempUnschedRulePayload
} from './credentialsBuilder'

export type CreateAccountCategory = 'oauth-based' | 'apikey' | 'upstream' | 'session' | 'bedrock'
export type AntigravityAccountType = 'oauth' | 'upstream'
export type GeminiOAuthType = 'code_assist' | 'google_one' | 'ai_studio'
export type GeminiGoogleOneTier = 'google_one_free' | 'google_ai_pro' | 'google_ai_ultra'
export type GeminiGcpTier = 'gcp_standard' | 'gcp_enterprise'
export type GeminiAIStudioTier = 'aistudio_free' | 'aistudio_paid'
export type BedrockAuthMode = 'sigv4' | 'apikey'

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

interface BuildCreateAnthropicOAuthAccountPayloadOptions {
  common: Omit<
    CreateAccountRequest,
    'credentials' | 'extra' | 'name' | 'platform' | 'type'
  >
  extra?: Record<string, unknown>
  interceptWarmupRequests: boolean
  name: string
  platform: AccountPlatform
  tempUnschedPayload?: TempUnschedRulePayload[]
  tokenInfo: Record<string, unknown>
  type: CreateAccountRequest['type']
}

interface BuildCreateOpenAICompatOAuthTargetOptions {
  baseName: string
  credentials: Record<string, unknown>
  fallbackBaseName?: string
  extra?: Record<string, unknown>
  index?: number
  platform: 'openai'
  total?: number
}

interface ResolveBatchCreateOutcomeOptions {
  failedCount: number
  successCount: number
  t: (key: string, values?: Record<string, unknown>) => string
}

interface BuildCreateCredentialResult {
  credentials?: Record<string, unknown>
  errorMessageKey?: string
}

interface ModelRestrictionOptions {
  allowedModels: string[]
  modelMappings: ModelMapping[]
  mode: 'whitelist' | 'mapping'
}

interface BuildCreateBedrockCredentialsOptions extends ModelRestrictionOptions {
  accessKeyId: string
  apiKey: string
  authMode: 'sigv4' | 'apikey'
  forceGlobal: boolean
  interceptWarmupRequests: boolean
  poolModeEnabled: boolean
  poolModeRetryCount: number
  region: string
  secretAccessKey: string
  sessionToken: string
}

interface BuildCreateAntigravityUpstreamCredentialsOptions {
  apiKey: string
  baseUrl: string
  interceptWarmupRequests: boolean
  modelMappings: ModelMapping[]
}

interface BuildCreateAnthropicOAuthCredentialsOptions {
  interceptWarmupRequests: boolean
  tempUnschedPayload?: TempUnschedRulePayload[]
  tokenInfo: Record<string, unknown>
}

interface BuildCreateAntigravityOAuthCredentialsOptions {
  interceptWarmupRequests: boolean
  modelMappings: ModelMapping[]
  tokenInfo: Record<string, unknown>
}

interface BuildCreateApiKeyCredentialsOptions extends ModelRestrictionOptions {
  apiKey: string
  baseUrl: string
  customErrorCodesEnabled: boolean
  geminiTierId: string
  interceptWarmupRequests: boolean
  isOpenAIModelRestrictionDisabled: boolean
  platform: AccountPlatform
  poolModeEnabled: boolean
  poolModeRetryCount: number
  selectedErrorCodes: number[]
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

export function buildCreateAccountRequest(
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

export function buildCreateAccountPayload(
  options: BuildCreateOAuthAccountPayloadOptions
): CreateAccountRequest {
  return buildCreateAccountRequest(options)
}

export function buildCreateOAuthAccountPayload(
  options: BuildCreateOAuthAccountPayloadOptions
): CreateAccountRequest {
  return buildCreateAccountRequest(options)
}

export function buildCreateAnthropicOAuthAccountPayload(
  options: BuildCreateAnthropicOAuthAccountPayloadOptions
): CreateAccountRequest {
  return buildCreateAccountRequest({
    common: options.common,
    name: options.name,
    platform: options.platform,
    type: options.type,
    credentials: buildCreateAnthropicOAuthCredentials({
      interceptWarmupRequests: options.interceptWarmupRequests,
      tempUnschedPayload: options.tempUnschedPayload,
      tokenInfo: options.tokenInfo
    }),
    extra: options.extra
  })
}

export function buildCreateOpenAICompatOAuthTarget(
  options: BuildCreateOpenAICompatOAuthTargetOptions
) {
  return {
    name: buildCreateBatchAccountName(
      options.baseName,
      options.index ?? 0,
      options.total ?? 1,
      options.fallbackBaseName
    ),
    platform: 'openai' as const,
    type: 'oauth' as const,
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

export function buildCreateBedrockCredentials(
  options: BuildCreateBedrockCredentialsOptions
): BuildCreateCredentialResult {
  const credentials: Record<string, unknown> = {
    auth_mode: options.authMode,
    aws_region: options.region.trim() || 'us-east-1'
  }

  if (options.authMode === 'sigv4') {
    if (!options.accessKeyId.trim()) {
      return { errorMessageKey: 'admin.accounts.bedrockAccessKeyIdRequired' }
    }
    if (!options.secretAccessKey.trim()) {
      return { errorMessageKey: 'admin.accounts.bedrockSecretAccessKeyRequired' }
    }

    credentials.aws_access_key_id = options.accessKeyId.trim()
    credentials.aws_secret_access_key = options.secretAccessKey.trim()
    if (options.sessionToken.trim()) {
      credentials.aws_session_token = options.sessionToken.trim()
    }
  } else {
    if (!options.apiKey.trim()) {
      return { errorMessageKey: 'admin.accounts.bedrockApiKeyRequired' }
    }
    credentials.api_key = options.apiKey.trim()
  }

  if (options.forceGlobal) {
    credentials.aws_force_global = 'true'
  }

  assignBuiltModelMapping(
    credentials,
    options.mode,
    options.allowedModels,
    options.modelMappings
  )

  if (options.poolModeEnabled) {
    credentials.pool_mode = true
    credentials.pool_mode_retry_count = normalizePoolModeRetryCount(options.poolModeRetryCount)
  }

  applyInterceptWarmup(credentials, options.interceptWarmupRequests, 'create')

  return { credentials }
}

export function buildCreateAntigravityUpstreamCredentials(
  options: BuildCreateAntigravityUpstreamCredentialsOptions
): BuildCreateCredentialResult {
  if (!options.baseUrl.trim()) {
    return { errorMessageKey: 'admin.accounts.upstream.pleaseEnterBaseUrl' }
  }
  if (!options.apiKey.trim()) {
    return { errorMessageKey: 'admin.accounts.upstream.pleaseEnterApiKey' }
  }

  const credentials: Record<string, unknown> = {
    base_url: options.baseUrl.trim(),
    api_key: options.apiKey.trim()
  }

  replaceAntigravityModelMapping(credentials, options.modelMappings)
  applyInterceptWarmup(credentials, options.interceptWarmupRequests, 'create')

  return { credentials }
}

export function buildCreateAnthropicOAuthCredentials(
  options: BuildCreateAnthropicOAuthCredentialsOptions
): Record<string, unknown> {
  const credentials: Record<string, unknown> = { ...options.tokenInfo }
  applyInterceptWarmup(credentials, options.interceptWarmupRequests, 'create')

  if (options.tempUnschedPayload && options.tempUnschedPayload.length > 0) {
    credentials.temp_unschedulable_enabled = true
    credentials.temp_unschedulable_rules = options.tempUnschedPayload
  }

  return credentials
}

export function buildCreateAntigravityOAuthCredentials(
  options: BuildCreateAntigravityOAuthCredentialsOptions
): Record<string, unknown> {
  const credentials: Record<string, unknown> = { ...options.tokenInfo }
  applyInterceptWarmup(credentials, options.interceptWarmupRequests, 'create')
  replaceAntigravityModelMapping(credentials, options.modelMappings)
  return credentials
}

export function buildCreateApiKeyCredentials(
  options: BuildCreateApiKeyCredentialsOptions
): BuildCreateCredentialResult {
  if (!options.apiKey.trim()) {
    return { errorMessageKey: 'admin.accounts.pleaseEnterApiKey' }
  }

  const trimmedBaseUrl = options.baseUrl.trim()
  const credentials: Record<string, unknown> = {
    base_url: trimmedBaseUrl || getDefaultBaseURL(options.platform),
    api_key: options.apiKey.trim()
  }

  if (options.platform === 'gemini') {
    credentials.tier_id = options.geminiTierId
  }

  if (!options.isOpenAIModelRestrictionDisabled) {
    assignBuiltModelMapping(
      credentials,
      options.mode,
      options.allowedModels,
      options.modelMappings
    )
  }

  if (options.poolModeEnabled) {
    credentials.pool_mode = true
    credentials.pool_mode_retry_count = normalizePoolModeRetryCount(options.poolModeRetryCount)
  }

  if (options.customErrorCodesEnabled) {
    credentials.custom_error_codes_enabled = true
    credentials.custom_error_codes = [...options.selectedErrorCodes]
  }

  applyInterceptWarmup(credentials, options.interceptWarmupRequests, 'create')

  return { credentials }
}
