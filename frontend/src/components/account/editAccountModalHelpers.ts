import type { AccountType } from '@/types'
import {
  OPENAI_WS_MODE_OFF,
  isOpenAIWSModeEnabled,
  resolveOpenAIWSModeFromExtra,
  type OpenAIWSMode
} from '@/utils/openaiWsMode'
import type { ModelMapping } from './credentialsBuilder'

export interface ModelRestrictionState {
  mode: 'whitelist' | 'mapping'
  allowedModels: string[]
  modelMappings: ModelMapping[]
}

export interface OpenAIExtraState {
  openaiPassthroughEnabled: boolean
  openaiOAuthResponsesWebSocketV2Mode: OpenAIWSMode
  openaiAPIKeyResponsesWebSocketV2Mode: OpenAIWSMode
  codexCLIOnlyEnabled: boolean
}

export interface AnthropicQuotaControlExtraOptions {
  baseRpm: number | null
  cacheTTLOverrideEnabled: boolean
  cacheTTLOverrideTarget: string
  customBaseUrl: string
  customBaseUrlEnabled: boolean
  maxSessions: number | null
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
}

export interface AntigravityExtraOptions {
  mixedScheduling: boolean
  allowOverages: boolean
}

export interface AnthropicAPIKeyExtraOptions {
  anthropicPassthroughEnabled: boolean
}

export interface OpenAIExtraOptions {
  accountType: AccountType
  codexCLIOnlyEnabled: boolean
  openaiAPIKeyResponsesWebSocketV2Mode: OpenAIWSMode
  openaiOAuthResponsesWebSocketV2Mode: OpenAIWSMode
  openaiPassthroughEnabled: boolean
}

export function createEmptyModelRestrictionState(): ModelRestrictionState {
  return {
    mode: 'whitelist',
    allowedModels: [],
    modelMappings: []
  }
}

export function deriveModelRestrictionStateFromMapping(rawMapping: unknown): ModelRestrictionState {
  if (!rawMapping || typeof rawMapping !== 'object') {
    return createEmptyModelRestrictionState()
  }

  const entries = Object.entries(rawMapping as Record<string, string>)
  if (entries.length === 0) {
    return createEmptyModelRestrictionState()
  }

  const isWhitelistMode = entries.every(([from, to]) => from === to)
  if (isWhitelistMode) {
    return {
      mode: 'whitelist',
      allowedModels: entries.map(([from]) => from),
      modelMappings: []
    }
  }

  return {
    mode: 'mapping',
    allowedModels: [],
    modelMappings: entries.map(([from, to]) => ({ from, to }))
  }
}

export function deriveAntigravityModelMappings(
  credentials: Record<string, unknown> | undefined
): ModelMapping[] {
  const rawMapping = credentials?.model_mapping
  if (rawMapping && typeof rawMapping === 'object') {
    return Object.entries(rawMapping as Record<string, string>).map(([from, to]) => ({
      from,
      to
    }))
  }

  const rawWhitelist = credentials?.model_whitelist
  if (Array.isArray(rawWhitelist) && rawWhitelist.length > 0) {
    return rawWhitelist
      .map((value) => String(value).trim())
      .filter((value) => value.length > 0)
      .map((model) => ({ from: model, to: model }))
  }

  return []
}

export function deriveOpenAIExtraState(
  accountType: AccountType,
  extra: Record<string, unknown> | undefined
): OpenAIExtraState {
  return {
    openaiPassthroughEnabled:
      extra?.openai_passthrough === true || extra?.openai_oauth_passthrough === true,
    openaiOAuthResponsesWebSocketV2Mode: resolveOpenAIWSModeFromExtra(extra, {
      modeKey: 'openai_oauth_responses_websockets_v2_mode',
      enabledKey: 'openai_oauth_responses_websockets_v2_enabled',
      fallbackEnabledKeys: ['responses_websockets_v2_enabled', 'openai_ws_enabled'],
      defaultMode: OPENAI_WS_MODE_OFF
    }),
    openaiAPIKeyResponsesWebSocketV2Mode: resolveOpenAIWSModeFromExtra(extra, {
      modeKey: 'openai_apikey_responses_websockets_v2_mode',
      enabledKey: 'openai_apikey_responses_websockets_v2_enabled',
      fallbackEnabledKeys: ['responses_websockets_v2_enabled', 'openai_ws_enabled'],
      defaultMode: OPENAI_WS_MODE_OFF
    }),
    codexCLIOnlyEnabled: accountType === 'oauth' && extra?.codex_cli_only === true
  }
}

