import { describe, expect, it } from 'vitest'
import {
  buildCreateAccountSharedPayload,
  buildCreateBatchAccountName,
  buildCreateAnthropicExtra,
  buildCreateAnthropicQuotaControlExtra,
  buildCreateAntigravityExtra,
  buildCreateOpenAIExtra,
  buildCreateOAuthAccountPayload,
  buildCreateSoraOAuthCredentials,
  buildCreateSoraExtra,
  resolveBatchCreateOutcome,
  resolveCreateAccountGeminiSelectedTier,
  resolveCreateAccountOAuthFlow
} from '../createAccountModalHelpers'

const t = (key: string, values?: Record<string, unknown>) =>
  values ? `${key}:${JSON.stringify(values)}` : key

describe('createAccountModalHelpers', () => {
  it('resolves oauth flow and gemini selected tier', () => {
    expect(
      resolveCreateAccountOAuthFlow({
        accountCategory: 'oauth-based',
        antigravityAccountType: 'oauth',
        platform: 'anthropic'
      })
    ).toBe(true)

    expect(
      resolveCreateAccountOAuthFlow({
        accountCategory: 'oauth-based',
        antigravityAccountType: 'upstream',
        platform: 'antigravity'
      })
    ).toBe(false)

    expect(
      resolveCreateAccountGeminiSelectedTier({
        accountCategory: 'oauth-based',
        geminiOAuthType: 'google_one',
        geminiTierAIStudio: 'aistudio_free',
        geminiTierGcp: 'gcp_standard',
        geminiTierGoogleOne: 'google_ai_pro',
        platform: 'gemini'
      })
    ).toBe('google_ai_pro')
  })

  it('builds create extra payloads', () => {
    expect(
      buildCreateAntigravityExtra({
        allowOverages: true,
        mixedScheduling: true
      })
    ).toEqual({
      allow_overages: true,
      mixed_scheduling: true
    })

    expect(
      buildCreateOpenAIExtra({
        accountCategory: 'oauth-based',
        codexCLIOnlyEnabled: true,
        openaiAPIKeyResponsesWebSocketV2Mode: 'off',
        openaiOAuthResponsesWebSocketV2Mode: 'passthrough',
        openaiPassthroughEnabled: true,
        platform: 'openai'
      })
    ).toMatchObject({
      openai_oauth_responses_websockets_v2_mode: 'passthrough',
      openai_oauth_responses_websockets_v2_enabled: true,
      openai_passthrough: true,
      codex_cli_only: true
    })

    expect(
      buildCreateAnthropicExtra({
        accountCategory: 'apikey',
        anthropicPassthroughEnabled: true,
        base: { foo: 'bar' },
        platform: 'anthropic'
      })
    ).toEqual({
      foo: 'bar',
      anthropic_passthrough: true
    })

    expect(
      buildCreateAnthropicQuotaControlExtra({
        baseExtra: { source: 'oauth' },
        baseRpm: null,
        cacheTTLOverrideEnabled: true,
        cacheTTLOverrideTarget: '10m',
        customBaseUrl: 'https://relay.example.com',
        customBaseUrlEnabled: true,
        maxSessions: 3,
        rpmLimitEnabled: true,
        rpmStickyBuffer: 7,
        rpmStrategy: 'sticky_exempt',
        sessionIdMaskingEnabled: true,
        sessionIdleTimeout: 9,
        sessionLimitEnabled: true,
        tlsFingerprintEnabled: true,
        tlsFingerprintProfileId: 42,
        userMsgQueueMode: 'serialize',
        windowCostEnabled: true,
        windowCostLimit: 12.5,
        windowCostStickyReserve: 4
      })
    ).toEqual({
      source: 'oauth',
      window_cost_limit: 12.5,
      window_cost_sticky_reserve: 4,
      max_sessions: 3,
      session_idle_timeout_minutes: 9,
      base_rpm: 15,
      rpm_strategy: 'sticky_exempt',
      rpm_sticky_buffer: 7,
      user_msg_queue_mode: 'serialize',
      enable_tls_fingerprint: true,
      tls_fingerprint_profile_id: 42,
      session_id_masking_enabled: true,
      cache_ttl_override_enabled: true,
      cache_ttl_override_target: '10m',
      custom_base_url_enabled: true,
      custom_base_url: 'https://relay.example.com'
    })
  })

  it('removes openai-only flags from sora extra', () => {
    expect(
      buildCreateSoraExtra(
        {
          openai_passthrough: true,
          codex_cli_only: true,
          openai_oauth_responses_websockets_v2_mode: 'passthrough',
          custom: 'value'
        },
        123
      )
    ).toEqual({
      custom: 'value',
      linked_openai_account_id: '123'
    })
  })

  it('builds shared payloads, derived names, and batch outcomes', () => {
    const common = buildCreateAccountSharedPayload({
      autoPauseOnExpired: true,
      concurrency: 5,
      expiresAt: 123,
      groupIds: [1, 2],
      loadFactor: 2,
      notes: 'note',
      priority: 9,
      proxyId: 3,
      rateMultiplier: 1.5
    })

    expect(common).toEqual({
      auto_pause_on_expired: true,
      concurrency: 5,
      expires_at: 123,
      group_ids: [1, 2],
      load_factor: 2,
      notes: 'note',
      priority: 9,
      proxy_id: 3,
      rate_multiplier: 1.5
    })

    expect(buildCreateBatchAccountName('Demo', 1, 3)).toBe('Demo #2')
    expect(buildCreateBatchAccountName('', 0, 1, 'Fallback')).toBe('Fallback')
    expect(buildCreateBatchAccountName('Demo', 0, 2, undefined, '(Sora)')).toBe('Demo #1 (Sora)')

    expect(
      buildCreateSoraOAuthCredentials({
        access_token: 'at',
        refresh_token: 'rt',
        client_id: 'client',
        expires_at: 10
      })
    ).toEqual({
      access_token: 'at',
      refresh_token: 'rt',
      client_id: 'client',
      expires_at: 10
    })

    expect(
      buildCreateOAuthAccountPayload({
        common,
        name: 'Demo',
        platform: 'openai',
        type: 'oauth',
        credentials: { token: 'x' },
        extra: { foo: 'bar' }
      })
    ).toMatchObject({
      name: 'Demo',
      platform: 'openai',
      type: 'oauth',
      credentials: { token: 'x' },
      extra: { foo: 'bar' },
      notes: 'note'
    })

    expect(resolveBatchCreateOutcome({ failedCount: 0, successCount: 2, t })).toEqual({
      type: 'success',
      message: 'admin.accounts.oauth.batchSuccess:{"count":2}',
      shouldClose: true,
      shouldEmitCreated: true
    })
    expect(resolveBatchCreateOutcome({ failedCount: 1, successCount: 1, t })).toEqual({
      type: 'warning',
      message: 'admin.accounts.oauth.batchPartialSuccess:{"success":1,"failed":1}',
      shouldClose: false,
      shouldEmitCreated: true
    })
    expect(resolveBatchCreateOutcome({ failedCount: 2, successCount: 0, t })).toEqual({
      type: 'error',
      message: 'admin.accounts.oauth.batchFailed',
      shouldClose: false,
      shouldEmitCreated: false
    })
  })
})
