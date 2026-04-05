import type {
  ApiKey,
  Group,
  GroupPlatform,
  SubscriptionType
} from '@/types'

export interface UserKeyGroupOption extends Record<string, unknown> {
  value: number
  label: string
  description: string | null
  rate: number
  userRate: number | null
  subscriptionType: SubscriptionType
  platform: GroupPlatform
}

export interface UserKeyFormData {
  name: string
  group_id: number | null
  status: 'active' | 'inactive'
  use_custom_key: boolean
  custom_key: string
  enable_ip_restriction: boolean
  ip_whitelist: string
  ip_blacklist: string
  enable_quota: boolean
  quota: number | null
  enable_rate_limit: boolean
  rate_limit_5h: number | null
  rate_limit_1d: number | null
  rate_limit_7d: number | null
  enable_expiration: boolean
  expiration_preset: '7' | '30' | '90' | 'custom'
  expiration_date: string
}

export function formatDateTimeLocal(isoDate: string): string {
  const date = new Date(isoDate)
  const pad = (value: number) => value.toString().padStart(2, '0')

  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`
}

export function buildDefaultUserKeyFormData(): UserKeyFormData {
  return {
    name: '',
    group_id: null,
    status: 'active',
    use_custom_key: false,
    custom_key: '',
    enable_ip_restriction: false,
    ip_whitelist: '',
    ip_blacklist: '',
    enable_quota: false,
    quota: null,
    enable_rate_limit: false,
    rate_limit_5h: null,
    rate_limit_1d: null,
    rate_limit_7d: null,
    enable_expiration: false,
    expiration_preset: '30',
    expiration_date: ''
  }
}

export function buildEditUserKeyFormData(key: ApiKey): UserKeyFormData {
  const hasIpRestriction =
    (key.ip_whitelist?.length > 0) || (key.ip_blacklist?.length > 0)
  const hasExpiration = Boolean(key.expires_at)

  return {
    name: key.name,
    group_id: key.group_id,
    status:
      key.status === 'quota_exhausted' || key.status === 'expired'
        ? 'inactive'
        : key.status,
    use_custom_key: false,
    custom_key: '',
    enable_ip_restriction: hasIpRestriction,
    ip_whitelist: (key.ip_whitelist || []).join('\n'),
    ip_blacklist: (key.ip_blacklist || []).join('\n'),
    enable_quota: key.quota > 0,
    quota: key.quota > 0 ? key.quota : null,
    enable_rate_limit:
      key.rate_limit_5h > 0 || key.rate_limit_1d > 0 || key.rate_limit_7d > 0,
    rate_limit_5h: key.rate_limit_5h || null,
    rate_limit_1d: key.rate_limit_1d || null,
    rate_limit_7d: key.rate_limit_7d || null,
    enable_expiration: hasExpiration,
    expiration_preset: 'custom',
    expiration_date: key.expires_at ? formatDateTimeLocal(key.expires_at) : ''
  }
}

export function buildUserKeyGroupOptions(
  groups: Group[],
  userGroupRates: Record<number, number>
): UserKeyGroupOption[] {
  return groups.map((group) => ({
    value: group.id,
    label: group.name,
    description: group.description,
    rate: group.rate_multiplier,
    userRate: userGroupRates[group.id] ?? null,
    subscriptionType: group.subscription_type,
    platform: group.platform
  }))
}

export function filterUserKeyGroupOptions(
  options: UserKeyGroupOption[],
  query: string
): UserKeyGroupOption[] {
  const normalizedQuery = query.trim().toLowerCase()
  if (!normalizedQuery) {
    return options
  }

  return options.filter((option) => {
    return (
      option.label.toLowerCase().includes(normalizedQuery) ||
      option.description?.toLowerCase().includes(normalizedQuery)
    )
  })
}

export function parseUserKeyIpList(text: string): string[] {
  return text
    .split('\n')
    .map((value) => value.trim())
    .filter((value) => value.length > 0)
}

export function resolveUserKeyQuotaValue(quota: number | null): number {
  return quota && quota > 0 ? quota : 0
}

export function buildUserKeyRateLimitPayload(form: UserKeyFormData): {
  rate_limit_5h: number
  rate_limit_1d: number
  rate_limit_7d: number
} {
  if (!form.enable_rate_limit) {
    return {
      rate_limit_5h: 0,
      rate_limit_1d: 0,
      rate_limit_7d: 0
    }
  }

  return {
    rate_limit_5h: form.rate_limit_5h && form.rate_limit_5h > 0 ? form.rate_limit_5h : 0,
    rate_limit_1d: form.rate_limit_1d && form.rate_limit_1d > 0 ? form.rate_limit_1d : 0,
    rate_limit_7d: form.rate_limit_7d && form.rate_limit_7d > 0 ? form.rate_limit_7d : 0
  }
}

export function buildUserKeyExpirationPayload(
  form: UserKeyFormData,
  isEditMode: boolean,
  now: Date = new Date()
): {
  expiresInDays?: number
  expiresAt?: string | null
} {
  if (form.enable_expiration && form.expiration_date) {
    if (!isEditMode) {
      const expirationDate = new Date(form.expiration_date)
      const diffDays = Math.ceil(
        (expirationDate.getTime() - now.getTime()) / (1000 * 60 * 60 * 24)
      )

      return {
        expiresInDays: diffDays > 0 ? diffDays : 1
      }
    }

    return {
      expiresAt: new Date(form.expiration_date).toISOString()
    }
  }

  if (isEditMode) {
    return {
      expiresAt: ''
    }
  }

  return {}
}

export function applyUserKeyExpirationPreset(
  form: UserKeyFormData,
  days: number,
  now: Date = new Date()
): UserKeyFormData {
  const expirationDate = new Date(now)
  expirationDate.setDate(expirationDate.getDate() + days)

  return {
    ...form,
    expiration_preset: String(days) as '7' | '30' | '90',
    expiration_date: formatDateTimeLocal(expirationDate.toISOString())
  }
}
