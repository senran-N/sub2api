import { describe, expect, it } from 'vitest'
import {
  buildOpenAICompatibleBaseUrlPresets,
  buildAccountOpenAIWSModeOptions,
  buildAccountQuotaExtra,
  buildAccountTempUnschedPresets,
  buildEditAccountBasePayload,
  buildAccountUmqModeOptions,
  buildMixedChannelDetails,
  createDefaultCreateAccountForm,
  createDefaultEditAccountForm,
  hydrateEditAccountForm,
  needsMixedChannelCheck,
  resetCreateAccountForm,
  resolveAccountApiKeyHint,
  resolveAccountApiKeyPlaceholder,
  resolveAccountBaseUrlHint,
  resolveAccountBaseUrlPlaceholder,
  resolveCreateAccountOAuthStepTitle,
  resolveMixedChannelWarningMessage
} from '../accountModalShared'

const t = (key: string, values?: Record<string, unknown>) =>
  values ? `${key}:${JSON.stringify(values)}` : key

describe('accountModalShared', () => {
  it('resolves platform-specific hints, placeholders, presets, and oauth titles', () => {
    expect(resolveAccountBaseUrlHint('openai', t)).toBe('admin.accounts.openai.baseUrlHint')
    expect(resolveAccountBaseUrlHint('gemini', t)).toBe('admin.accounts.gemini.baseUrlHint')
    expect(resolveAccountBaseUrlHint('anthropic', t)).toBe('admin.accounts.baseUrlHint')
    expect(resolveAccountBaseUrlPlaceholder('openai', t)).toBe(
      'admin.accounts.openai.baseUrlPlaceholder'
    )
    expect(resolveAccountBaseUrlPlaceholder('antigravity', t)).toBe(
      'https://cloudcode-pa.googleapis.com'
    )

    expect(resolveAccountApiKeyHint('openai', t)).toBe('admin.accounts.openai.apiKeyHint')
    expect(resolveAccountApiKeyHint('gemini', t)).toBe('admin.accounts.gemini.apiKeyHint')
    expect(resolveAccountApiKeyPlaceholder('openai', t)).toBe(
      'admin.accounts.openai.apiKeyPlaceholder'
    )
    expect(resolveAccountApiKeyPlaceholder('gemini', t)).toBe('AIza...')
    expect(buildOpenAICompatibleBaseUrlPresets(t)).toEqual([
      {
        label: 'admin.accounts.openai.baseUrlPresets.official',
        value: 'https://api.openai.com'
      },
      {
        label: 'admin.accounts.openai.baseUrlPresets.xai',
        value: 'https://api.x.ai'
      }
    ])
    expect(resolveCreateAccountOAuthStepTitle('antigravity', t)).toBe(
      'admin.accounts.oauth.antigravity.title'
    )
  })

  it('builds modal option lists', () => {
    expect(buildAccountUmqModeOptions(t).map((item) => item.value)).toEqual([
      '',
      'throttle',
      'serialize'
    ])
    expect(buildAccountOpenAIWSModeOptions(t).map((item) => item.value)).toEqual([
      'off',
      'ctx_pool',
      'passthrough'
    ])
    expect(buildAccountTempUnschedPresets(t)).toHaveLength(3)
  })

  it('builds mixed channel details and messages', () => {
    const details = buildMixedChannelDetails({
      has_risk: true,
      message: 'warning',
      details: {
        group_name: 'Alpha',
        current_platform: 'anthropic',
        other_platform: 'antigravity'
      }
    } as any)

    expect(details).toEqual({
      groupName: 'Alpha',
      currentPlatform: 'anthropic',
      otherPlatform: 'antigravity'
    })
    expect(needsMixedChannelCheck('anthropic')).toBe(true)
    expect(needsMixedChannelCheck('openai')).toBe(false)
    expect(
      resolveMixedChannelWarningMessage({
        details,
        rawMessage: 'raw warning',
        t
      })
    ).toContain('admin.accounts.mixedChannelWarning')
    expect(
      resolveMixedChannelWarningMessage({
        details: null,
        rawMessage: 'raw warning',
        t
      })
    ).toBe('raw warning')
  })

  it('applies quota fields onto extra state and removes stale reset keys', () => {
    expect(
      buildAccountQuotaExtra(
        {
          keep: true,
          quota_limit: 1,
          quota_daily_reset_mode: 'fixed',
          quota_daily_reset_hour: 6,
          quota_reset_timezone: 'Asia/Shanghai'
        },
        {
          dailyResetHour: null,
          dailyResetMode: null,
          quotaDailyLimit: null,
          quotaLimit: 8,
          quotaWeeklyLimit: 16,
          resetTimezone: null,
          weeklyResetDay: 5,
          weeklyResetHour: 9,
          weeklyResetMode: 'fixed'
        }
      )
    ).toEqual({
      keep: true,
      quota_limit: 8,
      quota_weekly_limit: 16,
      quota_weekly_reset_mode: 'fixed',
      quota_weekly_reset_day: 5,
      quota_weekly_reset_hour: 9,
      quota_reset_timezone: 'UTC'
    })
  })

  it('builds, resets, hydrates, and normalizes shared account form state', () => {
    const createForm = createDefaultCreateAccountForm()
    createForm.name = 'Changed'
    resetCreateAccountForm(createForm)
    expect(createForm).toEqual({
      name: '',
      notes: '',
      platform: 'anthropic',
      type: 'oauth',
      credentials: {},
      proxy_id: null,
      concurrency: 10,
      load_factor: null,
      priority: 1,
      rate_multiplier: 1,
      group_ids: [],
      expires_at: null
    })

    const editForm = createDefaultEditAccountForm()
    hydrateEditAccountForm(editForm, {
      name: 'Demo',
      notes: null,
      proxy_id: null,
      concurrency: 4,
      load_factor: null,
      priority: 8,
      rate_multiplier: null,
      status: 'error',
      group_ids: [1, 2],
      expires_at: 123
    } as any)

    expect(editForm).toEqual({
      name: 'Demo',
      notes: '',
      proxy_id: null,
      concurrency: 4,
      load_factor: null,
      priority: 8,
      rate_multiplier: 1,
      status: 'error',
      group_ids: [1, 2],
      expires_at: 123
    })

    expect(
      buildEditAccountBasePayload(
        {
          ...editForm,
          proxy_id: null,
          expires_at: null,
          load_factor: -1
        },
        true
      )
    ).toEqual({
      name: 'Demo',
      notes: '',
      proxy_id: 0,
      concurrency: 4,
      load_factor: 0,
      priority: 8,
      rate_multiplier: 1,
      status: 'error',
      group_ids: [1, 2],
      expires_at: 0,
      auto_pause_on_expired: true
    })
  })
})
