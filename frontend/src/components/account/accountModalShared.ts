import type {
  Account,
  AccountPlatform,
  AccountType,
  CheckMixedChannelResponse
} from '@/types'
import type { TempUnschedRuleForm } from './credentialsBuilder'
import {
  OPENAI_WS_MODE_OFF,
  OPENAI_WS_MODE_PASSTHROUGH,
  type OpenAIWSMode
} from '@/utils/openaiWsMode'

export type Translate = (key: string, values?: Record<string, unknown>) => string
type QuotaResetMode = 'rolling' | 'fixed' | null

interface AccountQuotaExtraOptions {
  dailyResetHour: number | null
  dailyResetMode: QuotaResetMode
  quotaDailyLimit: number | null
  quotaLimit: number | null
  quotaWeeklyLimit: number | null
  resetTimezone: string | null
  weeklyResetDay: number | null
  weeklyResetHour: number | null
  weeklyResetMode: QuotaResetMode
}

type EditableAccountStatus = Extract<Account['status'], 'active' | 'inactive' | 'error'>

export interface AccountBaseFormFields {
  name: string
  notes: string
  proxy_id: number | null
  concurrency: number
  load_factor: number | null
  priority: number
  rate_multiplier: number
  group_ids: number[]
  expires_at: number | null
}

export interface CreateAccountForm extends AccountBaseFormFields {
  platform: AccountPlatform
  type: AccountType
  credentials: Record<string, unknown>
}

export interface EditAccountForm extends AccountBaseFormFields {
  status: EditableAccountStatus
}

export function createDefaultCreateAccountForm(): CreateAccountForm {
  return {
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
  }
}

export function resetCreateAccountForm(form: CreateAccountForm): void {
  Object.assign(form, createDefaultCreateAccountForm())
}

export function createDefaultEditAccountForm(): EditAccountForm {
  return {
    name: '',
    notes: '',
    proxy_id: null,
    concurrency: 1,
    load_factor: null,
    priority: 1,
    rate_multiplier: 1,
    status: 'active',
    group_ids: [],
    expires_at: null
  }
}

export function hydrateEditAccountForm(
  form: EditAccountForm,
  account: Pick<
    Account,
    | 'name'
    | 'notes'
    | 'proxy_id'
    | 'concurrency'
    | 'load_factor'
    | 'priority'
    | 'rate_multiplier'
    | 'status'
    | 'group_ids'
    | 'expires_at'
  >
): void {
  Object.assign(form, createDefaultEditAccountForm(), {
    name: account.name,
    notes: account.notes || '',
    proxy_id: account.proxy_id,
    concurrency: account.concurrency,
    load_factor: account.load_factor ?? null,
    priority: account.priority,
    rate_multiplier: account.rate_multiplier ?? 1,
    status: normalizeEditableAccountStatus(account.status),
    group_ids: account.group_ids || [],
    expires_at: account.expires_at ?? null
  })
}

export function buildEditAccountBasePayload(
  form: EditAccountForm,
  autoPauseOnExpired: boolean
): Record<string, unknown> {
  const payload: Record<string, unknown> = {
    ...form,
    proxy_id: form.proxy_id ?? 0,
    expires_at: form.expires_at ?? 0,
    auto_pause_on_expired: autoPauseOnExpired
  }

  if (
    form.load_factor == null ||
    Number.isNaN(form.load_factor) ||
    form.load_factor <= 0
  ) {
    payload.load_factor = 0
  }

  return payload
}

export function resolveAccountBaseUrlHint(platform: AccountPlatform | null | undefined, t: Translate) {
  if (platform === 'openai' || platform === 'sora') {
    return t('admin.accounts.openai.baseUrlHint')
  }
  if (platform === 'gemini') {
    return t('admin.accounts.gemini.baseUrlHint')
  }
  return t('admin.accounts.baseUrlHint')
}

export function resolveAccountApiKeyHint(platform: AccountPlatform | null | undefined, t: Translate) {
  if (platform === 'openai' || platform === 'sora') {
    return t('admin.accounts.openai.apiKeyHint')
  }
  if (platform === 'gemini') {
    return t('admin.accounts.gemini.apiKeyHint')
  }
  return t('admin.accounts.apiKeyHint')
}

export function resolveCreateAccountOAuthStepTitle(
  platform: AccountPlatform,
  t: Translate
) {
  if (platform === 'openai' || platform === 'sora') {
    return t('admin.accounts.oauth.openai.title')
  }
  if (platform === 'gemini') {
    return t('admin.accounts.oauth.gemini.title')
  }
  if (platform === 'antigravity') {
    return t('admin.accounts.oauth.antigravity.title')
  }
  return t('admin.accounts.oauth.title')
}

export function buildAccountUmqModeOptions(t: Translate) {
  return [
    { value: '', label: t('admin.accounts.quotaControl.rpmLimit.umqModeOff') },
    { value: 'throttle', label: t('admin.accounts.quotaControl.rpmLimit.umqModeThrottle') },
    { value: 'serialize', label: t('admin.accounts.quotaControl.rpmLimit.umqModeSerialize') }
  ]
}

