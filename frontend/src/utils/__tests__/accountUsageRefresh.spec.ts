import { describe, expect, it } from 'vitest'
import {
  ACCOUNT_LIST_REFRESH_SCHEMA,
  ACCOUNT_LIST_REFRESH_DERIVED_SCHEMA,
  ACCOUNT_LIST_REFRESH_RUNTIME_SCHEMA,
  ACCOUNT_STATUS_REFRESH_BASE_SCHEMA,
  ACCOUNT_STATUS_REFRESH_SCHEMA,
  ACCOUNT_STATUS_RUNTIME_EXTENSION_SCHEMA,
  ACCOUNT_USAGE_PROVIDER_EXTENSION_SCHEMA,
  ACCOUNT_USAGE_REFRESH_BASE_SCHEMA,
  ACCOUNT_USAGE_REFRESH_SCHEMA,
  OPENAI_USAGE_REFRESH_SCHEMA,
  OPENAI_USAGE_REFRESH_BASE_SCHEMA,
  OPENAI_USAGE_REFRESH_EXTENSION_SCHEMA,
  buildAccountListRefreshKey,
  buildAccountStatusRefreshKey,
  buildAccountUsageRefreshKey,
  buildOpenAIUsageRefreshKey
} from '../accountUsageRefresh'

describe('buildOpenAIUsageRefreshKey', () => {
  it('会在 codex 快照变化时生成不同 key', () => {
    const base = {
      id: 1,
      platform: 'openai',
      type: 'oauth',
      updated_at: '2026-03-07T10:00:00Z',
      last_used_at: '2026-03-07T09:59:00Z',
      extra: {
        codex_usage_updated_at: '2026-03-07T10:00:00Z',
        codex_5h_used_percent: 0,
        codex_7d_used_percent: 0
      }
    } as any

    const next = {
      ...base,
      extra: {
        ...base.extra,
        codex_usage_updated_at: '2026-03-07T10:01:00Z',
        codex_5h_used_percent: 100
      }
    }

    expect(buildOpenAIUsageRefreshKey(base)).not.toBe(buildOpenAIUsageRefreshKey(next))
  })

  it('会在 last_used_at 变化时生成不同 key', () => {
    const base = {
      id: 3,
      platform: 'openai',
      type: 'oauth',
      updated_at: '2026-03-07T10:00:00Z',
      last_used_at: '2026-03-07T10:00:00Z',
      extra: {
        codex_usage_updated_at: '2026-03-07T10:00:00Z',
        codex_5h_used_percent: 12,
        codex_7d_used_percent: 24
      }
    } as any

    const next = {
      ...base,
      last_used_at: '2026-03-07T10:02:00Z'
    }

    expect(buildOpenAIUsageRefreshKey(base)).not.toBe(buildOpenAIUsageRefreshKey(next))
  })

  it('非 OpenAI OAuth 账号返回空 key', () => {
    expect(buildOpenAIUsageRefreshKey({
      id: 2,
      platform: 'anthropic',
      type: 'oauth',
      updated_at: '2026-03-07T10:00:00Z',
      last_used_at: '2026-03-07T10:00:00Z',
      extra: {}
    } as any)).toBe('')
  })
})

