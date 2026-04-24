import { describe, expect, it } from 'vitest'
import type { Account } from '@/types'
import {
  buildBulkAccountMutationPayload,
  buildCreateAccountMutationPayload,
  buildEditAccountMutationPayload,
  type BuildBulkAccountMutationPayloadOptions,
  type BuildEditAccountMutationPayloadOptions
} from '../accountMutationPayload'

const quotaOptions = {
  dailyResetHour: null,
  dailyResetMode: null,
  quotaDailyLimit: null,
  quotaLimit: null,
  quotaWeeklyLimit: null,
  quotaNotifyDailyEnabled: null,
  quotaNotifyDailyThreshold: null,
  quotaNotifyDailyThresholdType: null,
  quotaNotifyWeeklyEnabled: null,
  quotaNotifyWeeklyThreshold: null,
  quotaNotifyWeeklyThresholdType: null,
  quotaNotifyTotalEnabled: null,
  quotaNotifyTotalThreshold: null,
  quotaNotifyTotalThresholdType: null,
  resetTimezone: null,
  weeklyResetDay: null,
  weeklyResetHour: null,
  weeklyResetMode: null
} satisfies BuildEditAccountMutationPayloadOptions['quota']

function createBulkOptions(
  overrides: Partial<BuildBulkAccountMutationPayloadOptions> = {}
): BuildBulkAccountMutationPayloadOptions {
  return {
    baseUrl: {
      enabled: false,
      value: ''
    },
    customErrorCodes: {
      enabled: false,
      selectedErrorCodes: []
    },
    groups: {
      enabled: false,
      groupIds: []
    },
    interceptWarmup: {
      enabled: false,
      value: false
    },
    loadFactor: {
      enabled: false,
      value: null
    },
    modelRestriction: {
      allowedModels: [],
      disabledByOpenAIPassthrough: false,
      enabled: false,
      mode: 'whitelist',
      modelMappings: []
    },
    openAI: {
      passthroughEnabled: false,
      passthroughValue: false,
      wsModeEnabled: false,
      wsModeValue: 'off'
    },
    proxy: {
      enabled: false,
      proxyId: null
    },
    rpmLimit: {
      baseRpm: null,
      enabled: false,
      rpmEnabled: false,
      stickyBuffer: null,
      strategy: 'tiered'
    },
    scalars: {
      enableConcurrency: false,
      enablePriority: false,
      enableRateMultiplier: false,
      enableStatus: false
    },
    userMsgQueueMode: null,
    ...overrides
  }
}

function createAccount(overrides: Partial<Account>): Account {
  return {
    id: 1,
    name: 'Test account',
    platform: 'openai',
    type: 'apikey',
    credentials: {},
    extra: {},
    status: 'active',
    ...overrides
  } as Account
}

function createEditOptions(
  overrides: Partial<BuildEditAccountMutationPayloadOptions> = {}
): BuildEditAccountMutationPayloadOptions {
  return {
    account: createAccount({}),
    anthropicAPIKeyExtra: {
      anthropicPassthroughEnabled: false
    },
    anthropicQuotaExtra: {
      baseRpm: null,
      cacheTTLOverrideEnabled: false,
      cacheTTLOverrideTarget: '5m',
      customBaseUrl: '',
      customBaseUrlEnabled: false,
      maxSessions: null,
      rpmLimitEnabled: false,
      rpmStickyBuffer: null,
      rpmStrategy: 'tiered',
      sessionIdMaskingEnabled: false,
      sessionIdleTimeout: null,
      sessionLimitEnabled: false,
      tlsFingerprintEnabled: false,
      tlsFingerprintProfileId: null,
      userMsgQueueMode: '',
      windowCostEnabled: false,
      windowCostLimit: null,
      windowCostStickyReserve: null
    },
    antigravity: {
      allowOverages: false,
      mixedScheduling: false,
      modelMappings: []
    },
    basePayload: {
      name: 'Test account',
      status: 'active'
    },
    bedrock: {
      accessKeyId: '',
      allowedModels: [],
      apiKeyInput: '',
      forceGlobal: false,
      isApiKeyMode: false,
      mode: 'whitelist',
      modelMappings: [],
      poolModeEnabled: false,
      poolModeRetryCount: 3,
      region: '',
      secretAccessKey: '',
      sessionToken: ''
    },
    compatible: {
      allowedModels: [],
      apiKeyInput: '',
      baseUrlInput: '',
      customErrorCodesEnabled: false,
      defaultBaseUrl: 'https://api.openai.com',
      isOpenAIModelRestrictionDisabled: false,
      mode: 'whitelist',
      modelMappings: [],
      poolModeEnabled: false,
      poolModeRetryCount: 3,
      selectedErrorCodes: []
    },
    currentCredentials: {},
    currentExtra: {},
    openAIExtra: {
      accountType: 'apikey',
      codexCLIOnlyEnabled: false,
      openaiAPIKeyResponsesWebSocketV2Mode: 'off',
      openaiOAuthResponsesWebSocketV2Mode: 'off',
      openaiPassthroughEnabled: false
    },
    quota: quotaOptions,
    sessionTokenInput: '',
    sharedCredentials: {
      interceptWarmupRequests: false,
      tempUnschedEnabled: false,
      tempUnschedRules: []
    },
    ...overrides
  }
}

