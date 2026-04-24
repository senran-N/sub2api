import { describe, expect, it } from 'vitest'
import {
  buildEditableBedrockCredentials,
  buildEditableCompatibleCredentials,
  buildUpdatedAnthropicAPIKeyExtra,
  buildUpdatedAnthropicQuotaControlExtra,
  buildUpdatedAntigravityExtra,
  buildUpdatedOpenAIExtra,
  createEmptyModelRestrictionState,
  deriveAntigravityModelMappings,
  deriveModelRestrictionStateFromMapping,
  deriveOpenAIExtraState
} from '../editAccountModalHelpers'

describe('editAccountModalHelpers', () => {
  it('derives model restriction state from mapping data', () => {
    expect(createEmptyModelRestrictionState()).toEqual({
      mode: 'whitelist',
      allowedModels: [],
      modelMappings: []
    })

    expect(
      deriveModelRestrictionStateFromMapping({
        'gpt-5': 'gpt-5',
        'gpt-4o': 'gpt-4o'
      })
    ).toEqual({
      mode: 'whitelist',
      allowedModels: ['gpt-5', 'gpt-4o'],
      modelMappings: []
    })

    expect(
      deriveModelRestrictionStateFromMapping({
        'gpt-5': 'gpt-5-mini'
      })
    ).toEqual({
      mode: 'mapping',
      allowedModels: [],
      modelMappings: [{ from: 'gpt-5', to: 'gpt-5-mini' }]
    })
  })

  it('derives antigravity mapping state from mapping or whitelist values', () => {
    expect(
      deriveAntigravityModelMappings({
        model_mapping: {
          claude: 'claude-pro'
        }
      })
    ).toEqual([{ from: 'claude', to: 'claude-pro' }])

    expect(
      deriveAntigravityModelMappings({
        model_whitelist: [' claude ', 'sonnet']
      })
    ).toEqual([
      { from: 'claude', to: 'claude' },
      { from: 'sonnet', to: 'sonnet' }
    ])
  })

  it('derives and updates openai extra state', () => {
    expect(
      deriveOpenAIExtraState('oauth', {
        openai_passthrough: true,
        openai_oauth_responses_websockets_v2_mode: 'passthrough',
        codex_cli_only: true
      })
    ).toEqual({
      openaiPassthroughEnabled: true,
      openaiOAuthResponsesWebSocketV2Mode: 'passthrough',
      openaiAPIKeyResponsesWebSocketV2Mode: 'off',
      codexCLIOnlyEnabled: true
    })

    expect(
      buildUpdatedOpenAIExtra(
        {
          codex_cli_only: true,
          responses_websockets_v2_enabled: true
        },
        {
          accountType: 'oauth',
          codexCLIOnlyEnabled: false,
          openaiAPIKeyResponsesWebSocketV2Mode: 'off',
          openaiOAuthResponsesWebSocketV2Mode: 'off',
          openaiPassthroughEnabled: false
        }
      )
    ).toEqual({
      openai_oauth_responses_websockets_v2_mode: 'off',
      openai_oauth_responses_websockets_v2_enabled: false,
      codex_cli_only: false
    })
  })

  it('updates antigravity and anthropic extra payloads', () => {
    expect(
      buildUpdatedAntigravityExtra(
        { stale: true, mixed_scheduling: true },
        { mixedScheduling: false, allowOverages: true }
      )
    ).toEqual({
      stale: true,
      allow_overages: true
    })

    expect(
      buildUpdatedAnthropicAPIKeyExtra(
        { anthropic_passthrough: true, stale: true },
        { anthropicPassthroughEnabled: false }
      )
    ).toEqual({
      stale: true
    })

    expect(
      buildUpdatedAnthropicQuotaControlExtra(
        {
          stale: true,
          user_msg_queue_enabled: true
        },
        {
          baseRpm: null,
          cacheTTLOverrideEnabled: true,
          cacheTTLOverrideTarget: '10m',
          customBaseUrl: 'https://relay.example.com',
          customBaseUrlEnabled: true,
          maxSessions: 3,
          rpmLimitEnabled: true,
          rpmStickyBuffer: 4,
          rpmStrategy: 'sticky_exempt',
          sessionIdMaskingEnabled: true,
          sessionIdleTimeout: 8,
          sessionLimitEnabled: true,
          tlsFingerprintEnabled: true,
          tlsFingerprintProfileId: 9,
          userMsgQueueMode: 'serialize',
          windowCostEnabled: true,
          windowCostLimit: 20,
          windowCostStickyReserve: 6
        }
      )
    ).toEqual({
      stale: true,
      window_cost_limit: 20,
      window_cost_sticky_reserve: 6,
      max_sessions: 3,
      session_idle_timeout_minutes: 8,
      base_rpm: 15,
      rpm_strategy: 'sticky_exempt',
      rpm_sticky_buffer: 4,
      user_msg_queue_mode: 'serialize',
      enable_tls_fingerprint: true,
      tls_fingerprint_profile_id: 9,
      session_id_masking_enabled: true,
      cache_ttl_override_enabled: true,
      cache_ttl_override_target: '10m',
      custom_base_url_enabled: true,
      custom_base_url: 'https://relay.example.com'
    })
  })

  it('builds editable compatible credentials without dropping the existing API key', () => {
    const result = buildEditableCompatibleCredentials({
      allowedModels: [],
      apiKeyInput: '',
      baseUrlInput: ' https://relay.example.com ',
      currentCredentials: {
        api_key: 'sk-existing',
        base_url: 'https://api.openai.com',
        model_mapping: { stale: 'stale' },
        pool_mode_retry_count: 1
      },
      customErrorCodesEnabled: true,
      defaultBaseUrl: 'https://api.openai.com',
      mode: 'mapping',
      modelMappings: [{ from: 'gpt-alias', to: 'gpt-5.4' }],
      poolModeEnabled: true,
      poolModeRetryCount: 99,
      preserveModelMappingWhenDisabled: true,
      selectedErrorCodes: [429, 503],
      shouldApplyModelMapping: true
    })

    expect(result.error).toBeUndefined()
    expect(result.credentials).toEqual({
      api_key: 'sk-existing',
      base_url: 'https://relay.example.com',
      model_mapping: {
        'gpt-alias': 'gpt-5.4'
      },
      pool_mode: true,
      pool_mode_retry_count: 10,
      custom_error_codes_enabled: true,
      custom_error_codes: [429, 503]
    })
  })

  it('preserves OpenAI model mapping while passthrough disables mapping edits', () => {
    const result = buildEditableCompatibleCredentials({
      allowedModels: ['gpt-5.4'],
      apiKeyInput: '',
      baseUrlInput: '',
      currentCredentials: {
        api_key: 'sk-existing',
        model_mapping: { kept: 'kept' }
      },
      customErrorCodesEnabled: false,
      defaultBaseUrl: 'https://api.openai.com',
      mode: 'whitelist',
      modelMappings: [],
      poolModeEnabled: false,
      poolModeRetryCount: 3,
      preserveModelMappingWhenDisabled: true,
      selectedErrorCodes: [],
      shouldApplyModelMapping: false
    })

    expect(result.credentials).toMatchObject({
      api_key: 'sk-existing',
      base_url: 'https://api.openai.com',
      model_mapping: { kept: 'kept' }
    })
  })

  it('builds editable Bedrock credentials while preserving blank secret fields', () => {
    expect(
      buildEditableBedrockCredentials({
        accessKeyId: ' AKIA-NEW ',
        allowedModels: ['anthropic.claude-3-sonnet'],
        apiKeyInput: '',
        currentCredentials: {
          auth_mode: 'sigv4',
          aws_secret_access_key: 'old-secret',
          aws_session_token: 'old-session',
          aws_force_global: 'true',
          pool_mode: true
        },
        forceGlobal: false,
        isApiKeyMode: false,
        mode: 'whitelist',
        modelMappings: [],
        poolModeEnabled: false,
        poolModeRetryCount: 3,
        region: ' us-west-2 ',
        secretAccessKey: '',
        sessionToken: ' new-session '
      })
    ).toEqual({
      auth_mode: 'sigv4',
      aws_region: 'us-west-2',
      aws_access_key_id: 'AKIA-NEW',
      aws_secret_access_key: 'old-secret',
      aws_session_token: 'new-session',
      model_mapping: {
        'anthropic.claude-3-sonnet': 'anthropic.claude-3-sonnet'
      }
    })
  })
})