describe('account refresh keys', () => {
  it('用基础段和扩展段组合状态/用量/列表刷新字段', () => {
    expect(OPENAI_USAGE_REFRESH_SCHEMA.map((entry) => entry.field)).toEqual([
      ...OPENAI_USAGE_REFRESH_BASE_SCHEMA.map((entry) => entry.field),
      ...OPENAI_USAGE_REFRESH_EXTENSION_SCHEMA.map((entry) => entry.field)
    ])
    expect(ACCOUNT_STATUS_REFRESH_SCHEMA.map((entry) => entry.field)).toEqual([
      ...ACCOUNT_STATUS_REFRESH_BASE_SCHEMA.map((entry) => entry.field),
      ...ACCOUNT_STATUS_RUNTIME_EXTENSION_SCHEMA.map((entry) => entry.field)
    ])
    expect(ACCOUNT_USAGE_REFRESH_SCHEMA.map((entry) => entry.field)).toEqual([
      ...ACCOUNT_USAGE_REFRESH_BASE_SCHEMA.map((entry) => entry.field),
      ...ACCOUNT_USAGE_PROVIDER_EXTENSION_SCHEMA.map((entry) => entry.field)
    ])
    expect(ACCOUNT_LIST_REFRESH_SCHEMA.map((entry) => entry.field)).toEqual([
      ...ACCOUNT_LIST_REFRESH_RUNTIME_SCHEMA.map((entry) => entry.field),
      ...ACCOUNT_LIST_REFRESH_DERIVED_SCHEMA.map((entry) => entry.field)
    ])

    expect(ACCOUNT_STATUS_REFRESH_SCHEMA.map((entry) => entry.field)).toContain('model_rate_limits')
    expect(ACCOUNT_USAGE_REFRESH_SCHEMA.map((entry) => entry.field)).toContain('load_code_assist')
    expect(ACCOUNT_LIST_REFRESH_SCHEMA.map((entry) => entry.field)).toEqual([
      'current_concurrency',
      'schedulable',
      'status_key',
      'usage_key'
    ])
  })

  it('会在模型限流状态变化时刷新状态 key', () => {
    const base = {
      id: 7,
      updated_at: '2026-03-07T10:00:00Z',
      status: 'active',
      error_message: null,
      rate_limit_reset_at: null,
      overload_until: null,
      temp_unschedulable_until: null,
      temp_unschedulable_reason: null,
      extra: {}
    } as any

    const next = {
      ...base,
      extra: {
        model_rate_limits: {
          'gpt-5': {
            rate_limited_at: '2026-03-07T10:01:00Z',
            rate_limit_reset_at: '2026-03-07T10:06:00Z'
          }
        }
      }
    }

    expect(buildAccountStatusRefreshKey(base)).not.toBe(buildAccountStatusRefreshKey(next))
  })

  it('会在配额或 tier 变化时刷新用量 key', () => {
    const base = {
      id: 9,
      platform: 'gemini',
      type: 'apikey',
      updated_at: '2026-03-07T10:00:00Z',
      last_used_at: null,
      status: 'active',
      rate_limit_reset_at: null,
      overload_until: null,
      temp_unschedulable_until: null,
      session_window_start: null,
      session_window_end: null,
      session_window_status: null,
      quota_limit: 100,
      quota_used: 20,
      quota_daily_limit: 10,
      quota_daily_used: 2,
      quota_weekly_limit: 50,
      quota_weekly_used: 8,
      current_window_cost: 0,
      active_sessions: 0,
      current_rpm: 0,
      extra: {
        tier_id: 'aistudio_free'
      }
    } as any

    const next = {
      ...base,
      quota_used: 21,
      extra: {
        tier_id: 'aistudio_paid'
      }
    }

    expect(buildAccountUsageRefreshKey(base)).not.toBe(buildAccountUsageRefreshKey(next))
  })

  it('会在列表运行时字段变化时刷新列表 key', () => {
    const base = {
      id: 11,
      platform: 'openai',
      type: 'oauth',
      updated_at: '2026-03-07T10:00:00Z',
      last_used_at: null,
      status: 'active',
      error_message: null,
      rate_limit_reset_at: null,
      overload_until: null,
      temp_unschedulable_until: null,
      temp_unschedulable_reason: null,
      session_window_start: null,
      session_window_end: null,
      session_window_status: null,
      quota_limit: null,
      quota_used: null,
      quota_daily_limit: null,
      quota_daily_used: null,
      quota_weekly_limit: null,
      quota_weekly_used: null,
      current_window_cost: 0,
      active_sessions: 0,
      current_rpm: 0,
      current_concurrency: 1,
      schedulable: true,
      extra: {}
    } as any

    const next = {
      ...base,
      current_concurrency: 2
    }

    expect(buildAccountListRefreshKey(base)).not.toBe(buildAccountListRefreshKey(next))
  })
})