describe('accountMutationPayload', () => {
  it('keeps empty bulk whitelist as an explicit empty model mapping', () => {
    expect(
      buildBulkAccountMutationPayload(
        createBulkOptions({
          modelRestriction: {
            allowedModels: [],
            disabledByOpenAIPassthrough: false,
            enabled: true,
            mode: 'whitelist',
            modelMappings: []
          }
        })
      )
    ).toEqual({
      credentials: {
        model_mapping: {}
      }
    })
  })

  it('builds bulk OpenAI passthrough and websocket extra payloads', () => {
    expect(
      buildBulkAccountMutationPayload(
        createBulkOptions({
          openAI: {
            passthroughEnabled: true,
            passthroughValue: false,
            wsModeEnabled: false,
            wsModeValue: 'off'
          }
        })
      )
    ).toEqual({
      extra: {
        openai_passthrough: false,
        openai_oauth_passthrough: false
      }
    })

    expect(
      buildBulkAccountMutationPayload(
        createBulkOptions({
          openAI: {
            passthroughEnabled: true,
            passthroughValue: true,
            wsModeEnabled: true,
            wsModeValue: 'ctx_pool'
          }
        })
      )
    ).toEqual({
      extra: {
        openai_passthrough: true,
        openai_oauth_responses_websockets_v2_mode: 'ctx_pool',
        openai_oauth_responses_websockets_v2_enabled: true,
        responses_websockets_v2_enabled: false,
        openai_ws_enabled: false
      }
    })
  })

  it('uses explicit bulk RPM reset sentinel values', () => {
    expect(
      buildBulkAccountMutationPayload(
        createBulkOptions({
          rpmLimit: {
            baseRpm: null,
            enabled: true,
            rpmEnabled: false,
            stickyBuffer: null,
            strategy: 'tiered'
          }
        })
      )
    ).toEqual({
      extra: {
        base_rpm: 0,
        rpm_strategy: '',
        rpm_sticky_buffer: 0
      }
    })
  })

  it('preserves an existing OpenAI API key mapping while passthrough disables mapping edits', () => {
    const currentCredentials = {
      api_key: 'existing-api-key',
      model_mapping: {
        kept: 'kept'
      }
    }
    const result = buildEditAccountMutationPayload(
      createEditOptions({
        account: createAccount({
          credentials: currentCredentials,
          platform: 'openai',
          type: 'apikey'
        }),
        compatible: {
          ...createEditOptions().compatible,
          allowedModels: ['gpt-5.4'],
          isOpenAIModelRestrictionDisabled: true
        },
        currentCredentials,
        openAIExtra: {
          ...createEditOptions().openAIExtra,
          accountType: 'apikey',
          openaiPassthroughEnabled: true
        }
      })
    )

    expect(result.error).toBeUndefined()
    expect(result.payload?.credentials).toMatchObject({
      api_key: 'existing-api-key',
      base_url: 'https://api.openai.com',
      model_mapping: {
        kept: 'kept'
      }
    })
  })

  it('returns Grok session token validation errors before building an edit payload', () => {
    const account = createAccount({
      credentials: {},
      platform: 'grok',
      type: 'session'
    })

    expect(
      buildEditAccountMutationPayload(
        createEditOptions({
          account,
          currentCredentials: {},
          sessionTokenInput: ''
        })
      ).error
    ).toBe('grok_session_token_required')

    expect(
      buildEditAccountMutationPayload(
        createEditOptions({
          account,
          currentCredentials: {},
          sessionTokenInput: 'abc'
        })
      ).error
    ).toBe('grok_session_token_invalid')
  })

  it('applies create quota fields only for profile-backed quota accounts', () => {
    const common = {
      auto_pause_on_expired: true,
      concurrency: 1,
      group_ids: [],
      priority: 1,
      proxy_id: null,
      rate_multiplier: 1
    }
    const quota = {
      ...quotaOptions,
      quotaLimit: 50
    }

    expect(
      buildCreateAccountMutationPayload({
        common,
        credentials: {
          refresh_token: 'rt'
        },
        extra: {
          source: 'oauth'
        },
        name: 'OpenAI OAuth',
        platform: 'openai',
        quota,
        type: 'oauth'
      }).extra
    ).toEqual({
      source: 'oauth'
    })

    expect(
      buildCreateAccountMutationPayload({
        common,
        credentials: {
          api_key: 'demo-api-key'
        },
        extra: {
          source: 'apikey'
        },
        name: 'OpenAI API Key',
        platform: 'openai',
        quota,
        type: 'apikey'
      }).extra
    ).toEqual({
      source: 'apikey',
      quota_limit: 50
    })
  })
})
