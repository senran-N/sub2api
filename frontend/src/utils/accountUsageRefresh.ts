import type { Account } from '@/types'

export interface RefreshKeySchemaEntry<T> {
  field: string
  select: (target: T) => unknown
}

type RefreshSchema<T> = readonly RefreshKeySchemaEntry<T>[]

const normalizeUsageRefreshValue = (value: unknown): string => {
  if (value == null) return ''
  return String(value)
}

const sortObjectEntries = <T>(record: Record<string, T> | null | undefined): Array<[string, T]> => {
  if (!record) return []
  return Object.entries(record).sort(([left], [right]) => left.localeCompare(right))
}

const serializeModelRateLimits = (extra: Account['extra']): string => {
  const modelRateLimits = extra?.model_rate_limits
  return sortObjectEntries(modelRateLimits).map(([model, info]) => (
    [
      model,
      info?.rate_limited_at,
      info?.rate_limit_reset_at
    ].map(normalizeUsageRefreshValue).join(':')
  )).join(',')
}

const serializeLoadCodeAssistState = (extra: Account['extra']): string => {
  const loadCodeAssist = extra?.load_code_assist
  if (!loadCodeAssist || typeof loadCodeAssist !== 'object') {
    return ''
  }

  const record = loadCodeAssist as Record<string, unknown>
  return [
    record.currentTier ? JSON.stringify(record.currentTier) : '',
    record.paidTier ? JSON.stringify(record.paidTier) : '',
    Array.isArray(record.ineligibleTiers) ? JSON.stringify(record.ineligibleTiers) : ''
  ].join('|')
}

const extendRefreshSchema = <T>(...schemas: RefreshSchema<T>[]): RefreshSchema<T> => {
  return schemas.flat()
}

const buildRefreshKey = <T>(target: T, schema: RefreshSchema<T>): string => {
  return schema.map(({ field, select }) => (
    `${field}=${normalizeUsageRefreshValue(select(target))}`
  )).join('|')
}

type OpenAIUsageRefreshAccount = Pick<
  Account,
  'id' | 'platform' | 'type' | 'updated_at' | 'last_used_at' | 'rate_limit_reset_at' | 'extra'
>

type AccountStatusRefreshAccount = Pick<
  Account,
  | 'id'
  | 'updated_at'
  | 'status'
  | 'error_message'
  | 'rate_limit_reset_at'
  | 'overload_until'
  | 'temp_unschedulable_until'
  | 'temp_unschedulable_reason'
  | 'extra'
>

type AccountUsageRefreshAccount = Pick<
  Account,
  | 'id'
  | 'platform'
  | 'type'
  | 'updated_at'
  | 'last_used_at'
  | 'status'
  | 'rate_limit_reset_at'
  | 'overload_until'
  | 'temp_unschedulable_until'
  | 'session_window_start'
  | 'session_window_end'
  | 'session_window_status'
  | 'quota_limit'
  | 'quota_used'
  | 'quota_daily_limit'
  | 'quota_daily_used'
  | 'quota_weekly_limit'
  | 'quota_weekly_used'
  | 'current_window_cost'
  | 'active_sessions'
  | 'current_rpm'
  | 'extra'
>

type AccountListRefreshAccount = Pick<
  Account,
  | 'current_concurrency'
  | 'schedulable'
  | keyof AccountStatusRefreshAccount
  | keyof AccountUsageRefreshAccount
>

export const OPENAI_USAGE_REFRESH_BASE_SCHEMA: RefreshSchema<OpenAIUsageRefreshAccount> = [
  { field: 'id', select: (account) => account.id },
  { field: 'updated_at', select: (account) => account.updated_at },
  { field: 'last_used_at', select: (account) => account.last_used_at },
  { field: 'rate_limit_reset_at', select: (account) => account.rate_limit_reset_at }
]

export const OPENAI_USAGE_REFRESH_EXTENSION_SCHEMA: RefreshSchema<OpenAIUsageRefreshAccount> = [
  { field: 'codex_usage_updated_at', select: (account) => account.extra?.codex_usage_updated_at },
  { field: 'codex_5h_used_percent', select: (account) => account.extra?.codex_5h_used_percent },
  { field: 'codex_5h_reset_at', select: (account) => account.extra?.codex_5h_reset_at },
  {
    field: 'codex_5h_reset_after_seconds',
    select: (account) => account.extra?.codex_5h_reset_after_seconds
  },
  {
    field: 'codex_5h_window_minutes',
    select: (account) => account.extra?.codex_5h_window_minutes
  },
  { field: 'codex_7d_used_percent', select: (account) => account.extra?.codex_7d_used_percent },
  { field: 'codex_7d_reset_at', select: (account) => account.extra?.codex_7d_reset_at },
  {
    field: 'codex_7d_reset_after_seconds',
    select: (account) => account.extra?.codex_7d_reset_after_seconds
  },
  {
    field: 'codex_7d_window_minutes',
    select: (account) => account.extra?.codex_7d_window_minutes
  }
]

export const OPENAI_USAGE_REFRESH_SCHEMA = extendRefreshSchema(
  OPENAI_USAGE_REFRESH_BASE_SCHEMA,
  OPENAI_USAGE_REFRESH_EXTENSION_SCHEMA
)

