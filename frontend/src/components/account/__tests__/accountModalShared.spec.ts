import { describe, expect, it } from 'vitest'
import {
  buildAccountOpenAIWSModeOptions,
  buildAccountQuotaExtra,
  buildAccountTempUnschedPresets,
  buildAccountUmqModeOptions,
  buildMixedChannelDetails,
  needsMixedChannelCheck,
  resolveAccountApiKeyHint,
  resolveAccountBaseUrlHint,
  resolveCreateAccountOAuthStepTitle,
  resolveMixedChannelWarningMessage
} from '../accountModalShared'

const t = (key: string, values?: Record<string, unknown>) =>
  values ? `${key}:${JSON.stringify(values)}` : key

describe('accountModalShared', () => {
  it('resolves platform-specific hints and oauth titles', () => {
    expect(resolveAccountBaseUrlHint('openai', t)).toBe('admin.accounts.openai.baseUrlHint')
    expect(resolveAccountBaseUrlHint('gemini', t)).toBe('admin.accounts.gemini.baseUrlHint')
    expect(resolveAccountBaseUrlHint('anthropic', t)).toBe('admin.accounts.baseUrlHint')

    expect(resolveAccountApiKeyHint('sora', t)).toBe('admin.accounts.openai.apiKeyHint')
    expect(resolveAccountApiKeyHint('gemini', t)).toBe('admin.accounts.gemini.apiKeyHint')
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
})