export function buildUpdatedOpenAIExtra(
  currentExtra: Record<string, unknown>,
  options: OpenAIExtraOptions
): Record<string, unknown> {
  const nextExtra: Record<string, unknown> = { ...currentExtra }
  const hadCodexCLIOnlyEnabled = currentExtra.codex_cli_only === true

  if (options.accountType === 'oauth') {
    nextExtra.openai_oauth_responses_websockets_v2_mode =
      options.openaiOAuthResponsesWebSocketV2Mode
    nextExtra.openai_oauth_responses_websockets_v2_enabled = isOpenAIWSModeEnabled(
      options.openaiOAuthResponsesWebSocketV2Mode
    )
  } else if (options.accountType === 'apikey') {
    nextExtra.openai_apikey_responses_websockets_v2_mode =
      options.openaiAPIKeyResponsesWebSocketV2Mode
    nextExtra.openai_apikey_responses_websockets_v2_enabled = isOpenAIWSModeEnabled(
      options.openaiAPIKeyResponsesWebSocketV2Mode
    )
  }

  delete nextExtra.responses_websockets_v2_enabled
  delete nextExtra.openai_ws_enabled

  if (options.openaiPassthroughEnabled) {
    nextExtra.openai_passthrough = true
  } else {
    delete nextExtra.openai_passthrough
    delete nextExtra.openai_oauth_passthrough
  }

  if (options.accountType === 'oauth') {
    if (options.codexCLIOnlyEnabled) {
      nextExtra.codex_cli_only = true
    } else if (hadCodexCLIOnlyEnabled) {
      nextExtra.codex_cli_only = false
    } else {
      delete nextExtra.codex_cli_only
    }
  }

  return nextExtra
}

export function buildUpdatedAntigravityExtra(
  currentExtra: Record<string, unknown>,
  options: AntigravityExtraOptions
): Record<string, unknown> {
  const nextExtra: Record<string, unknown> = { ...currentExtra }

  if (options.mixedScheduling) {
    nextExtra.mixed_scheduling = true
  } else {
    delete nextExtra.mixed_scheduling
  }

  if (options.allowOverages) {
    nextExtra.allow_overages = true
  } else {
    delete nextExtra.allow_overages
  }

  return nextExtra
}

export function buildUpdatedAnthropicQuotaControlExtra(
  currentExtra: Record<string, unknown>,
  options: AnthropicQuotaControlExtraOptions
): Record<string, unknown> {
  const nextExtra: Record<string, unknown> = { ...currentExtra }

  if (options.windowCostEnabled && options.windowCostLimit != null && options.windowCostLimit > 0) {
    nextExtra.window_cost_limit = options.windowCostLimit
    nextExtra.window_cost_sticky_reserve = options.windowCostStickyReserve ?? 10
  } else {
    delete nextExtra.window_cost_limit
    delete nextExtra.window_cost_sticky_reserve
  }

  if (options.sessionLimitEnabled && options.maxSessions != null && options.maxSessions > 0) {
    nextExtra.max_sessions = options.maxSessions
    nextExtra.session_idle_timeout_minutes = options.sessionIdleTimeout ?? 5
  } else {
    delete nextExtra.max_sessions
    delete nextExtra.session_idle_timeout_minutes
  }

  if (options.rpmLimitEnabled) {
    const DEFAULT_BASE_RPM = 15
    nextExtra.base_rpm = options.baseRpm != null && options.baseRpm > 0 ? options.baseRpm : DEFAULT_BASE_RPM
    nextExtra.rpm_strategy = options.rpmStrategy
    if (options.rpmStickyBuffer != null && options.rpmStickyBuffer > 0) {
      nextExtra.rpm_sticky_buffer = options.rpmStickyBuffer
    } else {
      delete nextExtra.rpm_sticky_buffer
    }
  } else {
    delete nextExtra.base_rpm
    delete nextExtra.rpm_strategy
    delete nextExtra.rpm_sticky_buffer
  }

  if (options.userMsgQueueMode) {
    nextExtra.user_msg_queue_mode = options.userMsgQueueMode
  } else {
    delete nextExtra.user_msg_queue_mode
  }
  delete nextExtra.user_msg_queue_enabled

  if (options.tlsFingerprintEnabled) {
    nextExtra.enable_tls_fingerprint = true
    if (options.tlsFingerprintProfileId) {
      nextExtra.tls_fingerprint_profile_id = options.tlsFingerprintProfileId
    } else {
      delete nextExtra.tls_fingerprint_profile_id
    }
  } else {
    delete nextExtra.enable_tls_fingerprint
    delete nextExtra.tls_fingerprint_profile_id
  }

  if (options.sessionIdMaskingEnabled) {
    nextExtra.session_id_masking_enabled = true
  } else {
    delete nextExtra.session_id_masking_enabled
  }

  if (options.cacheTTLOverrideEnabled) {
    nextExtra.cache_ttl_override_enabled = true
    nextExtra.cache_ttl_override_target = options.cacheTTLOverrideTarget
  } else {
    delete nextExtra.cache_ttl_override_enabled
    delete nextExtra.cache_ttl_override_target
  }

  if (options.customBaseUrlEnabled && options.customBaseUrl.trim()) {
    nextExtra.custom_base_url_enabled = true
    nextExtra.custom_base_url = options.customBaseUrl.trim()
  } else {
    delete nextExtra.custom_base_url_enabled
    delete nextExtra.custom_base_url
  }

  return nextExtra
}

export function buildUpdatedAnthropicAPIKeyExtra(
  currentExtra: Record<string, unknown>,
  options: AnthropicAPIKeyExtraOptions
): Record<string, unknown> {
  const nextExtra: Record<string, unknown> = { ...currentExtra }

  if (options.anthropicPassthroughEnabled) {
    nextExtra.anthropic_passthrough = true
  } else {
    delete nextExtra.anthropic_passthrough
  }

  return nextExtra
}