export const ACCOUNT_STATUS_REFRESH_BASE_SCHEMA: RefreshSchema<AccountStatusRefreshAccount> = [
  { field: 'id', select: (account) => account.id },
  { field: 'updated_at', select: (account) => account.updated_at },
  { field: 'status', select: (account) => account.status },
  { field: 'error_message', select: (account) => account.error_message },
  { field: 'rate_limit_reset_at', select: (account) => account.rate_limit_reset_at },
  { field: 'overload_until', select: (account) => account.overload_until },
  {
    field: 'temp_unschedulable_until',
    select: (account) => account.temp_unschedulable_until
  },
  {
    field: 'temp_unschedulable_reason',
    select: (account) => account.temp_unschedulable_reason
  }
]

export const ACCOUNT_STATUS_RUNTIME_EXTENSION_SCHEMA: RefreshSchema<AccountStatusRefreshAccount> = [
  { field: 'allow_overages', select: (account) => account.extra?.allow_overages },
  {
    field: 'model_rate_limits',
    select: (account) => serializeModelRateLimits(account.extra)
  }
]

export const ACCOUNT_STATUS_REFRESH_SCHEMA = extendRefreshSchema(
  ACCOUNT_STATUS_REFRESH_BASE_SCHEMA,
  ACCOUNT_STATUS_RUNTIME_EXTENSION_SCHEMA
)

export const ACCOUNT_USAGE_REFRESH_BASE_SCHEMA: RefreshSchema<AccountUsageRefreshAccount> = [
  { field: 'id', select: (account) => account.id },
  { field: 'platform', select: (account) => account.platform },
  { field: 'type', select: (account) => account.type },
  { field: 'updated_at', select: (account) => account.updated_at },
  { field: 'last_used_at', select: (account) => account.last_used_at },
  { field: 'status', select: (account) => account.status },
  { field: 'rate_limit_reset_at', select: (account) => account.rate_limit_reset_at },
  { field: 'overload_until', select: (account) => account.overload_until },
  {
    field: 'temp_unschedulable_until',
    select: (account) => account.temp_unschedulable_until
  },
  { field: 'session_window_start', select: (account) => account.session_window_start },
  { field: 'session_window_end', select: (account) => account.session_window_end },
  { field: 'session_window_status', select: (account) => account.session_window_status },
  { field: 'quota_limit', select: (account) => account.quota_limit },
  { field: 'quota_used', select: (account) => account.quota_used },
  { field: 'quota_daily_limit', select: (account) => account.quota_daily_limit },
  { field: 'quota_daily_used', select: (account) => account.quota_daily_used },
  { field: 'quota_weekly_limit', select: (account) => account.quota_weekly_limit },
  { field: 'quota_weekly_used', select: (account) => account.quota_weekly_used },
  { field: 'current_window_cost', select: (account) => account.current_window_cost },
  { field: 'active_sessions', select: (account) => account.active_sessions },
  { field: 'current_rpm', select: (account) => account.current_rpm }
]

export const ACCOUNT_USAGE_PROVIDER_EXTENSION_SCHEMA: RefreshSchema<AccountUsageRefreshAccount> = [
  { field: 'oauth_type', select: (account) => account.extra?.oauth_type },
  { field: 'tier_id', select: (account) => account.extra?.tier_id },
  {
    field: 'load_code_assist',
    select: (account) => serializeLoadCodeAssistState(account.extra)
  },
  {
    field: 'model_rate_limits',
    select: (account) => serializeModelRateLimits(account.extra)
  },
  {
    field: 'openai_usage',
    select: (account) => buildOpenAIUsageRefreshKey(account)
  }
]

export const ACCOUNT_USAGE_REFRESH_SCHEMA = extendRefreshSchema(
  ACCOUNT_USAGE_REFRESH_BASE_SCHEMA,
  ACCOUNT_USAGE_PROVIDER_EXTENSION_SCHEMA
)

export const ACCOUNT_LIST_REFRESH_RUNTIME_SCHEMA: RefreshSchema<AccountListRefreshAccount> = [
  { field: 'current_concurrency', select: (account) => account.current_concurrency },
  { field: 'schedulable', select: (account) => account.schedulable }
]

export const ACCOUNT_LIST_REFRESH_DERIVED_SCHEMA: RefreshSchema<AccountListRefreshAccount> = [
  { field: 'status_key', select: (account) => buildAccountStatusRefreshKey(account) },
  { field: 'usage_key', select: (account) => buildAccountUsageRefreshKey(account) }
]

export const ACCOUNT_LIST_REFRESH_SCHEMA = extendRefreshSchema(
  ACCOUNT_LIST_REFRESH_RUNTIME_SCHEMA,
  ACCOUNT_LIST_REFRESH_DERIVED_SCHEMA
)

export const buildOpenAIUsageRefreshKey = (account: OpenAIUsageRefreshAccount): string => {
  if (account.platform !== 'openai' || account.type !== 'oauth') {
    return ''
  }

  return buildRefreshKey(account, OPENAI_USAGE_REFRESH_SCHEMA)
}

export const buildAccountStatusRefreshKey = (account: AccountStatusRefreshAccount): string => {
  return buildRefreshKey(account, ACCOUNT_STATUS_REFRESH_SCHEMA)
}

export const buildAccountUsageRefreshKey = (account: AccountUsageRefreshAccount): string => {
  return buildRefreshKey(account, ACCOUNT_USAGE_REFRESH_SCHEMA)
}

export const buildAccountListRefreshKey = (account: AccountListRefreshAccount): string => {
  return buildRefreshKey(account, ACCOUNT_LIST_REFRESH_SCHEMA)
}