export function buildAccountOpenAIWSModeOptions(t: Translate): Array<{ value: OpenAIWSMode; label: string }> {
  return [
    { value: OPENAI_WS_MODE_OFF, label: t('admin.accounts.openai.wsModeOff') },
    { value: OPENAI_WS_MODE_PASSTHROUGH, label: t('admin.accounts.openai.wsModePassthrough') }
  ]
}

export function buildAccountTempUnschedPresets(
  t: Translate
): Array<{ label: string; rule: TempUnschedRuleForm }> {
  return [
    {
      label: t('admin.accounts.tempUnschedulable.presets.overloadLabel'),
      rule: {
        error_code: 529,
        keywords: 'overloaded, too many',
        duration_minutes: 60,
        description: t('admin.accounts.tempUnschedulable.presets.overloadDesc')
      }
    },
    {
      label: t('admin.accounts.tempUnschedulable.presets.rateLimitLabel'),
      rule: {
        error_code: 429,
        keywords: 'rate limit, too many requests',
        duration_minutes: 10,
        description: t('admin.accounts.tempUnschedulable.presets.rateLimitDesc')
      }
    },
    {
      label: t('admin.accounts.tempUnschedulable.presets.unavailableLabel'),
      rule: {
        error_code: 503,
        keywords: 'unavailable, maintenance',
        duration_minutes: 30,
        description: t('admin.accounts.tempUnschedulable.presets.unavailableDesc')
      }
    }
  ]
}

export function buildAccountQuotaExtra(
  baseExtra: Record<string, unknown> | undefined,
  options: AccountQuotaExtraOptions
) {
  const nextExtra: Record<string, unknown> = { ...(baseExtra || {}) }

  if (options.quotaLimit != null && options.quotaLimit > 0) {
    nextExtra.quota_limit = options.quotaLimit
  } else {
    delete nextExtra.quota_limit
  }

  if (options.quotaDailyLimit != null && options.quotaDailyLimit > 0) {
    nextExtra.quota_daily_limit = options.quotaDailyLimit
  } else {
    delete nextExtra.quota_daily_limit
  }

  if (options.quotaWeeklyLimit != null && options.quotaWeeklyLimit > 0) {
    nextExtra.quota_weekly_limit = options.quotaWeeklyLimit
  } else {
    delete nextExtra.quota_weekly_limit
  }

  if (options.dailyResetMode === 'fixed') {
    nextExtra.quota_daily_reset_mode = 'fixed'
    nextExtra.quota_daily_reset_hour = options.dailyResetHour ?? 0
  } else {
    delete nextExtra.quota_daily_reset_mode
    delete nextExtra.quota_daily_reset_hour
  }

  if (options.weeklyResetMode === 'fixed') {
    nextExtra.quota_weekly_reset_mode = 'fixed'
    nextExtra.quota_weekly_reset_day = options.weeklyResetDay ?? 1
    nextExtra.quota_weekly_reset_hour = options.weeklyResetHour ?? 0
  } else {
    delete nextExtra.quota_weekly_reset_mode
    delete nextExtra.quota_weekly_reset_day
    delete nextExtra.quota_weekly_reset_hour
  }

  if (options.dailyResetMode === 'fixed' || options.weeklyResetMode === 'fixed') {
    nextExtra.quota_reset_timezone = options.resetTimezone || 'UTC'
  } else {
    delete nextExtra.quota_reset_timezone
  }

  return nextExtra
}

export function needsMixedChannelCheck(platform: AccountPlatform) {
  return platform === 'antigravity' || platform === 'anthropic'
}

export function buildMixedChannelDetails(resp?: CheckMixedChannelResponse) {
  const details = resp?.details
  if (!details) {
    return null
  }

  return {
    groupName: details.group_name || 'Unknown',
    currentPlatform: details.current_platform || 'Unknown',
    otherPlatform: details.other_platform || 'Unknown'
  }
}

export function resolveMixedChannelWarningMessage(options: {
  details: ReturnType<typeof buildMixedChannelDetails>
  rawMessage: string
  t: Translate
}) {
  if (options.details) {
    return options.t('admin.accounts.mixedChannelWarning', options.details)
  }
  return options.rawMessage
}

export const geminiQuotaDocs = {
  codeAssist: 'https://developers.google.com/gemini-code-assist/resources/quotas',
  aiStudio: 'https://ai.google.dev/pricing',
  vertex: 'https://cloud.google.com/vertex-ai/generative-ai/docs/quotas'
}

export const geminiHelpLinks = {
  apiKey: 'https://aistudio.google.com/app/apikey',
  aiStudioPricing: 'https://ai.google.dev/pricing',
  gcpProject: 'https://console.cloud.google.com/welcome/new',
  geminiWebActivation: 'https://gemini.google.com/gems/create?hl=en-US&pli=1',
  countryCheck: 'https://policies.google.com/terms',
  countryChange: 'https://policies.google.com/country-association-form'
}

function normalizeEditableAccountStatus(status: Account['status']): EditableAccountStatus {
  if (status === 'inactive' || status === 'error') {
    return status
  }

  return 'active'
}
